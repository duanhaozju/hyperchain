//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"hyperchain/consensus/events"
	"time"
	"hyperchain/core/types"

	"hyperchain/common"
	"fmt"
	"github.com/golang/protobuf/proto"
)

// batchManager manage basic batch issues
// exp:
// 	1.pushEvent
//      2.batch events timer management
type batchManager struct{
	batchSize        	int
	batchStore       	[]*types.Transaction            //ordered message batch
	batchEventsManager  	events.Manager //pbft.batchManager => batchManager
	batchTimer       	events.Timer
	batchTimerActive 	bool
	batchTimeout     	time.Duration

	etf                 events.TimerFactory //batch event timer factory
	pbftId              uint64

	pbftEventQueue      events.Queue
	txEventQueue        events.Queue
}

//batchValidator used to manager batch validate issues.
type batchValidator struct {
	vid                 	uint64                       // track the validate sequence number
	lastVid             	uint64                       // track the last validate batch seqNo
	currentVid          	*uint64                      // track the current validate batch seqNo

	validatedBatchStore 	map[string]*TransactionBatch // track the validated transaction batch
	cacheValidatedBatch 	map[string]*cacheBatch       // track the cached validated batch

	validateTimer		events.Timer
	validateTimeout		time.Duration
	preparedCert            map[msgID]string             // track the prepared cert to help validate

	pbftId                  uint64
}

func (bv *batchValidator) setVid(vid uint64)  {
	bv.vid = vid
}

//incVid increase vid.
func (bv *batchValidator) incVid()  {
	bv.vid = bv.vid + 1
}

func (bv *batchValidator) setLastVid(lvid uint64)  {
	bv.lastVid = lvid
}

//saveToCVB save the cacheBatch into cacheValidatedBatch.
func (bv *batchValidator) saveToCVB(digest string, cb *cacheBatch)  {
	bv.cacheValidatedBatch[digest] = cb
}

//saveToVBS save the transaction into validatedBatchStore.
func (bv *batchValidator) saveToVBS(digest string, tx *TransactionBatch)  {
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
func (bv *batchValidator) updateLCVid()  {
	bv.lastVid = *bv.currentVid
	bv.currentVid = nil
}

func (bv *batchValidator) setCurrentVid(cvid *uint64)  {
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
func (bv *batchValidator) deleteCacheFromCVB(digest string)  {
	delete(bv.cacheValidatedBatch, digest)
}

//deleteTxFromVBS delete transaction from validatedBatchStore.
func (bv *batchValidator) deleteTxFromVBS(digest string)  {
	delete(bv.validatedBatchStore, digest)
}

//emptyVBS empty the validatedBatchStore.
func (bv *batchValidator) emptyVBS()  {
	bv.validatedBatchStore = make(map[string]*TransactionBatch)
}

//vbsSize return the size of validatedBatchStore.
func (bv *batchValidator) vbsSize() int {
	return len(bv.validatedBatchStore)
}


func newBatchValidator(conf *common.Config, pbft *pbftImpl) *batchValidator {

	bv := &batchValidator{}
	bv.validatedBatchStore = make(map[string]*TransactionBatch)
	bv.cacheValidatedBatch = make(map[string]*cacheBatch)
	bv.preparedCert = make(map[msgID]string)

	logger.Infof("PBFT Batch size = %d", pbft.batchMgr.batchSize)
	logger.Infof("PBFT Batch timeout = %v", pbft.batchMgr.batchTimeout)

	bv.pbftId = pbft.id
	return bv
}

// newBatchManager init a instance of batchManager.
func newBatchManager(conf *common.Config, pbft *pbftImpl) *batchManager {
	bm := &batchManager{}
	bm.batchEventsManager = events.NewManagerImpl()
	bm.batchEventsManager.SetReceiver(pbft)

	bm.etf = events.NewTimerFactoryImpl(bm.batchEventsManager) //eft EventTimerFactory
	pbft.reqEventQueue = events.GetQueue(bm.batchEventsManager.Queue())

	bm.pbftEventQueue = pbft.pbftEventQueue
	bm.txEventQueue = pbft.reqEventQueue
	// initialize the batchTimeout

	bm.batchTimer = bm.etf.CreateTimer()
	bm.batchSize = conf.GetInt(PBFT_BATCH_SIZE)
	bm.batchStore = nil
	var err error
	bm.batchTimeout, err = time.ParseDuration(conf.GetString(PBFT_BATCH_TIMEOUT))
	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}

	if bm.batchTimeout >= pbft.pbftTimerMgr.requestTimeout {//TODO: change the pbftTimerMgr to batchTimerMgr
		pbft.pbftTimerMgr.requestTimeout = 3 * bm.batchTimeout / 2
		logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", pbft.pbftTimerMgr.requestTimeout)
	}
	bm.pbftId = pbft.id

	return bm
}


//createTimer create batch events related timer.
func (bm *batchManager) createTimer() events.Timer {
	return bm.etf.CreateTimer()
}

func (bm *batchManager) start()  {
	bm.batchEventsManager.Start()
}

func (bm *batchManager) batchStoreSize() int {
	return len(bm.batchStore)
}

func (bm *batchManager) isBatchStoreEmpty() bool  {
	return bm.batchStoreSize() == 0
}

func (bm *batchManager) setBatchStore(bs []*types.Transaction)  {
	bm.batchStore = bs
}

func (bm *batchManager) addTransaction(tx *types.Transaction)  {
	bm.batchStore = append(bm.batchStore, tx)
}

func (bm *batchManager) isBatchTimerActive() bool  {
	return bm.batchTimerActive
}

func (bm *batchManager) canSendBatch() bool {
	return bm.batchStoreSize() >= bm.batchSize
}

//pushEvent push the event into the batch events queue.
func (bm *batchManager) pushEvent(event interface{})  {
	logger.Debugf("send event into batch event queue, %v", event)
	bm.batchEventsManager.Queue() <- event
}

//startBatchTimer stop the batch event timer.
func (bm *batchManager) startBatchTimer()  {
	batchTimerEvent := &LocalEvent{
		Service:CORE_PBFT_SERVICE,
		EventType:CORE_BATCH_TIMER_EVENT,
	}
	bm.batchTimer.Reset(bm.batchTimeout, batchTimerEvent)
	bm.batchTimerActive = true
	logger.Debugf("Replica %d started the batch timer", bm.pbftId)
}

//stopBatchTimer stop batch Timer.
func (bm *batchManager) stopBatchTimer() {
	bm.batchTimer.Stop()
	bm.batchTimerActive = false
	logger.Debugf("Replica %d stpbftped the batch timer", bm.pbftId)
}

//sendBatchRequest send batch request into pbft event queue.
func (bm *batchManager) sendBatchRequest() error {
	bm.stopBatchTimer()

	if bm.isBatchStoreEmpty() {
		logger.Error("Told to send an empty batch store for ordering, ignoring")
		return nil
	}

	reqBatch := &TransactionBatch{
		Batch:     bm.batchStore,
		Timestamp: time.Now().UnixNano(),
	}
	payload, err := proto.Marshal(reqBatch)
	if err != nil {
		logger.Errorf("ConsensusMessage_TRANSACTION Marshal Error", err)
		return nil
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_TRANSACTION,
		Payload: payload,
	}

	bm.setBatchStore(nil)
	logger.Infof("Creating batch with %d requests", len(reqBatch.Batch))

	go bm.txEventQueue.Push(consensusMsg)

	return nil
}

// recvTransaction receive transaction from client.
func (bm *batchManager) recvTransaction(tx *types.Transaction) error {
	bm.addTransaction(tx)

	if !bm.isBatchTimerActive() {
		bm.startBatchTimer()
	}

	if bm.canSendBatch() {
		return bm.sendBatchRequest()
	}

	return nil
}

func (pbft *pbftImpl) primaryValidateBatch(txBatch *TransactionBatch, vid uint64) {

	newBatch, txStore := pbft.removeDuplicate(txBatch)
	if txStore.Len() == 0 {
		logger.Warningf("Primary %d get empty batch after check duplicate", pbft.id)
		return
	}

	// for keep the previous vid before viewchange
	var n uint64
	if vid != 0 {
		n = vid
	} else {
		n = pbft.batchVdr.vid + 1
	}

	pbft.batchVdr.vid = n
	pbft.duplicator[n] = txStore

	logger.Debugf("Primary %d try to validate batch for view=%d/vid=%d, batch size: %d", pbft.id, pbft.view, pbft.batchVdr.vid, txStore.Len())
	pbft.helper.ValidateBatch(newBatch.Batch, newBatch.Timestamp, n, pbft.view, true)

}

func (pbft *pbftImpl) validatePending() {

	if pbft.batchVdr.currentVid != nil {
		logger.Debugf("Backup %d not attempting to send validate because it is currently validate %d", pbft.id, pbft.batchVdr.currentVid)
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
		logger.Debugf("Backup %d already call validate for batch view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		return false
	}

	if idx.n != pbft.batchVdr.lastVid+1 {
		logger.Debugf("Backup %d gets validateBatch seqNo=%d, but expect seqNo=%d", pbft.id, idx.n, pbft.batchVdr.lastVid+1)
		return false
	}

	currentVid := idx.n
	pbft.batchVdr.currentVid = &currentVid

	txStore, err := pbft.checkDuplicate(cert.prePrepare.TransactionBatch)
	if err != nil {
		logger.Warningf("Backup %d find duplicate transaction in the batch for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
		pbft.sendViewChange()
		return true
	}
	logger.Debugf("Backup %d cache duplicator for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
	pbft.duplicator[idx.n] = txStore
	pbft.execValidate(cert.prePrepare.TransactionBatch, idx)
	cert.sentValidate = true

	return true
}

func (pbft *pbftImpl) execValidate(txBatch *TransactionBatch, idx msgID) {

	logger.Debugf("Backup %d try to validate batch for view=%d/seqNo=%d, batch size: %d", pbft.id, idx.v, idx.n, len(txBatch.Batch))

	pbft.helper.ValidateBatch(txBatch.Batch, txBatch.Timestamp, idx.n, idx.v, false)
	delete(pbft.batchVdr.preparedCert, idx)
	pbft.batchVdr.lastVid = *pbft.batchVdr.currentVid
	pbft.batchVdr.currentVid = nil

	pbft.validatePending()
}
