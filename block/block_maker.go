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
	blockMaker := &BlockMaker{
		eth: eth,
		eventMux:     eventMux,
		events:       eventMux.Subscribe(event.ChainHeadEvent{}, event.GasPriceChanged{}, event.RemovedTransactionEvent{}),

		, canStart: 1}



	return blockMaker
}
func (self *BlockMaker) update() {
	//events := self.mux.Subscribe(StartEvent{}, DoneEvent{},FailedEvent{})
	//defer self.wg.Done()

	for ev := range self.events.Chan() {
		switch ev.Data.(type) {
		case event.ChainHeadEvent:
			//TODO
		}
	}
}



func (self *BlockMaker) Start(coinbase common.Address, threads int) {
	//TODO
	go self.update()

}




