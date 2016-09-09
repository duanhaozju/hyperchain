// implement ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-31
package manager

import (
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/core"
	"hyperchain/consensus"
	"hyperchain/crypto"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"fmt"
	"sync"

	"hyperchain/protos"
	"time"

	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("manager")
}

type ProtocolManager struct {
	serverPort   int
	blockPool    *core.BlockPool
	fetcher      *core.Fetcher
	peerManager  p2p.PeerManager
	consenter    consensus.Consenter
	//encryption   crypto.Encryption
	accountManager	*accounts.AccountManager
	commonHash   crypto.CommonHash

	noMorePeers  chan struct{}
	eventMux     *event.TypeMux
	txSub        event.Subscription
	newBlockSub  event.Subscription
	consensusSub event.Subscription

	aLiveSub     event.Subscription
	quitSync     chan struct{}

	wg           sync.WaitGroup
}

var eventMuxAll *event.TypeMux
var countBlock int

func NewProtocolManager(blockPool *core.BlockPool,peerManager p2p.PeerManager, eventMux *event.TypeMux, fetcher *core.Fetcher, consenter consensus.Consenter,
//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
am *accounts.AccountManager, commonHash crypto.CommonHash) (*ProtocolManager) {
	log.Debug("enter parotocol manager")
	manager := &ProtocolManager{


		blockPool: blockPool,
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:consenter,
		peerManager:  peerManager,
		fetcher:fetcher,
		//encryption:encryption,
		accountManager:am,
		commonHash:commonHash,


	}
	eventMuxAll = eventMux
	return manager
}


func GetEventObject() *event.TypeMux {
	return eventMuxAll
}


// start listen new block msg and consensus msg
func (pm *ProtocolManager) Start() {


	pm.wg.Add(1)
	go pm.fetcher.Start()
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{}, event.BroadcastConsensusEvent{}, event.NewTxEvent{})
	pm.newBlockSub = pm.eventMux.Subscribe(event.NewBlockEvent{})
	go pm.NewBlockLoop()
	go pm.ConsensusLoop()



	pm.wg.Wait()

}



// listen block msg
func (self *ProtocolManager) NewBlockLoop() {

	for obj := range self.newBlockSub.Chan() {

		switch  ev :=obj.Data.(type) {
		case event.NewBlockEvent:
			//accept msg from consensus module
			//commit block into block pool

			log.Debug("write block success")
			self.commitNewBlock(ev.Payload,ev.CommitTime)
		//self.fetcher.Enqueue(ev.Payload)

		}
	}
}

//listen consensus msg
func (self *ProtocolManager) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.BroadcastConsensusEvent:
			log.Debug("######enter broadcast")

			go self.BroadcastConsensus(ev.Payload)
		case event.NewTxEvent:
			log.Debug("######receiver new tx")
			//call consensus module
			//send msg to consensus
			//for i:=0;i<10000;i+=1{
			//	go self.sendMsg(ev.Payload)
			//	//time.Sleep(100*time.Microsecond)
			//}

			go self.sendMsg(ev.Payload)

		case event.ConsensusEvent:
			//call consensus module
			log.Debug("###### enter ConsensusEvent")
			//logger.GetLogger().Println("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)


		}

	}
}

func (self *ProtocolManager)sendMsg(payload []byte)  {
	//Todo sign tx
	payLoad:=self.transformTx(payload)
	if payLoad==nil{
		//log.Fatal("payLoad nil")
		log.Error("payLoad nil")
		return
	}
	msg := &protos.Message{
		Type: protos.Message_TRANSACTION,
		//Payload: payload,
		Payload: payLoad,
		Timestamp: time.Now().UnixNano(),
		Id: 0,
	}
	msgSend, _ := proto.Marshal(msg)
	self.consenter.RecvMsg(msgSend)
}


// Broadcast consensus msg to a batch of peers not knowing about it
func (pm *ProtocolManager) BroadcastConsensus(payload []byte) {
	log.Debug("begin call broadcast")
	pm.peerManager.BroadcastPeers(payload)

}

//receive tx from web,sign it and marshal it,then give it to consensus module
func (pm *ProtocolManager)transformTx(payload []byte) []byte {

	//var transaction types.Transaction
	transaction := &types.Transaction{}
	//decode tx
	proto.Unmarshal(payload, transaction)
	//hash tx
	h := transaction.SighHash(pm.commonHash)
	addrHex := string(transaction.From)
	addr := common.HexToAddress(addrHex)

	sign, err := pm.accountManager.Sign(addr,h[:])
	if err != nil {
		fmt.Print(err)
	}
	transaction.Signature = sign
	//encode tx
	payLoad, err := proto.Marshal(transaction)
	if err != nil {
		return nil
	}
	return payLoad

}
//func (pm *ProtocolManager)transformTx(payload []byte) []byte {
//
//	//var transaction types.Transaction
//	transaction := &types.Transaction{}
//	//decode tx
//	proto.Unmarshal(payload, transaction)
//	//hash tx
//	h := transaction.SighHash(pm.commonHash)
//	addrHex := string(transaction.From)
//	addr := common.HexToAddress(addrHex)
//	account := accounts.Account{
//		Address:addr,
//		File:pm.accountManager.KeyStore.JoinPath(accounts.KeyFileName(addr[:])),
//	}
//	time1 := time.Now().UnixNano()
//	key, err := pm.accountManager.GetDecryptedKeyCache(account)
//	time2 := time.Now().UnixNano()
//	log.Info(">>>>>>>>>>>>>>",time2-time1)
//	if err!=nil{
//		log.Error(err)
//		return nil
//	}
//	//key, err := pm.encryption.GetKey()
//	//switch key.(type){
//	switch key.PrivateKey.(type){
//	//case ecdsa.PrivateKey:
//	case *ecdsa.PrivateKey:
//		//actualKey := key.(ecdsa.PrivateKey)
//		//sign, err := pm.encryption.Sign(h[:], actualKey)
//		actualKey := key.PrivateKey.(*ecdsa.PrivateKey)
//		sign, err := pm.accountManager.Encryption.Sign(h[:], actualKey)
//		if err != nil {
//			fmt.Print(err)
//		}
//		transaction.Signature = sign
//		//encode tx
//		payLoad, err := proto.Marshal(transaction)
//		if err != nil {
//			return nil
//		}
//		return payLoad
//
//
//	}
//	if err != nil {
//		return nil
//	}
//	return nil
//
//}


// add new block into block pool
func (pm *ProtocolManager) commitNewBlock(payload[]byte,commitTime int64) {

	msgList := &protos.ExeMessage{}
	proto.Unmarshal(payload, msgList)
	block := new(types.Block)
	for _, item := range msgList.Batch {
		tx := &types.Transaction{}

		proto.Unmarshal(item.Payload, tx)

		block.Transactions = append(block.Transactions, tx)
	}
	block.Timestamp = msgList.Timestamp
	//block.CommitTime =commitTime

	block.Number=msgList.No

	log.Info("now is ",msgList.No)
	pm.blockPool.AddBlock(block,pm.commonHash,commitTime)
	//core.WriteBlock(*block)

}




