// implementblock pool
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-09-01
package core

import (
	"sync"
	"hyperchain/event"

	"hyperchain/core/types"
	"hyperchain/crypto"
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
	stateLock    sync.Mutex
	wg           sync.WaitGroup // for shutdown sync
}

func NewBlockPool(eventMux *event.TypeMux) *BlockPool {
	pool := &BlockPool{
		eventMux:     eventMux,

		queue  :make(map[uint64]*types.Block),

		events:       eventMux.Subscribe(event.NewBlockPoolEvent{}),
	}


	//pool.wg.Add(1)
	//go pool.eventLoop()

	currentChain := GetChainCopy()
	pool.demandNumber=currentChain.Height+1
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


//check block sequence and validate in chain
func (pool *BlockPool)AddBlock(block *types.Block,commonHash crypto.CommonHash) {

	if (block.Number == 0) {
		WriteBlock(block,commonHash)
		return
	}

	if (block.Number > pool.maxNum) {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number ]; ok {
		log.Info("replated block number,number is: ",block.Number)
		return
	}


	log.Info("number is ",block.Number)

	currentChain := GetChainCopy()

	if (currentChain.Height>=block.Number) {

		log.Info("replated block number,number is: ",block.Number)
		return
	}

	if(pool.demandNumber==block.Number) {

		pool.mu.RLock()
		pool.demandNumber+=1
		log.Info("current demandNumber is ",pool.demandNumber)

		WriteBlock(block,commonHash)


		pool.mu.RUnlock()

			for i := block.Number + 1; i <= pool.maxNum; i += 1 {
				if _, ok := pool.queue[i]; ok {//存在}

					pool.mu.RLock()
					pool.demandNumber+=1
					log.Info("current demandNumber is ",pool.demandNumber)
					WriteBlock(pool.queue[i],commonHash)
					delete(pool.queue,i)
					pool.mu.RUnlock()


				} else {
					break
				}

			}

		return
	} else {

			pool.queue[block.Number] = block




	}

}

