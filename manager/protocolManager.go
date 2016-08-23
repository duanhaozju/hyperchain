package manager

import (

	"sync"


	"hyperchain-alpha/event"

	"hyperchain-alpha/common"

	"hyperchain-alpha/block"

	"hyperchain-alpha/core/types"

	"hyperchain-alpha/transaction"
)

type peerSet struct {
	peers  map[string]*peer
	lock   sync.RWMutex
	closed bool
}
type ProtocolManager struct {
	networkId int

	txpool      *transaction.TxPool


	fetcher    *Fetcher
	peers      *peerSet


	BlockMaker *block.BlockMaker
	newPeerCh   chan *peer
	noMorePeers chan struct{}

	eventMux      *event.TypeMux
	txSub         event.Subscription
	minedBlockSub event.Subscription

			// channels for fetcher, syncer, txsyncLoop

	quitSync    chan struct{}

			// wait group is used for graceful shutdowns during downloading
			// and processing
	wg sync.WaitGroup

	badBlockReportingEnabled bool
}

func NewProtocolManager( mux *event.TypeMux, txpool txPool) (*ProtocolManager) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		eventMux:    mux,
		txpool:      txpool,
		quitSync:    make(chan struct{}),
	}
	return manager
}


//启动内部监听
func (pm *ProtocolManager) Start() {



	go pm.fetcher.Start()
/*	pm.txSub = pm.eventMux.Subscribe(event.TxPreEvent{})
	go pm.txBroadcastLoop()*/
	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(event.NewMinedBlockEvent{},event.TxPreEvent{})
	go pm.BroadcastLoop()
	go pm.BlockMaker.Start()
	go pm.txpool.Start()



}



// Mined broadcast loop
func (self *ProtocolManager) BroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.NewMinedBlockEvent:
			self.fetcher.Enqueue( ev.Block)
			self.BroadcastBlock(ev.Block)
		case event.TxPreEvent:
			//event := obj.Data.(event.TxPreEvent)
			self.BroadcastTx(ev.Tx.Hash(), ev.Tx)
		case event.ConsensusEvent:
			//event := obj.Data.(event.TxPreEvent)
			self.BroadcastConsensusEvent(ev.Msg)
		}
		//pm.fetcher.Enqueue(p.id, request.Block)
	}
}



func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *types.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it


}


func (pm *ProtocolManager) BroadcastConsensusEvent(msg *types.Msg) {
	// Broadcast transaction to a batch of peers not knowing about it


}

//广播发送消息给其它连接的节点
func (pm *ProtocolManager) BroadcastBlock(block *types.Block) {

}
