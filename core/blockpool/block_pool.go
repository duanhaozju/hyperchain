// implementblock pool
// author: Lizhong kuang
// date: 2016-08-29
// last modified:2016-09-01
package blockpool

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/recovery"
	"hyperchain/trie"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"sort"
)

var (
	public_batch hyperdb.Batch
	log          *logging.Logger // package-level logger
)

func init() {
	log = logging.MustGetLogger("block-pool")
}

type BlockRecord struct {
	TxRoot      []byte
	ReceiptRoot []byte
	MerkleRoot  []byte
	InvalidTxs  []*types.InvalidTransactionRecord
	ValidTxs    []*types.Transaction
	SeqNo       uint64
}

type BlockPool struct {
	demandNumber        uint64
	demandSeqNo         uint64
	maxNum              uint64
	maxSeqNo            uint64
	consenter           consensus.Consenter
	eventMux            *event.TypeMux
	stateLock           sync.Mutex
	wg                  sync.WaitGroup // for shutdown sync
	lastValidationState common.Hash
	blockCache          *common.Cache
	validationQueue     *common.Cache
	queue               *common.Cache
}

func NewBlockPool(eventMux *event.TypeMux, consenter consensus.Consenter) *BlockPool {
	blockcache, _ := common.NewCache()
	queue, _ := common.NewCache()
	validationqueue, _ := common.NewCache()
	pool := &BlockPool{
		eventMux:        eventMux,
		consenter:       consenter,
		queue:           queue,
		validationQueue: validationqueue,
		blockCache:      blockcache,
	}

	currentChain := core.GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
	pool.demandSeqNo = currentChain.Height + 1
	db, _ := hyperdb.GetLDBDatabase()
	blk, _ := core.GetBlock(db, currentChain.LatestBlockHash)
	pool.lastValidationState = common.BytesToHash(blk.MerkleRoot)

	log.Noticef("Block pool Initialize demandNumber :%d, demandseqNo: %d\n", pool.demandNumber, pool.demandSeqNo)
	return pool
}

func (pool *BlockPool) SetDemandNumber(number uint64) {
	atomic.StoreUint64(&pool.demandNumber, number)
}
func (pool *BlockPool) SetDemandSeqNo(seqNo uint64) {
	atomic.StoreUint64(&pool.demandSeqNo, seqNo)
}

// Validate is an entry of `validate process`
// When a validationEvent received, put it into the validationQueue
// If the demand ValidationEvent arrived, call `PreProcess` function
// IMPORTANT this function called in parallelly, Make sure all the variable are thread-safe
func (pool *BlockPool) Validate(validationEvent event.ExeTxsEvent, commonHash crypto.CommonHash, encryption crypto.Encryption, peerManager p2p.PeerManager) {
	if validationEvent.SeqNo > pool.maxSeqNo {
		atomic.StoreUint64(&pool.maxSeqNo, validationEvent.SeqNo)
	}

	if _, existed := pool.validationQueue.Get(validationEvent.SeqNo); existed {
		log.Error("Receive Repeat ValidationEvent, ", validationEvent.SeqNo)
		return
	}

	// (1) Check SeqNo
	if validationEvent.SeqNo < pool.demandSeqNo {
		// Receive repeat ValidationEvent
		log.Error("Receive Repeat ValidationEvent, seqno less than demandseqNo, ", validationEvent.SeqNo)
		return
	} else if validationEvent.SeqNo == pool.demandSeqNo {
		// Process
		if _, success := pool.PreProcess(validationEvent, commonHash, encryption, peerManager); success {
			atomic.AddUint64(&pool.demandSeqNo, 1)
			log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
		}
		judge := func(key interface{}, iterKey interface{}) bool {
			id := key.(uint64)
			iterId := iterKey.(uint64)
			if id >= iterId {
				return true
			}
			return false
		}
		pool.validationQueue.RemoveWithCond(validationEvent.SeqNo, judge)

		// Process remain event
		for i := validationEvent.SeqNo + 1; i <= atomic.LoadUint64(&pool.maxSeqNo); i += 1 {
			if ret, existed := pool.validationQueue.Get(i); existed {
				ev := ret.(event.ExeTxsEvent)
				if _, success := pool.PreProcess(ev, commonHash, encryption, peerManager); success {
					pool.validationQueue.Remove(i)
					atomic.AddUint64(&pool.demandSeqNo, 1)
					log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
				}
			} else {
				break
			}
		}
		return
	} else {
		log.Notice("Receive ValidationEvent which is not demand, ", validationEvent.SeqNo, "save into cache temporarily")
		pool.validationQueue.Add(validationEvent.SeqNo, validationEvent)
	}
}

// Process an ValidationEvent
func (pool *BlockPool) PreProcess(validationEvent event.ExeTxsEvent, commonHash crypto.CommonHash, encryption crypto.Encryption, peerManager p2p.PeerManager) (error, bool) {
	var validTxSet []*types.Transaction
	var invalidTxSet []*types.InvalidTransactionRecord
	var index []int
	if validationEvent.IsPrimary {
		invalidTxSet, index = pool.CheckSign(validationEvent.Transactions, commonHash, encryption)
	} else {
		validTxSet = validationEvent.Transactions
	}

	if len(index) > 0 {
		sort.Ints(index)
		count := 0
		set := validationEvent.Transactions
		for i := range index {
			i = i-count
			set = append(set[:i-1], set[i+1:]...)
			count++
		}
		validTxSet = set
	} else {
		validTxSet = validationEvent.Transactions
	}
	err, _, merkleRoot, txRoot, receiptRoot, validTxSet, invalidTxSet := pool.ProcessBlockInVm(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		return err, false
	}
	hash := commonHash.Hash([]interface{}{
		merkleRoot,
		txRoot,
		receiptRoot,
	})

	if len(validTxSet) != 0 {
		pool.blockCache.Add(hash.Hex(), BlockRecord{
			TxRoot:      txRoot,
			ReceiptRoot: receiptRoot,
			MerkleRoot:  merkleRoot,
			InvalidTxs:  invalidTxSet,
			ValidTxs:    validTxSet,
			SeqNo:       validationEvent.SeqNo,
		})
	}
	log.Info("Invalid Tx number: ", len(invalidTxSet))
	log.Info("Valid Tx number: ", len(validTxSet))
	// Communicate with PBFT
	pool.consenter.RecvLocal(event.ValidatedTxs{
		Transactions: validTxSet,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Hex(),
		Timestamp:    validationEvent.Timestamp,
	})

	// empty block generated, throw all invalid transactions back to original node directly
	if validationEvent.IsPrimary && len(validTxSet) == 0 {
		for _, t := range invalidTxSet {
			payload, err := proto.Marshal(t)
			if err != nil {
				log.Error("Marshal tx error")
			}
			if t.Tx.Id == uint64(peerManager.GetNodeId()) {
				pool.StoreInvalidResp(event.RespInvalidTxsEvent{
					Payload: payload,
				})
				continue
			}
			var peers []uint64
			peers = append(peers, t.Tx.Id)
			peerManager.SendMsgToPeers(payload, peers, recovery.Message_INVALIDRESP)
		}
	}
	return nil, true
}

// check the sender's signature of the transaction
func (pool *BlockPool) CheckSign(txs []*types.Transaction, commonHash crypto.CommonHash, encryption crypto.Encryption) ([]*types.InvalidTransactionRecord, []int) {
	var invalidTxSet []*types.InvalidTransactionRecord
	// (1) check signature for each transaction
	var wg sync.WaitGroup
	var index []int
	for i, tx := range txs {
		wg.Add(1)
		go func(tx *types.Transaction){
			if !tx.ValidateSign(encryption, commonHash) {
				log.Notice("Validation, found invalid signature, send from :", tx.Id)
				invalidTxSet = append(invalidTxSet, &types.InvalidTransactionRecord{
					Tx:      tx,
					ErrType: types.InvalidTransactionRecord_SIGFAILED,
				})
				index = append(index, i)
			}
			wg.Done()
		}(tx)
	}
	wg.Wait()
	return invalidTxSet, index
	//return nil, nil
}

// Put all transactions into the virtual machine and execute
// Return the execution result, such as txs' merkle root, receipts' merkle root, accounts' merkle root and so on
func (pool *BlockPool) ProcessBlockInVm(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, []byte, []byte, []byte, []byte, []*types.Transaction, []*types.InvalidTransactionRecord) {
	var validtxs []*types.Transaction
	var (
		env = make(map[string]string)
	)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	txTrie, _ := trie.New(common.Hash{}, db)
	receiptTrie, _ := trie.New(common.Hash{}, db)
	statedb, err := state.New(pool.lastValidationState, db)
	if err != nil {
		log.Error("New StateDB ERROR")
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	public_batch = db.NewBatch()
	for i, tx := range txs {
		statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(*tx, vmenv)
		if err != nil && core.IsValueTransferErr(err) {
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
			})
			continue
		}
		// save to DB
		txValue, _ := proto.Marshal(tx)
		if err := public_batch.Put(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...), txValue); err != nil {
			return err, nil, nil, nil, nil, nil, invalidTxs
		}

		receiptValue, _ := proto.Marshal(receipt)
		if err := public_batch.Put(append(core.ReceiptsPrefix, receipt.TxHash...), receiptValue); err != nil {
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		// set temporarily
		// for primary node, the seqNo can be invalid. remove the incorrect txmeta info when commit block to avoid this error
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		metaValue, _ := proto.Marshal(meta)
		if err := public_batch.Put(append(tx.GetTransactionHash().Bytes(), core.TxMetaSuffix...), metaValue); err != nil {
			log.Error("Put txmeta into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		// Update trie
		txTrie.Update(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...), txValue)
		receiptTrie.Update(append(core.ReceiptsPrefix, receipt.TxHash...), receiptValue)
		validtxs = append(validtxs, tx)
	}
	root, _ := statedb.Commit()
	merkleRoot := root.Bytes()
	txRoot := txTrie.Hash().Bytes()
	receiptRoot := receiptTrie.Hash().Bytes()
	pool.lastValidationState = root
	go public_batch.Write()
	return nil, nil, merkleRoot, txRoot, receiptRoot, validtxs, invalidTxs
}

// When receive an CommitOrRollbackBlockEvent, if flag is true, generate a block and call AddBlock function
// CommitBlock function is just an entry of the commit logic
func (pool *BlockPool) CommitBlock(ev event.CommitOrRollbackBlockEvent, commonHash crypto.CommonHash, peerManager p2p.PeerManager) {
	ret, existed := pool.blockCache.Get(ev.Hash)
	if !existed {
		log.Notice("No record found when commit block, block hash:", ev.Hash)
		return
	}
	record := ret.(BlockRecord)
	if ev.Flag {
		// 1.generate a new block with the argument in cache
		newBlock := new(types.Block)
		newBlock.Transactions = make([]*types.Transaction, len(record.ValidTxs))
		copy(newBlock.Transactions, record.ValidTxs)
		currentChain := core.GetChainCopy()
		newBlock.ParentHash = currentChain.LatestBlockHash
		newBlock.MerkleRoot = record.MerkleRoot
		newBlock.TxRoot = record.TxRoot
		newBlock.ReceiptRoot = record.ReceiptRoot
		newBlock.Timestamp = ev.Timestamp
		newBlock.CommitTime = ev.CommitTime
		newBlock.Number = ev.SeqNo
		newBlock.WriteTime = time.Now().UnixNano()
		newBlock.EvmTime = time.Now().UnixNano()
		newBlock.BlockHash = newBlock.Hash(commonHash).Bytes()

		vid := record.SeqNo
		// 2.save block and update chain
		pool.AddBlock(newBlock, commonHash, vid, ev.IsPrimary)
		// 3.throw invalid tx back to origin node if current peer is primary
		if ev.IsPrimary {
			for _, t := range record.InvalidTxs {
				payload, err := proto.Marshal(t)
				if err != nil {
					log.Error("Marshal tx error")
				}
				if t.Tx.Id == uint64(peerManager.GetNodeId()) {
					pool.StoreInvalidResp(event.RespInvalidTxsEvent{
						Payload: payload,
					})
					continue
				}
				var peers []uint64
				peers = append(peers, t.Tx.Id)
				peerManager.SendMsgToPeers(payload, peers, recovery.Message_INVALIDRESP)
			}
		}
	} else {
		// this branch will never activated
		// instead of send an CommitOrRollbackBlockEvent with `false` flag, PBFT send a `viewchange` or `self recovery`
		// message to handle this issue
		// TODO remove this branch
		db, _ := hyperdb.GetLDBDatabase()
		for _, t := range record.InvalidTxs {
			db.Delete(append(core.TransactionPrefix, t.Tx.GetTransactionHash().Bytes()...))
			db.Delete(append(core.ReceiptsPrefix, t.Tx.GetTransactionHash().Bytes()...))
		}
	}
	pool.blockCache.Remove(ev.Hash)
}

// Put a new generated block into pool, handle the block saved in queue serially
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool) {
	if block.Number == 0 {
		WriteBlock(block, commonHash, 0, false)
		return
	}

	if block.Number > pool.maxNum {
		atomic.StoreUint64(&pool.maxNum, block.Number)
	}

	if _, existed := pool.queue.Get(block.Number); existed {
		log.Info("repeat block number,number is: ", block.Number)
		return
	}

	log.Info("number is ", block.Number)

	if pool.demandNumber == block.Number {
		WriteBlock(block, commonHash, vid, primary)
		atomic.AddUint64(&pool.demandNumber, 1)
		log.Info("current demandNumber is ", pool.demandNumber)

		for i := block.Number + 1; i <= atomic.LoadUint64(&pool.maxNum); i += 1 {
			if ret, existed := pool.queue.Get(i); existed {
				blk := ret.(*types.Block)
				WriteBlock(blk, commonHash, vid, primary)
				pool.queue.Remove(i)
				atomic.AddUint64(&pool.demandNumber, 1)
				log.Info("current demandNumber is ", pool.demandNumber)
			} else {
				break
			}
		}
		return
	} else {
		pool.queue.Add(block.Number, block)
	}
}

// WriteBlock: save block into database
func WriteBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool) {
	log.Info("block number is ", block.Number)
	core.UpdateChain(block, false)

	db, _ := hyperdb.GetLDBDatabase()
	// for primary node, check whether vid equal to block's number
	if primary && vid != block.Number {
		log.Info("Replace invalid txmeta data, block number:", block.Number)
		batch := db.NewBatch()
		for i, tx := range block.Transactions {
			meta := &types.TransactionMeta{
				BlockIndex: block.Number,
				Index:      int64(i),
			}
			metaValue, _ := proto.Marshal(meta)
			if err := batch.Put(append(tx.GetTransactionHash().Bytes(), core.TxMetaSuffix...), metaValue); err != nil {
				log.Error("Put txmeta into database failed! error msg, ", err.Error())
				return
			}
		}
		batch.Write()
	}

	err := core.PutBlockTx(db, commonHash, block.BlockHash, block)
	if err != nil {
		log.Error("Put block into database failed! error msg, ", err.Error())
	}

	if block.Number%10 == 0 && block.Number != 0 {
		core.WriteChainChan()
	}

	newChain := core.GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))
}

// save the invalid transaction into database for client query
func (pool *BlockPool) StoreInvalidResp(ev event.RespInvalidTxsEvent) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(ev.Payload, invalidTx)
	if err != nil {
		log.Error("Unmarshal Payload failed")
	}
	// save to db
	log.Notice("invalidTx", common.BytesToHash(invalidTx.Tx.TransactionHash).Hex())
	db, _ := hyperdb.GetLDBDatabase()
	db.Put(append(core.InvalidTransactionPrefix, invalidTx.Tx.TransactionHash...), ev.Payload)
}

// reset blockchain to a stable checkpoint status when `viewchange` occur
func (pool *BlockPool) ResetStatus(ev event.VCResetEvent) {
	tmpDemandNumber := atomic.LoadUint64(&pool.demandNumber)
	// 1. Reset demandNumber , demandSeqNo and lastValidationState
	atomic.StoreUint64(&pool.demandNumber, ev.SeqNo)
	atomic.StoreUint64(&pool.demandSeqNo, ev.SeqNo)
	atomic.StoreUint64(&pool.maxSeqNo, ev.SeqNo-1)

	db, _ := hyperdb.GetLDBDatabase()
	block, _ := core.GetBlockByNumber(db, ev.SeqNo-1)
	pool.lastValidationState = common.BytesToHash(block.MerkleRoot)
	// 2. Delete Invalid Stuff
	for i := ev.SeqNo; i < tmpDemandNumber; i += 1 {
		// delete tx, txmeta and receipt
		block, err := core.GetBlockByNumber(db, i)
		if err != nil {
			log.Errorf("ViewChange, miss block %d ,error msg %s", i, err.Error())
		}

		for _, tx := range block.Transactions {
			if err := db.Delete(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...)); err != nil {
				log.Errorf("ViewChange, delete useless tx in block %d failed, error msg %s", i, err.Error())
			}
			if err := db.Delete(append(core.ReceiptsPrefix, tx.GetTransactionHash().Bytes()...)); err != nil {
				log.Errorf("ViewChange, delete useless receipt in block %d failed, error msg %s", i, err.Error())
			}
			if err := db.Delete(append(tx.GetTransactionHash().Bytes(), core.TxMetaSuffix...)); err != nil {
				log.Errorf("ViewChange, delete useless txmeta in block %d failed, error msg %s", i, err.Error())
			}
		}
		// delete block
		if err := core.DeleteBlockByNum(db, i); err != nil {
			log.Errorf("ViewChange, delete useless block %d failed, error msg %s", i, err.Error())
		}

	}
	// 3. Delete from blockcache
	keys := pool.blockCache.Keys()
	for _, key := range keys {
		ret, _ := pool.blockCache.Get(key)
		record := ret.(BlockRecord)
		for i, tx := range record.ValidTxs {
			if err := db.Delete(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...)); err != nil {
				log.Errorf("ViewChange, delete useless tx in block %d failed, error msg %s", i, err.Error())
			}
			if err := db.Delete(append(core.ReceiptsPrefix, tx.GetTransactionHash().Bytes()...)); err != nil {
				log.Errorf("ViewChange, delete useless receipt in block %d failed, error msg %s", i, err.Error())
			}
			if err := db.Delete(append(tx.GetTransactionHash().Bytes(), core.TxMetaSuffix...)); err != nil {
				log.Errorf("ViewChange, delete useless txmeta in block %d failed, error msg %s", i, err.Error())
			}
		}
	}
	pool.blockCache.Purge()
	// 4. Reset chain
	isGenesis := (block.Number == 0)
	core.UpdateChain(block, isGenesis)

}
