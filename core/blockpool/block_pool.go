//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package blockpool

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	statedb "hyperchain/core/state"
	"hyperchain/core/hyperstate"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"sync"
	"sync/atomic"
	"errors"
	"hyperchain/core/vm"
	"hyperchain/tree/bucket"
)

var (
	log                          *logging.Logger // package-level logger
	globalState                  vm.Database
)

func init() {
	log = logging.MustGetLogger("block-pool")
}

// represent a validation result
type BlockRecord struct {
	TxRoot      []byte                             // hash of a batch of transactions
	ReceiptRoot []byte                             // hash of a batch of receipts
	MerkleRoot  []byte                             // hash of state
	InvalidTxs  []*types.InvalidTransactionRecord  // invalid transaction list
	ValidTxs    []*types.Transaction               // valid transaction list
	Receipts    []*types.Receipt                   // receipt list
	SeqNo       uint64                             // temp block number for this batch
	VID         uint64                             // validation ID. may larger than SeqNo
}

// block pool configuration
type BlockPoolConf struct {
	BlockVersion       string                     // block structure version
	TransactionVersion string                     // transaction structure version
	StateType          string                     // state type identifier, "rawstate"  or "hyperstate"
}

type BlockPool struct {
	// IMPORTANT atomic variable. make sure use atomic operation to use those variable
	demandNumber        uint64              // current demand number for commit
	demandSeqNo         uint64              // current demand seqNo for validation
	maxNum              uint64              // max block number in queue cache for commit
	maxSeqNo            uint64              // max validation event number in validation queue
	lastValidationState atomic.Value        // latest state root hash
	// external stuff
	consenter           consensus.Consenter // consensus module handler
	eventMux            *event.TypeMux      // message queue
	stateLock           sync.Mutex          // block pool lock
	wg                  sync.WaitGroup      // for shutdown sync
	// thread safe cache
	blockCache          *common.Cache       // cache for validation result
	validationQueue     *common.Cache       // cache for storing validation event
	queue               *common.Cache       // cache for storing commit event
	// config
	conf                BlockPoolConf       // block configuration
	bucketTreeConf      bucket.Conf         // bucket tree configuration for hyperstate use only
	// buffer
	transactionCalculator   interface{}       // a batch of transactions calculator
	receiptCalculator       interface{}       // a batch of receipts calculator
	transactionBuffer       [][]byte          // transaction buffer
	receiptBuffer           [][]byte          // receipt buffer
	// status variant
	tempBlockNumber         uint64            // temporarily block number
}

func NewBlockPool(eventMux *event.TypeMux, consenter consensus.Consenter, conf BlockPoolConf, bktConf bucket.Conf) *BlockPool {
	var err error
	blockcache, err := common.NewCache()
	if err != nil {return nil}
	queue, err := common.NewCache()
	if err != nil {return nil}
	validationqueue, err := common.NewCache()
	if err != nil {return nil}

	pool := &BlockPool{
		eventMux:        eventMux,
		consenter:       consenter,
		queue:           queue,
		validationQueue: validationqueue,
		blockCache:      blockcache,
		conf:            conf,
		bucketTreeConf:  bktConf,
	}
	// 1. set demand number and demand seqNo
	currentChain := core.GetChainCopy()
	atomic.StoreUint64(&pool.demandNumber, currentChain.Height+1)
	atomic.AddUint64(&pool.demandSeqNo, currentChain.Height+1)
	// initialize temp block number as current height plus 1
	pool.tempBlockNumber = currentChain.Height + 1
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {return nil}
	// get latest block
	blk, err := core.GetBlock(db, currentChain.LatestBlockHash)
	if err != nil {return nil}
	// 2. set current state root hash
	pool.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
	log.Noticef("Block pool Initialize demandNumber :%d, demandseqNo: %d\n", pool.demandNumber, pool.demandSeqNo)
	return pool
}
// obtain block pool configuration
func (pool *BlockPool) GetConfig() BlockPoolConf {
	return pool.conf
}
// set demand number
func (pool *BlockPool) SetDemandNumber(number uint64) {
	atomic.StoreUint64(&pool.demandNumber, number)
}
// set demand seqNo
func (pool *BlockPool) SetDemandSeqNo(seqNo uint64) {
	atomic.StoreUint64(&pool.demandSeqNo, seqNo)
}
// obtain state handler via configuration in block.conf
// two state: (1)raw state (2) hyper state are supported
func (pool *BlockPool) GetStateInstance(root common.Hash, db hyperdb.Database) (vm.Database, error) {
	switch pool.conf.StateType {
	case "rawstate":
		return statedb.New(root, db)
	case "hyperstate":
		// IMPORTANT initialize hyperstate only once
		if globalState == nil {
			var err error
			height := core.GetHeightOfChain()
			globalState, err = hyperstate.New(root, db, pool.bucketTreeConf, height)
			return globalState, err
		} else {
			return globalState, nil
		}
	default:
		return nil, errors.New("no state type specified")
	}
}
// create a latest state for simulate usage
// different with function `GetStateInstance`, this function will create a new instance each time when got invocation
func (pool *BlockPool) GetStateInstanceForSimulate(root common.Hash, db hyperdb.Database) (vm.Database, error) {
	switch pool.conf.StateType {
	case "rawstate":
		return statedb.New(root, db)
	case "hyperstate":
		height := core.GetHeightOfChain()
		return hyperstate.New(root, db, pool.bucketTreeConf, height)
	default:
		return nil, errors.New("no state type specified")
	}
}

