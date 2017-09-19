//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"fmt"
	"time"

	"hyperchain/consensus/events"
	"hyperchain/consensus/txpool"
	"hyperchain/manager/event"
)

// batchManager manages basic batch issues, including:
// 1. txPool which manages all transactions received from client or
// 2. batch events timer management
type batchManager struct {
	txPool           txpool.TxPool
	eventMux         *event.TypeMux
	batchSub         event.Subscription	// subscription channel for batch event posted from txPool module
	close            chan bool
	pbftQueue        events.Queue
	batchTimerActive bool			// track the batch timer event, true means there exists an undergoing batch timer event
}

// batchValidator manages batch validate issues
type batchValidator struct {
	lastVid             uint64                 // track the last validate batch seqNo
	currentVid          *uint64                // track the current validate batch seqNo
	cacheValidatedBatch map[string]*cacheBatch // track the cached validated batch

	validateTimer   events.Timer
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
// and push it to pbftQueue for primary to construct TransactionBatch for consensus
func newBatchManager(pbft *pbftImpl) *batchManager {
	bm := &batchManager{}

	// subscribe TxHashBatch
	bm.eventMux = new(event.TypeMux)
	bm.batchSub = bm.eventMux.Subscribe(txpool.TxHashBatch{})
	bm.close = make(chan bool)

	batchSize := pbft.config.GetInt(PBFT_BATCH_SIZE)
	poolSize := pbft.config.GetInt(PBFT_POOL_SIZE)

	batchTimeout, err := time.ParseDuration(pbft.config.GetString(PBFT_BATCH_TIMEOUT))
	if err != nil {
		pbft.logger.Criticalf("Cannot parse batch timeout: %s", err)
	}
	if batchTimeout >= pbft.timerMgr.requestTimeout {
		pbft.timerMgr.requestTimeout = 3 * batchTimeout / 2
		pbft.logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", pbft.timerMgr.requestTimeout)
	}

	// new instance for txPool
	bm.txPool, err = txpool.NewTxPool(pbft.namespace, poolSize, bm.eventMux, batchSize)
	if err != nil {
		panic(fmt.Errorf("Cannot create txpool: %s", err))
	}

	pbft.logger.Infof("PBFT Batch size = %d", batchSize)
	pbft.logger.Infof("PBFT Batch timeout = %v", batchTimeout)

	return bm
}

// start starts a go-routine to listen TxPool Event which continuously waits for TxHashBatch
func (bm *batchManager) start(queue events.Queue) {
	bm.pbftQueue = queue
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
				go bm.pbftQueue.Push(ev)
			}
		}
	}
}

// isBatchTimerActive returns if the batch timer is active or not
func (bm *batchManager) isBatchTimerActive() bool {
	return bm.batchTimerActive
}

// startBatchTimer starts the batch timer and sets the batchTimerActive to true
func (pbft *pbftImpl) startBatchTimer() {
	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}

	pbft.timerMgr.startTimer(BATCH_TIMER, event, pbft.pbftEventQueue)
	pbft.batchMgr.batchTimerActive = true
	pbft.logger.Debugf("Primary %d started the batch timer", pbft.id)
}

// stopBatchTimer stops the batch timer and reset the batchTimerActive to false
func (pbft *pbftImpl) stopBatchTimer() {
	pbft.timerMgr.stopTimer(BATCH_TIMER)
	pbft.batchMgr.batchTimerActive = false
	pbft.logger.Debugf("Primary %d stopped the batch timer", pbft.id)
}

// restartBatchTimer restarts the batch timer
func (pbft *pbftImpl) restartBatchTimer() {
	pbft.timerMgr.stopTimer(BATCH_TIMER)

	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}

	pbft.timerMgr.startTimer(BATCH_TIMER, event, pbft.pbftEventQueue)
	pbft.batchMgr.batchTimerActive = true
	pbft.logger.Debugf("Primary %d restarted the batch timer", pbft.id)
}

// primaryValidateBatch used by primary helps primary pre-validate the batch and stores this TransactionBatch
func (pbft *pbftImpl) primaryValidateBatch(digest string, batch *TransactionBatch, seqNo uint64) {
	// for keep the previous vid before viewchange
	// will specifies the vid to start validate batch)
	var n uint64
	if seqNo != 0 {
		n = seqNo
	} else {
		n = pbft.seqNo + 1
	}

	pbft.seqNo = n

	// store batch to outstandingReqBatches until execute this batch
	pbft.storeMgr.outstandingReqBatches[digest] = batch
	pbft.storeMgr.txBatchStore[digest] = batch

	pbft.logger.Debugf("Primary %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, pbft.view, n, len(batch.HashList))
	// here we soft start a new view timer with requestTimeout+validateTimeout, if primary cannot execute this batch
	// during that timeout, we think there may exist some problems with this primary which will trigger viewchange
	pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout+pbft.timerMgr.getTimeoutValue(VALIDATE_TIMER),
		fmt.Sprintf("new request batch for view=%d/seqNo=%d", pbft.view, n))
	pbft.helper.ValidateBatch(digest, batch.TxList, batch.Timestamp, n, pbft.view, true)

}

// validatePending used by backup nodes validates pending batched stored in preparedCert
func (pbft *pbftImpl) validatePending() {
	// avoid validate multi batches simultaneously
	if pbft.batchVdr.currentVid != nil {
		pbft.logger.Debugf("Backup %d not attempting to send validate because it is currently validate %d", pbft.id, *pbft.batchVdr.currentVid)
		return
	}

	for stop := false; !stop; {
		if find, digest, txBatch, idx := pbft.findNextValidateBatch(); find {
			pbft.execValidate(digest, txBatch, idx)
			cert := pbft.storeMgr.getCert(idx.view, idx.seqNo, digest)
			cert.sentValidate = true
		} else {
			stop = true
		}
	}

}

func (pbft *pbftImpl) findNextValidateBatch() (find bool, digest string, txBatch *TransactionBatch, idx vidx) {

	for idx, digest = range pbft.batchVdr.preparedCert {
		cert := pbft.storeMgr.getCert(idx.view, idx.seqNo, digest)

		if idx.seqNo != pbft.batchVdr.lastVid+1 {
			pbft.logger.Debugf("Backup %d gets validateBatch seqNo=%d, but expect seqNo=%d", pbft.id, idx.seqNo, pbft.batchVdr.lastVid+1)
			continue
		}

		if cert.prePrepare == nil {
			pbft.logger.Warningf("Replica %d get pre-prepare failed for view=%d/seqNo=%d/digest=%s",
				pbft.id, idx.view, idx.seqNo, digest)
			continue
		}
		preprep := cert.prePrepare

		batch, missing, err := pbft.batchMgr.txPool.GetTxsByHashList(digest, preprep.HashBatch.List)
		if err != nil {
			pbft.logger.Warningf("Replica %d get error when get txlist, err: %v", pbft.id, err)
			pbft.sendViewChange()
			return
		}
		if missing != nil {
			pbft.fetchMissingTransaction(preprep, missing)
			return
		}

		currentVid := idx.seqNo
		pbft.batchVdr.setCurrentVid(&currentVid)

		txBatch = &TransactionBatch{
			TxList:    batch,
			HashList:  preprep.HashBatch.List,
			Timestamp: preprep.HashBatch.Timestamp,
		}
		pbft.storeMgr.txBatchStore[preprep.BatchDigest] = txBatch
		pbft.storeMgr.outstandingReqBatches[preprep.BatchDigest] = txBatch

		find = true
		break
	}
	return
}

// execValidate used by backup nodes actually sends validate event
func (pbft *pbftImpl) execValidate(digest string, txBatch *TransactionBatch, idx vidx) {

	pbft.logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, idx.view, idx.seqNo, len(txBatch.TxList))

	pbft.helper.ValidateBatch(digest, txBatch.TxList, txBatch.Timestamp, idx.seqNo, idx.view, false)
	delete(pbft.batchVdr.preparedCert, idx)
	pbft.batchVdr.updateLCVid()

}

// handleTransactionsAfterAbnormal handles the transactions put in txPool during
// viewChange, updateN and recovery
func (pbft *pbftImpl) handleTransactionsAfterAbnormal() {

	// backup does not need to process it
	if pbft.primary(pbft.view) != pbft.id {
		return
	}

	// if primary has transactions in txPool, generate batches of the transactions
	if pbft.batchMgr.txPool.HasTxInPool() {
		pbft.batchMgr.txPool.GenerateTxBatch()
	}

}
