package executor

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/core/bloom"
	edb "hyperchain/core/ledger/db_utils"
	"hyperchain/core/ledger/state"
	"hyperchain/core/types"
	"hyperchain/hyperdb/db"
	"hyperchain/manager/event"
	"time"
)

// CommitBlock - the entry function of the commit process.
// Receive the commit event as a parameter and cached them in the channel
// if the pressure is too high.
func (executor *Executor) CommitBlock(ev event.CommitEvent) {
	executor.addCommitEvent(ev)
}

// listenCommitEvent - commit backend process, use to listen new commit event and dispatch it to the processor.
func (executor *Executor) listenCommitEvent() {
	executor.logger.Notice("commit backend start")
	for {
		select {
		case <-executor.getExit(IDENTIFIER_COMMIT):
			executor.logger.Notice("commit backend exit")
			return
		case v := <-executor.getSuspend(IDENTIFIER_COMMIT):
			if v {
				executor.logger.Notice("pause commit process")
				executor.pauseCommit()
			}
		case ev := <-executor.fetchCommitEvent():
			if success := executor.processCommitEvent(ev, executor.processCommitDone); success == false {
				executor.logger.Errorf("commit block #%d failed, system crush down.", ev.SeqNo)
			}
		}
	}
}

// processCommitEvent - handler of the commit process.
// consume commit event from the channel, exec the commit logic
// and notify backend via callback function.
func (executor *Executor) processCommitEvent(ev event.CommitEvent, done func()) bool {
	executor.markCommitBusy()
	defer executor.markCommitIdle()
	defer done()
	// Legitimacy validation
	if !executor.commitValidationCheck(ev) {
		executor.logger.Errorf("commit event %d not satisfy the demand", ev.SeqNo)
		return false
	}
	block := executor.constructBlock(ev)
	if block == nil {
		executor.logger.Errorf("construct new block for %d commit event failed.", ev.SeqNo)
		return false
	}
	record := executor.getValidateRecord(ValidationTag{ev.Hash, ev.SeqNo})
	if record == nil {
		executor.logger.Errorf("no validation record for #%d found", ev.SeqNo)
		return false
	}
	// write block data to database in a atomic operation
	if err := executor.writeBlock(block, record); err != nil {
		executor.logger.Errorf("write block for #%d failed. err %s", ev.SeqNo, err.Error())
		return false
	}
	// throw all invalid transactions back.
	if ev.IsPrimary {
		// TODO save invalid transaction by itself
		executor.throwInvalidTransactionBack(record.InvalidTxs)
	}
	executor.incDemandNumber()
	executor.cache.validationResultCache.Remove(ValidationTag{ev.Hash, ev.SeqNo})
	return true
}

// writeBlock - flush a block into database.
func (executor *Executor) writeBlock(block *types.Block, record *ValidationResultRecord) error {
	var filterLogs []*types.Log
	// fetch the relative db batch obj from the batch buffer
	// attention: state changes has already been push into the batch obj
	batch := executor.statedb.FetchBatch(record.SeqNo, state.BATCH_NORMAL)
	// persist transaction data
	if err := executor.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		executor.logger.Errorf("persist transactions of #%d failed.", block.Number)
		return err
	}
	// persist receipt data
	if err, ret := executor.persistReceipts(batch, record.ValidTxs, record.Receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		executor.logger.Errorf("persist receipts of #%d failed.", block.Number)
		return err
	} else {
		filterLogs = ret
	}
	// persist block data
	if err, _ := edb.PersistBlock(batch, block, false, false); err != nil {
		executor.logger.Errorf("persist block #%d into database failed.", block.Number, err.Error())
		return err
	}
	// persist chain data
	if err := edb.UpdateChain(executor.namespace, batch, block, false, false, false); err != nil {
		executor.logger.Errorf("update chain to #%d failed.", block.Number, err.Error())
		return err
	}
	// write bloom filter first
	if _, err := bloom.WriteTxBloomFilter(executor.namespace, block.Transactions); err != nil {
		executor.logger.Warning("write tx to bloom filter failed", err.Error())
	}

	// flush the whole batch obj.
	// the database atomic operation of the guarantee is by leveldb batch
	// look the doc https://godoc.org/github.com/syndtr/goleveldb/leveldb#Batch for detail
	if err := batch.Write(); err != nil {
		executor.logger.Errorf("commit #%d changes failed.", block.Number, err.Error())
		return err
	}
	// reset state, notify to remove some cached stuff
	executor.statedb.MarkProcessFinish(record.SeqNo)
	executor.statedb.MakeArchive(record.SeqNo)
	// notify consensus module if it is a checkpoint
	if block.Number%10 == 0 && block.Number != 0 {
		edb.WriteChainChan(executor.namespace)
	}
	executor.logger.Noticef("Block number %d", block.Number)
	executor.logger.Noticef("Block hash %s", hex.EncodeToString(block.BlockHash))
	// told consenus to remove Cached Transactions which used to check transaction duplication
	executor.TransitVerifiedBlock(block)

	// push feed data to event system.
	// external subscribers can access these internal messages through a messaging subscription system
	go executor.filterFeedback(block, filterLogs)
	return nil
}

// getValidateRecord - get validate record with given hash identification.
// nil will be return if no record been found.
func (executor *Executor) getValidateRecord(tag ValidationTag) *ValidationResultRecord {
	ret, existed := executor.fetchValidationResult(tag)
	if !existed {
		executor.logger.Noticef("no validation result found when commit block, hash %s, seqNo %d", tag.hash, tag.seqNo)
		return nil
	}
	return ret
}

// generateBlock - generate a block with given data.
func (executor *Executor) constructBlock(ev event.CommitEvent) *types.Block {
	record := executor.getValidateRecord(ValidationTag{ev.Hash, ev.SeqNo})
	if record == nil {
		return nil
	}
	// 1.generate a new block with the argument in cache
	bloom, err := types.CreateBloom(record.Receipts)
	if err != nil {
		return nil
	}
	newBlock := &types.Block{
		Transactions: nil,
		ParentHash:   edb.GetLatestBlockHash(executor.namespace),
		MerkleRoot:   record.MerkleRoot,
		TxRoot:       record.TxRoot,
		ReceiptRoot:  record.ReceiptRoot,
		Timestamp:    ev.Timestamp,
		CommitTime:   ev.Timestamp,
		Number:       ev.SeqNo,
		WriteTime:    time.Now().UnixNano(),
		EvmTime:      time.Now().UnixNano(),
		Bloom:        bloom,
	}
	if len(record.ValidTxs) != 0 {
		newBlock.Transactions = make([]*types.Transaction, len(record.ValidTxs))
		copy(newBlock.Transactions, record.ValidTxs)
	}
	newBlock.BlockHash = newBlock.Hash().Bytes()
	return newBlock
}

// commitValidationCheck - check whether this commit event is the demand one.
func (executor *Executor) commitValidationCheck(ev event.CommitEvent) bool {
	// 1. verify that the block height is consistent
	if !executor.isDemandNumber(ev.SeqNo) {
		executor.logger.Errorf("receive a commit event %d which is not demand, drop it.", ev.SeqNo)
		return false
	}
	// 2. verify whether validation result exist
	record := executor.getValidateRecord(ValidationTag{ev.Hash, ev.SeqNo})
	if record == nil {
		return false
	}
	// 3. verify whether ev's seqNo equal to record seqNo which acts as block number
	if record.SeqNo != ev.SeqNo {
		executor.logger.Errorf("miss match validation seqNo<#%d>and commit seqNo<#%d>  commit for block failed",
			record.SeqNo, ev.SeqNo)
		return false
	}
	return true
}

func (executor *Executor) persistTransactions(batch db.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
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
func (executor *Executor) persistReceipts(batch db.Batch, transaction []*types.Transaction, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) (error, []*types.Log) {
	var filterLogs []*types.Log
	if len(transaction) != len(receipts) {
		// short circuit if the number of transactions and receipt are not equal
		return errors.New("the number of transactions not equal to receipt"), nil
	}
	for idx, receipt := range receipts {
		logs, err := receipt.RetrieveLogs()
		if err != nil {
			return err, nil
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		filterLogs = append(filterLogs, logs...)

		if transaction[idx].Version != nil {
			if _, err := edb.PersistReceipt(batch, receipt, false, false, string(transaction[idx].Version)); err != nil {
				return err, nil
			}
		} else {
			if _, err := edb.PersistReceipt(batch, receipt, false, false); err != nil {
				return err, nil
			}
		}
	}
	return nil, filterLogs
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

func (executor *Executor) pauseCommit() {
	for {
		if v := <-executor.getSuspend(IDENTIFIER_COMMIT); !v {
			executor.logger.Notice("un-pause commit process")
			return
		}
	}
}

// filterFeedback - push latest block data, contract event data to subscription system.
func (executor *Executor) filterFeedback(block *types.Block, filterLogs []*types.Log) {
	if err := executor.sendFilterEvent(FILTER_NEW_BLOCK, block); err != nil {
		executor.logger.Warningf("send new block event failed. error detail: %s", err.Error())
	}
	go executor.snapshotReg.notifyNewBlock(block.Number)
	if len(filterLogs) > 0 {
		if err := executor.sendFilterEvent(FILTER_NEW_LOG, filterLogs); err != nil {
			executor.logger.Warningf("send new log event failed. error detail: %s", err.Error())
		}
	}
}
