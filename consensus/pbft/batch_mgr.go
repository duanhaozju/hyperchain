//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"hyperchain/consensus/events"
	"time"

	"hyperchain/common"
	"fmt"
	"sync/atomic"
	"hyperchain/consensus/txpool"
)

// batchManager manage basic batch issues
// exp:
// 	1.pushEvent
//      2.batch events timer management
type batchManager struct {
	txPool 		   txpool.TxPool
	batchEventsManager events.Manager       //pbft.batchManager => batchManager
}

//batchValidator used to manager batch validate issues.
type batchValidator struct {
	vid                 	uint64                       // track the validate sequence number
	lastVid             	uint64                       // track the last validate batch seqNo
	currentVid          	*uint64                      // track the current validate batch seqNo
	validateCount		int32
	validatedBatchStore 	map[string]*TransactionBatch // track the validated transaction batch
	cacheValidatedBatch 	map[string]*cacheBatch       // track the cached validated batch

	validateTimer		events.Timer
	validateTimeout		time.Duration
	preparedCert            map[vidx]string             // track the prepared cert to help validate

	pbftId                  uint64
}

func (bv *batchValidator) setVid(vid uint64) {
	bv.vid = vid
}

//incVid increase vid.
func (bv *batchValidator) incVid() {
	bv.vid = bv.vid + 1
}

func (bv *batchValidator) setLastVid(lvid uint64) {
	bv.lastVid = lvid
}

//saveToCVB save the cacheBatch into cacheValidatedBatch.
func (bv *batchValidator) saveToCVB(digest string, cb *cacheBatch) {
	bv.cacheValidatedBatch[digest] = cb
}

//saveToVBS save the transaction into validatedBatchStore.
func (bv *batchValidator) saveToVBS(digest string, tx *TransactionBatch) {
	bv.validatedBatchStore[digest] = tx
}

//containsInCVB judge whether the cache in cacheValidatedBatch.
func (bv *batchValidator) containsInCVB(digest string) bool {
	_, ok := bv.cacheValidatedBatch[digest]
	return ok
}

//containsInVBS judge whether the cache in validatedBatchStore.
func (bv *batchValidator) containsInVBS(digest string) bool {
	_, ok := bv.validatedBatchStore[digest]
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

//getTxFromVBS get transaction from validatedBatchStore.
func (bv *batchValidator) getTxBatchFromVBS(digest string) *TransactionBatch {
	return bv.validatedBatchStore[digest]
}

//deleteCacheFromCVB delete cacheBatch from cachedValidatedBatch.
func (bv *batchValidator) deleteCacheFromCVB(digest string) {
	delete(bv.cacheValidatedBatch, digest)
}

//deleteTxFromVBS delete transaction from validatedBatchStore.
func (bv *batchValidator) deleteTxFromVBS(digest string) {
	delete(bv.validatedBatchStore, digest)
}

//emptyVBS empty the validatedBatchStore.
func (bv *batchValidator) emptyVBS() {
	bv.validatedBatchStore = make(map[string]*TransactionBatch)
}

//vbsSize return the size of validatedBatchStore.
func (bv *batchValidator) vbsSize() int {
	return len(bv.validatedBatchStore)
}

func newBatchValidator(pbft *pbftImpl) *batchValidator {

	bv := &batchValidator{}
	bv.validatedBatchStore = make(map[string]*TransactionBatch)
	bv.cacheValidatedBatch = make(map[string]*cacheBatch)
	bv.preparedCert = make(map[vidx]string)
	atomic.StoreInt32(&bv.validateCount, 0)
	bv.pbftId = pbft.id
	return bv
}

// newBatchManager init a instance of batchManager.
func newBatchManager(conf *common.Config, pbft *pbftImpl) *batchManager {
	bm := &batchManager{}
	bm.batchEventsManager = events.NewManagerImpl(conf.GetString(common.NAMESPACE))
	bm.batchEventsManager.SetReceiver(pbft)
	pbft.reqEventQueue = events.GetQueue(bm.batchEventsManager.Queue())


	batchSize := conf.GetInt(PBFT_BATCH_SIZE)
	poolSize := conf.GetInt(PBFT_POOL_SIZE)
	batchTimeout, err := time.ParseDuration(conf.GetString(PBFT_BATCH_TIMEOUT))
	if err != nil {
		pbft.logger.Criticalf("Cannot parse batch timeout: %s", err)
	}
	bm.txPool, err = txpool.NewTxPool(poolSize, pbft.reqEventQueue, batchTimeout, batchSize)
	if err != nil {
		panic(fmt.Errorf("Cannot create txpool: %s", err))
	}


	if batchTimeout >= pbft.timerMgr.requestTimeout {//TODO: change the pbftTimerMgr to batchTimerMgr
		pbft.timerMgr.requestTimeout = 3 * batchTimeout / 2
		pbft.logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", pbft.timerMgr.requestTimeout)
	}

	pbft.logger.Infof("PBFT Batch size = %d", batchSize)
	pbft.logger.Infof("PBFT Batch timeout = %v", batchTimeout)

	return bm
}

func (bm *batchManager) start() {
	bm.batchEventsManager.Start()
}

func (bm *batchManager) stop() {
	bm.batchEventsManager.Stop()
}


//pushEvent push the event into the batch events queue.
func (bm *batchManager) pushEvent(event interface{}) {
	//pbft.logger.Debugf("send event into batch event queue, %v", event)
	bm.batchEventsManager.Queue() <- event
}


func (pbft *pbftImpl) primaryValidateBatch(digest string, batch *TransactionBatch, vid uint64) {


	// for keep the previous vid before viewchange
	var n uint64
	if vid != 0 {
		n = vid
	} else {
		n = pbft.batchVdr.vid + 1
	}

	// ignore too many validated batch as we limited the high watermark in send pre-prepare

	count := atomic.LoadInt32(&pbft.batchVdr.validateCount)
	if uint64(count) >= pbft.L {
		pbft.logger.Warningf("Primary %d try to validate batch for vid=%d, but we had already send %d ValidateEvent", pbft.id, n, count)
		return
	}

	pbft.batchVdr.vid = n

	atomic.AddInt32(&pbft.batchVdr.validateCount, 1)

	pbft.logger.Debugf("Primary %d try to validate batch for view=%d/vid=%d, batch size: %d", pbft.id, pbft.view, pbft.batchVdr.vid, len(batch.HashList))
	pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout + pbft.timerMgr.getTimeoutValue(VALIDATE_TIMER),
		fmt.Sprintf("new request batch for view=%d/vid=%d", pbft.view, pbft.batchVdr.vid))
	pbft.helper.ValidateBatch(digest, batch.TxList, batch.Timestamp, uint64(0), n, pbft.view, true)

}

func (pbft *pbftImpl) validatePending() {

	if pbft.batchVdr.currentVid != nil {
		pbft.logger.Debugf("Backup %d not attempting to send validate because it is currently validate %d", pbft.id, *pbft.batchVdr.currentVid)
		return
	}

	for idx := range pbft.batchVdr.preparedCert {
		if pbft.preValidate(idx) {
			break
		}
	}
}

func (pbft *pbftImpl) preValidate(idx msgID) bool {

	cert := pbft.storeMgr.certStore[idx]

	if cert == nil || cert.prePrepare == nil {
		pbft.logger.Debugf("Backup %d already call validate for batch view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		return false
	}

	if idx.n != pbft.batchVdr.lastVid+1 {
		pbft.logger.Debugf("Backup %d gets validateBatch seqNo=%d, but expect seqNo=%d", pbft.id, idx.n, pbft.batchVdr.lastVid+1)
		return false
	}

	currentVid := idx.n
	pbft.batchVdr.currentVid = &currentVid

	txStore, err := pbft.checkDuplicate(cert.prePrepare.TransactionBatch)
	if err != nil {
		pbft.logger.Warningf("Backup %d find duplicate transaction in the batch for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		pbft.sendViewChange()
		return true
	}
	pbft.logger.Debugf("Backup %d cache duplicator for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)

	pbft.dupLock.Lock()
	pbft.duplicator[idx.n] = txStore
	pbft.dupLock.Unlock()

	pbft.execValidate(cert.prePrepare.TransactionBatch, idx)
	cert.sentValidate = true

	return true
}

func (pbft *pbftImpl) execValidate(txBatch *TransactionBatch, idx msgID) {

	pbft.logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, idx.v, idx.n, len(txBatch.TxList))

	pbft.helper.ValidateBatch(txBatch.TxList, txBatch.Timestamp, idx.n, idx.v, false)
	delete(pbft.batchVdr.preparedCert, idx)
	pbft.batchVdr.lastVid = *pbft.batchVdr.currentVid
	pbft.batchVdr.currentVid = nil

	pbft.validatePending()
}
