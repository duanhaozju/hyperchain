package executor

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"time"
	"sync/atomic"
	"hyperchain/hyperdb/db"
)

func (executor *Executor) CommitBlock(ev event.CommitOrRollbackBlockEvent, peerManager p2p.PeerManager) {
	executor.commitQueue <- ev
	if executor.peerManager == nil {
		executor.peerManager = peerManager
	}
	atomic.AddInt32(&executor.commitQueueLen, 1)
}

func (executor *Executor) commitBackendLoop() {
	for ev := range executor.commitQueue {
		success := executor.consumeCommitEvent(ev)
		if !success {
			log.Errorf("commit block #%d failed, system crush down.", ev.SeqNo)
		}
		atomic.AddInt32(&executor.commitQueueLen, -1)
	}
}

// consumeCommitEvent - consume commit event from channel.
func (executor *Executor) consumeCommitEvent(ev event.CommitOrRollbackBlockEvent) bool {
	atomic.StoreInt32(&executor.commitInProgress, PROGRESS_TRUE)
	defer atomic.StoreInt32(&executor.commitInProgress, PROGRESS_FALSE)
	if executor.commitValidationCheck(ev) == false {
		log.Errorf("commit event %d not satisfied demand", ev.SeqNo)
		return false
	}
	block := executor.generateBlock(ev)
	if block == nil {
		log.Errorf("generate new block for %d commit event failed.")
		return false
	}
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		log.Errorf("no validation record for #%d found", ev.SeqNo)
		return false
	}
	if err := executor.writeBlock(block, record); err != nil {
		log.Errorf("write block for #%d failed. err %s", ev.SeqNo, err.Error())
		return false
	}
	// throw all invalid transactions back.
	executor.notifyInvalidTransactions(record.InvalidTxs, ev.IsPrimary, executor.peerManager)
	executor.increaseDemandBlockNumber()
	executor.blockCache.Remove(ev.Hash)
	return true
}

// writeBlock - flush a block into disk.
func (executor *Executor) writeBlock(block *types.Block, record *BlockRecord) error {
	state, err := executor.GetStateInstance()
	if err != nil {
		log.Errorf("get state instance failed when write #%d", block.Number)
		return err
	}
	batch := state.FetchBatch(block.Number)
	if err := executor.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		log.Errorf("persist transactions of #%d failed.", block.Number)
		return err
	}
	if err := executor.persistReceipts(batch, record.Receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		log.Errorf("persist receipts of #%d failed.", block.Number)
		return err
	}
	if err, _ := core.PersistBlock(batch, block, executor.GetBlockVersion(), false, false); err != nil {
		log.Errorf("persist block #%d into database failed! error msg, ", block.Number, err.Error())
		return err
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
	msg := protos.RemoveCache{Vid: record.VID}
	executor.consenter.RecvLocal(msg)
	return nil
}

// getValidateRecord - get validate record with given hash identification.
// nil will be return if no record been found.
func (executor *Executor) getValidateRecord(hash string) *BlockRecord {
	ret, existed := executor.blockCache.Get(hash)
	if !existed {
		log.Notice("No record found when commit block, record hash:", hash)
		return nil
	}
	record := ret.(BlockRecord)
	return &record
}

// generateBlock - generate a block with given data.
func (executor *Executor) generateBlock(ev event.CommitOrRollbackBlockEvent) *types.Block {
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		return nil
	}
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
	newBlock.BlockHash = newBlock.Hash(executor.commonHash).Bytes()
	return newBlock
}

// commitValidationCheck - check whether this commit event satisfy demand.
func (executor *Executor) commitValidationCheck(ev event.CommitOrRollbackBlockEvent) bool {
	// 1. check whether this ev is the demand one
	if ev.SeqNo != executor.demandNumber {
		log.Errorf("receive a commit event %d which is not demand, drop it.", ev.SeqNo)
		return false
	}
	// 2. check whether validation result exist
	ret, existed := executor.blockCache.Get(ev.Hash)
	if !existed {
		log.Notice("No record found when commit block, record hash:", ev.Hash)
		return false
	}
	record := ret.(BlockRecord)
	// 3. check whether ev's seqNo equal to record seqNo which act as block number
	vid := record.VID
	tempBlockNumber := record.SeqNo
	if tempBlockNumber != ev.SeqNo {
		log.Errorf("miss match temp block number<#%d>and actually block number<#%d> for vid #%d validation. commit for block #%d failed",
			tempBlockNumber, ev.SeqNo, vid, ev.SeqNo)
		return false
	}
	return true
}

// notifyInvalidTransactions - notify sender peer for invalid transactions.
func (executor *Executor) notifyInvalidTransactions(invalidTransactions []*types.InvalidTransactionRecord, primary bool, peerManager p2p.PeerManager) {
	if primary {
		for _, t := range invalidTransactions {
			payload, err := proto.Marshal(t)
			if err != nil {
				log.Error("Marshal tx error")
			}
			if t.Tx.Id == uint64(peerManager.GetNodeId()) {
				executor.StoreInvalidResp(event.RespInvalidTxsEvent{
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

// increaseDemandBlockNumber - increase current demand block number for commit.
func (executor *Executor) increaseDemandBlockNumber() {
	executor.demandNumber += 1
	log.Noticef("demand block number %d", executor.demandNumber)
}

func (executor *Executor) persistTransactions(batch db.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
		if err, _ := core.PersistTransaction(batch, transaction, executor.GetTransactionVersion(), false, false); err != nil {
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
func (executor *Executor) persistReceipts(batch db.Batch, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) error {
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
		if err, _ := core.PersistReceipt(batch, receipt, executor.GetTransactionVersion(), false, false); err != nil {
			log.Error("Put receipt into database failed! error msg, ", err.Error())
			return err
		}
	}
	return nil
}


// save the invalid transaction into database for client query
func (executor *Executor) StoreInvalidResp(ev event.RespInvalidTxsEvent) {
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
