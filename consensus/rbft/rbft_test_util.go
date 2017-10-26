package rbft

import (
	"errors"
	"github.com/facebookgo/ensure"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"hyperchain/common"
	"hyperchain/consensus/helper"
	"hyperchain/core/ledger/chain"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	pb "hyperchain/manager/protos"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

//path struct should match
// /namespaces/(namespace)/config/namespace.toml
func TNewConfig(fatherpath, name string, nodeId int) *common.Config {

	path := filepath.Join(fatherpath, name, "/config/namespace.toml")
	conf := common.NewConfig(path)
	common.InitHyperLoggerManager(conf)
	peerConfigPath := conf.GetString(common.PEER_CONFIG_PATH)

	peerConfigPath = filepath.Join(fatherpath, name, peerConfigPath)
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		panic("err " + err.Error())
	}
	//set node self.id
	if nodeId != 0 {
		conf.Set(common.C_NODE_ID, strconv.Itoa(nodeId))
	} else {
		conf.Set(common.C_NODE_ID, peerViper.GetInt("self.id"))
	}

	conf.Set("self.N", peerViper.GetInt("self.N"))
	namespace := name + "-" + strconv.Itoa(conf.GetInt(common.C_NODE_ID))
	conf.Set(common.NAMESPACE, namespace)
	common.InitHyperLogger(namespace, conf)
	return conf
}

//Return a pbft according to config path
//path for example
// ../../configuration/namespaces/
// If the namespace is global
// The directory of namesapces.toml should be   ../../configuration/namespaces/global/config/namespace.toml
func TNewRbft(dbpath, path, namespace string, nodeId int, t *testing.T) (*rbftImpl, *common.Config, error) {
	conf := TNewConfig(path, namespace, nodeId)
	namespace = conf.GetString(common.NAMESPACE)
	if dbpath != "" {
		conf.Set("database.leveldb.path", dbpath)
	}
	err := chain.InitDBForNamespace(conf, namespace)
	if err != nil {
		t.Errorf("init db for namespace: %s error, %v", namespace, err)
	}
	h := helper.NewHelper(new(event.TypeMux), new(event.TypeMux))
	pbft, err := newRBFT(namespace, conf, h, conf.GetInt("self.N"))
	return pbft, conf, err
}

/////////////////////////////////////////////////////
////implement new helper for Test		 ////
////////////////////////////////////////////////////

//InnerBroadcast(msg *pb.Message) error
//InnerUnicast(msg *pb.Message, to uint64) error
//Execute(seqNo uint64, hash string, flag bool, isPrimary bool, time int64) error
//UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error
//ValidateBatch(txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error
//VcReset(seqNo uint64) error
//InformPrimary(primary uint64) error
//BroadcastAddNode(msg *pb.Message) error
//BroadcastDelNode(msg *pb.Message) error
//UpdateTable(payload []byte, flag bool) error
//SendFilterEvent(informType int, message ...interface{}) error

type MessageChanel struct {
	MsgChan chan *pb.Message
}

func (MS *MessageChanel) Start(pbft *rbftImpl) {
	for msg := range MS.MsgChan {
		message, err := proto.Marshal(msg)
		if err != nil {
			pbft.logger.Error("Marshal failed")
		}
		pbft.RecvMsg(message)
	}
}

type TestHelp struct {
	PbftList     []*MessageChanel //save pbftHandle for communicate. Because the id of node is from 1,so the Pbftlist[0] always nil
	PbftID       int
	PbftLen      int
	namespace    string
	batchMap     map[common.Hash]*protos.ValidatedTxs
	batchMapLock sync.Mutex //no currency read, so don;t use RWMutex
	rbft         *rbftImpl
}

func (TH *TestHelp) InnerBroadcast(msg *pb.Message) error {
	for i := 1; i <= TH.PbftLen; i++ {
		if i != TH.PbftID {
			if TH.PbftList[i] == nil {
				continue
			}
			TH.rbft.logger.Debugf("broadcast to %v", i)
			TH.PbftList[i].MsgChan <- msg
		}
	}
	return nil
}

func (TH *TestHelp) InnerUnicast(msg *pb.Message, to uint64) error {
	to1 := int(to)
	if to1 <= TH.PbftLen {
		if TH.PbftList[to1] != nil {
			TH.PbftList[to1].MsgChan <- msg
		}
	}
	return nil
}

func (TH *TestHelp) ValidateBatch(digest string, txs []*types.Transaction, timeStamp int64, seqNo uint64, view uint64, isPrimary bool) error {
	go func() {
		time.Sleep(0)
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		validTxSet := make([][]byte, len(txs))
		for i := 0; i < len(txs); i++ {
			validTxSet[i] = txs[i].TransactionHash
		}
		hash := kec256Hash.Hash(validTxSet)
		vtx := protos.ValidatedTxs{
			SeqNo:        seqNo,
			View:         view,
			Hash:         hash.Hex(),
			Transactions: txs,
			Digest:       digest,
		}
		if TH.batchMap[hash] != nil {
			TH.batchMapLock.Lock()
			TH.batchMap[hash] = &vtx
			TH.batchMapLock.Unlock()
		}

		event := &LocalEvent{
			Service:   CORE_RBFT_SERVICE,
			EventType: CORE_VALIDATED_TXS_EVENT,
			Event:     vtx,
		}
		TH.rbft.RecvLocal(event)
	}()
	return nil
}

func (TH *TestHelp) Execute(seqNo uint64, hashS string, flag bool, isPrimary bool, time int64) error {
	hash := common.StringToHash(hashS)
	TH.batchMapLock.Lock()
	if TH.batchMap[hash] == nil {
		TH.rbft.logger.Error("miss commit block")
		TH.batchMapLock.Unlock()
		return nil
	}
	vtx := TH.batchMap[hash]
	TH.batchMapLock.Unlock()

	db, err := hyperdb.GetDBDatabaseByNamespace(TH.namespace)
	if err != nil {
		TH.rbft.logger.Error(err.Error())
	}

	batch := db.NewBatch()

	block := &types.Block{
		ParentHash:   edb.GetLatestBlockHash(TH.namespace),
		Transactions: vtx.Transactions,
		Number:       vtx.SeqNo,
	}
	if _, err := edb.PersistBlock(batch, block, false, false); err != nil {
		TH.rbft.logger.Errorf("persist block #%d into database failed.", block.Number, err.Error())
		return nil
	}

	if err := edb.UpdateChain(TH.namespace, batch, block, false, false, false); err != nil {
		TH.rbft.logger.Errorf("update chain to #%d failed.", block.Number, err.Error())
		return nil
	}
	err = batch.Write()
	if err != nil {
		TH.rbft.logger.Error(err.Error())
	}
	return nil
}

func (TH *TestHelp) UpdateState(myId uint64, height uint64, blockHash []byte, replicas []event.SyncReplica) error {
	return nil
}

func (TH *TestHelp) VcReset(seqNo uint64) error {
	//TODO vcReset
	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_VC_RESET_DONE_EVENT,
		Event:     protos.VcResetDone{SeqNo: seqNo},
	}
	TH.rbft.RecvLocal(event)
	return nil
}
func (TH *TestHelp) InformPrimary(primary uint64) error                           { return nil }
func (TH *TestHelp) BroadcastAddNode(msg *pb.Message) error                       { return nil }
func (TH *TestHelp) BroadcastDelNode(msg *pb.Message) error                       { return nil }
func (TH *TestHelp) UpdateTable(payload []byte, flag bool) error                  { return nil }
func (TH *TestHelp) SendFilterEvent(informType int, message ...interface{}) error { return nil }

type RBFTNode struct {
	nodeId    int
	dbPath    string
	confPath  string
	namesapce string
}

func CreatRBFT(t *testing.T, N int, dbPath string, confPath string, namespace string, nodes []*RBFTNode) (rbftList []*rbftImpl) {

	if N < 4 {
		t.Error("N is too small to create PBFT network")
	}
	rbftList = make([]*rbftImpl, N+1) //pbft id start from 1 not zero ,so Create N+1 slice
	mcList := make([]*MessageChanel, N+1)
	thList := make([]*TestHelp, N+1)
	var err error
	for i := 0; i < len(nodes); i++ {

		if nodes[i].nodeId > N {
			t.Log("node id is greater then N")
			continue
		}
		rbftList[nodes[i].nodeId], _, err = TNewRbft(nodes[i].dbPath, nodes[i].confPath, nodes[i].namesapce, 0, t)
		ensure.Nil(t, err)
	}

	//init node not in nodes
	for i := 1; i < N+1; i++ {
		if rbftList[i] == nil {
			rbftList[i], _, err = TNewRbft(dbPath, confPath, namespace, i, t)
			ensure.Nil(t, err)
		}
	}
	logger := common.GetLogger("system", "test")
	for i := 1; i < N+1; i++ {
		mcList[i] = &MessageChanel{
			MsgChan: make(chan *pb.Message),
		}
		go mcList[i].Start(rbftList[i])
	}

	//init helper and replace the helper in pbft
	for i := 1; i < N+1; i++ {
		thList[i] = &TestHelp{
			rbft:      rbftList[i],
			PbftID:    i,
			PbftList:  mcList,
			namespace: rbftList[i].namespace,
			batchMap:  make(map[common.Hash]*protos.ValidatedTxs),
			PbftLen:   N,
		}
		rbftList[i].helper = thList[i]
	}

	logger.Notice("Full system initialization completed. Now try to start pbft")
	for i := 1; i < N+1; i++ {
		rbftList[i].Start()
	}

	logger.Debug("start negotiate view")
	for i := 1; i < N+1; i++ {
		negoView := &protos.Message{
			Type:      protos.Message_NEGOTIATE_VIEW,
			Timestamp: time.Now().UnixNano(),
			Payload:   nil,
			Id:        0,
		}
		rbftList[i].RecvLocal(negoView)
	}
	return
	//pbft1,_,err:=TNewPbft("./Testdatabase/","../../build/node1/namespaces/","global",0,t)
	//ensure.Nil(t,err)
	//pbft2,_,err:=TNewPbft("./Testdatabase/","../../build/node2/namespaces/","global",0,t)
	//ensure.Nil(t,err)
	//pbft3,_,err:=TNewPbft("./Testdatabase/","../../build/node3/namespaces/","global",0,t)
	//ensure.Nil(t,err)
	//pbft4,_,err:=TNewPbft("./Testdatabase/","../../build/node4/namespaces/","global",0,t)
	//ensure.Nil(t,err)
	//
	//mc1:=&MessageChanel{
	//	MsgChan:make(chan *pb.Message),
	//}
	//go mc1.Start(pbft1)
	//
	//mc2:=&MessageChanel{
	//	MsgChan:make(chan *pb.Message),
	//}
	//go mc2.Start(pbft2)
	//
	//mc3:=&MessageChanel{
	//	MsgChan:make(chan *pb.Message),
	//}
	//go mc3.Start(pbft3)
	//
	//mc4:=&MessageChanel{
	//	MsgChan:make(chan *pb.Message),
	//}
	//go mc4.Start(pbft4)
	//
	//mcList :=make([]*MessageChanel,4+1)
	//mcList[1]=mc1
	//mcList[2]=mc2
	//mcList[3]=mc3
	//mcList[4]=mc4
	//th1:=&TestHelp{
	//	pbft:pbft1,
	//	PbftID:1,
	//	PbftList:mcList,
	//	namespace:pbft1.namespace,
	//	batchMap:make(map[common.Hash]*protos.ValidatedTxs),
	//	PbftLen:4,
	//}
	//
	//th2:=&TestHelp{
	//	pbft:pbft2,
	//	PbftID:2,
	//	PbftList:mcList,
	//	namespace:pbft2.namespace,
	//	batchMap:make(map[common.Hash]*protos.ValidatedTxs),
	//	PbftLen:4,
	//}
	//
	//th3:=&TestHelp{
	//	pbft:pbft3,
	//	PbftID:3,
	//	PbftList:mcList,
	//	namespace:pbft3.namespace,
	//	batchMap:make(map[common.Hash]*protos.ValidatedTxs),
	//	PbftLen:4,
	//}
	//
	//
	//th4:=&TestHelp{
	//	pbft:pbft4,
	//	PbftID:4,
	//	PbftList:mcList,
	//	namespace:pbft4.namespace,
	//	batchMap:make(map[common.Hash]*protos.ValidatedTxs),
	//	PbftLen:4,
	//}
	//
	//pbft1.helper=th1
	//pbft2.helper=th2
	//pbft3.helper=th3
	//pbft4.helper=th4
	//pbft1.Start()
	//pbft2.Start()
	//pbft3.Start()
	//pbft4.Start()
	//negoView1 := &protos.Message{
	//	Type:      protos.Message_NEGOTIATE_VIEW,
	//	Timestamp: time.Now().UnixNano(),
	//	Payload:   nil,
	//	Id:        0,
	//}
	//negoView2 := &protos.Message{
	//	Type:      protos.Message_NEGOTIATE_VIEW,
	//	Timestamp: time.Now().UnixNano(),
	//	Payload:   nil,
	//	Id:        0,
	//}
	//negoView3 := &protos.Message{
	//	Type:      protos.Message_NEGOTIATE_VIEW,
	//	Timestamp: time.Now().UnixNano(),
	//	Payload:   nil,
	//	Id:        0,
	//}
	//negoView4 := &protos.Message{
	//	Type:      protos.Message_NEGOTIATE_VIEW,
	//	Timestamp: time.Now().UnixNano(),
	//	Payload:   nil,
	//	Id:        0,
	//}
	// pbft1.RecvLocal(negoView1)
	// pbft2.RecvLocal(negoView2)
	// pbft3.RecvLocal(negoView3)
	// pbft4.RecvLocal(negoView4)
}

//
//func MsgForRecvMsg(msgType int32,)[]byte{
//
//}

//remove the data and namespace directory in ./
func CleanData(namespace string) error {
	hyperdb.StopDatabase(namespace)
	err := os.RemoveAll("./data")
	if err != nil {
		return err
	}
	err = os.RemoveAll("./namespaces")
	return err
}

func checkNilElems(i interface{}) (string, []string, error) {
	typ := reflect.TypeOf(i)
	value := reflect.Indirect(reflect.ValueOf(i))

	if typ.Kind() != reflect.Ptr {
		return "", nil, errors.New("Got a non-ptr to check if has nil elements.")
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return "", nil, errors.New("Got a non-struct to check if has nil elements.")
	}

	structName := typ.Name()
	nilElems := []string{}
	hasNil := false

	for i := 0; i < typ.NumField(); i++ {
		kind := typ.Field(i).Type.Kind()
		if kind == reflect.Chan || kind == reflect.Map {
			elemName := typ.Field(i).Name
			if value.FieldByName(elemName).IsNil() {
				nilElems = append(nilElems, elemName)
				hasNil = true
			}
		}
	}
	if hasNil {
		return structName, nilElems, nil
	}
	return structName, nil, nil
}
