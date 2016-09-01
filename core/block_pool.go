package core

import (
	"sync"
	"hyperchain/event"

	"hyperchain/core/types"
)

const (
	maxQueued = 64 // max limit of queued block in pool
)

type BlockPool struct {
	demandNumber uint64
	maxNum       uint64

	queue        map[uint64]*types.Block
	eventMux     *event.TypeMux
	events       event.Subscription
	mu           sync.RWMutex
	wg           sync.WaitGroup // for shutdown sync
}

func NewBlockPool(eventMux *event.TypeMux) *BlockPool {
	pool := &BlockPool{
		eventMux:     eventMux,

		events:       eventMux.Subscribe(event.NewBlockPoolEvent{}),
	}

	pool.wg.Add(1)
	go pool.eventLoop()

	return pool
}

func (pool *BlockPool) eventLoop() {
	defer pool.wg.Done()

	for ev := range pool.events.Chan() {
		switch  ev.Data.(type) {
		case event.NewBlockPoolEvent:
			pool.mu.Lock()
			/*if ev.Block != nil && pool.config.IsHomestead(ev.Block.Number()) {
				pool.homestead = true
			}

			pool.resetState()*/
			pool.mu.Unlock()

		}
	}
}

func (pool *BlockPool)AddBlock(block *types.Block) {
	if (block.Number == 0) {
		WriteBlock(*block)
		return
	}
	if (block.Number > pool.maxNum) {
		pool.maxNum = block.Number
	}
	currentChain := GetChainCopy()

	if(currentChain.Height>block.Number - 1) {

		WriteBlock(*block)
		if (pool.demandNumber == block.Number) {
			pool.demandNumber == block.Number - 1
			for i := block.Number + 1; i < pool.maxNum; i += 1 {
				if (pool.queue[block.Number + 1]) {
					WriteBlock(pool.queue[block.Number + 1])

				} else {
					pool.demandNumber = i
					break
				}

			}

		}

		return
	} else {
		if (pool.demandNumber == block.Number) {
			pool.demandNumber == block.Number - 1

		}
		pool.queue[block.Number] = block

	}

}