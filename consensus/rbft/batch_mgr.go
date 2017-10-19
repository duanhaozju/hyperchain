//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"time"

	"hyperchain/consensus/txpool"
	"hyperchain/manager/event"
)

// batchManager manages basic batch issues, including:
// 1. txPool which manages all transactions received from client or rpc layer
// 2. batch events timer management
type batchManager struct {
	txPool           txpool.TxPool
	eventMux         *event.TypeMux
	batchSub         event.Subscription // subscription channel for batch event posted from txPool module
	close            chan bool
	rbftQueue        *event.TypeMux
	batchTimerActive bool // track the batch timer event, true means there exists an undergoing batch timer event
}

// batchValidator manages batch validate issues
type batchValidator struct {
	lastVid             uint64  // track the last validate batch seqNo
	currentVid          *uint64 // track the current validate batch seqNo
	validateCount       int32
	cacheValidatedBatch map[string]*cacheBatch // track the cached validated batch

	validateTimeout time.Duration
	preparedCert    map[vidx]string // track the prepared cert to help validate
}

// setLastVid sets the lastVid to lvid
func (bv *batchValidator) setLastVid(lvid uint64) {
	bv.lastVid = lvid
}

// saveToCVB saves the cacheBatch into cacheValidatedBatch.
func (bv *batchValidator) saveToCVB(digest string, cb *cacheBatch) {
	bv.cacheValidatedBatch[digest] = cb
}

// containsInCVB judges whether the digest is in cacheValidatedBatch or not.
func (bv *batchValidator) containsInCVB(digest string) bool {
	_, ok := bv.cacheValidatedBatch[digest]
	return ok
}

// updateLCVid updates lastVid to the value of currentVid and reset currentVid to nil
func (bv *batchValidator) updateLCVid() {
	bv.lastVid = *bv.currentVid
	bv.currentVid = nil
}

// setCurrentVid sets the value of currentVid
func (bv *batchValidator) setCurrentVid(cvid *uint64) {
	bv.currentVid = cvid
}

// getCVB gets cacheValidatedBatch
func (bv *batchValidator) getCVB() map[string]*cacheBatch {
	return bv.cacheValidatedBatch
}

// getCacheBatchFromCVB gets cacheBatch from cacheValidatedBatch with specified digest
func (bv *batchValidator) getCacheBatchFromCVB(digest string) *cacheBatch {
	return bv.cacheValidatedBatch[digest]
}

// deleteCacheFromCVB deletes cacheBatch from cachedValidatedBatch with specified digest
func (bv *batchValidator) deleteCacheFromCVB(digest string) {
	delete(bv.cacheValidatedBatch, digest)
}

// newBatchValidator initializes an instance of batchValidator
func newBatchValidator() *batchValidator {
	bv := &batchValidator{}
	bv.cacheValidatedBatch = make(map[string]*cacheBatch)
	bv.preparedCert = make(map[vidx]string)
	return bv
}

// newBatchManager initializes an instance of batchManager. batchManager subscribes TxHashBatch from txPool module
// and push it to rbftQueue for primary to construct TransactionBatch for consensus
func newBatchManager(rbft *rbftImpl) *batchManager {
	bm := &batchManager{}

	// subscribe TxHashBatch
	bm.eventMux = new(event.TypeMux)
	bm.batchSub = bm.eventMux.Subscribe(txpool.TxHashBatch{})
	bm.close = make(chan bool)

	batchSize := rbft.config.GetInt(RBFT_BATCH_SIZE)
	poolSize := rbft.config.GetInt(RBFT_POOL_SIZE)

	batchTimeout, err := time.ParseDuration(rbft.config.GetString(RBFT_BATCH_TIMEOUT))
	if err != nil {
		rbft.logger.Criticalf("Cannot parse batch timeout: %s", err)
	}
	if batchTimeout >= rbft.timerMgr.requestTimeout {
		rbft.timerMgr.requestTimeout = 3 * batchTimeout / 2
		rbft.logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", rbft.timerMgr.requestTimeout)
	}

	// new instance for txPool
	bm.txPool, err = txpool.NewTxPool(rbft.namespace, poolSize, bm.eventMux, batchSize)
	if err != nil {
		panic(fmt.Errorf("Cannot create txpool: %s", err))
	}

	rbft.logger.Infof("RBFT Batch size = %d", batchSize)
	rbft.logger.Infof("RBFT Batch timeout = %v", batchTimeout)

	return bm
}

// start starts a go-routine to listen TxPool Event which continuously waits for TxHashBatch
func (bm *batchManager) start(queue *event.TypeMux) {
	bm.rbftQueue = queue
	go bm.listenTxPoolEvent()
}

// stop closes the bm.close channel which will stop the listening go-routine
func (bm *batchManager) stop() {
	if bm.close != nil {
		close(bm.close)
		bm.close = nil
	}
}

// listenTxPoolEvent continuously listens the TxHashBatch event sent from txPool or the bm.close flag which will stop
// this listening go-routine
func (bm *batchManager) listenTxPoolEvent() {
	for {
		select {
		case <-bm.close:
			return
		case obj := <-bm.batchSub.Chan():
			switch ev := obj.Data.(type) {
			case txpool.TxHashBatch:
				go bm.rbftQueue.Post(ev)
			}
		}
	}
}

// isBatchTimerActive returns if the batch timer is active or not
func (bm *batchManager) isBatchTimerActive() bool {
	return bm.batchTimerActive
}

// startBatchTimer starts the batch timer and sets the batchTimerActive to true
func (rbft *rbftImpl) startBatchTimer() {
	event := &LocalEvent{
		Service:   CORE_RBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}

	rbft.timerMgr.startTimer(BATCH_TIMER, event, rbft.eventMux)
	rbft.batchMgr.batchTimerActive = true
	rbft.logger.Debugf("Primary %d started the batch timer", rbft.id)
}

// stopBatchTimer stops the batch timer and reset the batchTimerActive to false
func (rbft *rbftImpl) stopBatchTimer() {
	rbft.timerMgr.stopTimer(BATCH_TIMER)
	rbft.batchMgr.batchTimerActive = false
	rbft.logger.Debugf("Primary %d stopped the batch timer", rbft.id)
}

// restartBatchTimer restarts the batch timer
func (rbft *rbftImpl) restartBatchTimer() {
	rbft.timerMgr.stopTimer(BATCH_TIMER)

	event := &LocalEvent{
		Service:   CORE_RBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}

	rbft.timerMgr.startTimer(BATCH_TIMER, event, rbft.eventMux)
	rbft.batchMgr.batchTimerActive = true
	rbft.logger.Debugf("Primary %d restarted the batch timer", rbft.id)
}

// primaryValidateBatch used by primary helps primary pre-validate the batch and stores this TransactionBatch
func (rbft *rbftImpl) primaryValidateBatch(digest string, batch *TransactionBatch, seqNo uint64) {
	// for keep the previous vid before viewchange
	// will specifies the vid to start validate batch)
	var n uint64
	if seqNo != 0 {
		n = seqNo
	} else {
		n = rbft.seqNo + 1
	}

	rbft.seqNo = n

	// store batch to outstandingReqBatches until execute this batch
	rbft.storeMgr.outstandingReqBatches[digest] = batch
	rbft.storeMgr.txBatchStore[digest] = batch

	rbft.logger.Debugf("Primary %d try to validate batch for view=%d/seqNo=%d, batch size: %d", rbft.id, rbft.view, n, len(batch.HashList))
	// here we soft start a new view timer with requestTimeout+validateTimeout, if primary cannot execute this batch
	// during that timeout, we think there may exist some problems with this primary which will trigger viewchange
	rbft.softStartNewViewTimer(rbft.timerMgr.requestTimeout+rbft.timerMgr.getTimeoutValue(VALIDATE_TIMER),
		fmt.Sprintf("new request batch for view=%d/seqNo=%d", rbft.view, n))
	rbft.helper.ValidateBatch(digest, batch.TxList, batch.Timestamp, n, rbft.view, true)

}

// validatePending used by backup nodes validates pending batched stored in preparedCert
func (rbft *rbftImpl) validatePending() {
	// avoid validate multi batches simultaneously
	if rbft.batchVdr.currentVid != nil {
		rbft.logger.Debugf("Backup %d not attempting to send validate because it is currently validate %d", rbft.id, *rbft.batchVdr.currentVid)
		return
	}

	for stop := false; !stop; {
		if find, digest, txBatch, idx := rbft.findNextValidateBatch(); find {
			rbft.execValidate(digest, txBatch, idx)
			cert := rbft.storeMgr.getCert(idx.view, idx.seqNo, digest)
			cert.sentValidate = true
		} else {
			stop = true
		}
	}

}

func (rbft *rbftImpl) findNextValidateBatch() (find bool, digest string, txBatch *TransactionBatch, idx vidx) {

	for idx, digest = range rbft.batchVdr.preparedCert {
		cert := rbft.storeMgr.getCert(idx.view, idx.seqNo, digest)

		if idx.seqNo != rbft.batchVdr.lastVid+1 {
			rbft.logger.Debugf("Backup %d gets validateBatch seqNo=%d, but expect seqNo=%d", rbft.id, idx.seqNo, rbft.batchVdr.lastVid+1)
			continue
		}

		if cert.prePrepare == nil {
			rbft.logger.Warningf("Replica %d get pre-prepare failed for view=%d/seqNo=%d/digest=%s",
				rbft.id, idx.view, idx.seqNo, digest)
			continue
		}
		preprep := cert.prePrepare

		batch, missing, err := rbft.batchMgr.txPool.GetTxsByHashList(digest, preprep.HashBatch.List)
		if err != nil {
			rbft.logger.Warningf("Replica %d get error when get txlist, err: %v", rbft.id, err)
			rbft.sendViewChange()
			return
		}
		if missing != nil {
			rbft.fetchMissingTransaction(preprep, missing)
			return
		}

		currentVid := idx.seqNo
		rbft.batchVdr.setCurrentVid(&currentVid)

		txBatch = &TransactionBatch{
			TxList:     batch,
			HashList:   preprep.HashBatch.List,
			Timestamp:  preprep.HashBatch.Timestamp,
			SeqNo:      preprep.SequenceNumber,
			ResultHash: preprep.ResultHash,
		}
		rbft.storeMgr.txBatchStore[preprep.BatchDigest] = txBatch
		rbft.storeMgr.outstandingReqBatches[preprep.BatchDigest] = txBatch

		find = true
		break
	}
	return
}

// execValidate used by backup nodes actually sends validate event
func (rbft *rbftImpl) execValidate(digest string, txBatch *TransactionBatch, idx vidx) {

	rbft.logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d, batch size: %d", rbft.id, idx.view, idx.seqNo, len(txBatch.TxList))

	rbft.helper.ValidateBatch(digest, txBatch.TxList, txBatch.Timestamp, idx.seqNo, idx.view, false)
	delete(rbft.batchVdr.preparedCert, idx)
	rbft.batchVdr.updateLCVid()

}

// handleTransactionsAfterAbnormal handles the transactions put in txPool during
// viewChange, updateN and recovery
func (rbft *rbftImpl) handleTransactionsAfterAbnormal() {

	// backup does not need to process it
	if !rbft.isPrimary(rbft.id) {
		return
	}

	// if primary has transactions in txPool, generate batches of the transactions
	if rbft.batchMgr.txPool.HasTxInPool() {
		rbft.batchMgr.txPool.GenerateTxBatch()
	}

}
