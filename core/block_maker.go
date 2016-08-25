package core

import (


	"hyperchain-alpha/event"
	"hyperchain-alpha/common"




	//"sync"
)


type BlockMaker struct {
	//mux *event.TypeMux
	eventMux     *event.TypeMux
	events       event.Subscription



	threads  int
	coinbase common.Address

}

func NewBlockMaker( eventMux *event.TypeMux) *BlockMaker {
	blockMaker := &BlockMaker{

		eventMux:     eventMux,
		events:       eventMux.Subscribe(event.ChainHeadEvent{}, event.GasPriceChanged{}, event.RemovedTransactionEvent{}),

		}

	return blockMaker
}


func (self *BlockMaker) Start(coinbase common.Address, threads int) {
	//TODO

	for ev := range self.events.Chan() {
		switch ev.Data.(type) {
		case event.ChainHeadEvent:
		//TODO
		}
	}

}




