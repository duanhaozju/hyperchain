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
	"crypto/ecdsa"
	"log"
	"hyperchain/protos"
	"time"

)

type ProtocolManager struct {
	serverPort   int
	fetcher      *core.Fetcher
	peerManager  p2p.PeerManager
	consenter    consensus.Consenter
	encryption   crypto.Encryption
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

func NewProtocolManager(peerManager p2p.PeerManager, eventMux *event.TypeMux, fetcher *core.Fetcher, consenter consensus.Consenter,
encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
	fmt.Println("enter parotocol manager")
	fmt.Println(consenter)
	manager := &ProtocolManager{

		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:consenter,
		peerManager:  peerManager,
		fetcher:fetcher,
		encryption:encryption,
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

	//commit block into local db


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
			//commit block into local db
			//log.Println(ev.Payload)

			countBlock=countBlock+1

			log.Println(time.Now().UnixNano())
			log.Println("block number is ",countBlock)
			log.Println("write block success")
			//ioutil.WriteFile("./123.txt",[]byte(strconv.FormatInt(time.Now().UnixNano(),10)+"\n"),os.ModeAppend)
			self.commitNewBlock(ev.Payload)
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
			log.Println("######enter broadcast")
			self.BroadcastConsensus(ev.Payload)
		case event.NewTxEvent:
			log.Println("######receiver new tx")
			//call consensus module
			//Todo


			/*payLoad:=self.transformTx(ev.Payload)
			if payLoad==nil{
				log.Fatal("payLoad nil")
			}*/

			//send msg to consensus
			/*msg := &protos.Message{
				Type: protos.Message_TRANSACTION,
				Payload: ev.Payload,
				Timestamp: time.Now().UnixNano(),
				Id: 0,
			}
			payload, _ := proto.Marshal(msg)
			self.consenter.RecvMsg(payload)*/
			for i:=0;i<200;i+=1{
				go self.sendMsg(ev.Payload)
				time.Sleep(100*time.Microsecond)
			}



		//sign tx


		case event.ConsensusEvent:
			//call consensus module
			//Todo
			log.Println("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)


		}

	}
}

func (self *ProtocolManager)sendMsg(payload []byte)  {
	msg := &protos.Message{
		Type: protos.Message_TRANSACTION,
		Payload: payload,
		Timestamp: time.Now().UnixNano(),
		Id: 0,
	}
	msgSend, _ := proto.Marshal(msg)
	self.consenter.RecvMsg(msgSend)
}


// Broadcast consensus msg to a batch of peers not knowing about it
func (pm *ProtocolManager) BroadcastConsensus(payload []byte) {
	pm.peerManager.BroadcastPeers(payload)

}

//receive tx from web,sign it and marshal it,then give it to consensus module
func (pm *ProtocolManager)transformTx(payLoad []byte) []byte {

	//var transaction types.Transaction
	transaction := &types.Transaction{}
	//decode tx
	proto.Unmarshal(payLoad, transaction)
	//hash tx
	h := transaction.SighHash(pm.commonHash)
	key, err := pm.encryption.GetKey()
	switch key.(type){
	case ecdsa.PrivateKey:
		actualKey := key.(ecdsa.PrivateKey)
		sign, err := pm.encryption.Sign(h[:], actualKey)
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
	if err != nil {
		return nil
	}
	return nil

}



func (pm *ProtocolManager) commitNewBlock(payload[]byte) {

	msgList := &protos.ExeMessage{}
	proto.Unmarshal(payload, msgList)
	block := new(types.Block)
	for _, item := range msgList.Batch {
		tx := &types.Transaction{}

		proto.Unmarshal(item.Payload, tx)
		block.Timestamp = item.Timestamp
		block.Transactions = append(block.Transactions, tx)
	}
	currentChain := core.GetChainCopy()
	block.Number = currentChain.Height + 1
	block.ParentHash = currentChain.LatestBlockHash

	//block.BlockHash=
	block.BlockHash = block.Hash(pm.commonHash).Bytes()
	//fmt.Println(block)

	core.WriteBlock(*block)

}




