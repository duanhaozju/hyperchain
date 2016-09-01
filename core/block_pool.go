package core

import (
	"sync"
	"hyperchain/event"

	"hyperchain/core/types"
	"hyperchain/logger"

	"log"
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
	//pool.stateLock.Lock()
	//defer pool.stateLock.Unlock()
	if (block.Number == 0) {
		WriteBlock(*block,commonHash)
		return
	}

	if (block.Number > pool.maxNum) {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number ]; ok {
		myLogger.GetLogger().Println("replated block number,number is: ",block.Number)
		return
	}


	log.Println("number is ",block.Number)

	currentChain := GetChainCopy()

	if (currentChain.Height>=block.Number) {

		myLogger.GetLogger().Println("replated block number,number is: ",block.Number)
		return
	}

	if(pool.demandNumber==block.Number) {


		pool.mu.RLock()

		pool.demandNumber+=1
		log.Println("current demandNumber is ",pool.demandNumber)
		WriteBlock(*block,commonHash)
		pool.mu.RUnlock()


			for i := block.Number + 1; i <= pool.maxNum; i += 1 {
				if _, ok := pool.queue[i]; ok {//存在}

					//if (pool.queue[block.Number + 1]) {
					pool.mu.RLock()
					pool.demandNumber+=1
					log.Println("current demandNumber is ",pool.demandNumber)
					WriteBlock(*pool.queue[i],commonHash)
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

/*//check block sequence and validate in chain
func (pool *BlockPool)AddBlock(block *types.Block) {
	pool.stateLock.Lock()
	defer pool.stateLock.Unlock()
	if (block.Number == 0) {
		WriteBlock(*block)
		return
	}

	if (block.Number > pool.maxNum) {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number ]; ok {
		myLogger.GetLogger().Println("replated block number,number is: ",block.Number)
		return
	}


	log.Println("number is ",block.Number)

	currentChain := GetChainCopy()
	log.Println("current height is ",currentChain.Height)
	if (currentChain.Height>=block.Number) {

		myLogger.GetLogger().Println("replated block number,number is: ",block.Number)
		return
	}

	if(currentChain.Height==block.Number - 1) {


		WriteBlock(*block)
		if (pool.demandNumber == block.Number) {
			pool.demandNumber = block.Number - 1
			for i := block.Number + 1; i < pool.maxNum; i += 1 {
				if _, ok := pool.queue[block.Number + 1]; ok {//存在}

				//if (pool.queue[block.Number + 1]) {
					WriteBlock(*pool.queue[block.Number + 1])

				} else {
					pool.demandNumber = i
					break
				}

			}

		}

		return
	} else {
		if (pool.demandNumber == block.Number) {
			pool.demandNumber = block.Number - 1
			pool.queue[block.Number] = block

		}else {
			*//*if _, ok := pool.queue[block.Number + 1]; ok {

			}*//*
			pool.queue[block.Number] = block
		}


	}

}*/
