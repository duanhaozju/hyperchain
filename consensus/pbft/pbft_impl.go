//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"

	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/core/types"
	"hyperchain/protos"
	"hyperchain/common"
	"sync/atomic"
	"bytes"
)

// batch is used to construct reqbatch, the middle layer between outer to pbft
type pbftImpl struct {
	activeView     uint32                       // view change happening
	f              int                          // max. number of faults we can tolerate
	N              int                          // max.number of validators in the network
	h              uint64                       // low watermark
	id             uint64                       // replica ID; PBFT `i`
	K              uint64                       // checkpoint period
	logMultiplier  uint64                       // use this value to calculate log size : k*logMultiplier
	L              uint64                       // log size
	seqNo          uint64                       // PBFT "n", strictly monotonic increasing sequence number
	view           uint64                       // current view
	nvInitialSeqNo uint64                       // initial seqNo in a new view

	status         status                       // basic status of pbft

	batchMgr       *batchManager                // manage batch related issues
	batchVdr       *batchValidator              // manage batch validate issues
	pbftTimerMgr   *timerManager                // manage pbft event timers
	storeMgr       *storeManager                // manage storage
	nodeMgr        *nodeManager                 // manage node delete or add
	recoveryMgr    *recoveryManager             // manage recovery issues
	vcMgr          *vcManager                   // manage viewchange issues
	exec           *executor                    // manage transaction exec

	helper         helper.Stack
	reqStore       *requestStore                // received messages
	duplicator     map[uint64]*transactionStore // currently executing request

	pbftManager    events.Manager               // manage pbft event

	reqEventQueue  events.Queue                 // transfer request transactions
	pbftEventQueue events.Queue                 // transfer PBFT related event

	config         *common.Config
}

//newPBFT init the PBFT instance
func newPBFT(config *common.Config, h helper.Stack) *pbftImpl {
	pbft := &pbftImpl{}
	pbft.helper = h
	pbft.config = config
	if !config.ContainsKey(common.C_NODE_ID) {
		panic(fmt.Errorf("No hyperchain id specified!, key: %s", common.C_NODE_ID))
	}
	pbft.id = uint64(config.GetInt64(common.C_NODE_ID))
	pbft.N = config.GetInt(PBFT_NODE_NUM)
	pbft.f = (pbft.N - 1) / 3
	pbft.K = uint64(10)
	pbft.logMultiplier = uint64(4)
	pbft.L = pbft.logMultiplier * pbft.K // log size

	//pbftManage manage consensus events
	pbft.pbftManager = events.NewManagerImpl()
	pbft.pbftManager.SetReceiver(pbft)
	pbft.pbftManager.Start()
	pbft.pbftEventQueue = events.GetQueue(pbft.pbftManager.Queue())// init pbftEventQueue

	pbft.initMsgEventMap()

	pbft.exec = newExecutor()
	pbft.pbftTimerMgr = newTimerMgr(pbft)

	pbft.initTimers()
	pbft.initStatus()

	if pbft.pbftTimerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		logger.Infof("PBFT null requests timeout = %v", pbft.pbftTimerMgr.getTimeoutValue(NULL_REQUEST_TIMER))
	} else {
		logger.Infof("PBFT null requests disabled")
	}

	pbft.vcMgr = newVcManager(pbft.pbftTimerMgr, pbft, config)
	// init the data logs
	pbft.storeMgr = newStoreMgr()

	// initialize state transfer
	pbft.nodeMgr = newNodeMgr()
	pbft.duplicator = make(map[uint64]*transactionStore)
	pbft.batchMgr = newBatchManager(config, pbft) 		// init after pbftEventQueue
	pbft.batchVdr = newBatchValidator(pbft)
	pbft.reqStore = newRequestStore()

	atomic.StoreUint32(&pbft.activeView, 1)

	logger.Infof("PBFT Max number of validating peers (N) = %v", pbft.N)
	logger.Infof("PBFT Max number of failing peers (f) = %v", pbft.f)
	logger.Infof("PBFT byzantine flag = %v", pbft.status[BYZANTINE])
	logger.Infof("PBFT request timeout = %v", pbft.pbftTimerMgr.requestTimeout)
	logger.Infof("PBFT Checkpoint period (K) = %v", pbft.K)
	logger.Infof("PBFT Log multiplier = %v", pbft.logMultiplier)
	logger.Infof("PBFT log size (L) = %v", pbft.L)

	return pbft
}

// =============================================================================
// general event process method
// =============================================================================

// ProcessEvent implements event.Receiver
func (pbft *pbftImpl) ProcessEvent(ee events.Event) events.Event {

	switch e := ee.(type) {
	case *types.Transaction: //local transaction
		tx := e
		return pbft.processTxEvent(tx)

	case protos.RemoveCache:
		vid := e.Vid
		ok := pbft.recvRemoveCache(vid)
		if !ok {
			logger.Debugf("Replica %d received local remove cached batch %d, but can not find mapping batch", pbft.id, vid)
		}
		return nil

	case protos.RoutersMessage:
		if len(e.Routers) == 0 {
			logger.Warningf("Replica %d received nil local routers", pbft.id)
			return nil
		}
		logger.Debugf("Replica %d received local routers %s", pbft.id, hashByte(e.Routers))
		pbft.nodeMgr.routers = e.Routers

	case *LocalEvent: //local event
		return pbft.dispatchLocalEvent(e)

	case *ConsensusMessage: //remote message
		next, _ := pbft.msgToEvent(e)
		return pbft.dispatchConsensusMsg(next)

	default:
		logger.Error("Can't recognize event type.")
		return pbft.dispatchConsensusMsg(ee) //TODO: fix this ...
		return nil
	}
	return nil
}

//dispatchCorePbftMsg dispatch core PBFT consensus messages from other peers.
func (pbft *pbftImpl) dispatchCorePbftMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *TransactionBatch:
		err := pbft.recvRequestBatch(et)
		if err != nil {
			logger.Warning(err.Error())
		}
	case *PrePrepare:
		return pbft.recvPrePrepare(et)
	case *Prepare:
		return pbft.recvPrepare(et)
	case *Commit:
		return pbft.recvCommit(et)
	case *Checkpoint:
		return pbft.recvCheckpoint(et)
	}
	return nil
}

// enqueueTx parse transaction msg and send it into request event queue.
func (pbft *pbftImpl) enqueueTx(msg *protos.Message) error {

	// Parse the transaction payload to transaction
	tx := &types.Transaction{}
	err := proto.Unmarshal(msg.Payload, tx)
	if err != nil {
		logger.Errorf("error: %v ,can not unmarshal protos.Message", err)
		return err
	}

	// Post a requestEvent
	go pbft.reqEventQueue.Push(tx)

	return nil
}

// enqueueConsensusMsg parse consensus msg and send it to the corresponding event queue.
func (pbft *pbftImpl) enqueueConsensusMsg(msg *protos.Message) error {
	consensus := &ConsensusMessage{}
	err := proto.Unmarshal(msg.Payload, consensus)
	if err != nil {
		logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
		return err
	}

	if consensus.Type == ConsensusMessage_TRANSACTION {
		tx := &types.Transaction{}
		err := proto.Unmarshal(consensus.Payload, tx)
		if err != nil {
			logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
			return err
		}
		go pbft.reqEventQueue.Push(tx)
	} else {
		go pbft.pbftEventQueue.Push(consensus)
	}
	return nil
}

//processStateUpdated process the state updated message.
func (pbft *pbftImpl) enqueueStateUpdatedMsg(msg *protos.Message) error {

	stateUpdatedMsg := &protos.StateUpdatedMessage{}
	err := proto.Unmarshal(msg.Payload, stateUpdatedMsg)

	if err != nil {
		logger.Errorf("processStateUpdate, unmarshal error: can not unmarshal UpdateStateMessage", err)
		return err
	}
	e := &LocalEvent{
		Service:CORE_PBFT_SERVICE,
		EventType:CORE_STATE_UPDATE_EVENT,
		Event:&stateUpdatedEvent{seqNo:stateUpdatedMsg.SeqNo},
	}
	go pbft.pbftEventQueue.Push(e)
	return nil
}

//=============================================================================
// null request methods
//=============================================================================

// processNullRequest process when a null request come
func (pbft *pbftImpl) processNullRequest(msg *protos.Message) error {
	if pbft.status[IN_NEGO_VIEW] {
		return nil
	}
	if pbft.primary(pbft.view) != pbft.id {
		pbft.pbftTimerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}

	logger.Infof("Replica %d received null request from primary %d", pbft.id, pbft.primary(pbft.view))
	pbft.nullReqTimerReset()
	return nil
}

//handleNullRequestEvent triggered by null request timer
func (pbft *pbftImpl) handleNullRequestTimerEvent() {

	if pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d try to nullRequestHandler, but it's in nego-view", pbft.id)
		return
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		return
	}

	if pbft.primary(pbft.view) != pbft.id {
		// backup expected a null request, but primary never sent one
		logger.Warningf("Replica %d null request timer expired, sending view change", pbft.id)
		pbft.sendViewChange()
	} else {
		logger.Infof("Primary %d null request timer expired, sending null request", pbft.id)
		pbft.sendNullRequest()
	}
}

// sendNullRequest is for primary peer to send null when nullRequestTimer booms
func (pbft *pbftImpl) sendNullRequest() {
	nullRequest := nullRequestMsgToPbMsg(pbft.id)
	pbft.helper.InnerBroadcast(nullRequest)
	pbft.nullReqTimerReset()
}

//=============================================================================
// Preprepare prepare commit methods
//=============================================================================

// trySendPrePrepares send all available PrePrepare messages.
func (pbft *pbftImpl) trySendPrePrepares() {

	if pbft.batchVdr.currentVid != nil {
		logger.Debugf("Replica %d not attempting to send pre-prepare bacause it is currently send %d, retry.", pbft.id, pbft.batchVdr.currentVid)
		return
	}

	logger.Debugf("Replica %d attempting to call sendPrePrepare", pbft.id)

	for stopTry := false; !stopTry;  {
		if find, txBatch, digest := pbft.findNextPrePrepareBatch(); find {
			pbft.sendPrePrepare(txBatch, digest)
			pbft.maybeSendCommit(digest, pbft.view, pbft.seqNo)
		}else {
			stopTry = true
		}
	}
}

//findNextPrePrepareBatch find next validated batch to send preprepare msg.
func (pbft *pbftImpl) findNextPrePrepareBatch() (bool, *TransactionBatch, string) {
	var find bool
	var nextPreprepareBatch *TransactionBatch
	var digest string
	for digest = range pbft.batchVdr.cacheValidatedBatch {
		cache := pbft.batchVdr.getCacheBatchFromCVB(digest)
		if cache == nil {
			logger.Debugf("Primary %d already call sendPrePrepare for batch: %d",
				pbft.batchVdr.pbftId, digest)
			continue
		}

		if cache.vid != pbft.batchVdr.lastVid + 1 {
			logger.Debugf("Primary %d hasn't done with last send pre-prepare, vid=%d",
				pbft.batchVdr.pbftId, pbft.batchVdr.lastVid)
			continue
		}

		currentVid := cache.vid
		pbft.batchVdr.setCurrentVid(&currentVid)

		if len(cache.batch.Batch) == 0 {
			logger.Warningf("Replica %d is primary, receives validated result %s that is empty",
				pbft.id, digest)
			pbft.deleteExistedTx(digest)
			continue
		}

		n := pbft.seqNo + 1

		// check for other PRE-PREPARE for same digest, but different seqNo
		if pbft.storeMgr.existedDigest(n, pbft.view, digest) {
			pbft.deleteExistedTx(digest)
			continue
		}

		if !pbft.inWV(pbft.view, n) {
			logger.Debugf("Replica %d is primary, not sending pre-prepare for request batch %s because " +
				"it is out of sequence numbers", pbft.id, digest)
			continue
		}

		find = true
		nextPreprepareBatch = cache.batch
	}
	return find, nextPreprepareBatch, digest
}

//sendPrePrepare send prepare message.
func (pbft *pbftImpl) sendPrePrepare(reqBatch *TransactionBatch, digest string) {

	logger.Debugf("Replica %d is primary, issuing pre-prepare for request batch %s", pbft.id, digest)

	n := pbft.seqNo + 1

	logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d", pbft.id, pbft.view, n)
	pbft.pbftTimerMgr.stopTimer(NULL_REQUEST_TIMER)
	pbft.seqNo = n

	preprepare := &PrePrepare{
		View:             pbft.view,
		SequenceNumber:   n,
		BatchDigest:      digest,
		TransactionBatch: reqBatch,
		ReplicaId:        pbft.id,
	}
	cert := pbft.storeMgr.getCert(pbft.view, n)
	cert.prePrepare = preprepare
	cert.digest = digest
	cert.sentValidate = true
	cert.validated = true
	pbft.batchVdr.deleteCacheFromCVB(digest)
	pbft.persistQSet(preprepare)

	payload, err := proto.Marshal(preprepare)
	if err != nil {
		logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
		return
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PRE_PREPARE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.batchVdr.updateLCVid()

	pbft.startNewViewTimer(pbft.pbftTimerMgr.requestTimeout, fmt.Sprintf("new request batch view=%d/seqNo=%d, hash=%s", pbft.view, n, digest))

}

//recvPrePrepare process logic for PrePrepare msg.
func (pbft *pbftImpl) recvPrePrepare(preprep *PrePrepare) error {

	logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest=%s ",
		pbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	pbft.stopFirstRequestTimer()

	if !pbft.isPrePrepareLegal(preprep) {
		return nil
	}

	pbft.pbftTimerMgr.stopTimer(NULL_REQUEST_TIMER)
	// add this for recovery, avoid saving batch with seqno that already executed
	if pbft.exec.currentExec != nil {
		if preprep.SequenceNumber <= *pbft.exec.currentExec {
			logger.Debugf("Replica %d reject out-of-date pre-prepare for seqNo=%d/view=%d", pbft.id, preprep.SequenceNumber, preprep.View)
			return nil
		}
	} else {
		if preprep.SequenceNumber <= pbft.exec.lastExec {
			logger.Debugf("Replica %d reject out-of-date pre-prepare for seqNo=%d/view=%d", pbft.id, preprep.SequenceNumber, preprep.View)
			return nil
		}
	}

	cert := pbft.storeMgr.getCert(preprep.View, preprep.SequenceNumber)

	if cert.digest != "" && cert.digest != preprep.BatchDigest {
		logger.Warningf("Pre-prepare found for same view/seqNo but different digest: received %s, stored %s",
			preprep.BatchDigest, cert.digest)
		return nil
	}

	cert.prePrepare = preprep
	cert.digest = preprep.BatchDigest

	// Store the request batch if, for whatever reason, we haven't received it from an earlier broadcast
	if preprep.BatchDigest != "" && !pbft.batchVdr.containsInVBS(preprep.BatchDigest) {
		digest := preprep.BatchDigest
		pbft.batchVdr.saveToVBS(digest, preprep.GetTransactionBatch())
		pbft.storeMgr.outstandingReqBatches[digest] = preprep.GetTransactionBatch()
	}

	if pbft.status.checkStatesAnd(!pbft.status[SKIP_IN_PROGRESS], !pbft.status[IN_RECOVERY]) {
		pbft.startNewViewTimer(pbft.pbftTimerMgr.requestTimeout,
			fmt.Sprintf("new pre-prepare for request batch view=%d/seqNo=%d, hash=%s", preprep.View, preprep.SequenceNumber, preprep.BatchDigest))
	}

	if pbft.primary(pbft.view) != pbft.id && pbft.prePrepared(preprep.BatchDigest, preprep.View, preprep.SequenceNumber) &&
		!cert.sentPrepare {
		cert.sentPrepare = true
		return pbft.sendPrepare(preprep)
	}

	return nil
}

//sendPrepare send prepare message.
func (pbft *pbftImpl) sendPrepare(preprep *PrePrepare) error {
	logger.Debugf("Backup %d broadcasting prepare for view=%d/seqNo=%d", pbft.id, preprep.View, preprep.SequenceNumber)
	prep := &Prepare{
		View:           preprep.View,
		SequenceNumber: preprep.SequenceNumber,
		BatchDigest:    preprep.BatchDigest,
		ReplicaId:      pbft.id,
	}
	pbft.persistQSet(preprep)
	pbft.recvPrepare(prep) // send to itself
	payload, err := proto.Marshal(prep)
	if err != nil {
		logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PREPARE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	return pbft.helper.InnerBroadcast(msg)
}

//recvPrepare process logic after receive prepare message
func (pbft *pbftImpl) recvPrepare(prep *Prepare) error {
	logger.Debugf("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if !pbft.isPrepareLegal(prep) {
		return nil
	}

	cert := pbft.storeMgr.getCert(prep.View, prep.SequenceNumber)

	ok := cert.prepare[*prep]

	if ok {
		logger.Warningf("Ignoring duplicate prepare from replica %d, view=%d/seqNo=%d",
			prep.ReplicaId, prep.View, prep.SequenceNumber)
		return nil
	}

	cert.prepare[*prep] = true
	cert.prepareCount++

	return pbft.maybeSendCommit(prep.BatchDigest, prep.View, prep.SequenceNumber)
}

//maybeSendCommit may send commit msg.
func (pbft *pbftImpl) maybeSendCommit(digest string, v uint64, n uint64) error {

	if !pbft.prepared(digest, v, n) {
		return nil
	}

	if pbft.status[SKIP_IN_PROGRESS] {
		logger.Debugf("Replica %d do not try to validate batch because it's in state update", pbft.id)
		return nil
	}

	cert := pbft.storeMgr.getCert(v, n)

	if ok, _ := pbft.isPrimary(); ok {
		return pbft.sendCommit(digest, v, n)
	} else {
		idx := msgID{v: v, n: n}

		if !cert.sentValidate {
			update, ok := pbft.nodeMgr.updateCertStore[idx]
			if ok && update.validated {
				pbft.batchVdr.vid = pbft.batchVdr.vid + 1
				pbft.batchVdr.lastVid = pbft.batchVdr.vid
				cert.sentValidate = true
				cert.validated = true
				return pbft.sendCommit(digest, v, n)
			}
			pbft.batchVdr.preparedCert[idx] = cert.digest
			pbft.validatePending()
		}
		return nil
	}
}

//sendCommit send commit message.
func (pbft *pbftImpl) sendCommit(digest string, v uint64, n uint64) error {

	cert := pbft.storeMgr.getCert(v, n)

	if !cert.sentCommit {
		logger.Debugf("Replica %d broadcasting commit for view=%d/seqNo=%d", pbft.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			BatchDigest:    digest,
			ReplicaId:      pbft.id,
		}
		cert.sentCommit = true

		pbft.persistPSet(v, n)
		payload, err := proto.Marshal(commit)
		if err != nil {
			logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		go pbft.pbftEventQueue.Push(consensusMsg)
		msg := cMsgToPbMsg(consensusMsg, pbft.id)
		return pbft.helper.InnerBroadcast(msg)
	}

	return nil
}

//recvCommit process logic after receive commit message.
func (pbft *pbftImpl) recvCommit(commit *Commit) error {
	logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !pbft.isCommitLegal(commit) {
		return nil
	}

	cert := pbft.storeMgr.getCert(commit.View, commit.SequenceNumber)

	ok := cert.commit[*commit]

	if ok {
		logger.Warningf("Ignoring duplicate commit from replica %d, view=%d/seqNo=%d",
			commit.ReplicaId, commit.View, commit.SequenceNumber)
		return nil
	}

	cert.commit[*commit] = true
	cert.commitCount++

	if pbft.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) {
		pbft.stopNewViewTimer()
		idx := msgID{v: commit.View, n: commit.SequenceNumber}
		if !cert.sentExecute && cert.validated {

			pbft.vcMgr.lastNewViewTimeout = pbft.pbftTimerMgr.getTimeoutValue(NEW_VIEW_TIMER)
			delete(pbft.storeMgr.outstandingReqBatches, commit.BatchDigest)
			update, ok := pbft.nodeMgr.updateCertStore[idx]
			if ok && update.sentExecute {
				pbft.exec.lastExec = pbft.exec.lastExec + 1
				cert.sentExecute = true
				return nil
			}
			pbft.storeMgr.committedCert[idx] = cert.digest
			pbft.commitTransactions()
			if commit.SequenceNumber == pbft.vcMgr.viewChangeSeqNo {
				logger.Warningf("Replica %d cycling view for seqNo=%d", pbft.id, commit.SequenceNumber)
				pbft.sendViewChange()
			}
		} else {
			logger.Debugf("Replica %d committed for seqNo: %d, but sentExecute: %v, validated: %v", pbft.id, commit.SequenceNumber, cert.sentExecute, cert.validated)
		}
	}

	return nil
}

//=============================================================================
// execute transactions
//=============================================================================

//commitTransactions commit all available transactions
func (pbft *pbftImpl) commitTransactions() {
	if pbft.exec.currentExec != nil {
		logger.Debugf("Replica %d not attempting to commitTransactions bacause it is currently executing %d",
			pbft.id, pbft.exec.currentExec)
	}
	logger.Debugf("Replica %d attempting to commitTransactions", pbft.id)

	for hasTxToExec := true; hasTxToExec; {
		if find, idx, cert := pbft.findNextCommitTx(); find{
			digest := cert.digest
			if digest == "" {
				logger.Infof("Replica %d executing null request for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			} else {
				logger.Noticef("======== Replica %d Call execute, view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
				isPrimary, _ := pbft.isPrimary()
				pbft.helper.Execute(idx.n, digest, true, isPrimary, cert.prePrepare.TransactionBatch.Timestamp)
				pbft.persistCSet(idx.v, idx.n)
			}
			cert.sentExecute = true
			pbft.afterCommitTx(idx)
		}else {
			hasTxToExec = false
		}
	}
	pbft.startTimerIfOutstandingRequests()
}

//findNextCommitTx find next msgID which is able to commit.
func (pbft *pbftImpl) findNextCommitTx() (bool, msgID, *msgCert) {
	var find bool = false
	var nextExecuteMsgId msgID
	var cert *msgCert

	for idx := range pbft.storeMgr.committedCert {
		cert = pbft.storeMgr.certStore[idx]

		if cert == nil || cert.prePrepare == nil {
			logger.Debugf("Replica %d already checkpoint for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			//break
			continue
		}

		// check if already executed
		if cert.sentExecute == true {
			logger.Debugf("Replica %d already execute for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			//break
			continue
		}

		if idx.n != pbft.exec.lastExec + 1 {
			logger.Debugf("Replica %d hasn't done with last execute %d, seq=%d", pbft.id, pbft.exec.lastExec, idx.n)
			//break
			continue
		}

		// skipInProgress == true, then this replica is in viewchange, not reply or execute
		if pbft.status[SKIP_IN_PROGRESS] {
			logger.Warningf("Replica %d currently picking a starting point to resume, will not execute", pbft.id)
			//break
			continue
		}

		digest := cert.digest

		// check if committed
		if !pbft.committed(digest, idx.v, idx.n) {
			//break
			continue
		}

		currentExec := idx.n
		pbft.exec.currentExec = &currentExec

		find = true
		nextExecuteMsgId = idx
		break
	}

	return find, nextExecuteMsgId, cert
}

//afterCommitTx after commit transaction.
func (pbft *pbftImpl) afterCommitTx(idx msgID) {

	if pbft.exec.currentExec != nil {
		logger.Debugf("Replica %d finish execution %d, trying next", pbft.id, *pbft.exec.currentExec)
		pbft.exec.lastExec = *pbft.exec.currentExec
		delete(pbft.storeMgr.committedCert, idx)
		if pbft.status[IN_RECOVERY] {
			if pbft.recoveryMgr.recoveryToSeqNo == nil {
				logger.Errorf("Replica %d in recovery execDoneSync but its recoveryToSeqNo is nil", pbft.id)
				return
			}
			if pbft.exec.lastExec == *pbft.recoveryMgr.recoveryToSeqNo {
				pbft.status.inActiveState(IN_RECOVERY)
				pbft.recoveryMgr.recoveryToSeqNo = nil
				pbft.pbftTimerMgr.stopTimer(RECOVERY_RESTART_TIMER)
				go pbft.pbftEventQueue.Push(&LocalEvent{
					Service:RECOVERY_SERVICE,
					EventType:RECOVERY_DONE_EVENT,
				})
			}
		}
		if pbft.exec.lastExec % pbft.K == 0 {
			bcInfo := getBlockchainInfo()
			height := bcInfo.Height
			if height == pbft.exec.lastExec {
				logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", pbft.exec.lastExec, height)
				//time.Sleep(3*time.Millisecond)
				pbft.checkpoint(pbft.exec.lastExec, bcInfo)
			} else {
				// reqBatch call execute but have not done with execute
				logger.Errorf("Fail to call the checkpoint, seqNo=%d, block height=%d", pbft.exec.lastExec, height)
				//pbft.retryCheckpoint(pbft.lastExec)
			}
		}
	} else {
		logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of data", pbft.id)
		pbft.status.activeState(SKIP_IN_PROGRESS)
	}

	pbft.exec.currentExec = nil
	// optimization: if we are in view changing waiting for executing to target seqNo,
	// one-time processNewView() is enough. No need to processNewView() every time in execDoneSync()
	if atomic.LoadUint32(&pbft.activeView) == 0 && pbft.exec.lastExec == pbft.nvInitialSeqNo {
		pbft.processNewView()
	}
}

//=============================================================================
// process methods
//=============================================================================

//processTxEvent process received transaction event
func (pbft *pbftImpl) processTxEvent(tx *types.Transaction) error {

	if atomic.LoadUint32(&pbft.activeView) == 0 || atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 ||
		pbft.status[IN_NEGO_VIEW] || pbft.status[IN_RECOVERY]  {
		pbft.reqStore.storeOutstanding(tx)
		return nil
	}
	//curr node is not primary
	if ok, currP := pbft.isPrimary(); !ok {
		//Broadcast request to primary
		payload, err := proto.Marshal(tx)
		if err != nil {
			logger.Errorf("C  ConsensusMessage_TRANSACTION Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_TRANSACTION,
			Payload: payload,
		}
		pbMsg := cMsgToPbMsg(consensusMsg, pbft.id)
		pbft.helper.InnerUnicast(pbMsg, currP)
		return nil
	}
	//curr node is primary
	return pbft.primaryProcessTx(tx)
}

//primaryProcessTx primary node use this method to handle transaction
func (pbft *pbftImpl) primaryProcessTx(tx *types.Transaction) error {

	return pbft.recvTransaction(tx)
}

//processRequestsDuringViewChange process requests received during view change.
func (pbft *pbftImpl) processRequestsDuringViewChange() error {
	if atomic.LoadUint32(&pbft.activeView) == 1 && atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 && !pbft.status[IN_RECOVERY] {
		pbft.processCachedTransactions()
	} else {
		logger.Warningf("Replica %d try to processReqDuringViewChange but view change is not finished or it's in recovery / updaingN", pbft.id)
	}
	return nil
}

//processCachedTransactions process cached tx.
func (pbft *pbftImpl) processCachedTransactions() {
	for pbft.reqStore.outstandingRequests.Len() != 0 {
		temp := pbft.reqStore.outstandingRequests.order.Front().Value
		reqc, ok := interface{}(temp).(requestContainer)
		if !ok {
			logger.Error("type assert error:", temp)
			return
		}
		req := reqc.req
		if req != nil {
			go pbft.reqEventQueue.Push(req)
		}
		pbft.reqStore.remove(req)
	}
}

//processRequestsDuringRecovery process requests
func (pbft *pbftImpl) processRequestsDuringRecovery() {
	if !pbft.status[IN_RECOVERY] && atomic.LoadUint32(&pbft.activeView) == 1 && atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 {
		pbft.processCachedTransactions()
	} else {
		logger.Warningf("Replica %d try to processRequestsDuringRecovery but recovery is not finished or it's in viewChange / updatingN", pbft.id)
	}
}

func (pbft *pbftImpl) recvStateUpdatedEvent(et *stateUpdatedEvent) error {

	if pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d try to recvStateUpdatedEvent, but it's in nego-view", pbft.id)
		return nil
	}

	pbft.status.inActiveState(STATE_TRANSFERRING)
	// If state transfer did not complete successfully, or if it did not reach our low watermark, do it again
	if et.seqNo < pbft.h {
		logger.Warningf("Replica %d recovered to seqNo %d but our low watermark has moved to %d", pbft.id, et.seqNo, pbft.h)
		if pbft.storeMgr.highStateTarget == nil {
			logger.Debugf("Replica %d has no state targets, cannot resume state transfer yet", pbft.id)
		} else if et.seqNo < pbft.storeMgr.highStateTarget.seqNo {
			logger.Debugf("Replica %d has state target for %d, transferring", pbft.id, pbft.storeMgr.highStateTarget.seqNo)
			pbft.retryStateTransfer(nil)
		} else {
			logger.Debugf("Replica %d has no state target above %d, highest is %d", pbft.id, et.seqNo, pbft.storeMgr.highStateTarget.seqNo)
		}
		return nil
	}

	logger.Infof("Replica %d application caught up via state transfer, lastExec now %d", pbft.id, et.seqNo)
	// XXX create checkpoint
	pbft.exec.setLastExec(et.seqNo)
	pbft.batchVdr.setVid(et.seqNo)
	pbft.batchVdr.setLastVid(et.seqNo)
	bcInfo := getCurrentBlockInfo()
	id, _ := proto.Marshal(bcInfo)
	pbft.persistCheckpoint(et.seqNo, id)
	pbft.moveWatermarks(pbft.exec.lastExec) // The watermark movement handles moving this to a checkpoint boundary
	pbft.status.inActiveState(SKIP_IN_PROGRESS)
	pbft.validateState()

	if pbft.status[IN_RECOVERY] {
		if pbft.recoveryMgr.recoveryToSeqNo == nil {
			logger.Errorf("Replica %d in recovery recvStateUpdatedEvent but " +
				"its recoveryToSeqNo is nil")
			return nil
		}
		if pbft.exec.lastExec == *pbft.recoveryMgr.recoveryToSeqNo {
			// This is a somewhat subtle situation, we are behind by checkpoint, but others are just on chkpt.
			// Hence, no need to fetch preprepare, prepare, commit
			pbft.status.inActiveState(IN_RECOVERY)
			pbft.recoveryMgr.recoveryToSeqNo = nil
			pbft.pbftTimerMgr.stopTimer(RECOVERY_RESTART_TIMER)
			go pbft.pbftEventQueue.Push(&LocalEvent{
				Service:RECOVERY_SERVICE,
				EventType:RECOVERY_DONE_EVENT,
			})
			return nil
		}

		event := &LocalEvent{
			Service:   RECOVERY_SERVICE,
			EventType: RECOVERY_RESTART_TIMER_EVENT,
		}

		af := func(){
			pbft.pbftEventQueue.Push(event)
		}

		pbft.pbftTimerMgr.startTimer(RECOVERY_RESTART_TIMER, af)

		if pbft.storeMgr.highStateTarget == nil {
			logger.Errorf("Try to fetch QPC, but highStateTarget is nil")
			return nil
		}
		peers := pbft.storeMgr.highStateTarget.replicas
		for idx := range pbft.storeMgr.certStore {
			pbft.persistDelQPCSet(idx.v, idx.n)
		}
		pbft.storeMgr.certStore = make(map[msgID]*msgCert)
		pbft.fetchRecoveryPQC(peers)
		return nil
	}else {
		pbft.executeAfterStateUpdate()
	}

	return nil
}

//recvRequestBatch handle logic after receive request batch
func (pbft *pbftImpl) recvRequestBatch(reqBatch *TransactionBatch) error {

	if pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d try to recvRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}

	digest := hash(reqBatch)
	logger.Debugf("Replica %d received request batch %s", pbft.id, digest)

	if atomic.LoadUint32(&pbft.activeView) == 1 && pbft.primary(pbft.view) == pbft.id {
		pbft.primaryValidateBatch(reqBatch, 0)
	} else {
		logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", pbft.id, digest)
	}

	return nil
}

func (pbft *pbftImpl) executeAfterStateUpdate() {

	logger.Debugf("Replica %d try to execute after state update", pbft.id)

	for idx, cert := range pbft.storeMgr.certStore {
		if idx.n > pbft.seqNo && pbft.prepared(cert.digest, idx.v, idx.n) && !cert.validated {
			logger.Debugf("Replica %d try to vaidate batch %s", pbft.id, cert.digest)
			pbft.batchVdr.preparedCert[idx] = cert.digest
			pbft.validatePending()
		}
	}

}

func (pbft *pbftImpl) checkpoint(n uint64, info *protos.BlockchainInfo) {

	if n % pbft.K != 0 {
		logger.Errorf("Attempted to checkpoint a sequence number (%d) which is not a multiple of the checkpoint interval (%d)", n, pbft.K)
		return
	}

	id, _ := proto.Marshal(info)
	idAsString := byteToString(id)
	seqNo := n

	logger.Infof("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64 id of %s",
		pbft.id, pbft.view, seqNo, idAsString)

	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      pbft.id,
		Id:             idAsString,
	}
	pbft.storeMgr.saveCheckpoint(seqNo, idAsString)

	pbft.persistCheckpoint(seqNo, id)
	pbft.recvCheckpoint(chkpt)
	payload, err := proto.Marshal(chkpt)
	if err != nil {
		logger.Errorf("ConsensusMessage_CHECKPOINT Marshal Error", err)
		return
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_CHECKPOINT,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
}

func (pbft *pbftImpl) recvCheckpoint(chkpt *Checkpoint) events.Event {

	logger.Debugf("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		pbft.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d try to recvCheckpoint, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.weakCheckpointSetOutOfRange(chkpt) {
		return nil
	}

	if !pbft.inW(chkpt.SequenceNumber) {
		if chkpt.SequenceNumber != pbft.h && !pbft.status[SKIP_IN_PROGRESS] {
			// It is perfectly normal that we receive checkpoints for the watermark we just raised, as we raise it after 2f+1, leaving f replies left
			logger.Warningf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		} else {
			logger.Debugf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		}
		return nil
	}

	cert := pbft.storeMgr.getChkptCert(chkpt.SequenceNumber, chkpt.Id)
	ok := cert.chkpts[*chkpt]

	if ok {
		logger.Warningf("Ignoring duplicate checkpoint from replica %d, seqNo=%d", chkpt.ReplicaId, chkpt.SequenceNumber)
		return nil
	}

	cert.chkpts[*chkpt] = true
	cert.chkptCount++
	pbft.storeMgr.checkpointStore[*chkpt] = true

	logger.Debugf("Replica %d found %d matching checkpoints for seqNo %d, digest %s",
		pbft.id, cert.chkptCount, chkpt.SequenceNumber, chkpt.Id)

	if cert.chkptCount == pbft.f + 1 {
		// We do have a weak cert
		pbft.witnessCheckpointWeakCert(chkpt)
	}

	if cert.chkptCount < pbft.intersectionQuorum() {
		// We do not have a quorum yet
		return nil
	}

	// It is actually just fine if we do not have this checkpoint
	// and should not trigger a state transfer
	// Imagine we are executing sequence number k-1 and we are slow for some reason
	// then everyone else finishes executing k, and we receive a checkpoint quorum
	// which we will agree with very shortly, but do not move our watermarks until
	// we have reached this checkpoint
	// Note, this is not divergent from the paper, as the paper requires that
	// the quorum certificate must contain 2f+1 messages, including its own

	chkptID, ok := pbft.storeMgr.chkpts[chkpt.SequenceNumber]
	if !ok {
		logger.Debugf("Replica %d found checkpoint quorum for seqNo %d, digest %s, but it has not reached this checkpoint itself yet",
			pbft.id, chkpt.SequenceNumber, chkpt.Id)
		if pbft.status[SKIP_IN_PROGRESS] {
			logSafetyBound := pbft.h + pbft.L / 2
			// As an optimization, if we are more than half way out of our log and in state transfer, move our watermarks so we don't lose track of the network
			// if needed, state transfer will restart on completion to a more recent point in time
			if chkpt.SequenceNumber >= logSafetyBound {
				logger.Debugf("Replica %d is in state transfer, but, the network seems to be moving on past %d, moving our watermarks to stay with it", pbft.id, logSafetyBound)
				pbft.moveWatermarks(chkpt.SequenceNumber)
			}
		}
		return nil
	}

	logger.Infof("Replica %d found checkpoint quorum for seqNo %d, digest %s",
		pbft.id, chkpt.SequenceNumber, chkpt.Id)

	if chkptID != chkpt.Id {
		logger.Criticalf("Replica %d generated a checkpoint of %s, but a quorum of the network agrees on %s. This is almost definitely non-deterministic chaincode.",
			pbft.id, chkptID, chkpt.Id)
		pbft.stateTransfer(nil)
	}

	pbft.moveWatermarks(chkpt.SequenceNumber)

	return nil
}

// used in view-change to fetch missing assigned, non-checkpointed requests
func (pbft *pbftImpl) fetchRequestBatches() (error) {

	for digest := range pbft.storeMgr.missingReqBatches {
		frb := &FetchRequestBatch{
			BatchDigest: digest,
			ReplicaId:   pbft.id,
		}
		payload, err := proto.Marshal(frb)
		if err != nil {
			logger.Errorf("ConsensusMessage_FRTCH_REQUEST_BATCH Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_FRTCH_REQUEST_BATCH,
			Payload: payload,
		}
		msg := cMsgToPbMsg(consensusMsg, pbft.id)
		pbft.helper.InnerBroadcast(msg)
	}

	return nil
}

func (pbft *pbftImpl) weakCheckpointSetOutOfRange(chkpt *Checkpoint) bool {
	H := pbft.h + pbft.L

	// Track the last observed checkpoint sequence number if it exceeds our high watermark, keyed by replica to prevent unbounded growth
	if chkpt.SequenceNumber < H {
		// For non-byzantine nodes, the checkpoint sequence number increases monotonically
		delete(pbft.storeMgr.hChkpts, chkpt.ReplicaId)
	} else {
		// We do not track the highest one, as a byzantine node could pick an arbitrarilly high sequence number
		// and even if it recovered to be non-byzantine, we would still believe it to be far ahead
		pbft.storeMgr.hChkpts[chkpt.ReplicaId] = chkpt.SequenceNumber

		// If f+1 other replicas have reported checkpoints that were (at one time) outside our watermarks
		// we need to check to see if we have fallen behind.
		if len(pbft.storeMgr.hChkpts) >= pbft.f + 1 {
			chkptSeqNumArray := make([]uint64, len(pbft.storeMgr.hChkpts))
			index := 0
			for replicaID, hChkpt := range pbft.storeMgr.hChkpts {
				chkptSeqNumArray[index] = hChkpt
				index++
				if hChkpt < H {
					delete(pbft.storeMgr.hChkpts, replicaID)
				}
			}
			sort.Sort(sortableUint64Slice(chkptSeqNumArray))

			// If f+1 nodes have issued checkpoints above our high water mark, then
			// we will never record 2f+1 checkpoints for that sequence number, we are out of date
			// (This is because all_replicas - missed - me = 3f+1 - f - 1 = 2f)
			if m := chkptSeqNumArray[len(chkptSeqNumArray) - (pbft.f + 1)]; m > H {
				logger.Warningf("Replica %d is out of date, f+1 nodes agree checkpoint with seqNo %d exists but our high water mark is %d", pbft.id, chkpt.SequenceNumber, H)
				// Discard all our requests, as we will never know which were executed, to be addressed in #394
				pbft.batchVdr.emptyVBS()
				pbft.moveWatermarks(m)
				pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
				pbft.status.activeState(SKIP_IN_PROGRESS)
				pbft.invalidateState()
				pbft.stopNewViewTimer()

				// TODO, reprocess the already gathered checkpoints, this will make recovery faster, though it is presently correct
				return true
			}
		}
	}

	return false
}

func (pbft *pbftImpl) witnessCheckpointWeakCert(chkpt *Checkpoint) {

	// Only ever invoked for the first weak cert, so guaranteed to be f+1
	checkpointMembers := make([]uint64, pbft.f + 1)
	i := 0
	for testChkpt := range pbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			checkpointMembers[i] = testChkpt.ReplicaId
			logger.Debugf("Replica %d adding replica %d (handle %v) to weak cert", pbft.id, testChkpt.ReplicaId, checkpointMembers[i])
			i++
		}
	}

	snapshotID, err := base64.StdEncoding.DecodeString(chkpt.Id)
	if err != nil {
		err = fmt.Errorf("Replica %d received a weak checkpoint cert which could not be decoded (%s)", pbft.id, chkpt.Id)
		logger.Error(err.Error())
		return
	}

	target := &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: chkpt.SequenceNumber,
			id:    snapshotID,
		},
		replicas: checkpointMembers,
	}
	pbft.updateHighStateTarget(target)

	if pbft.status[SKIP_IN_PROGRESS] {
		logger.Infof("Replica %d is catching up and witnessed a weak certificate for checkpoint %d, weak cert attested to by %d of %d (%v)",
			pbft.id, chkpt.SequenceNumber, i, pbft.N, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		pbft.retryStateTransfer(target)
	}
}

func (pbft *pbftImpl) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / pbft.K * pbft.K

	if pbft.h > n {
		logger.Critical("Replica %d movewatermark but pbft.h>n", pbft.id)
		return
	}

	for idx, cert := range pbft.storeMgr.certStore {
		if idx.n <= h {
			logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				pbft.id, idx.v, idx.n)
			pbft.batchVdr.deleteTxFromVBS(cert.digest)
			delete(pbft.storeMgr.outstandingReqBatches, cert.digest)
			delete(pbft.storeMgr.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n)
			delete(pbft.nodeMgr.updateCertStore, idx)
		}
	}

	for testChkpt := range pbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber <= h {
			logger.Debugf("Replica %d cleaning checkpoint message from replica %d, seqNo %d, b64 snapshot id %s",
				pbft.id, testChkpt.ReplicaId, testChkpt.SequenceNumber, testChkpt.Id)
			delete(pbft.storeMgr.checkpointStore, testChkpt)
		}
	}

	for cid := range pbft.storeMgr.chkptCertStore {
		if cid.n <= h {
			logger.Debugf("Replica %d cleaning checkpoint message, seqNo %d, b64 snapshot id %s",
				pbft.id, cid.n, cid.id)
			delete(pbft.storeMgr.chkptCertStore, cid)
		}
	}

	pbft.storeMgr.moveWatermarks(pbft, h)

	pbft.h = h

	logger.Infof("Replica %d updated low watermark to %d",
		pbft.id, pbft.h)

	pbft.resubmitRequestBatches()
}

func (pbft *pbftImpl) updateHighStateTarget(target *stateUpdateTarget) {
	if pbft.storeMgr.highStateTarget != nil && pbft.storeMgr.highStateTarget.seqNo >= target.seqNo {
		logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d",
			pbft.id, target.seqNo, pbft.storeMgr.highStateTarget.seqNo)
		return
	}

	pbft.storeMgr.highStateTarget = target
}

func (pbft *pbftImpl) stateTransfer(optional *stateUpdateTarget) {

	if !pbft.status[SKIP_IN_PROGRESS] {
		logger.Debugf("Replica %d is out of sync, pending state transfer", pbft.id)
		pbft.status.activeState(SKIP_IN_PROGRESS)
		pbft.invalidateState()
	}

	pbft.retryStateTransfer(optional)
}

func (pbft *pbftImpl) retryStateTransfer(optional *stateUpdateTarget) {

	if pbft.status[STATE_TRANSFERRING] {
		logger.Debugf("Replica %d is currently mid state transfer, it must wait for this state transfer to complete before initiating a new one", pbft.id)
		return
	}

	target := optional
	if target == nil {
		if pbft.storeMgr.highStateTarget == nil {
			logger.Debugf("Replica %d has no targets to attempt state transfer to, delaying", pbft.id)
			return
		}
		target = pbft.storeMgr.highStateTarget
	}

	pbft.status.activeState(STATE_TRANSFERRING)

	logger.Infof("Replica %d is initiating state transfer to seqNo %d", pbft.id, target.seqNo)

	//pbft.batch.pbftManager.Queue() <- stateUpdateEvent // Todo for stateupdate
	//pbft.consumer.skipTo(target.seqNo, target.id, target.replicas)

	pbft.skipTo(target.seqNo, target.id, target.replicas)

}

func (pbft *pbftImpl) resubmitRequestBatches() {
	if pbft.primary(pbft.view) != pbft.id {
		return
	}

	var submissionOrder []*TransactionBatch

	outer:
	for d, reqBatch := range pbft.storeMgr.outstandingReqBatches {
		for _, cert := range pbft.storeMgr.certStore {
			if cert.digest == d {
				logger.Debugf("Replica %d already has certificate for request batch %s - not going to resubmit", pbft.id, d)
				continue outer
			}
		}
		logger.Infof("Replica %d has detected request batch %s must be resubmitted", pbft.id, d)
		submissionOrder = append(submissionOrder, reqBatch)
	}

	if len(submissionOrder) == 0 {
		return
	}

	for _, reqBatch := range submissionOrder {
		// This is a request batch that has not been pre-prepared yet
		// Trigger request batch processing again
		pbft.recvRequestBatch(reqBatch)
	}
}

func (pbft *pbftImpl) skipTo(seqNo uint64, id []byte, replicas []uint64) {
	info := &protos.BlockchainInfo{}
	err := proto.Unmarshal(id, info)
	if err != nil {
		logger.Error(fmt.Sprintf("Error unmarshaling: %s", err))
		return
	}
	//pbft.UpdateState(&checkpointMessage{seqNo, id}, info, replicas)
	logger.Debug("seqNo: ", seqNo, "id: ", id, "replicas: ", replicas)
	pbft.updateState(seqNo, id, replicas)
}

// updateState attempts to synchronize state to a particular target, implicitly calls rollback if needed
func (pbft *pbftImpl) updateState(seqNo uint64, targetId []byte, replicaId []uint64) {
	//if pbft.valid {
	//	logger.Warning("State transfer is being called for, but the state has not been invalidated")
	//}

	updateStateMsg := stateUpdateHelper(pbft.id, seqNo, targetId, replicaId)
	pbft.helper.UpdateState(updateStateMsg) // TODO: stateUpdateEvent

}

func (pbft *pbftImpl) processNegotiateView() error {
	if !pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d try to negotiateView, but it's not inNegoView. This indicates a bug", pbft.id)
		return nil
	}

	logger.Debugf("Replica %d now negotiate view...", pbft.id)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT,
	}

	af := func(){
		pbft.pbftEventQueue.Push(event)
	}

	pbft.pbftTimerMgr.startTimer(NEGO_VIEW_RSP_TIMER, af)

	pbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)

	// broadcast the negotiate message to other replica
	negoViewMsg := &NegotiateView{
		ReplicaId: pbft.id,
	}
	payload, err := proto.Marshal(negoViewMsg)
	if err != nil {
		logger.Errorf("Marshal NegotiateView Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	logger.Debugf("Replica %d broadcast negociate view message", pbft.id)

	// post the negotiate message event to myself
	nvr := &NegotiateViewResponse{
		ReplicaId: pbft.id,
		View:      pbft.view,
		N:         uint64(pbft.N),
		Routers:   pbft.nodeMgr.routers,
	}
	consensusPayload, err := proto.Marshal(nvr)
	if err != nil {
		logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	responseMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: consensusPayload,
	}
	go pbft.pbftEventQueue.Push(responseMsg)

	return nil
}

func (pbft *pbftImpl) recvNegoView(nv *NegotiateView) events.Event {
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		return nil
	}
	sender := nv.ReplicaId
	logger.Debugf("Replica %d receive negotiate view from %d", pbft.id, sender)

	if pbft.nodeMgr.routers == nil {
		logger.Debugf("Replica %d ignore negotiate view from %d since has not received local msg", pbft.id, sender)
		return nil
	}

	negoViewRsp := &NegotiateViewResponse{
		ReplicaId: pbft.id,
		View:      pbft.view,
		N:         uint64(pbft.N),
		Routers:   pbft.nodeMgr.routers,
	}
	payload, err := proto.Marshal(negoViewRsp)
	if err != nil {
		logger.Errorf("Marshal NegotiateViewResponse Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW_RESPONSE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerUnicast(msg, sender)
	return nil
}

func (pbft *pbftImpl) recvNegoViewRsp(nvr *NegotiateViewResponse) events.Event {
	if !pbft.status[IN_NEGO_VIEW] {
		logger.Debugf("Replica %d already finished nego-view, ignore incoming nego-view response", pbft.id)
		return nil
	}

	//rspId, rspView := nvr.ReplicaId, nvr.View
	if _, ok := pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId]; ok {
		logger.Warningf("Already recv view number from %d, ignore it", nvr.ReplicaId)
		return nil
	}

	logger.Debugf("Replica %d receive nego-view response from %d, view: %d, N: %d", pbft.id, nvr.ReplicaId, nvr.View, nvr.N)

	pbft.recoveryMgr.negoViewRspStore[nvr.ReplicaId] = nvr

	if len(pbft.recoveryMgr.negoViewRspStore) > 2*pbft.f+1 {
		// Reason for not using '> pbft.N-pbft.f': if N==5, we are require more than we need
		// Reason for not using '≥ pbft.N-pbft.f': if self is wrong, then we are impossible to find 2f+1 same view
		// can we find same view from 2f+1 peers?
		type resp struct {
			n uint64
			view uint64
			routers string
		}
		viewCount := make(map[resp]uint64)
		var result resp
		canFind := false
		for _, rs := range pbft.recoveryMgr.negoViewRspStore {
			r := byteToString(rs.Routers)
			ret := resp{rs.N, rs.View, r}
			if _, ok := viewCount[ret]; ok {
				viewCount[ret]++
			} else {
				viewCount[ret] = uint64(1)
			}
			if viewCount[ret] >= uint64(2*pbft.f+1) {
				// yes we find the view
				result = ret
				canFind = true
				break
			}
		}

		if canFind {
			pbft.pbftTimerMgr.stopTimer(NEGO_VIEW_RSP_TIMER)
			pbft.view = result.view
			pbft.N = int(result.n)
			routers, _ := stringToByte(result.routers)
			if !bytes.Equal(routers, pbft.nodeMgr.routers) && !pbft.status[IS_NEW_NODE] {
				pbft.nodeMgr.routers = routers
				logger.Debugf("Replica %d update routing table according to nego result", pbft.id)
				pbft.helper.NegoRouters(routers)
			}
			pbft.status.inActiveState(IN_NEGO_VIEW)
			if atomic.LoadUint32(&pbft.activeView) == 0 {
				atomic.StoreUint32(&pbft.activeView, 1)
			}
			return &LocalEvent{
				Service:RECOVERY_SERVICE,
				EventType:RECOVERY_NEGO_VIEW_DONE_EVENT,
			}
		} else if len(pbft.recoveryMgr.negoViewRspStore) >= 2*pbft.f+2 {
			event := &LocalEvent{
				Service:   RECOVERY_SERVICE,
				EventType: RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT,
			}

			af := func(){
				pbft.pbftEventQueue.Push(event)
			}

			pbft.pbftTimerMgr.startTimer(NEGO_VIEW_RSP_TIMER, af)

			logger.Warningf("pbft recv at least N-f nego-view responses, but cannot find same view from 2f+1.")
		}
	}
	return nil
}

func (pbft *pbftImpl) restartNegoView() {
	logger.Debugf("Replica %d restart negotiate view", pbft.id)
	pbft.processNegotiateView()
}

// =============================================================================
// receive local message methods
// =============================================================================
func (pbft *pbftImpl) recvValidatedResult(result protos.ValidatedTxs) error {

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		logger.Debugf("Primary %d received validated batch for view=%d/vid=%d, batch hash: %s", pbft.id, result.View, result.SeqNo, result.Hash)

		if atomic.LoadUint32(&pbft.activeView) == 0 {
			logger.Debugf("Replica %d ignoring ValidatedResult as we sre in view change", pbft.id)
			return nil
		}

		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
			logger.Debugf("Replica %d ignoring ValidatedResult as we sre in updating N", pbft.id)
			return nil
		}

		if !pbft.inWV(result.View, result.SeqNo) {
			logger.Debugf("Replica %d receives validated result %s that is out of sequence numbers", pbft.id, result.Hash)
			return nil
		}

		batch := &TransactionBatch{
			Batch:     result.Transactions,
			Timestamp: result.Timestamp,
		}
		digest := result.Hash
		pbft.batchVdr.saveToVBS(digest, batch)

		pbft.storeMgr.outstandingReqBatches[digest] = batch
		cache := &cacheBatch{
			batch: batch,
			vid:   result.SeqNo,
		}
		pbft.batchVdr.saveToCVB(digest, cache)
		pbft.trySendPrePrepares()
	} else {
		logger.Debugf("Replica %d recived validated batch for view=%d/sqeNo=%d, batch hash: %s", pbft.id, result.View, result.SeqNo, result.Hash)

		if !pbft.inWV(result.View, result.SeqNo) {
			logger.Debugf("Replica %d receives validated result %s that is out of sequence numbers", pbft.id, result.Hash)
			return nil
		}

		cert := pbft.storeMgr.getCert(result.View, result.SeqNo)

		digest := result.Hash
		if digest == cert.digest {
			cert.validated = true
			pbft.sendCommit(digest, result.View, result.SeqNo)
		} else {
			logger.Warningf("Relica %d cannot agree with the validate result for view=%d/seqNo=%d sent from primary, self: %s, primary: %s", pbft.id, result.View, result.SeqNo, result.Hash, cert.digest)
			pbft.sendViewChange()
		}
	}
	return nil
}

func (pbft *pbftImpl) recvRemoveCache(vid uint64) bool {

	if vid <= 10 {
		logger.Debugf("Replica %d received remove cached batch %d <= 10, retain it until 11", pbft.id, vid)
		return true
	}
	id := vid - 10
	_, ok := pbft.duplicator[id]
	if ok {
		logger.Debugf("Replica %d received remove cached batch %d, and remove batch %d", pbft.id, vid, id)
		delete(pbft.duplicator, id)
	}

	return ok
}
