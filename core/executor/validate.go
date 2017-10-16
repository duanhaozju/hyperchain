package executor

import (
	"hyperchain/common"
	er "hyperchain/core/errors"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"sort"
	"sync"
)

// ValidationTag unique identification for validation result.
type ValidationTag struct {
	hash  string
	seqNo uint64
}

// represent a validation result collection
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

type SubscriptionData struct {
	Block *types.Block
	Logs  []*types.Log
}

// Validate - the entry function of the validation process.
// Receive the validation event as a parameter and cached them in the channel
// if the pressure is too high.
func (executor *Executor) Validate(validationEvent event.ValidationEvent) {
	executor.addValidationEvent(validationEvent)
}

// listenValidationEvent - validation backend process, use to listen new validation event and dispatch it to the processor.
func (executor *Executor) listenValidationEvent() {
	executor.logger.Notice("validation backend start")
	for {
		select {
		case <-executor.getExit(IDENTIFIER_VALIDATION):
			executor.logger.Notice("validation backend exit")
			return
		case v := <-executor.getSuspend(IDENTIFIER_VALIDATION):
			if v {
				executor.logger.Notice("pause validation process")
				executor.pauseValidation()
			}
		case ev := <-executor.fetchValidationEvent():
			if executor.isReadyToValidation() {
				if success := executor.processValidationEvent(ev, executor.processValidationDone); success == false {
					executor.logger.Errorf("validate #%d failed, system crush down.", ev.SeqNo)
				}
			} else {
				executor.dropValdiateEvent(ev, executor.processValidationDone)
			}
		}
	}
}

// processValidationEvent - process validation event, return true if process successfully, otherwise false will been returned.
func (executor *Executor) processValidationEvent(validationEvent event.ValidationEvent, done func()) bool {
	executor.markValidationBusy()
	defer executor.markValidationIdle()
	if !executor.isDemandSeqNo(validationEvent.SeqNo) {
		executor.addPendingValidationEvent(validationEvent)
		return true
	}
	if _, success := executor.process(validationEvent, done); success == false {
		return false
	}
	executor.incDemandSeqNo()
	return executor.processPendingValidationEvent(done)
}

// processPendingValidationEvent - handle the validation event that is cached in the queue.
func (executor *Executor) processPendingValidationEvent(done func()) bool {
	if executor.cache.pendingValidationEventQ.Len() > 0 {
		// there is still some events remain.
		for {
			if executor.cache.pendingValidationEventQ.Contains(executor.getDemandSeqNo()) {
				ev, _ := executor.fetchPendingValidationEvent(executor.getDemandSeqNo())
				if _, success := executor.process(ev, done); success == false {
					return false
				} else {
					executor.incDemandSeqNo()
					executor.cache.pendingValidationEventQ.RemoveWithCond(ev.SeqNo, RemoveLessThan)
				}
			} else {
				break
			}
		}
	}
	return true
}

// dropValdiateEvent - this function do nothing but consume a validation event.
func (executor *Executor) dropValdiateEvent(validationEvent event.ValidationEvent, done func()) {
	executor.markValidationBusy()
	defer executor.markValidationIdle()
	defer done()
	executor.logger.Noticef("[Namespace = %s] drop validation event %d", executor.namespace, validationEvent.SeqNo)
}

// process - Specific implementation logic, including the signature checking,
// the execution of transactions, ledger hash re-computation and etc.
func (executor *Executor) process(validationEvent event.ValidationEvent, done func()) (error, bool) {
	// invoke the callback no matter success or failed
	defer done()

	var (
		validtxs   []*types.Transaction
		invalidtxs []*types.InvalidTransactionRecord
	)

	invalidtxs, validtxs = executor.checkSign(validationEvent.Transactions)
	err, validateResult := executor.applyTransactions(validtxs, invalidtxs, validationEvent.SeqNo)
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

// checkSign - check the sender's signature validity.
func (executor *Executor) checkSign(txs []*types.Transaction) ([]*types.InvalidTransactionRecord, []*types.Transaction) {
	var (
		invalidtxs []*types.InvalidTransactionRecord
		wg         sync.WaitGroup
		index      []int
		mu         sync.Mutex
	)

	// check signature for each transaction parallelly
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
	// remove invalid transaction from transaction list
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

// applyTransactions - execute transaction one by one.
func (executor *Executor) applyTransactions(txs []*types.Transaction, invalidTxs []*types.InvalidTransactionRecord, seqNo uint64) (error, *ValidationResultRecord) {
	var (
		validtxs []*types.Transaction
		receipts []*types.Receipt
	)

	executor.initCalculator()
	executor.statedb.MarkProcessStart(seqNo)
	// execute transaction one by one
	for i, tx := range txs {
		receipt, _, _, err := executor.ExecTransaction(executor.statedb, tx, i, seqNo)
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
	err, merkleRoot, txRoot, receiptRoot := executor.submitValidationResult()
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

// classifyInvalid - classify invalid transaction via error type.
func (executor *Executor) classifyInvalid(err error) types.InvalidTransactionRecord_ErrType {
	var errType types.InvalidTransactionRecord_ErrType
	if er.IsValueTransferErr(err) {
		errType = types.InvalidTransactionRecord_OUTOFBALANCE
	} else if er.IsExecContractErr(err) {
		tmp := err.(*er.ExecContractError)
		if tmp.GetType() == 0 {
			errType = types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED
		} else if tmp.GetType() == 1 {
			errType = types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED
		}
	} else if er.IsInvalidInvokePermissionErr(err) {
		errType = types.InvalidTransactionRecord_INVALID_PERMISSION
	}
	return errType
}

// submitValidationResult - submit state changes to batch.
func (executor *Executor) submitValidationResult() (error, []byte, []byte, []byte) {
	// flush all state change
	root, err := executor.statedb.Commit()
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] commit state db failed! error msg, ", executor.namespace, err.Error())
		return err, nil, nil, nil
	}
	merkleRoot := root.Bytes()
	res, _ := executor.calculateTransactionsFingerprint(nil, true)
	txRoot := res.Bytes()
	res, _ = executor.calculateReceiptFingerprint(nil, nil, true)
	receiptRoot := res.Bytes()
	executor.recordStateHash(root)
	return nil, merkleRoot, txRoot, receiptRoot
}

func (executor *Executor) resetStateDb() {
	executor.statedb.Reset()
}

// throwInvalidTransactionBack - send all invalid transaction to its birth place.
func (executor *Executor) throwInvalidTransactionBack(invalidtxs []*types.InvalidTransactionRecord) {
	for _, t := range invalidtxs {
		executor.informP2P(NOTIFY_UNICAST_INVALID, t)
	}
}

// saveValidationResult - save validation result to cache.
func (executor *Executor) saveValidationResult(res *ValidationResultRecord, seqNo uint64, hash common.Hash) {
	executor.addValidationResult(ValidationTag{hash.Hex(), seqNo}, res)
}

// sendValidationResult - send validation result to consensus module.
func (executor *Executor) sendValidationResult(res *ValidationResultRecord, ev event.ValidationEvent, hash common.Hash) {
	executor.informConsensus(NOTIFY_VALIDATION_RES, protos.ValidatedTxs{
		Transactions: res.ValidTxs,
		SeqNo:        ev.SeqNo,
		View:         ev.View,
		Hash:         hash.Hex(),
		Timestamp:    ev.Timestamp,
		Digest:       ev.Digest,
	})
}

// pauseValidation - stop validation process for a while.
// continue by a signal invocation.
func (executor *Executor) pauseValidation() {
	for {
		if v := <-executor.getSuspend(IDENTIFIER_VALIDATION); !v {
			executor.logger.Notice("un-pause validation process")
			return
		}
	}
}
