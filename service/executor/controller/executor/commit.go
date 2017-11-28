// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	"encoding/hex"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/bloom"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/manager/event"

	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"strconv"
)

// CommitBlock is the entry function of the commit process.
// Receive the commit event as a parameter and cached them in the channel
// if the pressure is too high.
func (e *Executor) CommitBlock(ev *event.CommitEvent) {
	e.addCommitEvent(ev)
}

// listenCommitEvent is commit backend process, use to listen new commit event and dispatch it to the processor.
func (e *Executor) listenCommitEvent() {
	e.logger.Notice("commit backend start")
	for {
		select {
		case <-e.context.exit:
			e.logger.Notice("commit backend exit")
			return
		case v := <-e.getSuspend(IDENTIFIER_COMMIT):
			if v {
				e.logger.Notice("pause commit process")
				e.pauseCommit()
			}
		case ev := <-e.fetchCommitEvent():
			if success := e.processCommitEvent(ev); success == false {
				e.logger.Errorf("commit block #%d failed, system crush down.", ev.SeqNo)
			}
		}
	}
}

// processCommitEvent is the handler of the commit process,
// consumes commit event from the channel, executes the commit logic
// and notifies backend via callback function.
func (e *Executor) processCommitEvent(ev *event.CommitEvent) bool {
	e.markCommitBusy()
	defer e.markCommitIdle()

	// Legitimacy validation
	if !e.commitValidationCheck(ev) {
		e.logger.Errorf("commit event %d not satisfy the demand", ev.SeqNo)
		return false
	}
	record := e.getValidateRecord(ValidationTag{ev.Hash, ev.SeqNo})
	if record == nil {
		e.logger.Errorf("no validation record for #%d found", ev.SeqNo)
		return false
	}
	// Construct the Block with given CommitEvent
	block := e.constructBlock(ev, record)
	if block == nil {
		e.logger.Errorf("construct new block for %d commit event failed.", ev.SeqNo)
		return false
	}
	// Write block data to database in an atomic operation
	if err := e.writeBlock(block, record); err != nil {
		e.logger.Errorf("write block for #%d failed. err %s", ev.SeqNo, err.Error())
		return false
	}
	// Throw all invalid transactions back.
	if ev.IsPrimary {
		// TODO save invalid transaction by itself
		// TODO(Xiaoyi Wang): handle invalid transaction
		e.throwInvalidTransactionBack(record.InvalidTxs)
	}

	e.cache.validationResultCache.Remove(ValidationTag{ev.Hash, ev.SeqNo})
	e.context.incDemand(DemandNumber)
	e.processCommitDone()

	return true
}

// writeBlock flushes a block into database.
func (e *Executor) writeBlock(block *types.Block, record *ValidationResultRecord) error {
	var filterLogs []*types.Log
	// Fetch the relative db batch obj from the batch buffer
	// Attention: state changes has already been push into the batch obj
	batch := e.statedb.FetchBatch(record.SeqNo, state.BATCH_NORMAL)
	// Persist transaction data
	if err := e.persistTransactions(batch, block.Transactions, block.Number); err != nil {
		e.logger.Errorf("persist transactions of #%d failed.", block.Number)
		return err
	}
	// Persist receipt data
	if ret, err := e.persistReceipts(batch, record.ValidTxs, record.Receipts, block.Number, common.BytesToHash(block.BlockHash)); err != nil {
		e.logger.Errorf("persist receipts of #%d failed.", block.Number)
		return err
	} else {
		filterLogs = ret
	}
	// Persist block data
	if _, err := chain.PersistBlock(batch, block, false, false); err != nil {
		e.logger.Errorf("persist block #%d into database failed.", block.Number, err.Error())
		return err
	}
	// persist chain data
	if err := chain.UpdateChain(e.namespace, batch, block, false, false, false); err != nil {
		e.logger.Errorf("update chain to #%d failed.", block.Number, err.Error())
		return err
	}
	// Write bloom filter first
	if _, err := bloom.WriteTxBloomFilter(e.namespace, block.Transactions); err != nil {
		e.logger.Warning("write tx to bloom filter failed", err.Error())
	}

	if d, ok := e.cache.opLogIndexCache.Get(block.Number); ok {
		if lid, ok := d.(uint64); ok {
			buf := make([]byte, 0)
			b := strconv.AppendUint(buf, lid, 10)
			key := []byte(fmt.Sprintf("op.lo.id.%d", block.Number))
			batch.Put(key, b)
		}
	}

	// Flush the whole batch obj
	// The database atomic operation of the guarantee is by leveldb batch,
	// look the doc https://godoc.org/github.com/syndtr/goleveldb/leveldb#Batch for detail.
	if err := batch.Write(); err != nil {
		e.logger.Errorf("commit #%d changes failed.", block.Number, err.Error())
		return err
	}
	// Reset state, notify to remove some cached stuff
	e.statedb.MarkProcessFinish(record.SeqNo)
	//TODO(Xiaoyi Wang): e.statedb.MakeArchive(record.SeqNo)
	// Notify consensus module if it is a checkpoint
	//TODO(Xiaoyi Wang): Add checkpoint
	//if block.Number%10 == 0 && block.Number != 0 {
	//	chain.WriteChainChan(e.namespace)
	//}
	e.logger.Noticef("Block number %d", block.Number)
	e.logger.Noticef("Block hash %s", hex.EncodeToString(block.BlockHash))
	//e.TransitVerifiedBlock(block)

	// Push feed data to event system
	// External subscribers can access these internal messages through a messaging subscription system.
	go e.filterFeedback(block, filterLogs)
	return nil
}

// getValidateRecord gets validated record with given hash identification,
// return nil if no record been found.
func (executor *Executor) getValidateRecord(tag ValidationTag) *ValidationResultRecord {
	ret, existed := executor.fetchValidationResult(tag)
	if !existed {
		executor.logger.Noticef("no validation result found when commit block, hash %s, seqNo %d", tag.hash, tag.seqNo)
		return nil
	}
	return ret
}

// constructBlock constructs a block with given data.
func (e *Executor) constructBlock(ev *event.CommitEvent, record *ValidationResultRecord) *types.Block {
	// Generate a new block with the argument in cache
	bloom, err := types.CreateBloom(record.Receipts)
	if err != nil {
		return nil
	}
	newBlock := &types.Block{
		Transactions: nil,
		ParentHash:   chain.GetLatestBlockHash(e.namespace),
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
	e.logger.Debugf("ParentHash: %v, Number: %v, Timestamp: %v, TxRoot: %v, ReceiptRoot: %v, MerkleRoot: %v",
		common.Bytes2Hex(newBlock.ParentHash), newBlock.Number, newBlock.Timestamp,
		common.Bytes2Hex(newBlock.TxRoot), common.Bytes2Hex(newBlock.ReceiptRoot), common.Bytes2Hex(newBlock.MerkleRoot))
	newBlock.BlockHash = newBlock.Hash().Bytes()
	return newBlock
}

// commitValidationCheck checks whether this commit event is the demand one.
func (e *Executor) commitValidationCheck(ev *event.CommitEvent) bool {
	// 1. Verify that the block height is consistent
	if !e.context.isDemand(DemandNumber, ev.SeqNo) {
		e.logger.Errorf("receive a commit event %d which is not demand, drop it.", ev.SeqNo)
		return false
	}
	// 2. Verify whether validation result exists
	record := e.getValidateRecord(ValidationTag{ev.Hash, ev.SeqNo})
	if record == nil {
		return false
	}
	// 3. Verify whether ev's seqNo equal to record's seqNo which acts as block number
	if record.SeqNo != ev.SeqNo {
		e.logger.Errorf("miss match validation seqNo<#%d>and commit seqNo<#%d>  commit for block failed",
			record.SeqNo, ev.SeqNo)
		return false
	}
	return true
}

func (e *Executor) persistTransactions(batch db.Batch, transactions []*types.Transaction, blockNumber uint64) error {
	for i, transaction := range transactions {
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: blockNumber,
			Index:      int64(i),
		}
		if err := chain.PersistTransactionMeta(batch, meta, transaction.GetHash(), false, false); err != nil {
			return err
		}
	}
	return nil
}

// persistReceipts assigns block hash and block number to transaction executor.loggers
// and persists the receipts into database.
func (executor *Executor) persistReceipts(batch db.Batch, transaction []*types.Transaction, receipts []*types.Receipt, blockNumber uint64, blockHash common.Hash) ([]*types.Log, error) {
	var filterLogs []*types.Log
	if len(transaction) != len(receipts) {
		// Short circuit if the number of transactions and receipt are not equal
		return nil, errors.New("the number of transactions not equal to receipt")
	}
	for idx, receipt := range receipts {
		logs, err := receipt.RetrieveLogs()
		if err != nil {
			return nil, err
		}
		for _, log := range logs {
			log.BlockHash = blockHash
			log.BlockNumber = blockNumber
		}
		receipt.SetLogs(logs)
		filterLogs = append(filterLogs, logs...)

		if transaction[idx].Version != nil {
			if _, err := chain.PersistReceipt(batch, receipt, false, false, string(transaction[idx].Version)); err != nil {
				return nil, err
			}
		} else {
			if _, err := chain.PersistReceipt(batch, receipt, false, false); err != nil {
				return nil, err
			}
		}
	}
	return filterLogs, nil
}

// StoreInvalidTransaction stores the invalid transaction into database for client query
func (executor *Executor) StoreInvalidTransaction(payload []byte) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(payload, invalidTx)
	if err != nil {
		executor.logger.Error("unmarshal invalid transaction record payload failed")
	}
	executor.logger.Noticef("invalid transaction %s", invalidTx.Tx.Hash().Hex())
	// Save to database
	err, _ = chain.PersistInvalidTransactionRecord(executor.db.NewBatch(), invalidTx, true, true)
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

// filterFeedback pushes latest block data, contract event data to subscription system.
func (executor *Executor) filterFeedback(block *types.Block, filterLogs []*types.Log) {
	if err := executor.sendFilterEvent(FILTER_NEW_BLOCK, block); err != nil {
		executor.logger.Warningf("send new block event failed. error detail: %s", err.Error())
	}
	//go executor.snapshotReg.notifyNewBlock(block.Number)
	if len(filterLogs) > 0 {
		if err := executor.sendFilterEvent(FILTER_NEW_LOG, filterLogs); err != nil {
			executor.logger.Warningf("send new log event failed. error detail: %s", err.Error())
		}
	}
}
