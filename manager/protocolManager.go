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
)

type ProtocolManager struct {

	fetcher      *core.Fetcher
	peerManager  p2p.PeerManager
	consenter    consensus.Consenter



	noMorePeers  chan struct{}
	eventMux     *event.TypeMux
	txSub        event.Subscription
	newBlockSub  event.Subscription
	consensusSub event.Subscription
	quitSync     chan struct{}

	wg           sync.WaitGroup
}

func NewProtocolManager(peerManager p2p.PeerManager, fetcher *core.Fetcher, consenter consensus.Consenter) (*ProtocolManager) {

	eventMux := new(event.TypeMux)
	manager := &ProtocolManager{
		eventMux:    eventMux,
		quitSync:    make(chan struct{}),
		consenter:consenter,
		peerManager:  peerManager,
		fetcher:fetcher,

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
			self.consenter.RecvMsg(ev.Payload)

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


