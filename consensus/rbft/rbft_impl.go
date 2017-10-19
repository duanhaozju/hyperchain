//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package rbft

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"hyperchain/common"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/helper/persist"
	"hyperchain/consensus/txpool"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/hyperdb"
)

// rbftImpl is the core struct of rbft module, which handles all functions about consensus
type rbftImpl struct {
	namespace     string // this node belongs to which namespace
	f             int    // max. number of byzantine validators we can tolerate
	N             int    // max. number of validators in the network
	h             uint64 // low watermark
	id            uint64 // replica ID; PBFT `i`
	K             uint64 // how long this checkpoint period is
	logMultiplier uint64 // use this value to calculate log size : k*logMultiplier
	L             uint64 // log size: k*logMultiplier
	seqNo         uint64 // PBFT "n", strictly monotonic increasing sequence number
	view          uint64 // current view

	status RbftStatus // keep all basic status of rbft in this object

	batchMgr    *batchManager    // manage batch related issues
	batchVdr    *batchValidator  // manage batch validate issues
	timerMgr    *timerManager    // manage rbft event timers
	storeMgr    *storeManager    // manage storage
	nodeMgr     *nodeManager     // manage node delete or add
	recoveryMgr *recoveryManager // manage recovery issues
	vcMgr       *vcManager       // manage viewchange issues
	exec        *executor        // manage transaction exec

	helper helper.Stack // send message to other components of system

	eventMux *event.TypeMux
	batchSub event.Subscription // subscription channel for all events posted from consensus sub-modules
	close    chan bool          // channel to close this event process

	config    *common.Config  // get configuration info
	logger    *logging.Logger // write logger to record some info
	persister persist.Persister

	normal   uint32 // system is normal or not
	poolFull uint32 // txPool is full or not
}

// newPBFT init the PBFT instance
func newRBFT(namespace string, config *common.Config, h helper.Stack, n int) (*rbftImpl, error) {
	var err error
	rbft := &rbftImpl{}
	rbft.logger = common.GetLogger(namespace, "consensus")

	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	rbft.persister = persist.New(db)

	rbft.namespace = namespace
	rbft.helper = h
	rbft.config = config
	if !config.ContainsKey(common.C_NODE_ID) {
		err = fmt.Errorf("No hyperchain id specified!, key: %s", common.C_NODE_ID)
		return nil, err
	}
	rbft.id = uint64(config.GetInt64(common.C_NODE_ID))
	rbft.N = n
	rbft.f = (rbft.N - 1) / 3
	rbft.K = uint64(10)
	rbft.logMultiplier = uint64(4)
	rbft.L = rbft.logMultiplier * rbft.K // log size

	rbft.initMsgEventMap()

	// new executor
	rbft.exec = newExecutor()
	//new timer manager
	rbft.timerMgr = newTimerMgr(rbft)

	rbft.initTimers()
	rbft.initStatus()

	if rbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		rbft.logger.Infof("PBFT null requests timeout = %v", rbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER))
	} else {
		rbft.logger.Infof("PBFT null requests disabled")
	}

	rbft.vcMgr = newVcManager(rbft)
	// init the data logs
	rbft.storeMgr = newStoreMgr()
	rbft.storeMgr.logger = rbft.logger

	// initialize state transfer
	rbft.nodeMgr = newNodeMgr()

	rbft.batchMgr = newBatchManager(rbft) // init after rbftEventQueue
	// new batch manager
	rbft.batchVdr = newBatchValidator()
	//rbft.reqStore = newRequestStore()
	rbft.recoveryMgr = newRecoveryMgr()

	rbft.status.inActiveState(&rbft.status.inViewChange)

	rbft.logger.Infof("PBFT Max number of validating peers (N) = %v", rbft.N)
	rbft.logger.Infof("PBFT Max number of failing peers (f) = %v", rbft.f)
	rbft.logger.Infof("PBFT byzantine flag = %v", rbft.status.getState(&rbft.status.byzantine))
	rbft.logger.Infof("PBFT request timeout = %v", rbft.timerMgr.requestTimeout)
	rbft.logger.Infof("PBFT Checkpoint period (K) = %v", rbft.K)
	rbft.logger.Infof("PBFT Log multiplier = %v", rbft.logMultiplier)
	rbft.logger.Infof("PBFT log size (L) = %v", rbft.L)

	atomic.StoreUint32(&rbft.normal, 1)
	atomic.StoreUint32(&rbft.poolFull, 0)

	return rbft, nil
}

// =============================================================================
// general event process method
// =============================================================================

// listenEvent listens and dispatches messages according to their types
func (rbft *rbftImpl) listenEvent() {
	for {
		select {
		case <-rbft.close:
			return
		case obj := <-rbft.batchSub.Chan():
			ee := obj.Data
			var next consensusEvent
			var ok bool
			if next, ok = ee.(consensusEvent); !ok {
				rbft.logger.Error("Can't recognize event type")
				return
			}
			for {
				next = rbft.processEvent(next)
				if next == nil {
					break
				}
			}

		}
	}
}

func (rbft *rbftImpl) processEvent(ee consensusEvent) consensusEvent {
	switch e := ee.(type) {
	case txRequest:
		return rbft.processTransaction(e)

	case txpool.TxHashBatch:
		err := rbft.recvRequestBatch(e)
		if err != nil {
			rbft.logger.Warning(err.Error())
		}

	case protos.RoutersMessage:
		if len(e.Routers) == 0 {
			rbft.logger.Warningf("Replica %d received nil local routers", rbft.id)
			return nil
		}
		rbft.logger.Debugf("Replica %d received local routers %s", rbft.id, hashByte(e.Routers))
		rbft.nodeMgr.routers = e.Routers

	case *LocalEvent: //local event
		return rbft.dispatchLocalEvent(e)

	case *ConsensusMessage: //remote message
		next, _ := rbft.msgToEvent(e)
		return rbft.dispatchConsensusMsg(next)

	default:
		rbft.logger.Errorf("Can't recognize event type of %v.", e)
		return nil
	}
	return nil
}

// dispatchCorePbftMsg dispatch core PBFT consensus messages.
func (rbft *rbftImpl) dispatchCorePbftMsg(e consensusEvent) consensusEvent {
	switch et := e.(type) {
	case *PrePrepare:
		return rbft.recvPrePrepare(et)
	case *Prepare:
		return rbft.recvPrepare(et)
	case *Commit:
		return rbft.recvCommit(et)
	case *Checkpoint:
		return rbft.recvCheckpoint(et)
	case *FetchMissingTransaction:
		return rbft.recvFetchMissingTransaction(et)
	case *ReturnMissingTransaction:
		return rbft.recvReturnMissingTransaction(et)
	}
	return nil
}

// enqueueConsensusMsg parse consensus msg and send it to the corresponding event queue.
func (rbft *rbftImpl) enqueueConsensusMsg(msg *protos.Message) error {
	consensus := &ConsensusMessage{}
	err := proto.Unmarshal(msg.Payload, consensus)
	if err != nil {
		rbft.logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
		return err
	}

	if consensus.Type == ConsensusMessage_TRANSACTION {
		tx := &types.Transaction{}
		err := proto.Unmarshal(consensus.Payload, tx)
		if err != nil {
			rbft.logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
			return err
		}
		req := txRequest{
			tx:  tx,
			new: false,
		}
		go rbft.eventMux.Post(req)
	} else {
		go rbft.eventMux.Post(consensus)
	}

	return nil
}

//=============================================================================
// null request methods
//=============================================================================

// processNullRequest process null request when it come
func (rbft *rbftImpl) processNullRequest(msg *protos.Message) error {
	if rbft.status.getState(&rbft.status.inNegoView) {
		return nil
	}

	if rbft.status.getState(&rbft.status.inViewChange) {
		rbft.logger.Warningf("Replica %d is in viewchange, reject null request from replica %d", rbft.id, msg.Id)
		return nil
	}

	if !rbft.isPrimary(msg.Id) { // only primary could send a null request
		rbft.logger.Warningf("Replica %d received null request from replica %d who is not primary", rbft.id, msg.Id)
		return nil
	}
	// if receiver is not primary, stop FIRST_REQUEST_TIMER started after this replica finished recovery
	if !rbft.isPrimary(rbft.id) {
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}

	rbft.logger.Infof("Replica %d received null request from primary %d", rbft.id, msg.Id)
	rbft.nullReqTimerReset()
	return nil
}

// handleNullRequestEvent triggered by null request timer, primary needs to send a null request
// and replica needs to send view change
func (rbft *rbftImpl) handleNullRequestTimerEvent() {

	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to nullRequestHandler, but it's in nego-view", rbft.id)
		return
	}

	if rbft.status.getState(&rbft.status.inViewChange) {
		return
	}

	if !rbft.isPrimary(rbft.id) {
		// replica expects a null request, but primary never sent one
		rbft.logger.Warningf("Replica %d null request timer expired, sending view change", rbft.id)
		rbft.sendViewChange()
	} else {
		rbft.logger.Infof("Primary %d null request timer expired, sending null request", rbft.id)
		rbft.sendNullRequest()
	}
}

// sendNullRequest is for primary peer to send null when nullRequestTimer booms
func (rbft *rbftImpl) sendNullRequest() {
	nullRequest := nullRequestMsgToPbMsg(rbft.id)
	rbft.helper.InnerBroadcast(nullRequest)
	rbft.nullReqTimerReset()
}

//=============================================================================
// Preprepare prepare commit methods
//=============================================================================

// trySendPrePrepares check whether there is a pre-prepare message in process, If not,
// send all available PrePrepare messages by order.
func (rbft *rbftImpl) sendPendingPrePrepares() {

	// if we find a batch in findNextPrePrepareBatch, currentVid would be set.
	// And currentVid would be set to nil after send a pre-prepare message.
	if rbft.batchVdr.currentVid != nil {
		rbft.logger.Debugf("Replica %d not attempting to send pre-prepare bacause it is currently send %d, retry.", rbft.id, *rbft.batchVdr.currentVid)
		return
	}

	rbft.logger.Debugf("Replica %d attempting to call sendPrePrepare", rbft.id)

	for stop := false; !stop; {
		if find, digest, resultHash := rbft.findNextPrePrepareBatch(); find {
			waitingBatch := rbft.storeMgr.outstandingReqBatches[digest]
			rbft.sendPrePrepare(*rbft.batchVdr.currentVid, digest, resultHash, waitingBatch)
		} else {
			stop = true
		}
	}
}

// findNextPrePrepareBatch find next validated batch to send preprepare msg.
func (rbft *rbftImpl) findNextPrePrepareBatch() (find bool, digest string, resultHash string) {

	for digest = range rbft.batchVdr.cacheValidatedBatch {
		cache := rbft.batchVdr.getCacheBatchFromCVB(digest)
		if cache == nil {
			rbft.logger.Debugf("Primary %d already call sendPrePrepare for batch: %s",
				rbft.id, digest)
			continue
		}

		if cache.seqNo != rbft.batchVdr.lastVid+1 { // find the batch whose vid is exactly lastVid+1
			rbft.logger.Debugf("Primary %d expect to send pre-prepare for seqNo=%d, not seqNo=%d", rbft.id, rbft.batchVdr.lastVid+1, cache.seqNo)
			continue
		}

		currentVid := cache.seqNo
		rbft.batchVdr.setCurrentVid(&currentVid)

		n := currentVid + 1
		// check for other PRE-PREPARE for same digest, but different seqNo
		if rbft.storeMgr.existedDigest(n, rbft.view, digest) {
			rbft.deleteExistedTx(digest)
			continue
		}

		// restrict the speed of sending prePrepare, and in view change, when send new view,
		// we need to assign seqNo and there is an upper limit. So we cannot break the high watermark.
		if !rbft.sendInW(n) {
			rbft.logger.Debugf("Replica %d is primary, not sending pre-prepare for request batch %s because "+
				"batch seqNo=%d is out of sequence numbers", rbft.id, digest, n)
			rbft.batchVdr.currentVid = nil // don't send this message this time, send it after move watermark
			break
		}

		find = true
		resultHash = cache.resultHash
		break
	}
	return
}

// sendPrePrepare send prePrepare message.
func (rbft *rbftImpl) sendPrePrepare(seqNo uint64, digest string, hash string, reqBatch *TransactionBatch) {

	rbft.logger.Debugf("Primary %d broadcasting pre-prepare for view=%d/seqNo=%d/digest=%s",
		rbft.id, rbft.view, seqNo, *rbft.batchVdr.currentVid, digest)

	hashBatch := &HashBatch{
		List:      reqBatch.HashList,
		Timestamp: reqBatch.Timestamp,
	}

	preprepare := &PrePrepare{
		View:           rbft.view,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ResultHash:     hash,
		HashBatch:      hashBatch,
		ReplicaId:      rbft.id,
	}

	cert := rbft.storeMgr.getCert(rbft.view, seqNo, digest)
	cert.prePrepare = preprepare
	cert.resultHash = hash
	cert.sentValidate = true
	cert.validated = true
	rbft.batchVdr.deleteCacheFromCVB(digest)
	rbft.persistQSet(preprepare)

	reqBatch.SeqNo = seqNo
	reqBatch.ResultHash = hash
	rbft.storeMgr.outstandingReqBatches[digest] = reqBatch
	rbft.storeMgr.txBatchStore[digest] = reqBatch
	rbft.persistTxBatch(digest)

	payload, err := proto.Marshal(preprepare)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_PRE_PREPARE Marshal Error", err)
		rbft.batchVdr.updateLCVid()
		return
	}

	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PRE_PREPARE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	rbft.batchVdr.updateLCVid()
}

// recvPrePrepare process logic for PrePrepare msg.
func (rbft *rbftImpl) recvPrePrepare(preprep *PrePrepare) error {

	rbft.logger.Debugf("Replica %d received pre-prepare from replica %d for view=%d/seqNo=%d, digest=%s ",
		rbft.id, preprep.ReplicaId, preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	if !rbft.isPrePrepareLegal(preprep) {
		return nil
	}

	// In recovery, we would fetch recovery PQC, and receive these PQC again,
	// and we cannot stop timer in this situation, so we check seqNo here.
	if preprep.SequenceNumber > rbft.exec.lastExec {
		rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)
		rbft.stopFirstRequestTimer()
	}

	cert := rbft.storeMgr.getCert(preprep.View, preprep.SequenceNumber, preprep.BatchDigest)

	cert.prePrepare = preprep
	cert.resultHash = preprep.ResultHash

	if !rbft.status.checkStatesOr(&rbft.status.skipInProgress, &rbft.status.inRecovery) &&
		preprep.SequenceNumber > rbft.exec.lastExec {
		rbft.softStartNewViewTimer(rbft.timerMgr.requestTimeout,
			fmt.Sprintf("new pre-prepare for request batch view=%d/seqNo=%d, hash=%s",
				preprep.View, preprep.SequenceNumber, preprep.BatchDigest))
	}

	rbft.persistQSet(preprep)

	if !rbft.isPrimary(rbft.id) && !cert.sentPrepare &&
		rbft.prePrepared(preprep.BatchDigest, preprep.View, preprep.SequenceNumber) {
		cert.sentPrepare = true
		return rbft.sendPrepare(preprep)
	}

	return nil
}

// sendPrepare send prepare message.
func (rbft *rbftImpl) sendPrepare(preprep *PrePrepare) error {
	rbft.logger.Debugf("Backup %d broadcasting prepare for view=%d/seqNo=%d", rbft.id, preprep.View, preprep.SequenceNumber)
	prep := &Prepare{
		View:           preprep.View,
		SequenceNumber: preprep.SequenceNumber,
		BatchDigest:    preprep.BatchDigest,
		ResultHash:     preprep.ResultHash,
		ReplicaId:      rbft.id,
	}
	rbft.recvPrepare(prep) // send to itself
	payload, err := proto.Marshal(prep)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_PREPARE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	return rbft.helper.InnerBroadcast(msg)
}

// recvPrepare process logic after receive prepare message
func (rbft *rbftImpl) recvPrepare(prep *Prepare) error {
	rbft.logger.Debugf("Replica %d received prepare from replica %d for view=%d/seqNo=%d",
		rbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)

	if !rbft.isPrepareLegal(prep) {
		return nil
	}

	cert := rbft.storeMgr.getCert(prep.View, prep.SequenceNumber, prep.BatchDigest)

	ok := cert.prepare[*prep]

	if ok {
		if rbft.status.checkStatesOr(&rbft.status.inRecovery) || prep.SequenceNumber <= rbft.exec.lastExec {
			// this is normal when in recovery
			rbft.logger.Debugf("Replica %d in recovery, received duplicate prepare from replica %d, view=%d/seqNo=%d",
				rbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber)
			return nil
		} else {
			// this is abnormal in common case
			rbft.logger.Infof("Ignoring duplicate prepare from replica %d, view=%d/seqNo=%d",
				prep.ReplicaId, prep.View, prep.SequenceNumber)
			return nil
		}
	}

	cert.prepare[*prep] = true

	return rbft.maybeSendCommit(prep.BatchDigest, prep.View, prep.SequenceNumber)
}

// maybeSendCommit check if we could send commit. if no problem,
// primary would send commit and replica would validate the batch.
func (rbft *rbftImpl) maybeSendCommit(digest string, v uint64, n uint64) error {
	cert := rbft.storeMgr.getCert(v, n, digest)

	if cert == nil {
		rbft.logger.Errorf("Replica %d can't get the cert for the view=%d/seqNo=%d/digest=%s", rbft.id, v, n, digest)
		return nil
	}

	if !rbft.prepared(digest, v, n) {
		return nil
	}

	if rbft.status.getState(&rbft.status.skipInProgress) {
		rbft.logger.Debugf("Replica %d do not try to validate batch because it's in state update", rbft.id)
		return nil
	}

	if rbft.isPrimary(rbft.id) {
		return rbft.sendCommit(digest, v, n)
	} else {
		idx := vidx{view: v, seqNo: n}
		if !cert.sentValidate {
			rbft.batchVdr.preparedCert[idx] = digest
			rbft.validatePending()
		}
		return nil
	}
}

// sendCommit send commit message.
func (rbft *rbftImpl) sendCommit(digest string, v uint64, n uint64) error {

	cert := rbft.storeMgr.getCert(v, n, digest)

	if !cert.sentCommit {
		rbft.logger.Debugf("Replica %d broadcasting commit for view=%d/seqNo=%d", rbft.id, v, n)
		commit := &Commit{
			View:           v,
			SequenceNumber: n,
			ResultHash:     cert.resultHash,
			BatchDigest:    digest,
			ReplicaId:      rbft.id,
		}
		cert.sentCommit = true

		rbft.persistPSet(v, n, digest)
		payload, err := proto.Marshal(commit)
		if err != nil {
			rbft.logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		go rbft.eventMux.Post(consensusMsg)
		msg := cMsgToPbMsg(consensusMsg, rbft.id)
		return rbft.helper.InnerBroadcast(msg)
	}

	return nil
}

// recvCommit process logic after receive commit message.
func (rbft *rbftImpl) recvCommit(commit *Commit) error {
	rbft.logger.Debugf("Replica %d received commit from replica %d for view=%d/seqNo=%d",
		rbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)

	if !rbft.isCommitLegal(commit) {
		return nil
	}

	cert := rbft.storeMgr.getCert(commit.View, commit.SequenceNumber, commit.BatchDigest)

	ok := cert.commit[*commit]

	if ok {
		if rbft.status.checkStatesOr(&rbft.status.inRecovery) || commit.SequenceNumber <= rbft.exec.lastExec {
			// this is normal when in recovery
			rbft.logger.Debugf("Replica %d in recovery, received commit from replica %d, view=%d/seqNo=%d",
				rbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber)
			return nil
		} else {
			// this is abnormal in common case
			rbft.logger.Infof("Ignoring duplicate commit from replica %d, view=%d/seqNo=%d",
				commit.ReplicaId, commit.View, commit.SequenceNumber)
			return nil
		}

	}

	cert.commit[*commit] = true

	if rbft.committed(commit.BatchDigest, commit.View, commit.SequenceNumber) {
		rbft.stopNewViewTimer() // stop new view timer which was started when recv prePrepare
		idx := msgID{v: commit.View, n: commit.SequenceNumber, d: commit.BatchDigest}
		if !cert.sentExecute && cert.validated {

			rbft.vcMgr.lastNewViewTimeout = rbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER)
			delete(rbft.storeMgr.outstandingReqBatches, commit.BatchDigest)
			rbft.storeMgr.committedCert[idx] = commit.BatchDigest
			rbft.commitPendingBlocks()
			if commit.SequenceNumber == rbft.vcMgr.viewChangeSeqNo {
				rbft.logger.Warningf("Replica %d cycling view for seqNo=%d", rbft.id, commit.SequenceNumber)
				rbft.sendViewChange()
			}
		} else {
			rbft.logger.Debugf("Replica %d committed for seqNo: %d, but sentExecute: %v, validated: %v", rbft.id, commit.SequenceNumber, cert.sentExecute, cert.validated)
			rbft.startTimerIfOutstandingRequests()
		}
	}

	return nil
}

// fetchMissingTransaction fetch missing transactions from primary which this node didn't receive but primary received
func (rbft *rbftImpl) fetchMissingTransaction(preprep *PrePrepare, missing map[uint64]string) error {

	rbft.logger.Debugf("Replica %d try to fetch missing txs for view=%d/seqNo=%d/digest=%s from primary %d",
		rbft.id, preprep.View, preprep.SequenceNumber, preprep.BatchDigest, preprep.ReplicaId)

	fetch := &FetchMissingTransaction{
		View:           preprep.View,
		SequenceNumber: preprep.SequenceNumber,
		BatchDigest:    preprep.BatchDigest,
		HashList:       missing,
		ReplicaId:      rbft.id,
	}

	payload, err := proto.Marshal(fetch)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_FETCH_MISSING_TRANSACTION Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_FETCH_MISSING_TRANSACTION,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)

	rbft.helper.InnerUnicast(msg, preprep.ReplicaId)

	return nil
}

// recvFetchMissingTransaction returns transactions to a node which didn't receive some transactions and ask primary for them.
func (rbft *rbftImpl) recvFetchMissingTransaction(fetch *FetchMissingTransaction) error {

	rbft.logger.Debugf("Primary %d received FetchMissingTransaction request for view=%d/seqNo=%d/digest=%s from replica %d",
		rbft.id, fetch.View, fetch.SequenceNumber, fetch.BatchDigest, fetch.ReplicaId)

	txList := make(map[uint64]*types.Transaction)
	var err error

	if batch := rbft.storeMgr.txBatchStore[fetch.BatchDigest]; batch != nil {
		batchLen := uint64(len(batch.HashList))
		for i, hash := range fetch.HashList {
			if i >= batchLen || batch.HashList[i] != hash {
				rbft.logger.Errorf("Primary %d finds mismatch tx hash when return fetch missing transactions", rbft.id)
				return nil
			}
			txList[i] = batch.TxList[i]
		}
	} else {
		txList, err = rbft.batchMgr.txPool.ReturnFetchTxs(fetch.BatchDigest, fetch.HashList)
		if err != nil {
			rbft.logger.Warningf("Primary %d cannot find the digest %d, missing txList: %v, err: %s",
				rbft.id, fetch.BatchDigest, fetch.HashList, err)
			return nil
		}
	}

	re := &ReturnMissingTransaction{
		View:           fetch.View,
		SequenceNumber: fetch.SequenceNumber,
		BatchDigest:    fetch.BatchDigest,
		HashList:       fetch.HashList,
		TxList:         txList,
		ReplicaId:      rbft.id,
	}

	payload, err := proto.Marshal(re)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_RETURN_MISSING_TRANSACTION Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RETURN_MISSING_TRANSACTION,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)

	rbft.helper.InnerUnicast(msg, fetch.ReplicaId)

	return nil
}

// recvReturnMissingTransaction processes ReturnMissingTransaction from primary.
// Add these transactions to txPool and see if it has correct transactions.
func (rbft *rbftImpl) recvReturnMissingTransaction(re *ReturnMissingTransaction) consensusEvent {

	rbft.logger.Debugf("Replica %d received ReturnMissingTransaction from replica %d", rbft.id, re.ReplicaId)

	if len(re.TxList) != len(re.HashList) {
		rbft.logger.Warningf("Replica %d received mismatch length return %v", rbft.id, re)
		return nil
	}

	if re.SequenceNumber <= rbft.batchVdr.lastVid {
		rbft.logger.Warningf("Replica %d received validated missing transactions, seqNo=%d <= lastVid=%d, ignore it",
			rbft.id, re.SequenceNumber, rbft.batchVdr.lastVid)
		return nil
	}

	cert := rbft.storeMgr.getCert(re.View, re.SequenceNumber, re.BatchDigest)
	if cert.prePrepare == nil {
		rbft.logger.Warningf("Replica %d had not received a pre-prepare before for view=%d/seqNo=%d",
			rbft.id, re.View, re.SequenceNumber)
		return nil
	}

	err := rbft.batchMgr.txPool.GotMissingTxs(re.BatchDigest, re.TxList)
	if err != nil {
		rbft.logger.Warningf("Replica %d find something wrong with the return of missing txs, error: %v",
			rbft.id, err)
		return nil
	}

	rbft.validatePending()
	return nil
}

//=============================================================================
// execute transactions
//=============================================================================

// commitPendingBlocks commit all available transactions by order
func (rbft *rbftImpl) commitPendingBlocks() {

	if rbft.exec.currentExec != nil {
		rbft.logger.Debugf("Replica %d not attempting to commitTransactions bacause it is currently executing %d",
			rbft.id, rbft.exec.currentExec)
	}
	rbft.logger.Debugf("Replica %d attempting to commitTransactions", rbft.id)

	for hasTxToExec := true; hasTxToExec; {
		if find, idx, cert := rbft.findNextCommitTx(); find {
			rbft.logger.Noticef("======== Replica %d Call execute, view=%d/seqNo=%d", rbft.id, idx.v, idx.n)
			rbft.persistCSet(idx.v, idx.n, idx.d)
			isPrimary := rbft.isPrimary(rbft.id)
			//rbft.vcMgr.vcResendCount = 0
			rbft.helper.Execute(idx.n, cert.resultHash, true, isPrimary, cert.prePrepare.HashBatch.Timestamp)
			cert.sentExecute = true
			rbft.afterCommitBlock(idx)
		} else {
			hasTxToExec = false
		}
	}
	rbft.startTimerIfOutstandingRequests()
}

//findNextCommitTx find next msgID which is able to commit.
func (rbft *rbftImpl) findNextCommitTx() (find bool, idx msgID, cert *msgCert) {

	for idx = range rbft.storeMgr.committedCert {
		cert = rbft.storeMgr.certStore[idx]

		if cert == nil || cert.prePrepare == nil {
			rbft.logger.Debugf("Replica %d already checkpoint for view=%d/seqNo=%d", rbft.id, idx.v, idx.n)
			//break
			continue
		}

		// check if already executed
		if cert.sentExecute == true {
			rbft.logger.Debugf("Replica %d already execute for view=%d/seqNo=%d", rbft.id, idx.v, idx.n)
			//break
			continue
		}

		if idx.n != rbft.exec.lastExec+1 {
			rbft.logger.Debugf("Replica %d expects to execute seq=%d, but get seq=%d", rbft.id, rbft.exec.lastExec+1, idx.n)
			//break
			continue
		}

		// skipInProgress == true, then this replica is in viewchange, not reply or execute
		if rbft.status.getState(&rbft.status.skipInProgress) {
			rbft.logger.Warningf("Replica %d currently picking a starting point to resume, will not execute", rbft.id)
			//break
			continue
		}

		// check if committed
		if !rbft.committed(idx.d, idx.v, idx.n) {
			//break
			continue
		}

		currentExec := idx.n
		rbft.exec.currentExec = &currentExec

		find = true
		break
	}

	return
}

// afterCommitTx processes logic after commit transaction, update lastExec,
// and generate checkpoint when lastExec % K == 0
func (rbft *rbftImpl) afterCommitBlock(idx msgID) {
	if rbft.exec.currentExec != nil {
		rbft.logger.Debugf("Replica %d finish execution %d, trying next", rbft.id, *rbft.exec.currentExec)
		rbft.exec.lastExec = *rbft.exec.currentExec
		delete(rbft.storeMgr.committedCert, idx)
		if rbft.exec.lastExec%rbft.K == 0 {
			bcInfo := rbft.getBlockchainInfo()
			height := bcInfo.Height
			if height == rbft.exec.lastExec {
				rbft.logger.Debugf("Call the checkpoint, seqNo=%d, block height=%d", rbft.exec.lastExec, height)
				//time.Sleep(3*time.Millisecond)
				rbft.checkpoint(rbft.exec.lastExec, bcInfo)
			} else {
				// reqBatch call execute but have not done with execute
				rbft.logger.Errorf("Fail to call the checkpoint, seqNo=%d, block height=%d", rbft.exec.lastExec, height)
				//rbft.retryCheckpoint(rbft.lastExec)
			}
		}
	} else {
		rbft.logger.Warningf("Replica %d had execDoneSync called, flagging ourselves as out of data", rbft.id)
		rbft.status.activeState(&rbft.status.skipInProgress)
	}

	rbft.exec.currentExec = nil
}

//=============================================================================
// process methods
//=============================================================================

//processTxEvent process received transaction event
func (rbft *rbftImpl) processTransaction(req txRequest) consensusEvent {

	var err error
	var isGenerated bool

	// this node is not normal, just add a transaction without generating batch.
	if rbft.status.getState(&rbft.status.inViewChange) ||
		rbft.status.getState(&rbft.status.inUpdatingN) ||
		rbft.status.checkStatesOr(&rbft.status.inNegoView) {
		_, err = rbft.batchMgr.txPool.AddNewTx(req.tx, false, req.new)
	} else {
		// primary nodes would check if this transaction triggered generating a batch or not
		if rbft.isPrimary(rbft.id) {
			if !rbft.batchMgr.isBatchTimerActive() { // start batch timer when this node receives the first transaction of a batch
				rbft.startBatchTimer()
			}
			isGenerated, err = rbft.batchMgr.txPool.AddNewTx(req.tx, true, req.new)
			if isGenerated { // If this transaction triggers generating a batch, stop batch timer
				rbft.stopBatchTimer()
			}
		} else {
			_, err = rbft.batchMgr.txPool.AddNewTx(req.tx, false, req.new)
		}
	}

	if err != nil {
		rbft.logger.Warningf(err.Error())
	}

	if rbft.batchMgr.txPool.IsPoolFull() {
		atomic.StoreUint32(&rbft.poolFull, 1)
	}

	return nil
}

// recvStateUpdatedEvent processes StateUpdatedMessage.
func (rbft *rbftImpl) recvStateUpdatedEvent(et protos.StateUpdatedMessage) error {
	rbft.status.inActiveState(&rbft.status.stateTransferring)
	// If state transfer did not complete successfully, or if it did not reach our low watermark, do it again
	// When this node moves watermark before this node receives StateUpdatedMessage, this would happen.
	if et.SeqNo < rbft.h {
		rbft.logger.Warningf("Replica %d recovered to seqNo %d but our low watermark has moved to %d", rbft.id, et.SeqNo, rbft.h)
		if rbft.storeMgr.highStateTarget == nil {
			rbft.logger.Debugf("Replica %d has no state targets, cannot resume state transfer yet", rbft.id)
		} else if et.SeqNo < rbft.storeMgr.highStateTarget.seqNo {
			rbft.logger.Debugf("Replica %d has state target for %d, transferring", rbft.id, rbft.storeMgr.highStateTarget.seqNo)
			rbft.retryStateTransfer(nil)
		} else {
			rbft.logger.Debugf("Replica %d has no state target above %d, highest is %d", rbft.id, et.SeqNo, rbft.storeMgr.highStateTarget.seqNo)
		}
		return nil
	}

	rbft.logger.Infof("Replica %d application caught up via state transfer, lastExec now %d", rbft.id, et.SeqNo)
	// XXX create checkpoint
	rbft.seqNo = et.SeqNo
	rbft.exec.setLastExec(et.SeqNo)
	rbft.batchVdr.setLastVid(et.SeqNo)
	rbft.status.inActiveState(&rbft.status.skipInProgress)
	rbft.validateState()
	if et.SeqNo%rbft.K == 0 {
		bcInfo := rbft.getCurrentBlockInfo()
		rbft.checkpoint(et.SeqNo, bcInfo)
	}

	if !rbft.status.getState(&rbft.status.inViewChange) || !rbft.status.getState(&rbft.status.inUpdatingN) &&
		!rbft.status.getState(&rbft.status.inNegoView) {
		atomic.StoreUint32(&rbft.normal, 1)
	}

	if rbft.status.getState(&rbft.status.inRecovery) {
		if rbft.recoveryMgr.recoveryToSeqNo == nil {
			rbft.logger.Warningf("Replica %d in recovery recvStateUpdatedEvent but "+
				"its recoveryToSeqNo is nil", rbft.id)
			return nil
		}
		if rbft.exec.lastExec >= *rbft.recoveryMgr.recoveryToSeqNo {
			// This is a somewhat subtle situation, we are behind by checkpoint, but others are just on chkpt.
			// Hence, no need to fetch preprepare, prepare, commit

			for idx := range rbft.storeMgr.certStore {
				if idx.n > rbft.exec.lastExec {
					delete(rbft.storeMgr.certStore, idx)
					rbft.persistDelQPCSet(idx.v, idx.n, idx.d)
				}
			}
			rbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)

			go rbft.eventMux.Post(&LocalEvent{
				Service:   RECOVERY_SERVICE,
				EventType: RECOVERY_DONE_EVENT,
			})
			return nil
		}

		rbft.restartRecovery()
		return nil
	} else {
		rbft.executeAfterStateUpdate()
	}

	return nil
}

//recvRequestBatch handle logic after receive request batch
func (rbft *rbftImpl) recvRequestBatch(reqBatch txpool.TxHashBatch) error {

	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvRequestBatch, but it's in nego-view", rbft.id)
		return nil
	}

	rbft.logger.Debugf("Replica %d received request batch %s", rbft.id, reqBatch.BatchHash)

	txBatch := &TransactionBatch{
		TxList:    reqBatch.TxList,
		HashList:  reqBatch.TxHashList,
		Timestamp: time.Now().UnixNano(),
	}

	if !rbft.status.getState(&rbft.status.inViewChange) && rbft.isPrimary(rbft.id) &&
		!rbft.status.checkStatesOr(&rbft.status.inNegoView, &rbft.status.inRecovery) {
		rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)
		rbft.primaryValidateBatch(reqBatch.BatchHash, txBatch, 0)
	} else {
		rbft.logger.Debugf("Replica %d is backup, not sending pre-prepare for request batch %s", rbft.id, reqBatch.BatchHash)
		rbft.batchMgr.txPool.GetOneTxsBack(reqBatch.BatchHash)
	}

	return nil
}

// executeAfterStateUpdate processes logic after state update
func (rbft *rbftImpl) executeAfterStateUpdate() {

	if rbft.isPrimary(rbft.id) {
		rbft.logger.Debugf("Replica %d is primary, not execute after state update", rbft.id)
		return
	}
	rbft.logger.Debugf("Replica %d try to execute after state update", rbft.id)

	for idx, cert := range rbft.storeMgr.certStore {
		// If this node is primary, it would validate first, then send prePrepare
		if idx.n > rbft.seqNo && rbft.prepared(idx.d, idx.v, idx.n) && !cert.validated {
			rbft.logger.Debugf("Replica %d try to vaidate batch %s", rbft.id, idx.d)
			id := vidx{idx.v, idx.n}
			rbft.batchVdr.preparedCert[id] = idx.d
			rbft.validatePending()
		}
	}

}

// checkpoint generate a checkpoint and broadcast it to outer.
func (rbft *rbftImpl) checkpoint(n uint64, info *protos.BlockchainInfo) {

	if n%rbft.K != 0 {
		rbft.logger.Errorf("Attempted to checkpoint a sequence number (%d) which is not a multiple of the checkpoint interval (%d)", n, rbft.K)
		return
	}

	id, _ := proto.Marshal(info)
	idAsString := byteToString(id)
	seqNo := n
	genesis := rbft.getGenesisInfo()

	rbft.logger.Infof("Replica %d preparing checkpoint for view=%d/seqNo=%d and b64Id=%s/genesis=%d",
		rbft.id, rbft.view, seqNo, idAsString, genesis)

	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      rbft.id,
		Id:             idAsString,
		Genesis:        genesis,
	}
	rbft.storeMgr.saveCheckpoint(seqNo, idAsString)

	rbft.persistCheckpoint(seqNo, id)
	rbft.recvCheckpoint(chkpt) // send to itself
	payload, err := proto.Marshal(chkpt)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_CHECKPOINT Marshal Error", err)
		return
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_CHECKPOINT,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
}

// recvCheckpoint processes logic after receive checkpoint.
func (rbft *rbftImpl) recvCheckpoint(chkpt *Checkpoint) consensusEvent {

	rbft.logger.Debugf("Replica %d received checkpoint from replica %d, seqNo %d, digest %s",
		rbft.id, chkpt.ReplicaId, chkpt.SequenceNumber, chkpt.Id)

	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvCheckpoint, but it's in nego-view", rbft.id)
		return nil
	}

	// weakCheckpointSetOutOfRange checks if this node is fell behind or not. If we receive f+1 checkpoints whose seqNo > H (for example 150),
	// move watermark to the smallest seqNo (150) among these checkpoints, because this node is fell behind at least 50 blocks.
	// Then when this node receives f+1 checkpoints whose seqNo (160) is larger than 150,
	// enter witnessCheckpointWeakCert and set highStateTarget to 160, then this node would find itself fell behind and trigger state update
	if rbft.weakCheckpointSetOutOfRange(chkpt) {
		return nil
	}

	// if chkpt.seqNo<=h, ignore it as we have reached a higher h, else, continue to find f+1 checkpoint messages
	// with the same seqNo and ID
	if !rbft.inW(chkpt.SequenceNumber) {
		if chkpt.SequenceNumber != rbft.h && !rbft.status.getState(&rbft.status.skipInProgress) {
			// It is perfectly normal that we receive checkpoints for the watermark we just raised, as we raise it after 2f+1, leaving f replies left
			rbft.logger.Warningf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, rbft.h)
		} else {
			rbft.logger.Debugf("Checkpoint sequence number outside watermarks: seqNo %d, low-mark %d", chkpt.SequenceNumber, rbft.h)
		}
		return nil
	}

	cert := rbft.storeMgr.getChkptCert(chkpt.SequenceNumber, chkpt.Id)
	ok := cert.chkpts[*chkpt]

	if ok {
		rbft.logger.Warningf("Ignoring duplicate checkpoint from replica %d, seqNo=%d", chkpt.ReplicaId, chkpt.SequenceNumber)
		return nil
	}

	cert.chkpts[*chkpt] = true
	cert.chkptCount++
	rbft.storeMgr.checkpointStore[*chkpt] = true

	rbft.logger.Debugf("Replica %d found %d matching checkpoints for seqNo %d, digest %s",
		rbft.id, cert.chkptCount, chkpt.SequenceNumber, chkpt.Id)

	if cert.chkptCount == rbft.oneCorrectQuorum() {
		// update state update target and state transfer to it if this node already fell behind
		rbft.witnessCheckpointWeakCert(chkpt)
	}

	if cert.chkptCount < rbft.commonCaseQuorum() {
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

	chkptID, ok := rbft.storeMgr.chkpts[chkpt.SequenceNumber]
	if !ok {
		rbft.logger.Debugf("Replica %d found checkpoint quorum for seqNo %d, digest %s, but it has not reached this checkpoint itself yet",
			rbft.id, chkpt.SequenceNumber, chkpt.Id)
		if rbft.status.getState(&rbft.status.skipInProgress) {
			// When this node started state update, it would set h to the target, and finally it would receive a StateUpdatedEvent whose seqNo is this h.
			if rbft.status.getState(&rbft.status.inRecovery) {
				// If this node is in recovery, it wants to state update to a latest checkpoint so it would not fall behind more than 10 block.
				// So if move watermarks here, this node would receive StateUpdatedEvent whose seqNo is smaller than h,
				// and it would retryStateTransfer.
				// If not move watermarks here, this node would fall behind more then ten block,
				// and this is different from what we want to do using recovery.
				rbft.moveWatermarks(chkpt.SequenceNumber)
			} else {
				// If this node is not in recovery, if this node just fell behind in 20 blocks, this node could just commit and execute.
				// If larger than 20 blocks, just state update.
				logSafetyBound := rbft.h + rbft.L/2
				// As an optimization, if we are more than half way out of our log and in state transfer, move our watermarks so we don't lose track of the network
				// if needed, state transfer will restart on completion to a more recent point in time
				if chkpt.SequenceNumber >= logSafetyBound {
					rbft.logger.Debugf("Replica %d is in state transfer, but, the network seems to be moving on past %d, moving our watermarks to stay with it", rbft.id, logSafetyBound)
					rbft.moveWatermarks(chkpt.SequenceNumber)
				}
			}
		}
		return nil
	}

	rbft.logger.Infof("Replica %d found checkpoint quorum for seqNo %d, digest %s",
		rbft.id, chkpt.SequenceNumber, chkpt.Id)

	// if we found self checkpoint ID is not the same as the quorum checkpoint ID, we will fetch from others until
	// self block hash is the same as other quorum replicas
	if chkptID != chkpt.Id {
		rbft.logger.Criticalf("Replica %d generated a checkpoint of %s, but a quorum of the network agrees on %s. This is almost definitely non-deterministic chaincode.",
			rbft.id, chkptID, chkpt.Id)
		rbft.stateTransfer(nil)
	}

	rbft.moveWatermarks(chkpt.SequenceNumber)

	return nil
}

// used in view-change to fetch missing assigned, non-checkpointed requests
func (rbft *rbftImpl) fetchRequestBatches() error {

	for digest := range rbft.storeMgr.missingReqBatches {
		frb := &FetchRequestBatch{
			BatchDigest: digest,
			ReplicaId:   rbft.id,
		}
		payload, err := proto.Marshal(frb)
		if err != nil {
			rbft.logger.Errorf("ConsensusMessage_FRTCH_REQUEST_BATCH Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_FRTCH_REQUEST_BATCH,
			Payload: payload,
		}
		msg := cMsgToPbMsg(consensusMsg, rbft.id)
		rbft.helper.InnerBroadcast(msg)
	}

	return nil
}

// weakCheckpointSetOutOfRange checks if this node is fell behind or not. If we receive f+1 checkpoints whose seqNo > H (for example 150),
// move watermark to the smallest seqNo (150) among these checkpoints, because this node is fell behind 5 blocks at least.
func (rbft *rbftImpl) weakCheckpointSetOutOfRange(chkpt *Checkpoint) bool {
	H := rbft.h + rbft.L

	// Track the last observed checkpoint sequence number if it exceeds our high watermark, keyed by replica to prevent unbounded growth
	if chkpt.SequenceNumber < H {
		// For non-byzantine nodes, the checkpoint sequence number increases monotonically
		delete(rbft.storeMgr.hChkpts, chkpt.ReplicaId)
	} else {
		// We do not track the highest one, as a byzantine node could pick an arbitrarilly high sequence number
		// and even if it recovered to be non-byzantine, we would still believe it to be far ahead
		rbft.storeMgr.hChkpts[chkpt.ReplicaId] = chkpt.SequenceNumber

		// If f+1 other replicas have reported checkpoints that were (at one time) outside our watermarks
		// we need to check to see if we have fallen behind.
		if len(rbft.storeMgr.hChkpts) >= rbft.oneCorrectQuorum() {
			chkptSeqNumArray := make([]uint64, len(rbft.storeMgr.hChkpts))
			index := 0
			for replicaID, hChkpt := range rbft.storeMgr.hChkpts {
				chkptSeqNumArray[index] = hChkpt
				index++
				if hChkpt < H {
					delete(rbft.storeMgr.hChkpts, replicaID)
				}
			}
			sort.Sort(sortableUint64Slice(chkptSeqNumArray))

			// If f+1 nodes have issued checkpoints above our high water mark, then
			// we will never record 2f+1 checkpoints for that sequence number, we are out of date
			// (This is because all_replicas - missed - me = 3f+1 - f - 1 = 2f)
			if m := chkptSeqNumArray[len(chkptSeqNumArray)-rbft.oneCorrectQuorum()]; m > H {
				if rbft.exec.lastExec >= chkpt.SequenceNumber {
					rbft.logger.Warningf("Replica %d is ahead of others, waiting others catch up", rbft.id)
					return true
				}
				rbft.logger.Warningf("Replica %d is out of date, f+1 nodes agree checkpoint with seqNo %d exists but our high water mark is %d", rbft.id, chkpt.SequenceNumber, H)
				// Discard all our requests, as we will never know which were executed, to be addressed in #394
				rbft.storeMgr.txBatchStore = make(map[string]*TransactionBatch)
				rbft.moveWatermarks(m)
				rbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
				rbft.status.activeState(&rbft.status.skipInProgress)
				rbft.invalidateState()
				rbft.stopNewViewTimer()

				// TODO: state update here, this will make recovery faster, though it is presently correct
				return true
			}
		}
	}

	return false
}

// witnessCheckpointWeakCert updates state update target and state transfer to it if this node already fell behind
func (rbft *rbftImpl) witnessCheckpointWeakCert(chkpt *Checkpoint) {

	// Only ever invoked for the first weak cert, so guaranteed to be f+1
	checkpointMembers := make([]replicaInfo, rbft.oneCorrectQuorum())
	i := 0
	for testChkpt := range rbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber == chkpt.SequenceNumber && testChkpt.Id == chkpt.Id {
			checkpointMembers[i] = replicaInfo{
				id:      testChkpt.ReplicaId,
				height:  testChkpt.SequenceNumber,
				genesis: testChkpt.Genesis,
			}
			rbft.logger.Debugf("Replica %d adding replica %d (handle %v) to weak cert", rbft.id, testChkpt.ReplicaId, checkpointMembers[i])
			i++
		}
	}

	snapshotID, err := base64.StdEncoding.DecodeString(chkpt.Id)
	if err != nil {
		rbft.logger.Errorf("Replica %d received a weak checkpoint cert whose ID(%s) could not be decoded", rbft.id, chkpt.Id)
		return
	}

	target := &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: chkpt.SequenceNumber,
			id:    snapshotID,
		},
		replicas: checkpointMembers,
	}
	rbft.updateHighStateTarget(target)

	if rbft.status.getState(&rbft.status.skipInProgress) {
		rbft.logger.Infof("Replica %d is catching up and witnessed a weak certificate for checkpoint %d, weak cert attested to by %d of %d (%v)",
			rbft.id, chkpt.SequenceNumber, i, rbft.N, checkpointMembers)
		// The view should not be set to active, this should be handled by the yet unimplemented SUSPECT, see https://github.com/hyperledger/fabric/issues/1120
		rbft.retryStateTransfer(target)
	}
}

// moveWatermarks move low watermark h to n, and clear all message whose seqNo is smaller than h.
// Then if this node is primary, try to send prePrepare.
func (rbft *rbftImpl) moveWatermarks(n uint64) {

	// round down n to previous low watermark
	h := n / rbft.K * rbft.K

	if rbft.h > n {
		rbft.logger.Criticalf("Replica %d movewatermark but rbft.h(h=%d)>n(n=%d)", rbft.id, rbft.h, n)
		return
	}

	for idx := range rbft.storeMgr.certStore {
		if idx.n <= h {
			rbft.logger.Debugf("Replica %d cleaning quorum certificate for view=%d/seqNo=%d",
				rbft.id, idx.v, idx.n)
			delete(rbft.storeMgr.certStore, idx)
			rbft.persistDelQPCSet(idx.v, idx.n, idx.d)
		}
	}

	var target uint64
	if h < 10 {
		target = 0
	} else {
		target = h - uint64(10)
	}

	var digestList []string
	for digest, batch := range rbft.storeMgr.txBatchStore {
		if batch.SeqNo <= target && batch.SeqNo != 0 {
			delete(rbft.storeMgr.txBatchStore, digest)
			rbft.persistDelTxBatch(digest)
			digestList = append(digestList, digest)
		}
	}
	rbft.batchMgr.txPool.RemoveBatchedTxs(digestList)

	if !rbft.batchMgr.txPool.IsPoolFull() {
		atomic.StoreUint32(&rbft.poolFull, 0)
	}

	for idx := range rbft.batchVdr.preparedCert {
		if idx.seqNo <= h {
			delete(rbft.batchVdr.preparedCert, idx)
		}
	}

	for idx := range rbft.storeMgr.committedCert {
		if idx.n <= h {
			delete(rbft.storeMgr.committedCert, idx)
		}
	}

	for testChkpt := range rbft.storeMgr.checkpointStore {
		if testChkpt.SequenceNumber <= h {
			rbft.logger.Debugf("Replica %d cleaning checkpoint message from replica %d, seqNo %d, b64 snapshot id %s",
				rbft.id, testChkpt.ReplicaId, testChkpt.SequenceNumber, testChkpt.Id)
			delete(rbft.storeMgr.checkpointStore, testChkpt)
		}
	}

	for cid := range rbft.storeMgr.chkptCertStore {
		if cid.n <= h {
			rbft.logger.Debugf("Replica %d cleaning checkpoint message, seqNo %d, b64 snapshot id %s",
				rbft.id, cid.n, cid.id)
			delete(rbft.storeMgr.chkptCertStore, cid)
		}
	}

	rbft.storeMgr.moveWatermarks(rbft, h)

	rbft.h = h

	err := rbft.persister.StoreState("rbft.h", []byte(strconv.FormatUint(h, 10)))
	if err != nil {
		panic("persist rbft.h failed " + err.Error())
	}

	// we should update the recovery target if system goes on
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.recoveryMgr.recoveryToSeqNo = &h
	}

	rbft.logger.Infof("Replica %d updated low watermark to %d",
		rbft.id, rbft.h)

	primary := rbft.primary(rbft.view)
	if primary == rbft.id {
		rbft.sendPendingPrePrepares()
	}
}

// updateHighStateTarget updates high state target
func (rbft *rbftImpl) updateHighStateTarget(target *stateUpdateTarget) {
	if !rbft.status.getState(&rbft.status.inViewChange) && rbft.storeMgr.highStateTarget != nil && rbft.storeMgr.highStateTarget.seqNo >= target.seqNo {
		rbft.logger.Infof("Replica %d not updating state target to seqNo %d, has target for seqNo %d",
			rbft.id, target.seqNo, rbft.storeMgr.highStateTarget.seqNo)
		return
	}

	rbft.storeMgr.highStateTarget = target
}

// stateTransfer state transfers to the target
func (rbft *rbftImpl) stateTransfer(optional *stateUpdateTarget) {

	if !rbft.status.getState(&rbft.status.skipInProgress) {
		rbft.logger.Debugf("Replica %d is out of sync, pending state transfer", rbft.id)
		rbft.status.activeState(&rbft.status.skipInProgress)
		rbft.invalidateState()
	}

	rbft.retryStateTransfer(optional)
}

// retryStateTransfer sets system abnormal and stateTransferring, then skips to target
func (rbft *rbftImpl) retryStateTransfer(optional *stateUpdateTarget) {

	atomic.StoreUint32(&rbft.normal, 0)

	if rbft.status.getState(&rbft.status.stateTransferring) {
		rbft.logger.Debugf("Replica %d is currently mid state transfer, it must wait for this state transfer to complete before initiating a new one", rbft.id)
		return
	}

	target := optional
	if target == nil {
		if rbft.storeMgr.highStateTarget == nil {
			rbft.logger.Debugf("Replica %d has no targets to attempt state transfer to, delaying", rbft.id)
			return
		}
		target = rbft.storeMgr.highStateTarget
	}

	rbft.status.activeState(&rbft.status.stateTransferring)

	rbft.logger.Infof("Replica %d is initiating state transfer to seqNo %d", rbft.id, target.seqNo)

	rbft.skipTo(target.seqNo, target.id, target.replicas)

}

// skipTo skips to seqNo with id
func (rbft *rbftImpl) skipTo(seqNo uint64, id []byte, replicas []replicaInfo) {
	info := &protos.BlockchainInfo{}
	err := proto.Unmarshal(id, info)
	if err != nil {
		rbft.logger.Error(fmt.Sprintf("Error unmarshaling: %s", err))
		return
	}
	//rbft.UpdateState(&checkpointMessage{seqNo, id}, info, replicas)
	rbft.logger.Debug("seqNo: ", seqNo, "id: ", id, "replicas: ", replicas)
	rbft.updateState(seqNo, info, replicas)
}

// updateState attempts to synchronize state to a particular target, implicitly calls rollback if needed
func (rbft *rbftImpl) updateState(seqNo uint64, info *protos.BlockchainInfo, replicas []replicaInfo) {
	var targets []event.SyncReplica
	for _, replica := range replicas {
		target := event.SyncReplica{
			Id:      replica.id,
			Height:  replica.height,
			Genesis: replica.genesis,
		}
		targets = append(targets, target)
	}
	rbft.helper.UpdateState(rbft.id, info.Height, info.CurrentBlockHash, targets)

}

// =============================================================================
// receive local message methods
// =============================================================================

// recvValidatedResult processes ValidatedResult
func (rbft *rbftImpl) recvValidatedResult(result protos.ValidatedTxs) error {
	if rbft.status.getState(&rbft.status.inViewChange) {
		rbft.logger.Debugf("Replica %d ignoring ValidatedResult as we are in view change", rbft.id)
		return nil
	}

	if rbft.status.getState(&rbft.status.inUpdatingN) {
		rbft.logger.Debugf("Replica %d ignoring ValidatedResult as we are in updating N", rbft.id)
		return nil
	}

	primary := rbft.primary(rbft.view)
	if primary == rbft.id {
		rbft.logger.Debugf("Primary %d received validated batch for view=%d/seqNo=%d, batch size: %d, hash: %s", rbft.id, result.View, result.SeqNo, len(result.Transactions), result.Hash)

		if !rbft.inV(result.View) {
			rbft.logger.Debugf("Replica %d receives validated result whose view is in old view, now view=%v", rbft.id, rbft.view)
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
		rbft.batchVdr.saveToCVB(result.Digest, cache)
		rbft.sendPendingPrePrepares()
	} else {
		rbft.logger.Debugf("Replica %d received validated batch for view=%d/seqNo=%d, batch size: %d, hash: %s",
			rbft.id, result.View, result.SeqNo, len(result.Transactions), result.Hash)

		if !rbft.inV(result.View) {
			rbft.logger.Debugf("Replica %d receives validated result %s that not in current view", rbft.id, result.Hash)
			return nil
		}

		cert := rbft.storeMgr.getCert(result.View, result.SeqNo, result.Digest)
		if cert.resultHash == "" {
			rbft.logger.Warningf("Replica %d has not store the resultHash or batchDigest for view=%d/seqNo=%d",
				rbft.id, result.View, result.SeqNo)
			return nil
		}
		if result.Hash == cert.resultHash {
			cert.validated = true
			batch := rbft.storeMgr.outstandingReqBatches[result.Digest]
			rbft.storeMgr.outstandingReqBatches[result.Digest] = batch
			rbft.storeMgr.txBatchStore[result.Digest] = batch
			rbft.persistTxBatch(result.Digest)
			rbft.sendCommit(result.Digest, result.View, result.SeqNo)
		} else {
			rbft.logger.Warningf("Relica %d cannot agree with the validate result for view=%d/seqNo=%d sent from primary, self: %s, primary: %s",
				rbft.id, result.View, result.SeqNo, result.Hash, cert.resultHash)
			rbft.sendViewChange()
		}
	}
	return nil
}
