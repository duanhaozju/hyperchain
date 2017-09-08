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

// batchManager manage basic batch issues
// exp:
//      1.batch events timer management
type batchManager struct {
	txPool           txpool.TxPool
	eventMux         *event.TypeMux
	batchSub         event.Subscription
	close            chan bool
	pbftQueue        events.Queue
	batchTimerActive bool
}

//batchValidator used to manager batch validate issues.
type batchValidator struct {
	lastVid             uint64                 // track the last validate batch seqNo
	currentVid          *uint64                // track the current validate batch seqNo
	cacheValidatedBatch map[string]*cacheBatch // track the cached validated batch

	validateTimer   events.Timer
	validateTimeout time.Duration
	preparedCert    map[vidx]string // track the prepared cert to help validate
}

func (bv *batchValidator) setLastVid(lvid uint64) {
	bv.lastVid = lvid
}

//saveToCVB save the cacheBatch into cacheValidatedBatch.
func (bv *batchValidator) saveToCVB(digest string, cb *cacheBatch) {
	bv.cacheValidatedBatch[digest] = cb
}

//containsInCVB judge whether the cache in cacheValidatedBatch.
func (bv *batchValidator) containsInCVB(digest string) bool {
	_, ok := bv.cacheValidatedBatch[digest]
	return ok
}

//update lastVid  and currentVid
func (bv *batchValidator) updateLCVid() {
	bv.lastVid = *bv.currentVid
	bv.currentVid = nil
}

func (bv *batchValidator) setCurrentVid(cvid *uint64) {
	bv.currentVid = cvid
}

//getCVB get cacheValidatedBatch
func (bv *batchValidator) getCVB() map[string]*cacheBatch {
	return bv.cacheValidatedBatch
}

//getCacheBatchFromCVB get cacheBatch form cacheValidatedBatch.
func (bv *batchValidator) getCacheBatchFromCVB(digest string) *cacheBatch {
	return bv.cacheValidatedBatch[digest]
}

//deleteCacheFromCVB delete cacheBatch from cachedValidatedBatch.
func (bv *batchValidator) deleteCacheFromCVB(digest string) {
	delete(bv.cacheValidatedBatch, digest)
}

func newBatchValidator() *batchValidator {

	bv := &batchValidator{}
	bv.cacheValidatedBatch = make(map[string]*cacheBatch)
	bv.preparedCert = make(map[vidx]string)
	return bv
}

// newBatchManager init a instance of batchManager.
func newBatchManager(pbft *pbftImpl) *batchManager {
	bm := &batchManager{}

	bm.eventMux = new(event.TypeMux)
	bm.batchSub = bm.eventMux.Subscribe(txpool.TxHashBatch{})
	bm.close = make(chan bool)

	batchSize := pbft.config.GetInt(PBFT_BATCH_SIZE)
	poolSize := pbft.config.GetInt(PBFT_POOL_SIZE)

	batchTimeout, err := time.ParseDuration(pbft.config.GetString(PBFT_BATCH_TIMEOUT))
	if err != nil {
		pbft.logger.Criticalf("Cannot parse batch timeout: %s", err)
	}
	if batchTimeout >= pbft.timerMgr.requestTimeout { //TODO: change the pbftTimerMgr to batchTimerMgr
		pbft.timerMgr.requestTimeout = 3 * batchTimeout / 2
		pbft.logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", pbft.timerMgr.requestTimeout)
	}

	bm.txPool, err = txpool.NewTxPool(pbft.namespace, poolSize, bm.eventMux, batchSize)
	if err != nil {
		panic(fmt.Errorf("Cannot create txpool: %s", err))
	}

	pbft.logger.Infof("PBFT Batch size = %d", batchSize)
	pbft.logger.Infof("PBFT Batch timeout = %v", batchTimeout)

	return bm
}

func (bm *batchManager) start(queue events.Queue) {
	bm.pbftQueue = queue
	go bm.listenTxPoolEvent()
}

func (bm *batchManager) stop() {
	if bm.close != nil {
		close(bm.close)
	}
}

// listenTxPoolEvent start the thread of listening the TxHashBatch sent by txpool
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

func (bm *batchManager) isBatchTimerActive() bool {
	return bm.batchTimerActive
}

// startBatchTimer start the batch timer
func (pbft *pbftImpl) startBatchTimer() {
	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}

	pbft.timerMgr.startTimer(BATCH_TIMER, event, pbft.pbftEventQueue)
	pbft.batchMgr.batchTimerActive = true
	pbft.logger.Debugf("Primary %d started the batch timer", pbft.id)
}

// stopBatchTimer stop the batch timer
func (pbft *pbftImpl) stopBatchTimer() {
	pbft.timerMgr.stopTimer(BATCH_TIMER)
	pbft.batchMgr.batchTimerActive = false
	pbft.logger.Debugf("Primary %d stopped the batch timer", pbft.id)
}

// restartBatchTimer restart the batch timer
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

func (pbft *pbftImpl) primaryValidateBatch(digest string, batch *TransactionBatch, seqNo uint64) {
	// for keep the previous vid before viewchange
	var n uint64
	if seqNo != 0 {
		n = seqNo
	} else {
		n = pbft.seqNo + 1
	}

	pbft.seqNo = n

	pbft.storeMgr.outstandingReqBatches[digest] = batch
	pbft.storeMgr.txBatchStore[digest] = batch

	pbft.logger.Debugf("Primary %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, pbft.view, n, len(batch.HashList))
	pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout+pbft.timerMgr.getTimeoutValue(VALIDATE_TIMER),
		fmt.Sprintf("new request batch for view=%d/seqNo=%d", pbft.view, n))
	pbft.helper.ValidateBatch(digest, batch.TxList, batch.Timestamp, n, pbft.view, true)

}

func (pbft *pbftImpl) validatePending() {

	if pbft.batchVdr.currentVid != nil {
		pbft.logger.Debugf("Backup %d not attempting to send validate because it is currently validate %d", pbft.id, *pbft.batchVdr.currentVid)
		return
	}

	for idx, digest := range pbft.batchVdr.preparedCert {
		if pbft.preValidate(idx, digest) {
			break
		}
	}
}

func (pbft *pbftImpl) preValidate(idx vidx, digest string) bool {
	if idx.seqNo != pbft.batchVdr.lastVid+1 {
		pbft.logger.Debugf("Backup %d gets validateBatch seqNo=%d, but expect seqNo=%d", pbft.id, idx.seqNo, pbft.batchVdr.lastVid+1)
		return false
	}

	cert := pbft.storeMgr.getCert(idx.view, idx.seqNo, digest)
	if cert.prePrepare == nil {
		pbft.logger.Warningf("Replica %d get pre-prepare failed for view=%d/seqNo=%d/digest=%s",
			pbft.id, idx.view, idx.seqNo, digest)
		return false
	}
	preprep := cert.prePrepare

	batch, missing, err := pbft.batchMgr.txPool.GetTxsByHashList(preprep.BatchDigest, preprep.HashBatch.List)
	if err != nil {
		pbft.logger.Warningf("Replica %d get error when get txlist, err: %v", pbft.id, err)
		pbft.sendViewChange()
		return false
	}
	if missing != nil {
		pbft.fetchMissingTransaction(preprep, missing)
		return false
	}

	txBatch := &TransactionBatch{
		TxList:    batch,
		HashList:  preprep.HashBatch.List,
		Timestamp: preprep.HashBatch.Timestamp,
	}
	pbft.storeMgr.txBatchStore[preprep.BatchDigest] = txBatch
	pbft.storeMgr.outstandingReqBatches[preprep.BatchDigest] = txBatch

	currentVid := idx.seqNo
	pbft.batchVdr.currentVid = &currentVid
	cert.sentValidate = true

	pbft.execValidate(digest, txBatch, idx)

	return true
}

func (pbft *pbftImpl) execValidate(digest string, txBatch *TransactionBatch, idx vidx) {

	pbft.logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, idx.view, idx.seqNo, len(txBatch.TxList))

	pbft.helper.ValidateBatch(digest, txBatch.TxList, txBatch.Timestamp, idx.seqNo, idx.view, false)
	delete(pbft.batchVdr.preparedCert, idx)
	pbft.batchVdr.updateLCVid()

	pbft.validatePending()
}
