package block

import (


	"hyperchain-alpha/event"
	"hyperchain-alpha/common"
	"hyperchain-alpha/core"

	"sync/atomic"



	//"sync"
)
type DoneEvent struct{}
type StartEvent struct{}
type FailedEvent struct{ Err error }

type BlockMaker struct {
	//mux *event.TypeMux
	eventMux     *event.TypeMux
	events       event.Subscription



	threads  int
	coinbase common.Address
	mining   int32
	eth      core.Ethereum
	//pow      pow.PoW
	//wg sync.WaitGroup // for shutdown sync

	canStart    int32 // can start indicates whether we can start the mining operation
	shouldStart int32 // should start indicates whether we should start after sync
}

func New(eth core.Ethereum,  eventMux *event.TypeMux) *BlockMaker {
	miner := &BlockMaker{
		eth: eth,
		eventMux:     eventMux,
		events:       eventMux.Subscribe(event.ChainHeadEvent{}, event.GasPriceChanged{}, event.RemovedTransactionEvent{}),

		, canStart: 1}

	go miner.update()

	return miner
}
func (self *BlockMaker) update() {
	//events := self.mux.Subscribe(StartEvent{}, DoneEvent{},FailedEvent{})
	//defer self.wg.Done()

	// Track chain events. When a chain events occurs (new chain canon block)
	// we need to know the new state. The new state will help us determine
	// the nonces in the managed state
	for ev := range self.events.Chan() {
		switch ev.Data.(type) {
		case StartEvent:
			atomic.StoreInt32(&self.canStart, 0)
			if self.Mining() {
				self.Stop()
				atomic.StoreInt32(&self.shouldStart, 1)
				}
		case DoneEvent, FailedEvent:
			shouldStart := atomic.LoadInt32(&self.shouldStart) == 1

			atomic.StoreInt32(&self.canStart, 1)
			atomic.StoreInt32(&self.shouldStart, 0)
			if shouldStart {
				self.Start(self.coinbase, self.threads)
			}
			// unsubscribe. we're only interested in this event once
			self.events.Unsubscribe()
			// stop immediately and ignore all further pending events
			break
		}
	}
}

func (self *BlockMaker) Stop() {
	//TODO
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.shouldStart, 0)
}

func (self *BlockMaker) Start(coinbase common.Address, threads int) {
	//TODO

}

func (self *BlockMaker) Mining() bool {
	return atomic.LoadInt32(&self.mining) > 0
}

//新增远程连接的peer时,同步区块
func ( self *BlockMaker) SyncWithPeer()  {
	self.eventMux.Post(StartEvent{})
	defer func() {
		// reset on error
		self.eventMux.Post(DoneEvent{})

	}()
	//TODO 同步新来的区块
}

