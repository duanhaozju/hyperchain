package manager

import (
	"github.com/ethereum/go-ethereum/core"
	"sync"
	"github.com/ethereum/go-ethereum/eth/fetcher"

	"hyperchain-alpha/event"

	"github.com/ethereum/go-ethereum/core/types"
	"hyperchain-alpha/common"

)


type ProtocolManager struct {
	networkId int

	txpool      txPool


	fetcher    *fetcher.Fetcher


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

func NewProtocolManager( mux *event.TypeMux, txpool txPool) (*ProtocolManager, error) {
	// Create the protocol manager with the base fields
	manager := &ProtocolManager{
		eventMux:    mux,
		txpool:      txpool,


		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
	}
	return manager, nil
}


//启动内部监听
func (pm *ProtocolManager) Start() {
	//监听本节点中产生的消息,然后广播给其它节点(主要广播)
	// broadcast transactions
	pm.txSub = pm.eventMux.Subscribe(core.TxPreEvent{})
	go pm.txBroadcastLoop()
	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(core.NewMinedBlockEvent{})
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
		case core.NewMinedBlockEvent:

			self.BroadcastBlock(ev.Block)
		}
	}
}
func (self *ProtocolManager) txBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range self.txSub.Chan() {
		event := obj.Data.(core.TxPreEvent)
		self.BroadcastTx(event.Tx.Hash(), event.Tx)
	}
}

//广播发送消息给其它连接的节点
func (pm *ProtocolManager) BroadcastBlock(block *types.Block) {

}
