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
	"time"
	"encoding/hex"
	"hyperchain/hyperdb"
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
func (pool *BlockPool)AddBlock(block *types.Block,commonHash crypto.CommonHash,commitTime int64) {

	if (block.Number == 0) {
		WriteBlock(block,commonHash,commitTime)
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

		WriteBlock(block,commonHash,commitTime)


		pool.mu.RUnlock()

			for i := block.Number + 1; i <= pool.maxNum; i += 1 {
				if _, ok := pool.queue[i]; ok {//存在}

					pool.mu.RLock()
					pool.demandNumber+=1
					log.Info("current demandNumber is ",pool.demandNumber)
					WriteBlock(pool.queue[i],commonHash,commitTime)
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


// WriteBlock need:
// 1. Put block into db
// 2. Put transactions in block into db  (-- cancel --)
// 3. Update chain
// 4. Update balance
func WriteBlock(block *types.Block, commonHash crypto.CommonHash,commitTime int64)  {

	log.Info("block number is ",block.Number)
	currentChain := GetChainCopy()
	block.ParentHash = currentChain.LatestBlockHash
	block.BlockHash = block.Hash(commonHash).Bytes()
	//block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}

	UpdateChain(block, false)
	balance, err := GetBalanceIns()
	if err != nil {
		log.Fatal(err)
	}

	newChain := GetChainCopy()
	log.Notice("Block number",newChain.Height)
	log.Notice("Block hash",hex.EncodeToString(newChain.LatestBlockHash))
	block.WriteTime = time.Now().UnixNano()




	balance.UpdateDBBalance(block)



	if block.Number%10==0 && block.Number!=0{
		WriteChainChan()

	}
	// update our stateObject and statedb to blockchain
	//ExecBlock(block)
	block.EvmTime=time.Now().UnixNano()
	err = PutBlock(db, block.BlockHash, block)
	// write transaction
	//PutTransactions(db, commonHash, block.Transactions)
	if err != nil {
		log.Fatal(err)
	}
	//TxSum.Add(TxSum,big.NewInt(int64(len(block.Transactions))))
	//CommitStatedbToBlockchain()
}




