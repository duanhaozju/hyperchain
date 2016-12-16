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
)

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
	}
	// instead of send an CommitOrRollbackBlockEvent with `false` flag, PBFT send a `viewchange` or `self recovery`
	// message to handle this issue
	pool.blockCache.Remove(ev.Hash)
}

// Put a new generated block into pool, handle the block saved in queue serially
func (pool *BlockPool) AddBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool) {
	if block.Number == 0 {
		pool.WriteBlock(block, commonHash, 0, false)
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
		pool.WriteBlock(block, commonHash, vid, primary)
		atomic.AddUint64(&pool.demandNumber, 1)
		log.Info("current demandNumber is ", pool.demandNumber)

		for i := block.Number + 1; i <= atomic.LoadUint64(&pool.maxNum); i += 1 {
			if ret, existed := pool.queue.Get(i); existed {
				blk := ret.(*types.Block)
				pool.WriteBlock(blk, commonHash, vid, primary)
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
func(pool *BlockPool) WriteBlock(block *types.Block, commonHash crypto.CommonHash, vid uint64, primary bool) {
	log.Info("block number is ", block.Number)
	core.UpdateChain(block, false)

	db, err := hyperdb.GetLDBDatabase()
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

	err, _ = core.PersistBlock(batch, block, pool.conf.BlockVersion, false, false)
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
		pool.consenter.RemoveCachedBatch(vid)
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
