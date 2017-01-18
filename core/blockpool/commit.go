package blockpool

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"sync/atomic"
	"time"
)

// When receive an CommitOrRollbackBlockEvent, if flag is true, generate a block and call AddBlock function
// CommitBlock function is just an entry of the commit logic
// TODO refactor
func (pool *BlockPool) CommitBlock(ev event.CommitOrRollbackBlockEvent, commonHash crypto.CommonHash, peerManager p2p.PeerManager) {
	ret, existed := pool.blockCache.Get(ev.Hash)
	if !existed {
		log.Notice("No record found when commit block, record hash:", ev.Hash)
		return
	}
	record := ret.(BlockRecord)
	// 1.generate a new block with the argument in cache
	newBlock := &types.Block{
		ParentHash:  core.GetChainCopy().LatestBlockHash,
		MerkleRoot:  record.MerkleRoot,
		TxRoot:      record.TxRoot,
		ReceiptRoot: record.ReceiptRoot,
		Timestamp:   ev.Timestamp,
		CommitTime:  ev.Timestamp,
		Number:      ev.SeqNo,
		WriteTime:   time.Now().UnixNano(),
		EvmTime:     time.Now().UnixNano(),
	}
	newBlock.Transactions = make([]*types.Transaction, len(record.ValidTxs))
	copy(newBlock.Transactions, record.ValidTxs)
	newBlock.BlockHash = newBlock.Hash(commonHash).Bytes()

	vid := record.VID
	tempBlockNumber := record.SeqNo
	// check block number validation
	if tempBlockNumber != ev.SeqNo {
		log.Errorf("miss match temp block number<#%d>and actually block number<#%d> for vid #%d validation. commit for block #%d failed",
			tempBlockNumber, ev.SeqNo, vid, ev.SeqNo)
		return
	}
	log.Debugf("commit for block #%d, parent hash %s, merkle root %s, tx root %s, receipt root %s, vid #%d", newBlock.Number,
		common.Bytes2Hex(newBlock.ParentHash), common.Bytes2Hex(newBlock.MerkleRoot), common.Bytes2Hex(newBlock.TxRoot),
		common.Bytes2Hex(newBlock.ReceiptRoot), vid)
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
func (pool *BlockPool) WriteBlock(block *types.Block, receipts []*types.Receipt, commonHash crypto.CommonHash, vid uint64, primary bool) {
	// for primary node, check whether vid equal to block's number
	state, err := pool.GetStateInstance()
	if err != nil {
		log.Errorf("get state instance failed when write #%d", block.Number)
		return
	}
	batch := state.FetchBatch(block.Number)
	if err := pool.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		log.Errorf("persist transactions of #%d failed.", block.Number)
		return
	}
	if err := pool.persistReceipts(batch, receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		log.Errorf("persist receipts of #%d failed.", block.Number)
		return
	}
	if err, _ := core.PersistBlock(batch, block, pool.GetBlockVersion(), false, false); err != nil {
		log.Errorf("persist block #%d into database failed! error msg, ", block.Number, err.Error())
		return
	}
	core.UpdateChain(batch, block, false, false, false)
	batch.Write()
	// mark the block process finish, remove some stuff avoid of memory leak
	// IMPORTANT this should be done after batch.Write been called
	state.MarkProcessFinish(block.Number)
	 //log.Critical("state #%d %s", vid, string(state.Dump()))
	// write checkpoint data
	if block.Number%10 == 0 && block.Number != 0 {
		core.WriteChainChan()
	}
	log.Notice("Block number", block.Number)
	log.Notice("Block hash", hex.EncodeToString(block.BlockHash))
	// remove Cached Transactions which used to check transaction duplication
	msg := protos.RemoveCache{Vid: vid}
	pool.consenter.RecvLocal(msg)
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
	db, err := hyperdb.GetDBDatabase()
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

func (pool *BlockPool) persistTransactions(batch hyperdb.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
		if err, _ := core.PersistTransaction(batch, transaction, pool.GetTransactionVersion(), false, false); err != nil {
			log.Error("put tx data into database failed! error msg, ", err.Error())
			return err
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: blockNumber,
			Index:      int64(i),
		}
		if err := core.PersistTransactionMeta(batch, meta, transaction.GetTransactionHash(), false, false); err != nil {
			log.Error("Put txmeta into database failed! error msg, ", err.Error())
			return err
		}
	}
	return nil
}

// re assign block hash and block number to transaction logs
// during the validation, block number and block hash can be incorrect
func (pool *BlockPool) persistReceipts(batch hyperdb.Batch, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) error {
	for _, receipt := range receipts {
		logs, err := receipt.GetLogs()
		if err != nil {
			log.Error("re assign transaction log, unmarshal receipt failed")
			return err
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		if err, _ := core.PersistReceipt(batch, receipt, pool.GetTransactionVersion(), false, false); err != nil {
			log.Error("Put receipt into database failed! error msg, ", err.Error())
			return err
		}
	}
	return nil
}
