// implement ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
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
)

type ProtocolManager struct {
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
func NewProtocolManager(peerManager p2p.PeerManager, eventMux *event.TypeMux, fetcher *core.Fetcher, consenter consensus.Consenter,
encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {

	manager := &ProtocolManager{
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:consenter,
		peerManager:  peerManager,
		fetcher:fetcher,
		encryption:encryption,
		commonHash:commonHash,


	}
	eventMuxAll=eventMux
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

	/*for i := 0; i < 100; i += 1 {
		fmt.Println("enenen")
		eventmux := new(event.TypeMux)
		eventmux.Post(event.BroadcastConsensusEvent{[]byte{0x00, 0x00, 0x03, 0xe8}})

	}*/

	pm.wg.Wait()

}



// listen block msg
func (self *ProtocolManager) NewBlockLoop() {

	for obj := range self.newBlockSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.NewBlockEvent:
			//commit block into local db
			self.fetcher.Enqueue(ev.Payload)

		}
	}
}

//listen consensus msg
func (self *ProtocolManager) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.BroadcastConsensusEvent:
			fmt.Println("enter broadcast")
			self.BroadcastConsensus(ev.Payload)
		case event.NewTxEvent:
			//call consensus module
			//Todo
			fmt.Println("get new TxEvent")
			var transaction *types.Transaction
			//decode tx
			proto.Unmarshal(ev.Payload, transaction)
			//hash tx
			h := transaction.SighHash(self.commonHash)
			key, err := self.encryption.GetKey()
			switch key.(type){
			case ecdsa.PrivateKey:
				actualKey:=key.(ecdsa.PrivateKey)
				sign, err := self.encryption.Sign(h[:], actualKey)
				if err != nil {
					fmt.Print(err)
				}
				transaction.Signature = sign
				//encode tx
				payLoad, err := proto.Marshal(transaction)
				if err != nil {
					return
				}

				self.consenter.RecvMsg(payLoad)
			}
			if err != nil {
				return
			}
			//sign tx


		case event.ConsensusEvent:
			//call consensus module
			//Todo
			self.consenter.RecvMsg(ev.Payload)


		}

	}
}



// Broadcast consensus msg to a batch of peers not knowing about it
func (pm *ProtocolManager) BroadcastConsensus(payload []byte) {
	pm.peerManager.BroadcastPeers(payload)

}




