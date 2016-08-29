// implement ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package manager

import "sync"



import (
	"hyperchain/event"

	"hyperchain/p2p"

	"hyperchain/core"
	"hyperchain/consensus"
	"hyperchain/crypto"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"fmt"
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
	quitSync     chan struct{}

	wg           sync.WaitGroup
}

func NewProtocolManager(peerManager p2p.PeerManager, fetcher *core.Fetcher, consenter consensus.Consenter,
encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {

	eventMux := new(event.TypeMux)
	manager := &ProtocolManager{
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:consenter,
		peerManager:  peerManager,
		fetcher:fetcher,
		encryption:encryption,
		commonHash:commonHash,


	}
	return manager
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
			self.BroadcastConsensus(ev.Payload)
		case event.NewTxEvent:
			//call consensus module
			//Todo
			var transaction *types.Transaction
			proto.Unmarshal(ev.Payload, &transaction)

			self.encryption.GetKey()

			sign, err := self.encryption.Sign(transaction.SighHash(
				self.commonHash), self.encryption.GetKey())
			if err != nil {
				fmt.Print(err)
			}
			transaction.Signature = sign
			proto.Marshal(transaction)

			self.consenter.RecvMsg(transaction)

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


