// implementblock pool
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-09-01
package core

import (
	"hyperchain/event"
	"sync"

	"encoding/hex"

	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"strconv"
	"time"
)

const (
	maxQueued = 64 // max limit of queued block in pool
)

type BlockPool struct {
	demandNumber uint64
	maxNum       uint64

	queue     map[uint64]*types.Block
	eventMux  *event.TypeMux
	events    event.Subscription
	mu        sync.RWMutex
	stateLock sync.Mutex
	wg        sync.WaitGroup // for shutdown sync
}

func NewBlockPool(eventMux *event.TypeMux) *BlockPool {
	pool := &BlockPool{
		eventMux: eventMux,

		queue: make(map[uint64]*types.Block),

		events: eventMux.Subscribe(event.NewBlockPoolEvent{}),
	}

	//pool.wg.Add(1)
	//go pool.eventLoop()

	currentChain := GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
	return pool
}

func (pool *BlockPool) eventLoop() {
	defer pool.wg.Done()

	for ev := range pool.events.Chan() {
		switch ev.Data.(type) {
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
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {

	if block.Number == 0 {
		WriteBlock(block, commonHash, commitTime)
		return
	}

	if block.Number > pool.maxNum {
		pool.maxNum = block.Number
	}
	if _, ok := pool.queue[block.Number]; ok {
		log.Info("replated block number,number is: ", block.Number)
		return
	}

	log.Info("number is ", block.Number)

	currentChain := GetChainCopy()

	if currentChain.Height >= block.Number {
		//todo view change ,delete block and rewrite block

		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Fatal(err)
		}

		block,_:=GetBlockByNumber(db,block.Number)
		//rollback chain height,latestHash
		UpdateChainByViewChange(block.Number-1,block.ParentHash)
		keyNum := strconv.FormatInt(int64(block.Number), 10)
		DeleteBlock(db,append(blockNumPrefix, keyNum...))
		WriteBlock(block, commonHash, commitTime)
		log.Notice("replated block number,number is: ", block.Number)
		return
	}

	if pool.demandNumber == block.Number {

		pool.mu.RLock()
		pool.demandNumber += 1
		log.Info("current demandNumber is ", pool.demandNumber)

		WriteBlock(block, commonHash, commitTime)

		pool.mu.RUnlock()

		for i := block.Number + 1; i <= pool.maxNum; i += 1 {
			if _, ok := pool.queue[i]; ok { //存在}

				pool.mu.RLock()
				pool.demandNumber += 1
				log.Info("current demandNumber is ", pool.demandNumber)
				WriteBlock(pool.queue[i], commonHash, commitTime)
				delete(pool.queue, i)
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
func WriteBlock(block *types.Block, commonHash crypto.CommonHash, commitTime int64) {
	log.Info("block number is ", block.Number)

	currentChain := GetChainCopy()

	block.ParentHash = currentChain.LatestBlockHash
	if err := ProcessBlock(block); err != nil {
		log.Fatal(err)
	}

	block.WriteTime = time.Now().UnixNano()
	block.CommitTime = commitTime
	block.BlockHash = block.Hash(commonHash).Bytes()

	UpdateChain(block, false)

	// update our stateObject and statedb to blockchain
	//ExecBlock(block)
	block.EvmTime = time.Now().UnixNano()

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := PutBlock(db, block.BlockHash, block); err != nil {
		log.Fatal(err)
	}
	// write transaction
	//PutTransactions(db, commonHash, block.Transactions)

	newChain := GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))

	if block.Number%10 == 0 && block.Number != 0 {
		WriteChainChan()
	}
	//TxSum.Add(TxSum,big.NewInt(int64(len(block.Transactions))))
	//CommitStatedbToBlockchain()
}



func ProcessBlock(block *types.Block) error {
	var (
		receipts types.Receipts
		env      = make(map[string]string)
	)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	parentBlock, _ := GetBlock(db, block.ParentHash)
	statedb, e := state.New(common.BytesToHash(parentBlock.MerkleRoot), db)
	//fmt.Printf("[Before Process %d] %s\n", block.Number, string(statedb.Dump()))
	if err != nil {
		return e
	}
	env["currentNumber"] = strconv.FormatUint(block.Number, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	for i, tx := range block.Transactions {
		statedb.StartRecord(tx.BuildHash(), common.Hash{}, i)
		receipt, _, _, _ := ExecTransaction(*tx, vmenv)
		receipts = append(receipts, receipt)
	}
	receiptInst, _ := GetReceiptInst()
	for _, receipt := range receipts {
		receiptInst.PutReceipt(common.BytesToHash(receipt.TxHash), receipt)
	}
		//WriteReceipts(receipts)

	root, _ := statedb.Commit()

	block.MerkleRoot = root.Bytes()
	//fmt.Printf("[After Process %d] %s\n", block.Number, string(statedb.Dump()))
	return nil
}
