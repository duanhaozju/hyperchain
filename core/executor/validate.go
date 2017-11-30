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
	"sort"
	"sync"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/errors"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"
)

// ValidationTag unique identification for validation result,
// Since we will write empty block, we cannot only use hash as the key
// (we may find same hash for several different validation records).
type ValidationTag struct {
	hash  string
	seqNo uint64
}

// ValidationResultRecord represents a validation result collection
type ValidationResultRecord struct {
	TxRoot      []byte                            // hash of a batch of transactions
	ReceiptRoot []byte                            // hash of a batch of receipts
	MerkleRoot  []byte                            // hash of current world state
	InvalidTxs  []*types.InvalidTransactionRecord // invalid transaction list
	ValidTxs    []*types.Transaction              // valid transaction list
	Receipts    []*types.Receipt                  // receipt list
	SeqNo       uint64                            // sequence number
	SubscriptionData
}

// SubscriptionData is specially used to store the filter data of block and log
type SubscriptionData struct {
	Block *types.Block
	Logs  []*types.Log
}

// Validate - the entry function of the validation process.
// Receive the validation event as a parameter and cached them in the channel
// if the pressure is too high.
func (executor *Executor) Validate(validationEvent event.TransactionBlock) {
	executor.addValidationEvent(validationEvent)
}

// listenValidationEvent starts the validation backend process,
// use to listen new validation event and dispatch it to the processor.
func (executor *Executor) listenValidationEvent() {
	executor.logger.Notice("validation backend start")
	for {
		select {
		case <-executor.context.exit:
			// Tell to Exit
			executor.logger.Notice("validation backend exit")
			return
		case v := <-executor.getSuspend(IDENTIFIER_VALIDATION):
			// Tell to pause
			if v {
				executor.logger.Notice("pause validation process")
				executor.pauseValidation()
			}
		case ev := <-executor.fetchValidationEvent():
			// Handle the received validation event
			if executor.isReadyToValidation() {
				// Normally process the event
				if success := executor.processValidationEvent(ev, executor.processValidationDone); success == false {
					executor.logger.Errorf("validate #%d failed, system crush down.", ev.SeqNo)
				}
			} else {
				// Drop the event if needed
				executor.dropValdiateEvent(ev, executor.processValidationDone)
			}
		}
	}
}

// processValidationEvent processes validation event,
// return true if process successfully, otherwise false will been returned.
func (executor *Executor) processValidationEvent(validationEvent event.TransactionBlock, done func()) bool {
	// Mark the busy of validation, and mark idle after finishing processing
	// This process cannot be stopped
	executor.markValidationBusy()
	defer executor.markValidationIdle()

	// If the event is not the current demand one, pending it
	if !executor.context.isDemand(DemandSeqNo, validationEvent.SeqNo) {
		executor.addPendingValidationEvent(validationEvent)
		return true
	}
	if _, success := executor.process(validationEvent, done); success == false {
		return false
	}
	executor.context.incDemand(DemandSeqNo)
	// Process all the pending events
	return executor.processPendingValidationEvent(done)
}

// processPendingValidationEvent handles all the validation events that is cached in the queue.
func (executor *Executor) processPendingValidationEvent(done func()) bool {
	if executor.cache.pendingValidationEventQ.Len() > 0 {
		// Handle all the remain events sequentially
		for executor.cache.pendingValidationEventQ.Contains(executor.context.getDemand(DemandSeqNo)) {
			ev, _ := executor.fetchPendingValidationEvent(executor.context.getDemand(DemandSeqNo))
			if _, success := executor.process(ev, done); success == false {
				return false
			} else {
				executor.context.incDemand(DemandSeqNo)
				executor.cache.pendingValidationEventQ.RemoveWithCond(ev.SeqNo, RemoveLessThan)
			}
		}
	}
	return true
}

// dropValdiateEvent does nothing but consume a validation event.
func (executor *Executor) dropValdiateEvent(validationEvent event.TransactionBlock, done func()) {
	executor.markValidationBusy()
	defer executor.markValidationIdle()
	defer done()
	executor.logger.Noticef("[Namespace = %s] drop validation event %d", executor.namespace, validationEvent.SeqNo)
}

// process specific implementation logic, including the signature checking,
// the execution of transactions, ledger hash re-computation and etc.
func (executor *Executor) process(validationEvent event.TransactionBlock, done func()) (error, bool) {
	// Invoke the callback no matter success or fail
	defer done()

	var (
		validtxs   []*types.Transaction
		invalidtxs []*types.InvalidTransactionRecord
	)

	invalidtxs, validtxs = executor.checkSign(validationEvent.Transactions)
	err, validateResult := executor.applyTransactions(validtxs, invalidtxs, validationEvent.SeqNo, validationEvent.Timestamp)
	if err != nil {
		// short circuit if error occur during the transaction execution
		executor.logger.Errorf("[Namespace = %s] process transaction batch #%d failed.", executor.namespace, validationEvent.SeqNo)
		return err, false
	}
	// calculate validation result hash for comparison
	hash := executor.calculateValidationResultHash(validateResult.MerkleRoot, validateResult.TxRoot, validateResult.ReceiptRoot)
	executor.logger.Debugf("[Namespace = %s] invalid transaction number %d", executor.namespace, len(validateResult.InvalidTxs))
	executor.logger.Debugf("[Namespace = %s] valid transaction number %d", executor.namespace, len(validateResult.ValidTxs))
	executor.saveValidationResult(validateResult, validationEvent.SeqNo, hash)
	executor.sendValidationResult(validateResult, validationEvent, hash)
	return nil, true
}

// checkSign checks the sender's signature validity.
func (executor *Executor) checkSign(txs []*types.Transaction) ([]*types.InvalidTransactionRecord, []*types.Transaction) {
	var (
		invalidtxs []*types.InvalidTransactionRecord
		wg         sync.WaitGroup
		index      []int
		mu         sync.Mutex
	)

	// Parallel check the signature of each transaction
	for i := range txs {
		wg.Add(1)
		go func(i int) {
			tx := txs[i]
			if !tx.ValidateSign(executor.encryption, executor.commonHash) {
				executor.logger.Warningf("[Namespace = %s] found invalid signature, send from : %v", executor.namespace, tx.Id)
				mu.Lock()
				invalidtxs = append(invalidtxs, &types.InvalidTransactionRecord{
					Tx:      tx,
					ErrType: types.InvalidTransactionRecord_SIGFAILED,
					ErrMsg:  []byte("Invalid signature"),
				})
				index = append(index, i)
				mu.Unlock()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Remove invalid transaction from transaction list
	// Keep the txs sequentially as origin
	if len(index) > 0 {
		sort.Ints(index)
		count := 0
		for _, idx := range index {
			idx = idx - count
			txs = append(txs[:idx], txs[idx+1:]...)
			count++
		}
	}

	return invalidtxs, txs
}

// applyTransactions executes transaction sequentially.
func (executor *Executor) applyTransactions(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64, timestamp int64) (error, *ValidationResultRecord) {
	var (
		validtxs []*types.Transaction
		receipts []*types.Receipt
	)

	executor.initCalculator()
	executor.statedb.MarkProcessStart(seqNo)

	// Execute transaction one by one
	for i, tx := range txs {
		receipt, _, _, err := executor.ExecTransaction(executor.statedb, tx, i, seqNo, timestamp)
		if err != nil {
			errType := executor.classifyInvalid(err)
			invalidTxs = append(invalidTxs, &types.InvalidTransactionRecord{
				Tx:      tx,
				ErrType: errType,
				ErrMsg:  []byte(err.Error()),
			})
			continue
		}
		executor.calculateTransactionsFingerprint(tx, false)
		executor.calculateReceiptFingerprint(tx, receipt, false)
		receipts = append(receipts, receipt)
		validtxs = append(validtxs, tx)
	}

	merkleRoot, txRoot, receiptRoot, err := executor.submitValidationResult()
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] submit validation result failed.", executor.namespace, err.Error())
		return err, nil
	}
	executor.resetStateDb()
	executor.logger.Debugf("[Namespace = %s] validate result seqNo #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		executor.namespace, seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	return nil, &ValidationResultRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
		Receipts:    receipts,
		ValidTxs:    validtxs,
		InvalidTxs:  invalidTxs,
		SeqNo:       seqNo,
	}
}

// classifyInvalid classifies invalid transaction via error type.
func (executor *Executor) classifyInvalid(err error) types.InvalidTransactionRecord_ErrType {
	var errType types.InvalidTransactionRecord_ErrType
	if errors.IsValueTransferErr(err) {
		errType = types.InvalidTransactionRecord_OUTOFBALANCE
	} else if errors.IsExecContractErr(err) {
		tmp := err.(*errors.ExecContractError)
		if tmp.GetType() == 0 {
			errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
		} else if tmp.GetType() == 1 {
			errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
		}
	} else if errors.IsInvalidInvokePermissionErr(err) {
		errType = types.InvalidTransactionRecord_INVALID_PERMISSION
	}
	return errType
}

// submitValidationResult submits state changes to batch,
// and return the merkleRoot, txRoot, receiptRoot.
func (executor *Executor) submitValidationResult() ([]byte, []byte, []byte, error) {
	// flush all state change
	root, err := executor.statedb.Commit()
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] commit state db failed! error msg, ", executor.namespace, err.Error())
		return nil, nil, nil, err
	}
	merkleRoot := root.Bytes()
	res, _ := executor.calculateTransactionsFingerprint(nil, true)
	txRoot := res.Bytes()
	res, _ = executor.calculateReceiptFingerprint(nil, nil, true)
	receiptRoot := res.Bytes()
	return merkleRoot, txRoot, receiptRoot, nil
}

// resetStateDb resets the statsdb
func (executor *Executor) resetStateDb() {
	executor.statedb.Reset()
}

// throwInvalidTransactionBack sends all invalid transaction to its birth place.
func (executor *Executor) throwInvalidTransactionBack(invalidtxs []*types.InvalidTransactionRecord) {
	for _, t := range invalidtxs {
		executor.informP2P(NOTIFY_UNICAST_INVALID, t)
	}
}

// saveValidationResult saves validation result to cache.
func (executor *Executor) saveValidationResult(res *ValidationResultRecord, seqNo uint64, hash common.Hash) {
	executor.addValidationResult(ValidationTag{hash.Hex(), seqNo}, res)
}

// sendValidationResult sends validation result to consensus module.
func (executor *Executor) sendValidationResult(res *ValidationResultRecord, ev event.TransactionBlock, hash common.Hash) {
	executor.informConsensus(NOTIFY_VALIDATION_RES, protos.ValidatedTxs{
		Transactions: res.ValidTxs,
		SeqNo:        ev.SeqNo,
		View:         ev.View,
		Hash:         hash.Hex(),
		Timestamp:    ev.Timestamp,
		Digest:       ev.Digest,
	})
}

// pauseValidation stops validation process for a while.
// continue by a signal invocation.
func (executor *Executor) pauseValidation() {
	for {
		if v := <-executor.getSuspend(IDENTIFIER_VALIDATION); !v {
			executor.logger.Notice("un-pause validation process")
			return
		}
	}
}
