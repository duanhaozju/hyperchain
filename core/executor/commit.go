package executor

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb/db"
	"hyperchain/protos"
	"time"
)

func (executor *Executor) CommitBlock(ev event.CommitEvent) {
	executor.addCommitEvent(ev)
}

func (executor *Executor) listenCommitEvent() {
	executor.logger.Notice("commit backend start")
	for {
		select {
		case <-executor.getExit(IDENTIFIER_COMMIT):
			executor.logger.Notice("commit backend exit")
			return
		case ev := <-executor.fetchCommitEvent():
			if success := executor.processCommitEvent(ev, executor.processCommitDone); success == false {
				executor.logger.Errorf("commit block #%d failed, system crush down.", ev.SeqNo)
			}
		}
	}
}

// processCommitEvent - consume commit event from channel.
func (executor *Executor) processCommitEvent(ev event.CommitEvent, done func()) bool {
	executor.markCommitBusy()
	defer executor.markCommitIdle()
	defer done()
	if !executor.commitValidationCheck(ev) {
		executor.logger.Errorf("commit event %d not satisfy the demand", ev.SeqNo)
		return false
	}
	block := executor.constructBlock(ev)
	if block == nil {
		executor.logger.Errorf("construct new block for %d commit event failed.", ev.SeqNo)
		return false
	}
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		executor.logger.Errorf("no validation record for #%d found", ev.SeqNo)
		return false
	}
	if err := executor.writeBlock(block, record); err != nil {
		executor.logger.Errorf("write block for #%d failed. err %s", ev.SeqNo, err.Error())
		return false
	}
	// throw all invalid transactions back.
	if ev.IsPrimary {
		executor.throwInvalidTransactionBack(record.InvalidTxs)
	}
	executor.incDemandNumber()
	executor.cache.validationResultCache.Remove(ev.Hash)
	return true
}

// writeBlock - flush a block into disk.
func (executor *Executor) writeBlock(block *types.Block, record *ValidationResultRecord) error {
	batch := executor.statedb.FetchBatch(record.SeqNo)
	if err := executor.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		executor.logger.Errorf("persist transactions of #%d failed.", block.Number)
		return err
	}
	if err := executor.persistReceipts(batch, record.Receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		executor.logger.Errorf("persist receipts of #%d failed.", block.Number)
		return err
	}
	if err, _ := edb.PersistBlock(batch, block, false, false); err != nil {
		executor.logger.Errorf("persist block #%d into database failed.", block.Number, err.Error())
		return err
	}
	edb.UpdateChain(executor.namespace, batch, block, false, false, false)
	batch.Write()
	executor.statedb.MarkProcessFinish(record.SeqNo)

	if block.Number%10 == 0 && block.Number != 0 {
		edb.WriteChainChan(executor.namespace)
	}
	executor.logger.Noticef("Block number %d", block.Number)
	executor.logger.Noticef("Block hash %s", hex.EncodeToString(block.BlockHash))
	// executor.logger.Notice(string(executor.statedb.Dump()))
	// remove Cached Transactions which used to check transaction duplication
	executor.informConsensus(NOTIFY_REMOVE_CACHE, protos.RemoveCache{Vid: record.VID})
	return nil
}

// getValidateRecord - get validate record with given hash identification.
// nil will be return if no record been found.
func (executor *Executor) getValidateRecord(hash string) *ValidationResultRecord {
	ret, existed := executor.fetchValidationResult(hash)
	if !existed {
		executor.logger.Noticef("no validation result found when commit block, hash %s", hash)
		return nil
	}
	return ret
}

// generateBlock - generate a block with given data.
func (executor *Executor) constructBlock(ev event.CommitEvent) *types.Block {
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		return nil
	}
	// 1.generate a new block with the argument in cache
	newBlock := &types.Block{
		ParentHash:  edb.GetLatestBlockHash(executor.namespace),
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
	newBlock.BlockHash = newBlock.Hash().Bytes()
	return newBlock
}

// commitValidationCheck - check whether this commit event satisfy demand.
func (executor *Executor) commitValidationCheck(ev event.CommitEvent) bool {
	// 1. check whether this ev is the demand one
	if !executor.isDemandNumber(ev.SeqNo) {
		executor.logger.Errorf("receive a commit event %d which is not demand, drop it.", ev.SeqNo)
		return false
	}
	// 2. check whether validation result exist
	record := executor.getValidateRecord(ev.Hash)
	if record == nil {
		return false
	}
	// 3. check whether ev's seqNo equal to record seqNo which act as block number
	vid := record.VID
	tempBlockNumber := record.SeqNo
	if tempBlockNumber != ev.SeqNo {
		executor.logger.Errorf("miss match temp block number<#%d>and actually block number<#%d> for vid #%d validation. commit for block #%d failed",
			tempBlockNumber, ev.SeqNo, vid, ev.SeqNo)
		return false
	}
	return true
}

func (executor *Executor) persistTransactions(batch db.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
		if err, _ := edb.PersistTransaction(batch, transaction, false, false); err != nil {
			return err
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: blockNumber,
			Index:      int64(i),
		}
		if err := edb.PersistTransactionMeta(batch, meta, transaction.GetHash(), false, false); err != nil {
			return err
		}
	}
	return nil
}

// re assign block hash and block number to transaction executor.loggers
// during the validation, block number and block hash can be incorrect
func (executor *Executor) persistReceipts(batch db.Batch, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) error {
	for _, receipt := range receipts {
		logs, err := receipt.RetrieveLogs()
		if err != nil {
			return err
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		if err, _ := edb.PersistReceipt(batch, receipt, false, false); err != nil {
			return err
		}
	}
	return nil
}

// save the invalid transaction into database for client query
func (executor *Executor) StoreInvalidTransaction(payload []byte) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(payload, invalidTx)
	if err != nil {
		executor.logger.Error("unmarshal invalid transaction record payload failed")
	}
	// save to db
	executor.logger.Noticef("invalid transaction %s", invalidTx.Tx.Hash().Hex())
	err, _ = edb.PersistInvalidTransactionRecord(executor.db.NewBatch(), invalidTx, true, true)
	if err != nil {
		executor.logger.Error("save invalid transaction record failed,", err.Error())
		return
	}
}
