//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package blockpool

import (
	"errors"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/hyperstate"
	statedb "hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"sync/atomic"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/crypto"
)

var (
	log         *logging.Logger // package-level logger
	globalState vm.Database
)

const (
	COMMITQUEUESIZE = 10
	VALIDATEQUEUESIZE = 40

	VALIDATEBEHAVETYPE_NORMAL = 0
	VALIDATEBEHAVETYPE_DROP = 1
)

func init() {
	log = logging.MustGetLogger("block-pool")
}

// represent a validation result
type BlockRecord struct {
	TxRoot      []byte                            // hash of a batch of transactions
	ReceiptRoot []byte                            // hash of a batch of receipts
	MerkleRoot  []byte                            // hash of state
	InvalidTxs  []*types.InvalidTransactionRecord // invalid transaction list
	ValidTxs    []*types.Transaction              // valid transaction list
	Receipts    []*types.Receipt                  // receipt list
	SeqNo       uint64                            // temp block number for this batch
	VID         uint64                            // validation ID. may larger than SeqNo
}

type BlockPool struct {
	demandNumber        uint64                // current demand number for commit
	demandSeqNo         uint64                // current demand seqNo for validation
	maxNum              uint64                // max block number in queue cache for commit
	maxSeqNo              uint64              // max validation event number in validation queue
	tempBlockNumber       uint64              // temporarily block number
	lastValidationState   atomic.Value        // latest state root hash
						  // external stuff
	consenter             consensus.Consenter // consensus module handler
	peerManager           p2p.PeerManager
	commonHash            crypto.CommonHash
	encryption            crypto.Encryption
						  // thread safe cache
	blockCache            *common.Cache       // cache for validation result
	validateEventQueue    *common.Cache       // cache for storing validation event
	queue                 *common.Cache       // cache for storing commit event
						  // config
	conf                  *common.Config      // block configuration
						  // hash utils
	transactionCalculator interface{}         // a batch of transactions calculator
	receiptCalculator     interface{}         // a batch of receipts calculator
	transactionBuffer     [][]byte            // transaction buffer
	receiptBuffer         [][]byte            // receipt buffer
						  // commit queue
	validateQueue         chan event.ExeTxsEvent
	commitQueue           chan event.CommitOrRollbackBlockEvent

	validateBehaveFlag    int32

}

func NewBlockPool(consenter consensus.Consenter, conf *common.Config, commonHash crypto.CommonHash, encryption crypto.Encryption) *BlockPool {
	var err error
	blockCache, err := common.NewCache()
	if err != nil {
		return nil
	}
	queue, err := common.NewCache()
	if err != nil {
		return nil
	}
	validationQueue, err := common.NewCache()
	if err != nil {
		return nil
	}

	pool := &BlockPool{
		consenter:       consenter,
		queue:           queue,
		validateEventQueue: validationQueue,
		blockCache:      blockCache,
		conf:            conf,
	}
	// 1. set demand number and demand seqNo
	currentChain := core.GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
	pool.demandSeqNo = currentChain.Height + 1
	pool.tempBlockNumber = currentChain.Height + 1

	pool.commonHash = commonHash
	pool.encryption = encryption
	pool.validateQueue = make(chan event.ExeTxsEvent, VALIDATEQUEUESIZE)
	pool.commitQueue = make(chan event.CommitOrRollbackBlockEvent, COMMITQUEUESIZE)
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return nil
	}
	// get latest block
	blk, err := core.GetBlock(db, currentChain.LatestBlockHash)
	if err != nil {
		log.Errorf("get block #%d failed.", blk.Number)
		pool.lastValidationState.Store(common.Hash{})
		return pool
	} else {
		log.Noticef("Block pool Initialize demandNumber :%d, demandseqNo: %d\n", pool.demandNumber, pool.demandSeqNo)
		pool.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
		return pool
	}

	// 2. set current state root hash
	log.Noticef("block pool Initialize. current chain height #%d, latest block hash %s, demandNumber #%d, demandseqNo #%d, temp block number #%d\n",
		currentChain.Height, common.Bytes2Hex(currentChain.LatestBlockHash), pool.demandNumber, pool.demandSeqNo, pool.tempBlockNumber)
	return pool
}

func (pool *BlockPool) Initialize() {
	go pool.commitBackendLoop()
	go pool.validateBackendLoop()
}

// SetDemandNumber - set demand number.
func (pool *BlockPool) SetDemandNumber(number uint64) {
	atomic.StoreUint64(&pool.demandNumber, number)
}

// SetDemandSeqNo - set demand seqNo.
func (pool *BlockPool) SetDemandSeqNo(seqNo uint64) {
	atomic.StoreUint64(&pool.demandSeqNo, seqNo)
}

// IncreaseTempBlockNumber - increase temporary block number.
func (pool *BlockPool) IncreaseTempBlockNumber() {
	pool.tempBlockNumber = pool.tempBlockNumber + 1
}

// SetTempBlockNumber - set temporary block number
func (pool *BlockPool) SetTempBlockNumber(seqNo uint64) {
	pool.tempBlockNumber = seqNo
}

// PurgeValidateQueue - clear validation event queue cache.
func (pool *BlockPool) PurgeValidateQueue() {
	pool.validateEventQueue.Purge()
}

// PurgeBlockCache - clear validation result cache
func (pool *BlockPool) PurgeBlockCache() {
	pool.blockCache.Purge()
}

// GetStateInstance - obtain state handler via configuration in block.conf
// two state: (1)raw state (2) hyper state are supported.
func (pool *BlockPool) GetStateInstance() (vm.Database, error) {
	// obtain latest root
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return nil, err
	}
	v := pool.lastValidationState.Load()
	latestRoot, ok := v.(common.Hash)
	if ok == false {
		return nil, err
	}
	switch pool.GetStateType() {
	case "rawstate":
		return statedb.New(latestRoot, db)
	case "hyperstate":
		// initialize hyperstate only once
		if globalState == nil {
			var err error
			height := core.GetHeightOfChain()
			globalState, err = hyperstate.New(latestRoot, db, pool.conf, height)
			return globalState, err
		} else {
			return globalState, nil
		}
	default:
		return nil, errors.New("no state type specified")
	}
}

// GetStateInstanceForSimulate - create a latest state for simulate usage
// different with function `GetStateInstance`, this function will create a new instance each time when got invocation.
func (pool *BlockPool) GetStateInstanceForSimulate(root common.Hash, db hyperdb.Database) (vm.Database, error) {

	switch pool.GetStateType() {
	case "rawstate":
		return statedb.New(root, db)
	case "hyperstate":
		height := core.GetHeightOfChain()
		return hyperstate.New(root, db, pool.conf, height)
	default:
		return nil, errors.New("no state type specified")
	}
}
