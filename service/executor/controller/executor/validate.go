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
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/errors"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	"time"
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
func (e *Executor) Validate(validationEvent *event.ValidationEvent) {
	e.addValidationEvent(validationEvent)
}

// listenValidationEvent starts the validation backend process,
// use to listen new validation event and dispatch it to the processor.
func (e *Executor) listenValidationEvent() {
	e.logger.Notice("validation backend start")
	for {
		select {
		case <-e.context.exit:
			// Tell to Exit
			e.logger.Notice("validation backend exit")
			return
		case v := <-e.getSuspend(IDENTIFIER_VALIDATION):
			// Tell to pause
			if v {
				e.logger.Notice("pause validation process")
				e.pauseValidation()
			}
		case ev := <-e.fetchValidationEvent():
			// Handle the received validation event
			if e.isReadyToValidation() {
				// Normally process the event
				if success := e.processValidationEvent(ev, e.processValidationDone); success == false {
					e.logger.Errorf("validate #%d failed, system crush down.", ev.SeqNo)
				}
			} else {
				// Drop the event if needed
				e.dropValidateEvent(ev, e.processValidationDone)
			}
		}
	}
}

// processValidationEvent processes validation event,
// return true if process successfully, otherwise false will been returned.
func (e *Executor) processValidationEvent(validationEvent *event.ValidationEvent, done func()) bool {
	// Mark the busy of validation, and mark idle after finishing processing
	// This process cannot be stopped
	e.markValidationBusy()
	defer e.markValidationIdle()

	// If the event is not the current demand one, pending it
	if !e.context.isDemand(DemandSeqNo, validationEvent.SeqNo) {
		e.addPendingValidationEvent(validationEvent)
		return true
	}
	if _, success := e.process(validationEvent, done); success == false {
		return false
	}
	e.context.incDemand(DemandSeqNo)
	// Process all the pending events
	return e.processPendingValidationEvent(done)
}

// processPendingValidationEvent handles all the validation events that is cached in the queue.
func (e *Executor) processPendingValidationEvent(done func()) bool {
	if e.cache.pendingValidationEventQ.Len() > 0 {
		// Handle all the remain events sequentially
		for e.cache.pendingValidationEventQ.Contains(e.context.getDemand(DemandSeqNo)) {
			ev, _ := e.fetchPendingValidationEvent(e.context.getDemand(DemandSeqNo))
			if _, success := e.process(ev, done); success == false {
				return false
			} else {
				e.context.incDemand(DemandSeqNo)
				e.cache.pendingValidationEventQ.RemoveWithCond(ev.SeqNo, RemoveLessThan)
			}
		}
	}
	return true
}

// dropValidateEvent does nothing but consume a validation event.
func (e *Executor) dropValidateEvent(validationEvent *event.ValidationEvent, done func()) {
	e.markValidationBusy()
	defer e.markValidationIdle()
	defer done()
	e.logger.Noticef("drop validation event %d", validationEvent.SeqNo)
}

// process specific implementation logic, including the signature checking,
// the execution of transactions, ledger hash re-computation and etc.
func (e *Executor) process(validationEvent *event.ValidationEvent, done func()) (error, bool) {
	// Invoke the callback no matter success or fail
	e.logger.Noticef("try to execute block %d", validationEvent.SeqNo)
	defer done()

	var (
		validtxs   []*types.Transaction
		invalidtxs []*types.InvalidTransactionRecord
	)

	validtxs, invalidtxs = validationEvent.Transactions, make([]*types.InvalidTransactionRecord, 0)
	err, validateResult := e.applyTransactions(validtxs, invalidtxs, validationEvent.SeqNo, validationEvent.Timestamp)
	if err != nil {
		// short circuit if error occur during the transaction execution
		e.logger.Errorf("process transaction batch #%d failed.", e.namespace, validationEvent.SeqNo)
		return err, false
	}
	// calculate validation result hash for comparison
	hash := e.calculateValidationResultHash(validateResult.MerkleRoot, validateResult.TxRoot, validateResult.ReceiptRoot)
	e.logger.Debugf("invalid transaction number %d", len(validateResult.InvalidTxs))
	e.logger.Debugf("valid transaction number %d", len(validateResult.ValidTxs))
	e.logger.Noticef("execute block %d done", validationEvent.SeqNo)

	e.saveValidationResult(validateResult, validationEvent.SeqNo, hash)
	e.tryToCommit(validationEvent, hash)
	return nil, true
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
func (e *Executor) saveValidationResult(res *ValidationResultRecord, seqNo uint64, hash common.Hash) {
	e.addValidationResult(ValidationTag{hash.Hex(), seqNo}, res)
}

// tryToCommit sends validation result to commit thread.
func (e *Executor) tryToCommit(ev *event.ValidationEvent, hash common.Hash) {
	e.CommitBlock(&event.CommitEvent{
		SeqNo:      ev.SeqNo,
		Timestamp:  ev.Timestamp,
		CommitTime: time.Now().UnixNano(),
		Flag:       true,
		Hash:       hash.Hex(),
		IsPrimary:  true,
	})
}

// pauseValidation stops validation process for a while.
// continue by a signal invocation.
func (e *Executor) pauseValidation() {
	for {
		if v := <-e.getSuspend(IDENTIFIER_VALIDATION); !v {
			e.logger.Notice("un-pause validation process")
			return
		}
	}
}
