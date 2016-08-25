package manager


import "sync"



import (




	"hyperchain-alpha/event"

	"hyperchain-alpha/common"


	"hyperchain-alpha/core/types"

	"hyperchain-alpha/p2p"


	"hyperchain-alpha/core"
	"hyperchain-alpha/node"
	"fmt"
)

type ProtocolManager struct {
	networkId int


	fetcher    *core.Fetcher
	peerManager           p2p.PeerManager
	node  node.Node


	newPeerCh   chan *p2p.Peer
	noMorePeers chan struct{}

	eventMux      *event.TypeMux
	txSub         event.Subscription
	newBlockSub event.Subscription

	consensusSub event.Subscription


	quitSync    chan struct{}

	wg sync.WaitGroup

}


func NewProtocolManager(mux *event.TypeMux, peerManager p2p.PeerManager,node node.Node,fetcher *core.Fetcher) (*ProtocolManager) {

	//eventmux:=new(event.TypeMux)
	manager := &ProtocolManager{
		eventMux:    mux,
		quitSync:    make(chan struct{}),

		peerManager:  peerManager,
		node:node,
		fetcher:fetcher,

	}
	return manager
}



func (pm *ProtocolManager) Start() {


	go pm.fetcher.Start()
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{},event.BroadcastConsensusEvent{})
	pm.newBlockSub = pm.eventMux.Subscribe(event.NewBlockEvent{})
	go pm.NewBlockLoop()
	go pm.ConsensusLoop()



}



//commit block into local db
func (self *ProtocolManager) NewBlockLoop() {

	for obj := range self.newBlockSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.NewBlockEvent:
			//self.eventMux.Post(event.ConsensusEvent{ev.Msg})
			self.fetcher.Enqueue(ev.Block)
			//self.BroadcastBlock(ev.Block)

		}
	}
}

func (self *ProtocolManager) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.BroadcastConsensusEvent:
			//event := obj.Data.(event.TxPreEvent)
			self.BroadcastConsensus(ev.Msg)
		case event.ConsensusEvent:
			fmt.Println("hahahahha")
			//call consensus module

		}
		//self.eventMux.Post(event.ConsensusEvent{ev.Msg})
		//pm.fetcher.Enqueue(p.id, request.Block)
	}
}


func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *types.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it


}


func (pm *ProtocolManager) BroadcastConsensus(msg *types.Msg) {
	pm.peerManager.BroadcastPeers(msg)


	// Broadcast transaction to a batch of peers not knowing about it


}


