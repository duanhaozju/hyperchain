package blockpool

import (
	"time"
	"github.com/golang/protobuf/proto"
	"hyperchain/recovery"
	"sync/atomic"
	"hyperchain/hyperdb"
	"encoding/hex"
	"hyperchain/event"
	"hyperchain/crypto"
	"hyperchain/p2p"
	"hyperchain/core/types"
	"hyperchain/core"
	"hyperchain/common"
	"hyperchain/core/hyperstate"
)

// When receive an CommitOrRollbackBlockEvent, if flag is true, generate a block and call AddBlock function
// CommitBlock function is just an entry of the commit logic
// TODO refactor
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
		newBlock.Timestamp = ev.Timestamp          // primary make batch time
		newBlock.CommitTime = ev.CommitTime        // local commit time
		newBlock.Number = ev.SeqNo                 // actually block number
		newBlock.WriteTime = time.Now().UnixNano() // local write time
		newBlock.EvmTime = time.Now().UnixNano()   // local write time
		newBlock.BlockHash = newBlock.Hash(commonHash).Bytes()

		vid := record.SeqNo
		// 2.save block and update chain
		pool.AddBlock(newBlock, record.Receipts, commonHash, vid, ev.IsPrimary)
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
func (pool *BlockPool) AddBlock(block *types.Block, receipts []*types.Receipt, commonHash crypto.CommonHash, vid uint64, primary bool) {
	if block.Number == 0 {
		pool.WriteBlock(block, receipts, commonHash, 0, false)
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
		pool.WriteBlock(block, receipts, commonHash, vid, primary)
		atomic.AddUint64(&pool.demandNumber, 1)
		log.Info("current demandNumber is ", pool.demandNumber)

		for i := block.Number + 1; i <= atomic.LoadUint64(&pool.maxNum); i += 1 {
			if ret, existed := pool.queue.Get(i); existed {
				blk := ret.(*types.Block)
				pool.WriteBlock(blk, receipts, commonHash, vid, primary)
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
func(pool *BlockPool) WriteBlock(block *types.Block, receipts []*types.Receipt, commonHash crypto.CommonHash, vid uint64, primary bool) {
	time.Sleep(1 * time.Second)
	log.Info("block number is ", block.Number)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database instance failed! error msg,", err.Error())
		return
	}
	// for primary node, check whether vid equal to block's number
	state, _ := pool.GetStateInstance(common.Hash{}, db)
	batch := state.FetchBatch(vid)
	if primary && vid != block.Number {
		log.Info("replace invalid txmeta data, block number:", block.Number)
		for i, tx := range block.Transactions {
			meta := &types.TransactionMeta{
				BlockIndex: block.Number,
				Index:      int64(i),
			}
			// replace with correct value
			if err := core.PersistTransactionMeta(batch, meta, tx.GetTransactionHash(), false, false); err != nil {
				log.Error("invalid txmeta sturct, marshal failed! error msg,", err.Error())
				return
			}
		}
	}
	// reassign receipt
	pool.reAssignTransactionLog(batch, receipts, block.Number, common.BytesToHash(block.BlockHash))
	err, _ = core.PersistBlock(batch, block, pool.conf.BlockVersion, false, false)
	if err != nil {
		log.Error("Put block into database failed! error msg, ", err.Error())
		return
	}
	core.UpdateChain(batch, block, false, false, false)
	// flush to disk
	// IMPORTANT never remove this statement, otherwise the whole batch of data will lose
	batch.Write()
	// mark the block process finish, remove some stuff avoid of memory leak
	// IMPORTANT this should be done after batch.Write been called
	state.MarkProcessFinish(vid)
	// write checkpoint data
	// FOR TEST
	log.Criticalf("state #%d %s", vid, string(state.Dump()))
	if block.Number %10 == 0 && block.Number != 0 {
		core.WriteChainChan()
	}
	newChain := core.GetChainCopy()
	log.Notice("Block number", newChain.Height)
	log.Notice("Block hash", hex.EncodeToString(newChain.LatestBlockHash))
	// remove Cached Transactions which used to check transaction duplication
	if primary {
		pool.consenter.RemoveCachedBatch(vid)
	}
	// FOR TEST
	// get journals
	// TODO journal prefix number is not correct
	j, err := db.Get(hyperstate.CompositeJournalKey(vid))
	if err != nil {
		return
	}
	journals, err := hyperstate.UnmarshalJournal(j)
	if err != nil {
		return
	}
	for _, entry := range journals.JournalList {
		log.Errorf("#%d journal %s", vid, entry)
	}
}

// save the invalid transaction into database for client query
func (pool *BlockPool) StoreInvalidResp(ev event.RespInvalidTxsEvent) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(ev.Payload, invalidTx)
	if err != nil {
		log.Error("unmarshal invalid transaction record payload failed")
	}
	// save to db
	log.Notice("invalid transaction", common.BytesToHash(invalidTx.Tx.TransactionHash).Hex())
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database instance failed! error msg,", err.Error())
		return
	}
	err, _ = core.PersistInvalidTransactionRecord(db.NewBatch(), invalidTx, true, true)
	if err != nil {
		log.Error("save invalid transaction record failed,", err.Error())
		return
	}
}


// re assign block hash and block number to transaction logs
// during the validation, block number and block hash can be incorrect
func (pool *BlockPool) reAssignTransactionLog(batch hyperdb.Batch, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) {
	for _, receipt := range receipts {
		logs, err := receipt.GetLogs()
		if err != nil {
			log.Error("re assign transaction log, unmarshal receipt failed")
			return
		}
		if len(logs) == 0 {
			continue
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		// replace original receipt
		core.PersistReceipt(batch, receipt, pool.conf.TransactionVersion, false, false)
	}
}
