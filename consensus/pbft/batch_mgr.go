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
	vid                 uint64                 // track the validate sequence number
	lastVid             uint64                 // track the last validate batch seqNo
	currentVid          *uint64                // track the current validate batch seqNo
	cacheValidatedBatch map[string]*cacheBatch // track the cached validated batch
	preparedCert    map[vidx]string            // track the prepared cert to help validate
	spNullRequest   map[msgID]*PrePrepare      // track the special pre-prepare received from primary
}

// setVid sets the value of vid
func (bv *batchValidator) setVid(vid uint64) {
	bv.vid = vid
}

// incVid increases vid.
func (bv *batchValidator) incVid() {
	bv.vid = bv.vid + 1
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
	bv.spNullRequest = make(map[msgID]*PrePrepare)
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
func (pbft *pbftImpl) primaryValidateBatch(digest string, batch *TransactionBatch, vid uint64) {
	// for keep the previous vid before viewchange (which means when viewchange or node addition happens, new primary
	// will specifies the vid to start validate batch)
	var n uint64
	if vid != 0 {
		n = vid
	} else {
		n = pbft.batchVdr.vid + 1
	}

	pbft.batchVdr.vid = n

	// store batch to outstandingReqBatches until execute this batch
	pbft.storeMgr.outstandingReqBatches[digest] = batch
	pbft.storeMgr.txBatchStore[digest] = batch

	pbft.logger.Debugf("Primary %d try to validate batch for view=%d/vid=%d, batch size: %d", pbft.id, pbft.view, pbft.batchVdr.vid, len(batch.HashList))
	// here we soft start a new view timer with requestTimeout+validateTimeout, if primary cannot execute this batch
	// during that timeout, we think there may exist some problems with this primary which will trigger viewchange
	pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout+pbft.timerMgr.getTimeoutValue(VALIDATE_TIMER),
		fmt.Sprintf("new request batch for view=%d/vid=%d", pbft.view, pbft.batchVdr.vid))
	pbft.helper.ValidateBatch(digest, batch.TxList, batch.Timestamp, uint64(0), n, pbft.view, true)
}

// validatePending used by backup nodes validates pending batched stored in preparedCert
func (pbft *pbftImpl) validatePending() {
	// avoid validate multi batches simultaneously
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

// preValidate used by backup nodes prepares to validate a batch:
// 1. judge whether vid equals to lastVid+1(expected validate id)
// 2. get all txs by hash list to construct TransactionBatch, if missing txs, fetch from primary
// 3. send execValidate
func (pbft *pbftImpl) preValidate(idx vidx, digest string) bool {
	if idx.vid != pbft.batchVdr.lastVid+1 {
		pbft.logger.Debugf("Backup %d gets validateBatch vid=%d, but expect vid=%d", pbft.id, idx.vid, pbft.batchVdr.lastVid+1)
		return false
	}

	var preprep *PrePrepare
	var ok bool

	if idx.seqNo == uint64(0) {
		// special batch only stored in spNullRequest
		id := msgID{idx.view, idx.vid, digest}
		preprep, ok = pbft.batchVdr.spNullRequest[id]
		if !ok {
			pbft.logger.Warningf("Replica %d get pre-prepare failed for special null-request view=%d/vid=%d/digest=%s",
				pbft.id, idx.view, idx.vid, digest)
			return false
		}
	} else {
		// normal batch stored in certStore
		cert := pbft.storeMgr.getCert(idx.view, idx.seqNo, digest)
		if cert.prePrepare == nil {
			pbft.logger.Warningf("Replica %d get pre-prepare failed for view=%d/seqNo=%d/digest=%s",
				pbft.id, idx.view, idx.seqNo, digest)
			return false
		}
		preprep = cert.prePrepare
	}

	// fetch all txs from txPool module
	batch, missing, err := pbft.batchMgr.txPool.GetTxsByHashList(preprep.BatchDigest, preprep.HashBatch.List)
	if err != nil {
		pbft.logger.Warningf("Replica %d get error when get txlist, err: %v", pbft.id, err)
		pbft.sendViewChange()
		return false
	}
	if missing != nil {
		// fetch missing txs from primary
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

	// set currentVid to avoid validate multi batches simultaneously
	currentVid := idx.vid
	pbft.batchVdr.currentVid = &currentVid

	if idx.seqNo != uint64(0) {
		cert := pbft.storeMgr.getCert(idx.view, idx.seqNo, digest)
		cert.sentValidate = true
	}

	pbft.execValidate(digest, txBatch, idx)

	return true
}

// execValidate used by backup nodes actually sends validate event
func (pbft *pbftImpl) execValidate(digest string, txBatch *TransactionBatch, idx vidx) {

	pbft.logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d/vid=%d, batch size: %d", pbft.id, idx.view, idx.seqNo, idx.vid, len(txBatch.TxList))

	pbft.helper.ValidateBatch(digest, txBatch.TxList, txBatch.Timestamp, idx.seqNo, idx.vid, idx.view, false)
	delete(pbft.batchVdr.preparedCert, idx)
	pbft.batchVdr.updateLCVid()

	pbft.validatePending()
}
