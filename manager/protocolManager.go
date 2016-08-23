package manager

import (

	"sync"
	"github.com/ethereum/go-ethereum/eth/fetcher"

	"hyperchain-alpha/event"

	"hyperchain-alpha/common"


	"hyperchain-alpha/block"

	"hyperchain-alpha/core/types"
)

type peerSet struct {
	peers  map[string]*peer
	lock   sync.RWMutex
	closed bool
}
type ProtocolManager struct {
	networkId int

	txpool      txPool


	fetcher    *fetcher.Fetcher
	peers      *peerSet


	BlockMaker *block.BlockMaker
	newPeerCh   chan *peer
	noMorePeers chan struct{}
	txsyncCh    chan *txsync
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


		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}
	return manager
}


//启动内部监听
func (pm *ProtocolManager) Start() {
	//监听本节点中产生的消息,然后广播给其它节点(主要广播)
	// broadcast transactions
	pm.txSub = pm.eventMux.Subscribe(event.TxPreEvent{})
	go pm.txBroadcastLoop()
	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(event.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()

	// start sync handlers,同步监听handle中传来的消息,然后处理
	go pm.syncer()
	go pm.txsyncLoop()
}

func (pm *ProtocolManager) BroadcastTx(hash common.Hash, tx *types.Transaction) {
	// Broadcast transaction to a batch of peers not knowing about it


}

// Mined broadcast loop
func (self *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.NewMinedBlockEvent:

			self.BroadcastBlock(ev.Block)
		}
	}
}
func (self *ProtocolManager) txBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.txSub.Chan() {
		event := obj.Data.(event.TxPreEvent)
		self.BroadcastTx(event.Tx.Hash(), event.Tx)
	}
}

//广播发送消息给其它连接的节点
func (pm *ProtocolManager) BroadcastBlock(block *types.Block) {

}
