//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus/helper"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	edb "github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"
	pb "github.com/hyperchain/hyperchain/manager/protos"

	"github.com/facebookgo/ensure"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
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
	common.SetLogLevel(namespace, "consensus", "DEBUG")
	return conf
}

//Return a rbft according to config path
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
	rbft, err := newRBFT(namespace, conf, h, conf.GetInt("self.N"))
	return rbft, conf, err
}

/////////////////////////////////////////////////////
////implement new helper for Test		 ////
////////////////////////////////////////////////////

type MessageChanel struct {
	MsgChan chan *pb.Message
}

func (MS *MessageChanel) Start(rbft *rbftImpl) {
	for msg := range MS.MsgChan {
		message, err := proto.Marshal(msg)
		if err != nil {
			rbft.logger.Error("Marshal failed")
		}
		rbft.RecvMsg(message)
	}
}

type TestHelp struct {
	RbftList     []*MessageChanel //save rbftHandle for communicate. Because the id of node is from 1,so the Rbftlist[0] always nil
	RbftID       int
	RbftLen      int
	namespace    string
	batchMap     map[common.Hash]*protos.ValidatedTxs
	batchMapLock sync.Mutex //no currency read, so don;t use RWMutex
	rbft         *rbftImpl
}

func (TH *TestHelp) InnerBroadcast(msg *pb.Message) error {
	for i := 1; i <= TH.RbftLen; i++ {
		if i != TH.RbftID {
			if TH.RbftList[i] == nil {
				continue
			}
			TH.rbft.logger.Debugf("broadcast to %v", i)
			TH.RbftList[i].MsgChan <- msg
		}
	}
	return nil
}

func (TH *TestHelp) InnerUnicast(msg *pb.Message, to uint64) error {
	to1 := int(to)
	if to1 <= TH.RbftLen {
		if TH.RbftList[to1] != nil {
			TH.RbftList[to1].MsgChan <- msg
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

	db, err := hyperdb.GetDBDatabaseByNamespace(TH.namespace, hcom.DBNAME_BLOCKCHAIN)
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

func (TH *TestHelp) VcReset(seqNo uint64, view uint64) error {
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
		t.Error("N is too small to create RBFT network")
	}
	rbftList = make([]*rbftImpl, N+1) // rbft id start from 1 not zero ,so Create N+1 slice
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

	//init helper and replace the helper in rbft
	for i := 1; i < N+1; i++ {
		thList[i] = &TestHelp{
			rbft:      rbftList[i],
			RbftID:    i,
			RbftList:  mcList,
			namespace: rbftList[i].namespace,
			batchMap:  make(map[common.Hash]*protos.ValidatedTxs),
			RbftLen:   N,
		}
		rbftList[i].helper = thList[i]
	}

	logger.Notice("Full system initialization completed. Now try to start rbft")
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
}

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

// checkNilElems checks if provided param has nil elements, returns error if provided
// param is not a struct pointer and returns nil elements' name if has.
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
