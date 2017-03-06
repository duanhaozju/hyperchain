//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package executor

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
	"hyperchain/hyperdb/db"
)

var (
	log         *logging.Logger // package-level logger
	globalState vm.Database
)

const (
	COMMITQUEUESIZE = 1024
	VALIDATEQUEUESIZE = 1024

	VALIDATEBEHAVETYPE_NORMAL = 0
	VALIDATEBEHAVETYPE_DROP = 1

	PROGRESS_TRUE = 1
	PROGRESS_FALSE = 0
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

type Executor struct {
	demandNumber        uint64                // current demand number for commit
	demandSeqNo         uint64                // current demand seqNo for validation
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
	validateInProgress    int32
	commitInProgress      int32
	validateQueueLen      int32
	commitQueueLen        int32
	helper                *Helper

}

func NewBlockExecutor(consenter consensus.Consenter, conf *common.Config, commonHash crypto.CommonHash, encryption crypto.Encryption, eventMux *event.TypeMux) *Executor {
	var err error
	blockCache, err := common.NewCache()
	if err != nil {
		return nil
	}
	validationQueue, err := common.NewCache()
	if err != nil {
		return nil
	}
	helper := NewHelper(eventMux)
	executor := &Executor{
		consenter:       consenter,
		validateEventQueue: validationQueue,
		blockCache:      blockCache,
		conf:            conf,
		helper:          helper,
	}
	// 1. set demand number and demand seqNo
	currentChain := core.GetChainCopy()
	executor.demandNumber = currentChain.Height + 1
	executor.demandSeqNo = currentChain.Height + 1
	executor.tempBlockNumber = currentChain.Height + 1

	executor.commonHash = commonHash
	executor.encryption = encryption
	executor.validateQueue = make(chan event.ExeTxsEvent, VALIDATEQUEUESIZE)
	executor.commitQueue = make(chan event.CommitOrRollbackBlockEvent, COMMITQUEUESIZE)
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return nil
	}
	// get latest block
	blk, err := core.GetBlock(db, currentChain.LatestBlockHash)
	if err != nil {
		log.Errorf("get block #%d failed.", blk.Number)
		executor.lastValidationState.Store(common.Hash{})
		return executor
	} else {
		log.Noticef("Block pool Initialize demandNumber :%d, demandseqNo: %d\n", executor.demandNumber, executor.demandSeqNo)
		executor.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
		return executor
	}

	// 2. set current state root hash
	log.Noticef("block pool Initialize. current chain height #%d, latest block hash %s, demandNumber #%d, demandseqNo #%d, temp block number #%d\n",
		currentChain.Height, common.Bytes2Hex(currentChain.LatestBlockHash), executor.demandNumber, executor.demandSeqNo, executor.tempBlockNumber)
	return executor
}

func (excutor *Executor) Initialize() {
	go excutor.commitBackendLoop()
	go excutor.validateBackendLoop()
}

// SetDemandNumber - set demand number.
func (excutor *Executor) SetDemandNumber(number uint64) {
	atomic.StoreUint64(&excutor.demandNumber, number)
}

// SetDemandSeqNo - set demand seqNo.
func (excutor *Executor) SetDemandSeqNo(seqNo uint64) {
	atomic.StoreUint64(&excutor.demandSeqNo, seqNo)
}

// IncreaseTempBlockNumber - increase temporary block number.
func (excutor *Executor) IncreaseTempBlockNumber() {
	excutor.tempBlockNumber = excutor.tempBlockNumber + 1
}

// SetTempBlockNumber - set temporary block number
func (excutor *Executor) SetTempBlockNumber(seqNo uint64) {
	excutor.tempBlockNumber = seqNo
}

// PurgeValidateQueue - clear validation event queue cache.
func (excutor *Executor) PurgeValidateQueue() {
	excutor.validateEventQueue.Purge()
}

// PurgeBlockCache - clear validation result cache
func (excutor *Executor) PurgeBlockCache() {
	excutor.blockCache.Purge()
}

// GetStateInstance - obtain state handler via configuration in block.conf
// two state: (1)raw state (2) hyper state are supported.
func (excutor *Executor) GetStateInstance() (vm.Database, error) {
	// obtain latest root
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return nil, err
	}
	v := excutor.lastValidationState.Load()
	latestRoot, ok := v.(common.Hash)
	if ok == false {
		return nil, err
	}
	switch excutor.GetStateType() {
	case "rawstate":
		return statedb.New(latestRoot, db)
	case "hyperstate":
		// initialize hyperstate only once
		if globalState == nil {
			var err error
			height := core.GetHeightOfChain()
			globalState, err = hyperstate.New(latestRoot, db, excutor.conf, height)
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
func (excutor *Executor) GetStateInstanceForSimulate(root common.Hash, db db.Database) (vm.Database, error) {

	switch excutor.GetStateType() {
	case "rawstate":
		return statedb.New(root, db)
	case "hyperstate":
		height := core.GetHeightOfChain()
		return hyperstate.New(root, db, excutor.conf, height)
	default:
		return nil, errors.New("no state type specified")
	}
}
