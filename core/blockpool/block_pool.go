//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
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
	"hyperchain/protos"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"sort"
	"errors"
	"os"
	"fmt"
)

var (
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

type BlockPoolConf struct {
	BlockVersion       string
	TransactionVersion string
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
	lastValidationState atomic.Value
	blockCache          *common.Cache
	validationQueue     *common.Cache
	queue               *common.Cache
	conf                BlockPoolConf
}

func NewBlockPool(eventMux *event.TypeMux, consenter consensus.Consenter, conf BlockPoolConf) *BlockPool {
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
	}
	currentChain := core.GetChainCopy()
	pool.demandNumber = currentChain.Height + 1
	pool.demandSeqNo = currentChain.Height + 1
	db, err := hyperdb.GetDBDatabase()
	if err != nil {return nil}

	blk, err := core.GetBlock(db, currentChain.LatestBlockHash)
	if err != nil {return nil}
	pool.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))

	log.Noticef("Block pool Initialize demandNumber :%d, demandseqNo: %d\n", pool.demandNumber, pool.demandSeqNo)
	return pool
}

func (pool *BlockPool) GetConfig() BlockPoolConf {
	return pool.conf
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
	if validationEvent.SeqNo > atomic.LoadUint64(&pool.maxSeqNo) {
		atomic.StoreUint64(&pool.maxSeqNo, validationEvent.SeqNo)
	}

	if _, existed := pool.validationQueue.Get(validationEvent.SeqNo); existed {
		log.Error("Receive Repeat ValidationEvent, ", validationEvent.SeqNo)
		return
	}

	// (1) Check SeqNo
	if validationEvent.SeqNo < atomic.LoadUint64(&pool.demandSeqNo)  {
		// Receive repeat ValidationEvent
		log.Error("Receive Repeat ValidationEvent, seqno less than demandseqNo, ", validationEvent.SeqNo)
		return
	} else if validationEvent.SeqNo == atomic.LoadUint64(&pool.demandSeqNo)  {
		// Process
		if _, success := pool.PreProcess(validationEvent, commonHash, encryption, peerManager); success {
			atomic.AddUint64(&pool.demandSeqNo, 1)
			log.Notice("Current demandSeqNo is, ", pool.demandSeqNo)
		} else {
			log.Error("PreProcess Failed")
			return
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
				} else {
					log.Error("PreProcess Failed")
					return
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
		for _, idx := range index {
			idx = idx-count
			if idx == 0 {
				set = set[1:]
				count++
			} else {
				set = append(set[:idx - 1], set[idx + 1:]...)
				count++
			}
		}
		validTxSet = set
	} else {
		validTxSet = validationEvent.Transactions
	}
	err, _, merkleRoot, txRoot, receiptRoot, validTxSet, invalidTxSet := pool.ProcessBlockInVm(validTxSet, invalidTxSet, validationEvent.SeqNo)
	if err != nil {
		log.Error("ProcessBlock Failed!, block number: ", validationEvent.SeqNo)
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
	pool.consenter.RecvLocal(protos.ValidatedTxs{
		Transactions: validTxSet,
		SeqNo:        validationEvent.SeqNo,
		View:         validationEvent.View,
		Hash:         hash.Hex(),
		Timestamp:    validationEvent.Timestamp,
	})

	// empty block generated, throw all invalid transactions back to original node directly
	if validationEvent.IsPrimary && len(validTxSet) == 0 {
		// 1. Remove all cached transaction in this block, because empty block won't enter network
		pool.consenter.RemoveCachedBatch(validationEvent.SeqNo)
		// 2. Throw all invalid transaction back to the origin node
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
	var mu sync.Mutex
	for i := range txs {
		wg.Add(1)
		go func(i int){
			tx := txs[i]
			if !tx.ValidateSign(encryption, commonHash) {
				log.Notice("Validation, found invalid signature, send from :", tx.Id)
				invalidTxSet = append(invalidTxSet, &types.InvalidTransactionRecord{
					Tx:      tx,
					ErrType: types.InvalidTransactionRecord_SIGFAILED,
					ErrMsg:  []byte("Invalid signature"),
				})
				mu.Lock()
				index = append(index, i)
				mu.Unlock()
			}
			wg.Done()
		}(i)
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
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	txTrie, err := trie.New(common.Hash{}, db)
	if err != nil {return err, nil, nil, nil, nil, nil, invalidTxs}
	receiptTrie, err := trie.New(common.Hash{}, db)
	if err != nil {return err, nil, nil, nil, nil, nil, invalidTxs}
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("Get StateDB Status Failed!"), nil, nil, nil, nil, nil, invalidTxs
	}
	statedb, err := state.New(initStatus, db)

	if err != nil {
		f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY|os.O_CREATE, 0644)
		if err1 != nil {
			fmt.Println("1.txt file create failed. err: " + err.Error())
		} else {
			// 查找文件末尾的偏移量
			n, _ := f.Seek(0, os.SEEK_END)
			// 从末尾的偏移量开始写入内容
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str:=newFormat+"block pool 302 the err of statebd :"+err.Error()+"\n"
			_, err = f.WriteAt([]byte(str), n)

			f.Close()
		}
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	env["currentNumber"] = strconv.FormatUint(seqNo, 10)
	env["currentGasLimit"] = "10000000"
	vmenv := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)

	public_batch := db.NewBatch()
	for i, tx := range txs {
		var err error
		var data []byte
		statedb.StartRecord(tx.GetTransactionHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(*tx, vmenv)
		if err != nil{
			var errType types.InvalidTransactionRecord_ErrType
			if core.IsValueTransferErr(err) {
				errType = types.InvalidTransactionRecord_OUTOFBALANCE
			} else if core.IsExecContractErr(err) {
				tmp := err.(*core.ExecContractError)
				if tmp.GetType() == 0 {
					errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
				} else if tmp.GetType() == 1{
					errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
				} else {
					// For extension
				}
			} else {
				// For extension
			}
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: errType,
				ErrMsg:  []byte(err.Error()),
			})
			continue
		}
		// Persist transaction
		tx.Version = []byte(pool.conf.TransactionVersion)
		err, data = core.PersistTransaction(public_batch, tx, pool.conf.TransactionVersion, false, false);
		if err != nil {
			log.Error("Put tx data into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		txTrie.Update(append(core.TransactionPrefix, tx.GetTransactionHash().Bytes()...), data)

		// Persist receipt
		receipt.Version = []byte(pool.conf.TransactionVersion)
		err, data = core.PersistReceipt(public_batch, receipt, pool.conf.TransactionVersion, false, false)
		if  err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}
		receiptTrie.Update(append(core.ReceiptsPrefix, receipt.TxHash...), data)

		// Persist transaction meta
		// set temporarily
		// for primary node, the seqNo can be invalid. remove the incorrect txmeta info when commit block to avoid this error
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(public_batch, meta, tx.GetTransactionHash(), false, false); err != nil {
			log.Error("Put txmeta into database failed! error msg, ", err.Error())
			return err, nil, nil, nil, nil, nil, invalidTxs
		}

		validtxs = append(validtxs, tx)
	}
	root, err := statedb.Commit()
	if err != nil {
		log.Error("Commit state db failed! error msg, ", err.Error())
		return err, nil, nil, nil, nil, nil, invalidTxs
	}
	merkleRoot := root.Bytes()
	txRoot := txTrie.Hash().Bytes()
	receiptRoot := receiptTrie.Hash().Bytes()
	pool.lastValidationState.Store(root)
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
		newBlock.Version = []byte(pool.conf.BlockVersion)

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
	}
	// instead of send an CommitOrRollbackBlockEvent with `false` flag, PBFT send a `viewchange` or `self recovery`
	// message to handle this issue
	pool.blockCache.Remove(ev.Hash)
}

// Put a new generated block into pool, handle the block saved in queue serially
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool) {
	if block.Number == 0 {
		WriteBlock(block, commonHash, 0, false, pool.consenter)
		return
	}

	if block.Number > pool.maxNum {
		atomic.StoreUint64(&pool.maxNum, block.Number)
	}

	if _, existed := pool.queue.Get(block.Number); existed {
		log.Info("repeat block number,number is: ", block.Number)
		return
	}

	if pool.demandNumber == block.Number {
		WriteBlock(block, commonHash, vid, primary, pool.consenter)
		atomic.AddUint64(&pool.demandNumber, 1)

		for i := block.Number + 1; i <= atomic.LoadUint64(&pool.maxNum); i += 1 {
			if ret, existed := pool.queue.Get(i); existed {
				blk := ret.(*types.Block)
				WriteBlock(blk, commonHash, vid, primary, pool.consenter)
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
func WriteBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool, consenter consensus.Consenter) {
	core.UpdateChain(block, false)

	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	// for primary node, check whether vid equal to block's number
	batch := db.NewBatch()
	if primary && vid != block.Number {
		log.Info("Replace invalid txmeta data, block number:", block.Number)
		for i, tx := range block.Transactions {
			meta := &types.TransactionMeta{
				BlockIndex: block.Number,
				Index:      int64(i),
			}
			// replace with correct value
			if err := core.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
				log.Error("Invalid txmeta sturct, marshal failed! error msg,", err.Error())
				return
			}
		}
	}


	err, _ = core.PersistBlock(batch, block, string(block.Version), false, false)
	if err != nil {
		log.Error("Put block into database failed! error msg, ", err.Error())
		return
	}
	// flush to disk
	// IMPORTANT never remove this statement, otherwise the whole batch of data will lose
	batch.Write()
	if block.Number%10 == 0 && block.Number != 0 {
		core.WriteChainChan()
	}
	newChain := core.GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))
	/*
		Remove Cached Transactions which used to check transaction duplication
	 */
	if primary {
		consenter.RemoveCachedBatch(vid)
	}
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
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	db.Put(append(core.InvalidTransactionPrefix, invalidTx.Tx.TransactionHash...), ev.Payload)
}

// reset blockchain to a stable checkpoint status when `viewchange` occur
func (pool *BlockPool) ResetStatus(ev event.VCResetEvent) {
	tmpDemandNumber := atomic.LoadUint64(&pool.demandNumber)
	// 1. Reset demandNumber , demandSeqNo and lastValidationState
	atomic.StoreUint64(&pool.demandNumber, ev.SeqNo)
	atomic.StoreUint64(&pool.demandSeqNo, ev.SeqNo)
	atomic.StoreUint64(&pool.maxSeqNo, ev.SeqNo-1)

	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}

	block, err := core.GetBlockByNumber(db, ev.SeqNo - 1)
	if err != nil {
		return
	}
	pool.lastValidationState.Store(common.BytesToHash(block.MerkleRoot))
	// 2. Delete Invalid Stuff
	pool.RemoveData(ev.SeqNo, tmpDemandNumber)

	// 3. Delete from blockcache
	keys := pool.blockCache.Keys()
	for _, key := range keys {
		ret, _ := pool.blockCache.Get(key)
		if ret == nil {
			continue
		}
		record := ret.(BlockRecord)
		for i, tx := range record.ValidTxs {
			if err := core.DeleteTransaction(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("ViewChange, delete useless tx in cache %d failed, error msg %s", i, err.Error())
			}

			if err := core.DeleteReceipt(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("ViewChange, delete useless receipt in cache %d failed, error msg %s", i, err.Error())
			}
			if err := core.DeleteTransactionMeta(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("ViewChange, delete useless txmeta in cache %d failed, error msg %s", i, err.Error())
			}
		}
	}
	pool.blockCache.Purge()
	pool.validationQueue.Purge()
	// 4. Reset chain
	isGenesis := (block.Number == 0)
	core.UpdateChain(block, isGenesis)
}
func (pool *BlockPool) RunInSandBox(tx *types.Transaction) error {
	// TODO add block number to specify the initial status
	var env = make(map[string]string)
	fakeBlockNumber := core.GetHeightOfChain()
	env["currentNumber"] = strconv.FormatUint(fakeBlockNumber, 10)
	env["currentGasLimit"] = "10000000"
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err
	}
	v := pool.lastValidationState.Load()
	initStatus, ok := v.(common.Hash)
	if ok == false {
		return errors.New("Get StateDB Status Failed!")
	}
	statedb, err := state.New(initStatus, db)
	if err != nil {
		return err
	}
	sandBox := core.NewEnvFromMap(core.RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)
	receipt, _, _, err := core.ExecTransaction(*tx, sandBox)
	if err != nil{
		var errType types.InvalidTransactionRecord_ErrType
		if core.IsValueTransferErr(err) {
			errType = types.InvalidTransactionRecord_OUTOFBALANCE
		} else if core.IsExecContractErr(err) {
			tmp := err.(*core.ExecContractError)
			if tmp.GetType() == 0 {
				errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
			} else if tmp.GetType() == 1{
				errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
			} else {
				// For extension
			}
		} else {
			// For extension
		}
		t :=  &types.InvalidTransactionRecord{
			Tx:      tx,
			ErrType: errType,
			ErrMsg:  []byte(err.Error()),
		}
		payload, err := proto.Marshal(t)
		if err != nil {
			log.Error("Marshal tx error")
			return nil
		}
		pool.StoreInvalidResp(event.RespInvalidTxsEvent{
			Payload: payload,
		})
		return nil
	} else {
		err, _ := core.PersistReceipt(db.NewBatch(), receipt, pool.conf.TransactionVersion, true, true)
		if err != nil {
			log.Error("Put receipt data into database failed! error msg, ", err.Error())
			return err
		}
		return nil
	}
}

func (pool *BlockPool) RemoveData(from, to uint64) {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}

	// delete tx, txmeta and receipt
	for i := from; i < to; i += 1 {
		block, err := core.GetBlockByNumber(db, i)
		if err != nil {
			log.Errorf("miss block %d ,error msg %s", i, err.Error())
			continue
		}

		for _, tx := range block.Transactions {
			if err := core.DeleteTransaction(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless tx in block %d failed, error msg %s", i, err.Error())
			}
			if err := core.DeleteReceipt(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless receipt in block %d failed, error msg %s", i, err.Error())
			}
			if err := core.DeleteTransactionMeta(db, tx.GetTransactionHash().Bytes()); err != nil {
				log.Errorf("delete useless txmeta in block %d failed, error msg %s", i, err.Error())
			}
		}
		// delete block
		if err := core.DeleteBlockByNum(db, i); err != nil {
			log.Errorf("ViewChange, delete useless block %d failed, error msg %s", i, err.Error())
		}
	}
}

func (pool *BlockPool) CutdownBlock(number uint64) {
	// 1. reset demand number and demand seqNo
	atomic.StoreUint64(&pool.demandNumber, number)
	atomic.StoreUint64(&pool.demandSeqNo, number)
	atomic.StoreUint64(&pool.maxSeqNo, number - 1)
	// 2. remove block releted data
	pool.RemoveData(number, number + 1)
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("Get Database Instance Failed! error msg,", err.Error())
		return
	}
	block, err := core.GetBlockByNumber(db, number - 1)
	if err != nil {
		log.Errorf("miss block %d ,error msg %s", number - 1, err.Error())
		return
	}
	// clear all stuff in block cache and validation cache
	pool.lastValidationState.Store(common.BytesToHash(block.MerkleRoot))
	core.UpdateChainByBlcokNum(db, block.Number - 1)
}
func (pool *BlockPool) PurgeValidateQueue() {
	pool.validationQueue.Purge()
}
func (pool *BlockPool) PurgeBlockCache() {
	pool.blockCache.Purge()
}

