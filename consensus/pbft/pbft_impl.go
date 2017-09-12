//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"hyperchain/common"
	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/helper/persist"
	"hyperchain/consensus/txpool"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
)

// batch is used to construct reqbatch, the middle layer between outer to pbft
type pbftImpl struct {
	namespace      string
	activeView     uint32 // view change happening
	f              int    // max. number of faults we can tolerate
	N              int    // max.number of validators in the network
	h              uint64 // low watermark
	id             uint64 // replica ID; PBFT `i`
	K              uint64 // checkpoint period
	logMultiplier  uint64 // use this value to calculate log size : k*logMultiplier
	L              uint64 // log size
	seqNo          uint64 // PBFT "n", strictly monotonic increasing sequence number
	view           uint64 // current view
	nvInitialSeqNo uint64 // initial seqNo in a new view
	cachetx        int

	status PbftStatus // basic status of pbft

	batchMgr    *batchManager    // manage batch related issues
	batchVdr    *batchValidator  // manage batch validate issues
	timerMgr    *timerManager    // manage pbft event timers
	storeMgr    *storeManager    // manage storage
	nodeMgr     *nodeManager     // manage node delete or add
	recoveryMgr *recoveryManager // manage recovery issues
	vcMgr       *vcManager       // manage viewchange issues
	exec        *executor        // manage transaction exec

	helper helper.Stack
	//reqStore       *requestStore                // received messages

	pbftManager    events.Manager // manage pbft event
	pbftEventQueue events.Queue   // transfer PBFT related event

	config *common.Config
	logger *logging.Logger

	normal   uint32
	poolFull uint32
}

//newPBFT init the PBFT instance
func newPBFT(namespace string, config *common.Config, h helper.Stack, n int) (*pbftImpl, error) {
	var err error
	pbft := &pbftImpl{}
	pbft.logger = common.GetLogger(namespace, "consensus")
	pbft.namespace = namespace
	pbft.helper = h
	pbft.config = config
	if !config.ContainsKey(common.C_NODE_ID) {
		err = fmt.Errorf("No hyperchain id specified!, key: %s", common.C_NODE_ID)
		return nil, err
	}
	pbft.id = uint64(config.GetInt64(common.C_NODE_ID))
	pbft.N = n
	pbft.f = (pbft.N - 1) / 3
	pbft.K = uint64(10)
	pbft.logMultiplier = uint64(4)
	pbft.L = pbft.logMultiplier * pbft.K // log size
	pbft.cachetx = config.GetInt("consensus.pbft.cachetx")
	pbft.logger.Noticef("Replica %d set cachetx %d", pbft.id, pbft.cachetx)

	//pbftManage manage consensus events
	pbft.pbftManager = events.NewManagerImpl(pbft.namespace)
	pbft.pbftManager.SetReceiver(pbft)

	pbft.initMsgEventMap()

	// new executor
	pbft.exec = newExecutor()
	//new timer manager
	pbft.timerMgr = newTimerMgr(pbft)

	pbft.initTimers()
	pbft.initStatus()

	if pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		pbft.logger.Infof("PBFT null requests timeout = %v", pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER))
	} else {
		pbft.logger.Infof("PBFT null requests disabled")
	}

	pbft.vcMgr = newVcManager(pbft)
	// init the data logs
	pbft.storeMgr = newStoreMgr()
	pbft.storeMgr.logger = pbft.logger

	// initialize state transfer
	pbft.nodeMgr = newNodeMgr()

	pbft.batchMgr = newBatchManager(pbft) // init after pbftEventQueue
	// new batch manager
	pbft.batchVdr = newBatchValidator()
	//pbft.reqStore = newRequestStore()
	pbft.recoveryMgr = newRecoveryMgr()

	atomic.StoreUint32(&pbft.activeView, 1)

	pbft.logger.Infof("PBFT Max number of validating peers (N) = %v", pbft.N)
	pbft.logger.Infof("PBFT Max number of failing peers (f) = %v", pbft.f)
	pbft.logger.Infof("PBFT byzantine flag = %v", pbft.status.getState(&pbft.status.byzantine))
	pbft.logger.Infof("PBFT request timeout = %v", pbft.timerMgr.requestTimeout)
	pbft.logger.Infof("PBFT Checkpoint period (K) = %v", pbft.K)
	pbft.logger.Infof("PBFT Log multiplier = %v", pbft.logMultiplier)
	pbft.logger.Infof("PBFT log size (L) = %v", pbft.L)

	atomic.StoreUint32(&pbft.normal, 1)
	atomic.StoreUint32(&pbft.poolFull, 0)

	return pbft, nil
}

// =============================================================================
// general event process method
// =============================================================================

// ProcessEvent implements event.Receiver
func (pbft *pbftImpl) ProcessEvent(ee events.Event) events.Event {

	switch e := ee.(type) {
	case *types.Transaction: //local transaction
		return pbft.processTransaction(e)

	case txpool.TxHashBatch:
		err := pbft.recvRequestBatch(e)
		if err != nil {
			pbft.logger.Warning(err.Error())
		}

	case protos.RoutersMessage:
		if len(e.Routers) == 0 {
			pbft.logger.Warningf("Replica %d received nil local routers", pbft.id)
			return nil
		}
		pbft.logger.Debugf("Replica %d received local routers %s", pbft.id, hashByte(e.Routers))
		pbft.nodeMgr.routers = e.Routers

	case *LocalEvent: //local event
		return pbft.dispatchLocalEvent(e)

	case *ConsensusMessage: //remote message
		next, _ := pbft.msgToEvent(e)
		return pbft.dispatchConsensusMsg(next)

	default:
		pbft.logger.Errorf("Can't recognize event type of %v.", e)
		return nil
	}
	return nil
}

//dispatchCorePbftMsg dispatch core PBFT consensus messages from other peers.
func (pbft *pbftImpl) dispatchCorePbftMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *PrePrepare:
		return pbft.recvPrePrepare(et)
	case *Prepare:
		return pbft.recvPrepare(et)
	case *Commit:
		return pbft.recvCommit(et)
	case *Checkpoint:
		return pbft.recvCheckpoint(et)
	case *FetchMissingTransaction:
		return pbft.recvFetchMissingTransaction(et)
	case *ReturnMissingTransaction:
		return pbft.recvReturnMissingTransaction(et)
	}
	return nil
}

// enqueueConsensusMsg parse consensus msg and send it to the corresponding event queue.
func (pbft *pbftImpl) enqueueConsensusMsg(msg *protos.Message) error {
	consensus := &ConsensusMessage{}
	err := proto.Unmarshal(msg.Payload, consensus)
	if err != nil {
		pbft.logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
		return err
	}

	if consensus.Type == ConsensusMessage_TRANSACTION {
		tx := &types.Transaction{}
		err := proto.Unmarshal(consensus.Payload, tx)
		if err != nil {
			pbft.logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
			return err
		}
		go pbft.pbftEventQueue.Push(tx)
	} else {
		go pbft.pbftEventQueue.Push(consensus)
	}

	return nil
}

//=============================================================================
// null request methods
//=============================================================================

// processNullRequest process when a null request come
func (pbft *pbftImpl) processNullRequest(msg *protos.Message) error {
	if pbft.status.getState(&pbft.status.inNegoView) {
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Warningf("Replica %d is in viewchange, reject null request from replica %d", pbft.id, msg.Id)
		return nil
	}

	if pbft.primary(pbft.view) != msg.Id {
		pbft.logger.Warningf("Replica %d received null request from replica %d who is not primary", pbft.id, msg.Id)
		return nil
	}

	if pbft.primary(pbft.view) != pbft.id {
		pbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}

	pbft.logger.Infof("Replica %d received null request from primary %d", pbft.id, msg.Id)
	pbft.nullReqTimerReset()
	return nil
}

//handleNullRequestEvent triggered by null request timer
func (pbft *pbftImpl) handleNullRequestTimerEvent() {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to nullRequestHandler, but it's in nego-view", pbft.id)
		return
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		return
	}

	if pbft.primary(pbft.view) != pbft.id {
		// backup expected a null request, but primary never sent one
		pbft.logger.Warningf("Replica %d null request timer expired, sending view change", pbft.id)
		pbft.sendViewChange()
	} else {
		pbft.logger.Infof("Primary %d null request timer expired, sending null request", pbft.id)
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
func (pbft *pbftImpl) sendPendingPrePrepares() {

	if pbft.batchVdr.currentVid != nil {
		pbft.logger.Debugf("Replica %d not attempting to send pre-prepare bacause it is currently send %d, retry.", pbft.id, *pbft.batchVdr.currentVid)
		return
	}

	pbft.logger.Debugf("Replica %d attempting to call sendPrePrepare", pbft.id)

	for stop := false; !stop; {
		if find, digest, resultHash := pbft.findNextPrePrepareBatch(); find {
			waitingBatch := pbft.storeMgr.outstandingReqBatches[digest]
			pbft.sendPrePrepare(*pbft.batchVdr.currentVid, digest, resultHash, waitingBatch)
		} else {
			stop = true
		}
	}
}

//findNextPrePrepareBatch find next validated batch to send preprepare msg.
func (pbft *pbftImpl) findNextPrePrepareBatch() (find bool, digest string, resultHash string) {

	for digest = range pbft.batchVdr.cacheValidatedBatch {
		cache := pbft.batchVdr.getCacheBatchFromCVB(digest)
		if cache == nil {
			pbft.logger.Debugf("Primary %d already call sendPrePrepare for batch: %s",
				pbft.id, digest)
			continue
		}

		if cache.seqNo != pbft.batchVdr.lastVid+1 {
			pbft.logger.Debugf("Primary %d expect to send pre-prepare for seqNo=%d, not seqNo=%d", pbft.id, pbft.batchVdr.lastVid+1, cache.seqNo)
			continue
		}

		currentVid := cache.seqNo
		pbft.batchVdr.setCurrentVid(&currentVid)

		n := currentVid + 1
		// check for other PRE-PREPARE for same digest, but different seqNo
		if pbft.storeMgr.existedDigest(n, pbft.view, digest) {
			pbft.deleteExistedTx(digest)
			continue
		}

		if !pbft.sendInWV(pbft.view, n) {
			pbft.logger.Debugf("Replica %d is primary, not sending pre-prepare for request batch %s because "+
				"batch seqNo=%d is out of sequence numbers", pbft.id, digest, n)
			pbft.batchVdr.currentVid = nil
			//ogger.Debugf("Replica %d broadcast FinishUpdate", pbft.id)
			break
		}

		find = true
		resultHash = cache.resultHash
		break
	}
	return
}

//sendPrePrepare send prepare message.
func (pbft *pbftImpl) sendPrePrepare(seqNo uint64, digest string, hash string, reqBatch *TransactionBatch) {

	pbft.logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d/digest=%s",
		pbft.id, pbft.view, seqNo, *pbft.batchVdr.currentVid, digest)

	hashBatch := &HashBatch{
		List:      reqBatch.HashList,
		Timestamp: reqBatch.Timestamp,
	}

	preprepare := &PrePrepare{
		View:           pbft.view,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ResultHash:     hash,
		HashBatch:      hashBatch,
		ReplicaId:      pbft.id,
	}

	cert := pbft.storeMgr.getCert(pbft.view, seqNo, digest)
	cert.prePrepare = preprepare
	cert.resultHash = hash
	cert.sentValidate = true
	cert.validated = true
	pbft.batchVdr.deleteCacheFromCVB(digest)
	pbft.persistQSet(preprepare)

	reqBatch.SeqNo = seqNo
	reqBatch.ResultHash = hash
	pbft.storeMgr.outstandingReqBatches[digest] = reqBatch
	pbft.storeMgr.txBatchStore[digest] = reqBatch
	pbft.persistTxBatch(digest)

	payload, err := proto.Marshal(preprepare)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
		pbft.batchVdr.updateLCVid()
		return
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PRE_PREPARE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.batchVdr.updateLCVid()
}

//recvPrePrepare process logic for PrePrepare msg.
func (pbft *pbftImpl) recvPrePrepare(preprep *PrePrepare) error {

	pbft.logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest=%s ",
		pbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if !pbft.isPrePrepareLegal(preprep) {
		return nil
	}

	if preprep.SequenceNumber > pbft.exec.lastExec {
		pbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)
		pbft.stopFirstRequestTimer()
	}

	cert := pbft.storeMgr.getCert(preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	cert.prePrepare = preprep
	cert.resultHash = preprep.ResultHash

	if !pbft.status.checkStatesOr(&pbft.status.skipInProgress, &pbft.status.inRecovery) &&
		preprep.SequenceNumber > pbft.exec.lastExec {
		pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout,
			fmt.Sprintf("new pre-prepare for request batch view=%d/seqNo=%d, hash=%s",
				preprep.View, preprep.SequenceNumber, preprep.BatchDigest))
	}

	pbft.persistQSet(preprep)

	if pbft.primary(pbft.view) != pbft.id && pbft.prePrepared(preprep.BatchDigest, preprep.View, preprep.SequenceNumber) &&
		!cert.sentPrepare {
		cert.sentPrepare = true
		return pbft.sendPrepare(preprep)
	}

	return nil
}

//sendPrepare send prepare message.
func (pbft *pbftImpl) sendPrepare(preprep *PrePrepare) error {
	pbft.logger.Debugf("Backup %d broadcasting prepare for view=%d/seqNo=%d", pbft.id, preprep.View, preprep.SequenceNumber)
	prep := &Prepare{
		View:           preprep.View,
		SequenceNumber: preprep.SequenceNumber,
		BatchDigest:    preprep.BatchDigest,
		ResultHash:     preprep.ResultHash,
		ReplicaId:      pbft.id,
	}
	pbft.recvPrepare(prep) // send to itself
	payload, err := proto.Marshal(prep)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
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
	pbft.logger.Debugf("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if !pbft.isPrepareLegal(prep) {
		return nil
	}

	cert := pbft.storeMgr.getCert(prep.View, prep.SequenceNumber, prep.BatchDigest)

	ok := cert.prepare[*prep]

	if ok {
		if pbft.status.checkStatesOr(&pbft.status.inRecovery) || prep.SequenceNumber <= pbft.exec.lastExec {
			// this is normal when in recovery
			pbft.logger.Debugf("Replica %d in recovery, received duplicate prepare from replica %d, view=%d/seqNo=%d",
				pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)
			return nil
		} else {
			// this is abnormal in common case
			pbft.logger.Warningf("Ignoring duplicate prepare from replica %d, view=%d/seqNo=%d",
				prep.ReplicaId, prep.View, prep.SequenceNumber)
			return nil
		}
	}

	cert.prepare[*prep] = true

	return pbft.maybeSendCommit(prep.BatchDigest, prep.View, prep.SequenceNumber)
}

//maybeSendCommit may send commit msg.
func (pbft *pbftImpl) maybeSendCommit(digest string, v uint64, n uint64) error {

	cert := pbft.storeMgr.getCert(v, n, digest)

	if cert == nil {
		pbft.logger.Errorf("Replica %d can't get the cert for the view=%d/seqNo=%d/digest=%s", pbft.id, v, n, digest)
		return nil
	}

	if !pbft.prepared(digest, v, n) {
		return nil
	}

	if pbft.status.getState(&pbft.status.skipInProgress) {
		pbft.logger.Debugf("Replica %d do not try to validate batch because it's in state update", pbft.id)
		return nil
	}

	if ok, _ := pbft.isPrimary(); ok {
		return pbft.sendCommit(digest, v, n)
	} else {
		idx := vidx{view: v, seqNo: n}

		if !cert.sentValidate {
			pbft.batchVdr.preparedCert[idx] = digest
			pbft.validatePending()
		}
		return nil
	}
}

//sendCommit send commit message.
func (pbft *pbftImpl) sendCommit(digest string, v uint64, n uint64) error {

	cert := pbft.storeMgr.getCert(v, n, digest)

	if !cert.sentCommit {
		pbft.logger.Debugf("Replica %d broadcasting commit for view=%d/seqNo=%d", pbft.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			ResultHash:     cert.resultHash,
			BatchDigest:    digest,
			ReplicaId:      pbft.id,
		}
		cert.sentCommit = true

		pbft.persistPSet(v, n, digest)
		payload, err := proto.Marshal(commit)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
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
	pbft.logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !pbft.isCommitLegal(commit) {
		return nil
	}

	cert := pbft.storeMgr.getCert(commit.View, commit.SequenceNumber, commit.BatchDigest)

	ok := cert.commit[*commit]

	if ok {
		if pbft.status.checkStatesOr(&pbft.status.inRecovery) || commit.SequenceNumber <= pbft.exec.lastExec {
			// this is normal when in recovery
			pbft.logger.Debugf("Replica %d in recovery, received commit from replica %d, view=%d/seqNo=%d",
				pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)
			return nil
		} else {
			// this is abnormal in common case
			pbft.logger.Warningf("Ignoring duplicate commit from replica %d, view=%d/seqNo=%d",
				commit.ReplicaId, commit.View, commit.SequenceNumber)
			return nil
		}

	}

	cert.commit[*commit] = true

	if pbft.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) {
		pbft.stopNewViewTimer()
		idx := msgID{v: commit.View, n: commit.SequenceNumber, d: commit.BatchDigest}
		if !cert.sentExecute && cert.validated {

			pbft.vcMgr.lastNewViewTimeout = pbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER)
			delete(pbft.storeMgr.outstandingReqBatches, commit.BatchDigest)
			pbft.storeMgr.committedCert[idx] = commit.BatchDigest
			pbft.commitPendingBlocks()
			if commit.SequenceNumber == pbft.vcMgr.viewChangeSeqNo {
				pbft.logger.Warningf("Replica %d cycling view for seqNo=%d", pbft.id, commit.SequenceNumber)
				pbft.sendViewChange()
			}
		} else {
			pbft.logger.Debugf("Replica %d committed for seqNo: %d, but sentExecute: %v, validated: %v", pbft.id, commit.SequenceNumber, cert.sentExecute, cert.validated)
			pbft.startTimerIfOutstandingRequests()
		}
	}

	return nil
}

func (pbft *pbftImpl) fetchMissingTransaction(preprep *PrePrepare, missing []string) error {

	pbft.logger.Debugf("Replica %d try to fetch missing txs for view=%d/seqNo=%d from primary %d",
		pbft.id, preprep.View, preprep.SequenceNumber, preprep.ReplicaId)

	fetch := &FetchMissingTransaction{
		View:           preprep.View,
		SequenceNumber: preprep.SequenceNumber,
		BatchDigest:    preprep.BatchDigest,
		HashList:       missing,
		ReplicaId:      pbft.id,
	}

	payload, err := proto.Marshal(fetch)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_FETCH_MISSING_TRANSACTION Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_FETCH_MISSING_TRANSACTION,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)

	pbft.helper.InnerUnicast(msg, preprep.ReplicaId)

	return nil
}

func (pbft *pbftImpl) recvFetchMissingTransaction(fetch *FetchMissingTransaction) error {

	pbft.logger.Debugf("Primary %d received FetchMissingTransaction request for view=%d/seqNo=%d from replica %d",
		pbft.id, fetch.View, fetch.SequenceNumber, fetch.ReplicaId)

	txList, err := pbft.batchMgr.txPool.ReturnFetchTxs(fetch.BatchDigest, fetch.HashList)
	if err != nil {
		pbft.logger.Warningf("Primary %d cannot find the digest %d, missing txs: %v",
			pbft.id, fetch.BatchDigest, fetch.HashList)
		return nil
	}

	re := &ReturnMissingTransaction{
		View:           fetch.View,
		SequenceNumber: fetch.SequenceNumber,
		BatchDigest:    fetch.BatchDigest,
		HashList:       fetch.HashList,
		TxList:         txList,
		ReplicaId:      pbft.id,
	}

	payload, err := proto.Marshal(re)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_RETURN_MISSING_TRANSACTION Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RETURN_MISSING_TRANSACTION,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)

	pbft.helper.InnerUnicast(msg, fetch.ReplicaId)

	return nil
}

func (pbft *pbftImpl) recvReturnMissingTransaction(re *ReturnMissingTransaction) events.Event {

	pbft.logger.Debugf("Replica %d received ReturnMissingTransaction from replica %d", pbft.id, re.ReplicaId)

	if len(re.TxList) != len(re.HashList) {
		pbft.logger.Warningf("Replica %d received mismatch length return %v", pbft.id, re)
		return nil
	}

	if re.SequenceNumber <= pbft.batchVdr.lastVid {
		pbft.logger.Warningf("Replica %d received validated missing transactions, seqNo=%d <= lastVid=%d, ignore it",
			pbft.id, re.SequenceNumber, pbft.batchVdr.lastVid)
		return nil
	}

	cert := pbft.storeMgr.getCert(re.View, re.SequenceNumber, re.BatchDigest)
	if cert.prePrepare == nil {
		pbft.logger.Warningf("Replica %d had not received a pre-prepare before for view=%d/seqNo=%d",
			pbft.id, re.View, re.SequenceNumber)
		return nil
	}

	_, err := pbft.batchMgr.txPool.GotMissingTxs(re.BatchDigest, re.TxList)
	if err != nil {
		pbft.logger.Warningf("Replica %d find something wrong with the return of missing txs, error: %v",
			pbft.id, err)
		return nil
	}

	pbft.validatePending()
	return nil
}

//=============================================================================
// execute transactions
//=============================================================================

//commitTransactions commit all available transactions
func (pbft *pbftImpl) commitPendingBlocks() {

	if pbft.exec.currentExec != nil {
		pbft.logger.Debugf("Replica %d not attempting to commitTransactions bacause it is currently executing %d",
			pbft.id, pbft.exec.currentExec)
	}
	pbft.logger.Debugf("Replica %d attempting to commitTransactions", pbft.id)

	for hasTxToExec := true; hasTxToExec; {
		if find, idx, cert := pbft.findNextCommitTx(); find {
			pbft.logger.Noticef("======== Replica %d Call execute, view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			pbft.persistCSet(idx.v, idx.n, idx.d)
			isPrimary, _ := pbft.isPrimary()
			//pbft.vcMgr.vcResendCount = 0
			pbft.helper.Execute(idx.n, cert.resultHash, true, isPrimary, cert.prePrepare.HashBatch.Timestamp)
			cert.sentExecute = true
			pbft.afterCommitBlock(idx)
		} else {
			hasTxToExec = false
		}
	}
	pbft.startTimerIfOutstandingRequests()
}

//findNextCommitTx find next msgID which is able to commit.
func (pbft *pbftImpl) findNextCommitTx() (find bool, idx msgID, cert *msgCert) {

	for idx = range pbft.storeMgr.committedCert {
		cert = pbft.storeMgr.certStore[idx]

		if cert == nil || cert.prePrepare == nil {
			pbft.logger.Debugf("Replica %d already checkpoint for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			//break
			continue
		}

		// check if already executed
		if cert.sentExecute == true {
			pbft.logger.Debugf("Replica %d already execute for view=%d/seqNo=%d", pbft.id, idx.v, idx.n)
			//break
			continue
		}

		if idx.n != pbft.exec.lastExec+1 {
			pbft.logger.Debugf("Replica %d expects to execute seq=%d, but get seq=%d", pbft.id, pbft.exec.lastExec+1, idx.n)
			//break
			continue
		}

		// skipInProgress == true, then this replica is in viewchange, not reply or execute
		if pbft.status.getState(&pbft.status.skipInProgress) {
			pbft.logger.Warningf("Replica %d currently picking a starting point to resume, will not execute", pbft.id)
			//break
			continue
		}

		// check if committed
		if !pbft.committed(idx.d, idx.v, idx.n) {
			//break
			continue
		}

		currentExec := idx.n
		pbft.exec.currentExec = &currentExec

		find = true
		break
	}

	return
}

//afterCommitTx after commit transaction.
func (pbft *pbftImpl) afterCommitBlock(idx msgID) {

	if pbft.exec.currentExec != nil {
		pbft.logger.Debugf("Replica %d finish execution %d, trying next", pbft.id, *pbft.exec.currentExec)
		pbft.exec.lastExec = *pbft.exec.currentExec
		delete(pbft.storeMgr.committedCert, idx)
		if pbft.exec.lastExec%pbft.K == 0 {
			bcInfo := pbft.getBlockchainInfo()
			height := bcInfo.Height
			if height == pbft.exec.lastExec {
				pbft.logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", pbft.exec.lastExec, height)
				//time.Sleep(3*time.Millisecond)
				pbft.checkpoint(pbft.exec.lastExec, bcInfo)
			} else {
				// reqBatch call execute but have not done with execute
				pbft.logger.Errorf("Fail to call the checkpoint, seqNo=%d, block height=%d", pbft.exec.lastExec, height)
				//pbft.retryCheckpoint(pbft.lastExec)
			}
		}
	} else {
		pbft.logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of data", pbft.id)
		pbft.status.activeState(&pbft.status.skipInProgress)
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
func (pbft *pbftImpl) processTransaction(tx *types.Transaction) events.Event {

	var err error
	var isGenerated bool

	if atomic.LoadUint32(&pbft.activeView) == 0 ||
		atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 ||
		pbft.status.checkStatesOr(&pbft.status.inNegoView) {
		_, err = pbft.batchMgr.txPool.AddNewTx(tx, false, true)
	}
	if ok, _ := pbft.isPrimary(); ok {
		if !pbft.batchMgr.isBatchTimerActive() {
			pbft.startBatchTimer()
		}
		isGenerated, err = pbft.batchMgr.txPool.AddNewTx(tx, true, true)
		if isGenerated {
			pbft.stopBatchTimer()
		}
	} else {
		_, err = pbft.batchMgr.txPool.AddNewTx(tx, false, true)
	}

	if pbft.batchMgr.txPool.IsPoolFull() {
		atomic.StoreUint32(&pbft.poolFull, 1)
	}

	if err != nil {
		pbft.logger.Warningf(err.Error())
	}

	return nil
}

func (pbft *pbftImpl) recvStateUpdatedEvent(et protos.StateUpdatedMessage) error {
	pbft.status.inActiveState(&pbft.status.stateTransferring)
	// If state transfer did not complete successfully, or if it did not reach our low watermark, do it again
	if et.SeqNo < pbft.h {
		pbft.logger.Warningf("Replica %d recovered to seqNo %d but our low watermark has moved to %d", pbft.id, et.SeqNo, pbft.h)
		if pbft.storeMgr.highStateTarget == nil {
			pbft.logger.Debugf("Replica %d has no state targets, cannot resume state transfer yet", pbft.id)
		} else if et.SeqNo < pbft.storeMgr.highStateTarget.seqNo {
			pbft.logger.Debugf("Replica %d has state target for %d, transferring", pbft.id, pbft.storeMgr.highStateTarget.seqNo)
			pbft.retryStateTransfer(nil)
		} else {
			pbft.logger.Debugf("Replica %d has no state target above %d, highest is %d", pbft.id, et.SeqNo, pbft.storeMgr.highStateTarget.seqNo)
		}
		return nil
	}

	pbft.logger.Infof("Replica %d application caught up via state transfer, lastExec now %d", pbft.id, et.SeqNo)
	// XXX create checkpoint
	pbft.seqNo = et.SeqNo
	pbft.exec.setLastExec(et.SeqNo)
	pbft.batchVdr.setLastVid(et.SeqNo)
	pbft.status.inActiveState(&pbft.status.skipInProgress)
	pbft.validateState()
	if et.SeqNo%pbft.K == 0 {
		bcInfo := pbft.getCurrentBlockInfo()
		pbft.checkpoint(et.SeqNo, bcInfo)
	}

	if atomic.LoadUint32(&pbft.activeView) == 1 || atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 &&
		!pbft.status.getState(&pbft.status.inNegoView) {
		atomic.StoreUint32(&pbft.normal, 1)
	}

	if pbft.status.getState(&pbft.status.inRecovery) {
		if pbft.recoveryMgr.recoveryToSeqNo == nil {
			pbft.logger.Warningf("Replica %d in recovery recvStateUpdatedEvent but "+
				"its recoveryToSeqNo is nil", pbft.id)
			return nil
		}
		if pbft.exec.lastExec >= *pbft.recoveryMgr.recoveryToSeqNo {
			// This is a somewhat subtle situation, we are behind by checkpoint, but others are just on chkpt.
			// Hence, no need to fetch preprepare, prepare, commit

			for idx := range pbft.storeMgr.certStore {
				if idx.n > pbft.exec.lastExec {
					delete(pbft.storeMgr.certStore, idx)
					pbft.persistDelQPCSet(idx.v, idx.n, idx.d)
				}
			}
			pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)

			go pbft.pbftEventQueue.Push(&LocalEvent{
				Service:   RECOVERY_SERVICE,
				EventType: RECOVERY_DONE_EVENT,
			})
			return nil
		}

		pbft.restartRecovery()
		return nil
	} else {
		pbft.executeAfterStateUpdate()
	}

	return nil
}

//recvRequestBatch handle logic after receive request batch
func (pbft *pbftImpl) recvRequestBatch(reqBatch txpool.TxHashBatch) error {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}

	pbft.logger.Debugf("Replica %d received request batch %s", pbft.id, reqBatch.BatchHash)

	txBatch := &TransactionBatch{
		TxList:    reqBatch.TxList,
		HashList:  reqBatch.TxHashList,
		Timestamp: time.Now().UnixNano(),
	}

	if atomic.LoadUint32(&pbft.activeView) == 1 && pbft.primary(pbft.view) == pbft.id &&
		!pbft.status.checkStatesOr(&pbft.status.inNegoView, &pbft.status.inRecovery) {
		pbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)
		pbft.primaryValidateBatch(reqBatch.BatchHash, txBatch, 0)
	} else {
		pbft.logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", pbft.id, reqBatch.BatchHash)
		pbft.batchMgr.txPool.GetOneTxsBack(reqBatch.BatchHash)
	}

	return nil
}

func (pbft *pbftImpl) executeAfterStateUpdate() {

	pbft.logger.Debugf("Replica %d try to execute after state update", pbft.id)

	for idx, cert := range pbft.storeMgr.certStore {
		if idx.n > pbft.seqNo && pbft.prepared(idx.d, idx.v, idx.n) && !cert.validated {
			pbft.logger.Debugf("Replica %d try to vaidate batch %s", pbft.id, idx.d)
			id := vidx{idx.v, idx.n}
			pbft.batchVdr.preparedCert[id] = idx.d
			pbft.validatePending()
		}
	}

}

func (pbft *pbftImpl) checkpoint(n uint64, info *protos.BlockchainInfo) {

	if n%pbft.K != 0 {
		pbft.logger.Errorf("Attempted to checkpoint a sequence number (%d) which is not a multiple of the checkpoint interval (%d)", n, pbft.K)
		return
	}

	id, _ := proto.Marshal(info)
	idAsString := byteToString(id)
	seqNo := n
	genesis := pbft.getGenesisInfo()

	pbft.logger.Infof("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64Id=%s/genesis=%d",
		pbft.id, pbft.view, seqNo, idAsString, genesis)

	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      pbft.id,
		Id:             idAsString,
		Genesis:        genesis,
	}
	pbft.storeMgr.saveCheckpoint(seqNo, idAsString)

	pbft.persistCheckpoint(seqNo, id)
	pbft.recvCheckpoint(chkpt)
	payload, err := proto.Marshal(chkpt)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_CHECKPOINT Marshal Error", err)
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

	pbft.logger.Debugf("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		pbft.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvCheckpoint, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.weakCheckpointSetOutOfRange(chkpt) {
		return nil
	}

	if !pbft.inW(chkpt.SequenceNumber) {
		if chkpt.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			// It is perfectly normal that we receive checkpoints for the watermark we just raised, as we raise it after 2f+1, leaving f replies left
			pbft.logger.Warningf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		} else {
			pbft.logger.Debugf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, pbft.h)
		}
		return nil
	}

	cert := pbft.storeMgr.getChkptCert(chkpt.SequenceNumber, chkpt.Id)
	ok := cert.chkpts[*chkpt]

	if ok {
		pbft.logger.Warningf("Ignoring duplicate checkpoint from replica %d, seqNo=%d", chkpt.ReplicaId, chkpt.SequenceNumber)
		return nil
	}

	cert.chkpts[*chkpt] = true
	cert.chkptCount++
	pbft.storeMgr.checkpointStore[*chkpt] = true

	pbft.logger.Debugf("Replica %d found %d matching checkpoints for seqNo %d, digest %s",
		pbft.id, cert.chkptCount, chkpt.SequenceNumber, chkpt.Id)

	if cert.chkptCount == pbft.oneCorrectQuorum() {
		// We do have a weak cert
		pbft.witnessCheckpointWeakCert(chkpt)
	}

	if cert.chkptCount < pbft.commonCaseQuorum() {
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
		pbft.logger.Debugf("Replica %d found checkpoint quorum for seqNo %d, digest %s, but it has not reached this checkpoint itself yet",
			pbft.id, chkpt.SequenceNumber, chkpt.Id)
		if pbft.status.getState(&pbft.status.skipInProgress) {
			if pbft.status.getState(&pbft.status.inRecovery) {
				pbft.moveWatermarks(chkpt.SequenceNumber)
			} else {
				logSafetyBound := pbft.h + pbft.L/2
				// As an optimization, if we are more than half way out of our log and in state transfer, move our watermarks so we don't lose track of the network
				// if needed, state transfer will restart on completion to a more recent point in time
				if chkpt.SequenceNumber >= logSafetyBound {
					pbft.logger.Debugf("Replica %d is in state transfer, but, the network seems to be moving on past %d, moving our watermarks to stay with it", pbft.id, logSafetyBound)
					pbft.moveWatermarks(chkpt.SequenceNumber)
				}
			}
		}
		return nil
	}

	pbft.logger.Infof("Replica %d found checkpoint quorum for seqNo %d, digest %s",
		pbft.id, chkpt.SequenceNumber, chkpt.Id)

	if chkptID != chkpt.Id {
		pbft.logger.Criticalf("Replica %d generated a checkpoint of %s, but a quorum of the network agrees on %s. This is almost definitely non-deterministic chaincode.",
			pbft.id, chkptID, chkpt.Id)
		pbft.stateTransfer(nil)
	}

	pbft.moveWatermarks(chkpt.SequenceNumber)

	return nil
}

// used in view-change to fetch missing assigned, non-checkpointed requests
func (pbft *pbftImpl) fetchRequestBatches() error {

	for digest := range pbft.storeMgr.missingReqBatches {
		frb := &FetchRequestBatch{
			BatchDigest: digest,
			ReplicaId:   pbft.id,
		}
		payload, err := proto.Marshal(frb)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_FRTCH_REQUEST_BATCH Marshal Error", err)
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
		if len(pbft.storeMgr.hChkpts) >= pbft.oneCorrectQuorum() {
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
			if m := chkptSeqNumArray[len(chkptSeqNumArray)-pbft.oneCorrectQuorum()]; m > H {
				if pbft.exec.lastExec >= chkpt.SequenceNumber {
					pbft.logger.Warningf("Replica %d is ahead of others, waiting others catch up", pbft.id)
					return true
				}
				pbft.logger.Warningf("Replica %d is out of date, f+1 nodes agree checkpoint with seqNo %d exists but our high water mark is %d", pbft.id, chkpt.SequenceNumber, H)
				// Discard all our requests, as we will never know which were executed, to be addressed in #394
				pbft.storeMgr.txBatchStore = make(map[string]*TransactionBatch)
				pbft.moveWatermarks(m)
				pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
				pbft.status.activeState(&pbft.status.skipInProgress)
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
	checkpointMembers := make([]replicaInfo, pbft.oneCorrectQuorum())
	i := 0
	for testChkpt := range pbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			checkpointMembers[i] = replicaInfo{
				id:      testChkpt.ReplicaId,
				height:  testChkpt.SequenceNumber,
				genesis: testChkpt.Genesis,
			}
			pbft.logger.Debugf("Replica %d adding replica %d (handle %v) to weak cert", pbft.id, testChkpt.ReplicaId, checkpointMembers[i])
			i++
		}
	}

	snapshotID, err := base64.StdEncoding.DecodeString(chkpt.Id)
	if err != nil {
		err = fmt.Errorf("Replica %d received a weak checkpoint cert which could not be decoded (%s)", pbft.id, chkpt.Id)
		pbft.logger.Error(err.Error())
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

	if pbft.status.getState(&pbft.status.skipInProgress) {
		pbft.logger.Infof("Replica %d is catching up and witnessed a weak certificate for checkpoint %d, weak cert attested to by %d of %d (%v)",
			pbft.id, chkpt.SequenceNumber, i, pbft.N, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		pbft.retryStateTransfer(target)
	}
}

func (pbft *pbftImpl) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / pbft.K * pbft.K

	if pbft.h > n {
		pbft.logger.Criticalf("Replica %d movewatermark but pbft.h(h=%d)>n(n=%d)", pbft.id, pbft.h, n)
		return
	}

	for idx := range pbft.storeMgr.certStore {
		if idx.n <= h {
			pbft.logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				pbft.id, idx.v, idx.n)
			delete(pbft.storeMgr.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n, idx.d)
		}
	}

	var target uint64
	if h < 10 {
		target = 0
	} else {
		target = h - uint64(10)
	}

	var digestList []string
	for digest, batch := range pbft.storeMgr.txBatchStore {
		if batch.SeqNo <= target && batch.SeqNo != 0 {
			delete(pbft.storeMgr.txBatchStore, digest)
			pbft.persistDelTxBatch(digest)
			digestList = append(digestList, digest)
		}
	}
	pbft.batchMgr.txPool.RemoveBatchedTxs(digestList)

	if !pbft.batchMgr.txPool.IsPoolFull() {
		atomic.StoreUint32(&pbft.poolFull, 0)
	}

	for idx := range pbft.batchVdr.preparedCert {
		if idx.seqNo <= h {
			delete(pbft.batchVdr.preparedCert, idx)
		}
	}

	for idx := range pbft.storeMgr.committedCert {
		if idx.n <= h {
			delete(pbft.storeMgr.committedCert, idx)
		}
	}

	for testChkpt := range pbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber <= h {
			pbft.logger.Debugf("Replica %d cleaning checkpoint message from replica %d, seqNo %d, b64 snapshot id %s",
				pbft.id, testChkpt.ReplicaId, testChkpt.SequenceNumber, testChkpt.Id)
			delete(pbft.storeMgr.checkpointStore, testChkpt)
		}
	}

	for cid := range pbft.storeMgr.chkptCertStore {
		if cid.n <= h {
			pbft.logger.Debugf("Replica %d cleaning checkpoint message, seqNo %d, b64 snapshot id %s",
				pbft.id, cid.n, cid.id)
			delete(pbft.storeMgr.chkptCertStore, cid)
		}
	}

	pbft.storeMgr.moveWatermarks(pbft, h)

	pbft.h = h
	err := persist.StoreState(pbft.namespace, "pbft.h", []byte(strconv.FormatUint(h, 10)))
	if err != nil {
		panic("persist pbft.h failed " + err.Error())
	}

	// we should update the recovery target if system goes on
	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.recoveryMgr.recoveryToSeqNo = &h
	}

	pbft.logger.Infof("Replica %d updated low watermark to %d",
		pbft.id, pbft.h)

	pbft.sendPendingPrePrepares()
}

func (pbft *pbftImpl) updateHighStateTarget(target *stateUpdateTarget) {
	if atomic.LoadUint32(&pbft.activeView) == 1 && pbft.storeMgr.highStateTarget != nil && pbft.storeMgr.highStateTarget.seqNo >= target.seqNo {
		pbft.logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d",
			pbft.id, target.seqNo, pbft.storeMgr.highStateTarget.seqNo)
		return
	}

	pbft.storeMgr.highStateTarget = target
}

func (pbft *pbftImpl) stateTransfer(optional *stateUpdateTarget) {

	if !pbft.status.getState(&pbft.status.skipInProgress) {
		pbft.logger.Debugf("Replica %d is out of sync, pending state transfer", pbft.id)
		pbft.status.activeState(&pbft.status.skipInProgress)
		pbft.invalidateState()
	}

	pbft.retryStateTransfer(optional)
}

func (pbft *pbftImpl) retryStateTransfer(optional *stateUpdateTarget) {

	atomic.StoreUint32(&pbft.normal, 0)

	if pbft.status.getState(&pbft.status.stateTransferring) {
		pbft.logger.Debugf("Replica %d is currently mid state transfer, it must wait for this state transfer to complete before initiating a new one", pbft.id)
		return
	}

	target := optional
	if target == nil {
		if pbft.storeMgr.highStateTarget == nil {
			pbft.logger.Debugf("Replica %d has no targets to attempt state transfer to, delaying", pbft.id)
			return
		}
		target = pbft.storeMgr.highStateTarget
	}

	pbft.status.activeState(&pbft.status.stateTransferring)

	pbft.logger.Infof("Replica %d is initiating state transfer to seqNo %d", pbft.id, target.seqNo)

	//pbft.batch.pbftManager.Queue() <- stateUpdateEvent // Todo for stateupdate
	//pbft.consumer.skipTo(target.seqNo, target.id, target.replicas)

	pbft.skipTo(target.seqNo, target.id, target.replicas)

}

func (pbft *pbftImpl) skipTo(seqNo uint64, id []byte, replicas []replicaInfo) {
	info := &protos.BlockchainInfo{}
	err := proto.Unmarshal(id, info)
	if err != nil {
		pbft.logger.Error(fmt.Sprintf("Error unmarshaling: %s", err))
		return
	}
	//pbft.UpdateState(&checkpointMessage{seqNo, id}, info, replicas)
	pbft.logger.Debug("seqNo: ", seqNo, "id: ", id, "replicas: ", replicas)
	pbft.updateState(seqNo, info, replicas)
}

// updateState attempts to synchronize state to a particular target, implicitly calls rollback if needed
func (pbft *pbftImpl) updateState(seqNo uint64, info *protos.BlockchainInfo, replicas []replicaInfo) {
	var targets []event.SyncReplica
	for _, replica := range replicas {
		target := event.SyncReplica{
			Id:      replica.id,
			Height:  replica.height,
			Genesis: replica.genesis,
		}
		targets = append(targets, target)
	}
	pbft.helper.UpdateState(pbft.id, info.Height, info.CurrentBlockHash, targets) // TODO: stateUpdateEvent

}

// =============================================================================
// receive local message methods
// =============================================================================
func (pbft *pbftImpl) recvValidatedResult(result protos.ValidatedTxs) error {
	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Debugf("Replica %d ignoring ValidatedResult as we sre in view change", pbft.id)
		return nil
	}

	if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
		pbft.logger.Debugf("Replica %d ignoring ValidatedResult as we are in updating N", pbft.id)
		return nil
	}

	primary := pbft.primary(pbft.view)
	if primary == pbft.id {
		pbft.logger.Debugf("Primary %d received validated batch for view=%d/seqNo=%d, batch size: %d, hash: %s", pbft.id, result.View, result.SeqNo, len(result.Transactions), result.Hash)

		if !pbft.inV(result.View) {
			pbft.logger.Debugf("Replica %d receives validated result whose view is in old view, now view=%v", pbft.id, pbft.view)
			return nil
		}
		batch := &TransactionBatch{
			TxList:    result.Transactions,
			Timestamp: result.Timestamp,
		}
		cache := &cacheBatch{
			batch:      batch,
			seqNo:      result.SeqNo,
			resultHash: result.Hash,
		}
		pbft.batchVdr.saveToCVB(result.Digest, cache)
		pbft.sendPendingPrePrepares()
	} else {
		pbft.logger.Debugf("Replica %d received validated batch for view=%d/seqNo=%d, batch size: %d, hash: %s",
			pbft.id, result.View, result.SeqNo, len(result.Transactions), result.Hash)

		if !pbft.inV(result.View) {
			pbft.logger.Debugf("Replica %d receives validated result %s that not in current view", pbft.id, result.Hash)
			return nil
		}

		cert := pbft.storeMgr.getCert(result.View, result.SeqNo, result.Digest)
		if cert.resultHash == "" {
			pbft.logger.Warningf("Replica %d has not store the resultHash or batchDigest for view=%d/seqNo=%d",
				pbft.id, result.View, result.SeqNo)
			return nil
		}
		if result.Hash == cert.resultHash {
			cert.validated = true
			batch := pbft.storeMgr.outstandingReqBatches[result.Digest]
			batch.SeqNo = result.SeqNo
			batch.ResultHash = result.Hash
			pbft.storeMgr.outstandingReqBatches[result.Digest] = batch
			pbft.storeMgr.txBatchStore[result.Digest] = batch
			pbft.persistTxBatch(result.Digest)
			pbft.sendCommit(result.Digest, result.View, result.SeqNo)
		} else {
			pbft.logger.Warningf("Relica %d cannot agree with the validate result for view=%d/seqNo=%d sent from primary, self: %s, primary: %s",
				pbft.id, result.View, result.SeqNo, result.Hash, cert.resultHash)
			pbft.sendViewChange()
		}
	}
	return nil
}
