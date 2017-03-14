//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"time"
	"fmt"

	"hyperchain/protos"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"sync/atomic"
)

// =============================================================================
// helper functions for sort
// =============================================================================
type sortableUint64Slice []uint64

func (a sortableUint64Slice) Len() int {
	return len(a)
}
func (a sortableUint64Slice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a sortableUint64Slice) Less(i, j int) bool {
	return a[i] < a[j]
}

// =============================================================================
// helper functions for PBFT
// =============================================================================

// Given a certain view v and replicaCount n, what is the expected primary?
func (pbft *pbftImpl) primary(v uint64) uint64 {
	return (v % uint64(pbft.N)) + 1
}

//isPrimary is current PBFT node the primary?
func (pbft *pbftImpl) isPrimary() (bool, uint64)  {
	primary := pbft.primary(pbft.view)
	return primary == pbft.id, primary
}

// Is the sequence number between watermarks?
func (pbft *pbftImpl) inW(n uint64) bool {
	return n > pbft.h && n-pbft.h <= pbft.L
}

// Is the view right? And is the sequence number between watermarks?
func (pbft *pbftImpl) inWV(v uint64, n uint64) bool {
	return pbft.view == v && pbft.inW(n)
}

// Given a ip/digest get the addnode Cert
func (pbft *pbftImpl) getAddNodeCert(addHash string) (cert *addNodeCert) {

	cert, ok := pbft.nodeMgr.addNodeCertStore[addHash]

	if ok {
		return
	}

	adds := make(map[AddNode]bool)
	cert = &addNodeCert{
		addNodes: adds,
	}
	pbft.nodeMgr.addNodeCertStore[addHash] = cert

	return
}

// Given a ip/digest get the addnode Cert
func (pbft *pbftImpl) getDelNodeCert(delHash string) (cert *delNodeCert) {

	cert, ok := pbft.nodeMgr.delNodeCertStore[delHash]

	if ok {
		return
	}

	dels := make(map[DelNode]bool)
	cert = &delNodeCert{
		delNodes: dels,
	}
	pbft.nodeMgr.delNodeCertStore[delHash] = cert

	return
}

func (pbft *pbftImpl) getAddNV() (n int64, v uint64) {

	n = int64(pbft.N) + 1
	if pbft.view < uint64(pbft.N) {
		v = pbft.view + uint64(n)
	} else {
		v = pbft.view + uint64(n) + 1
	}

	return
}

func (pbft *pbftImpl) getDelNV(del uint64) (n int64, v uint64) {

	n = int64(pbft.N) - 1
	if pbft.primary(pbft.view) > del {
		v = pbft.view % uint64(pbft.N) - 1 + (uint64(pbft.N) - 1) * (pbft.view / uint64(pbft.N) + 1)
	} else {
		logger.Debug("N: ", pbft.N, " view: ", pbft.view, " del: ", del)
		v = pbft.view % uint64(pbft.N) + (uint64(pbft.N) - 1) * (pbft.view / uint64(pbft.N) + 1)
	}

	return
}

// =============================================================================
// prepare/commit quorum checks helper
// =============================================================================

func (pbft *pbftImpl) preparedReplicasQuorum() int {
	return (2 * pbft.f)
}

func (pbft *pbftImpl) committedReplicasQuorum() int {
	return (2*pbft.f + 1)
}

// intersectionQuorum returns the number of replicas that have to
// agree to guarantee that at least one correct replica is shared by
// two intersection quora
func (pbft *pbftImpl) intersectionQuorum() int {
	return (pbft.N + pbft.f + 2) / 2
}

func (pbft *pbftImpl) allCorrectReplicasQuorum() int {
	return (pbft.N - pbft.f)
}

func (pbft *pbftImpl) minimumCorrectQuorum() int {
	return pbft.f + 1
}

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

func (pbft *pbftImpl) prePrepared(digest string, v uint64, n uint64) bool {

	if digest != "" && !pbft.batchVdr.containsInVBS(digest) {
		logger.Debugf("Replica %d havan't store the reqBatch", pbft.id)
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest && cert.digest == digest {
			return true
		}
	}

	logger.Debugf("Replica %d does not have view=%d/seqNo=%d pre-prepared",
		pbft.id, v, n)

	return false
}

//prepared judge whether collect enough prepare messages.
func (pbft *pbftImpl) prepared(digest string, v uint64, n uint64) bool {

	if !pbft.prePrepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cert.prepareCount)

	return cert.prepareCount >= pbft.preparedReplicasQuorum()
}

func (pbft *pbftImpl) committed(digest string, v uint64, n uint64) bool {

	if !pbft.prepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cert.commitCount)

	return cert.commitCount >= pbft.intersectionQuorum()
}

// =============================================================================
// helper functions for transfer message
// =============================================================================
// consensusMsgHelper help convert the ConsensusMessage to pb.Message
func cMsgToPbMsg(msg *ConsensusMessage, id uint64) *protos.Message {

	msgPayload, err := proto.Marshal(msg)

	if err != nil {
		logger.Errorf("ConsensusMessage Marshal Error", err)
		return nil
	}

	pbMsg := &protos.Message{
		Type:      protos.Message_CONSENSUS,
		Payload:   msgPayload,
		Timestamp: time.Now().UnixNano(),
		Id:        id,
	}

	return pbMsg
}

// nullRequestMsgToPbMsg help convert the nullRequestMessage to pb.Message.
func nullRequestMsgToPbMsg(id uint64) *protos.Message {
	pbMsg := &protos.Message{
		Type:      protos.Message_NULL_REQUEST,
		Payload:   nil,
		Timestamp: time.Now().UnixNano(),
		Id:        id,
	}

	return pbMsg
}

// StateUpdateHelper help convert checkPointInfo, blockchainInfo, replicas to pb.UpdateStateMessage
func stateUpdateHelper(myId uint64, seqNo uint64, id []byte, replicaId []uint64) *protos.UpdateStateMessage {

	stateUpdateMsg := &protos.UpdateStateMessage{
		Id:       myId,
		SeqNo:    seqNo,
		TargetId: id,
		Replicas: replicaId,
	}
	return stateUpdateMsg
}

func (pbft *pbftImpl) getBlockchainInfo() *protos.BlockchainInfo {

	bcInfo := persist.GetBlockchainInfo(pbft.namespace)

	height := bcInfo.Height
	curBlkHash := bcInfo.LatestBlockHash
	preBlkHash := bcInfo.ParentBlockHash

	return &protos.BlockchainInfo{
		Height:	           height,
		CurrentBlockHash:  curBlkHash,
		PreviousBlockHash: preBlkHash,
	}
}

func (pbft *pbftImpl) getCurrentBlockInfo() *protos.BlockchainInfo {
	height, curHash, prevHash := persist.GetCurrentBlockInfo(pbft.namespace)
	return &protos.BlockchainInfo{
		Height:            height,
		CurrentBlockHash:  curHash,
		PreviousBlockHash: prevHash,
	}
}

// =============================================================================
// helper functions for timer
// =============================================================================

func (pbft *pbftImpl) startTimerIfOutstandingRequests() {
	if pbft.status.getState(&pbft.status.skipInProgress) || pbft.exec.currentExec != nil {
		// Do not start the view change timer if we are executing or state transferring, these take arbitrarilly long amounts of time
		return
	}

	if len(pbft.storeMgr.outstandingReqBatches) > 0 {
		getOutstandingDigests := func() []string {
			var digests []string
			for digest := range pbft.storeMgr.outstandingReqBatches {
				digests = append(digests, digest)
			}
			return digests
		}()
		pbft.softStartNewViewTimer(pbft.pbftTimerMgr.requestTimeout, fmt.Sprintf("outstanding request batches num=%v", len(getOutstandingDigests)))
	} else if pbft.pbftTimerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		pbft.nullReqTimerReset()
	}
}

func (pbft *pbftImpl) nullReqTimerReset() {
	timeout := pbft.pbftTimerMgr.getTimeoutValue(NULL_REQUEST_TIMER)
	if pbft.primary(pbft.view) != pbft.id {
		// we're waiting for the primary to deliver a null request - give it a bit more time
		timeout += pbft.pbftTimerMgr.requestTimeout
	}

	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_NULL_REQUEST_TIMER_EVENT,
	}

	af := func(){
		pbft.pbftEventQueue.Push(event)
	}

	//logger.Errorf("null request time out is %v", pbft.pbftTimerMgr.getTimeoutValue(NULL_REQUEST_TIMER))
	//logger.Errorf("request time out is %v", pbft.pbftTimerMgr.requestTimeout)
	//logger.Errorf("reset null request timeout to %v", timeout)

	pbft.pbftTimerMgr.startTimerWithNewTT(NULL_REQUEST_TIMER, timeout, af)
}

//stopFirstRequestTimer
func (pbft *pbftImpl) stopFirstRequestTimer()  {
	if ok, _ := pbft.isPrimary(); !ok {
		pbft.pbftTimerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}
}

// =============================================================================
// helper functions for validateState
// =============================================================================
// invalidateState is invoked to tell us that consensus realizes the ledger is out of sync
func (pbft *pbftImpl) invalidateState() {
	logger.Debug("Invalidating the current state")
	pbft.status.inActiveState(&pbft.status.valid)
}

// validateState is invoked to tell us that consensus has the ledger back in sync
func (pbft *pbftImpl) validateState() {
	logger.Debug("Validating the current state")
	pbft.status.activeState(&pbft.status.valid)
}

//deleteExistedTx delete existed transaction.
func (pbft *pbftImpl) deleteExistedTx(digest string) {
	pbft.batchVdr.updateLCVid()
	pbft.batchVdr.deleteCacheFromCVB(digest)
	pbft.batchVdr.deleteTxFromVBS(digest)
	delete(pbft.storeMgr.outstandingReqBatches, digest)
}

//isPrePrepareLegal both PBFT state PrePrepare message are legal.
func (pbft *pbftImpl) isPrePrepareLegal(preprep *PrePrepare) bool {
	if pbft.status.getState(&pbft.status.inNegoView) {
		logger.Debugf("Replica %d try recvPrePrepare, but it's in nego-view", pbft.id)
		return false
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		logger.Debugf("Replica %d ignoring pre-prepare as we are in view change", pbft.id)
		return false
	}

	if pbft.primary(pbft.view) != preprep.ReplicaId {
		logger.Warningf("Pre-prepare from other than primary: got %d, should be %d",
			preprep.ReplicaId, pbft.primary(pbft.view))
		return false
	}

	if !pbft.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			logger.Warningf("Replica %d pre-prepare view different, or sequence number outside " +
				"watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d",
				pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: " +
				"preprep.View %d, expected.View %d, seqNo %d, low-mark %d",
				pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		}
		return false
	}
	return true
}

//isPrepareLegal both PBFT state Prepare message are legal.
func (pbft *pbftImpl) isPrepareLegal(prep *Prepare) bool  {
	if pbft.status.getState(&pbft.status.inNegoView) {
		logger.Debugf("Replica %d try to recvPrepare, but it's in nego-view", pbft.id)
		return false
	}
	//TODO: need !pbft.status[IN_RECOVERY] ?
	if pbft.primary(prep.View) == prep.ReplicaId && !pbft.status.getState(&pbft.status.inRecovery) {
		logger.Warningf("Replica %d received prepare from primary, ignoring", pbft.id)
		return false
	}

	if !pbft.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			logger.Warningf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring prepare for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		}

		return false
	}
	return true
}

//isPrepareLegal both PBFT state Commit message are legal.
func (pbft *pbftImpl) isCommitLegal(commit *Commit) bool  {
	if pbft.status.getState(&pbft.status.inNegoView) {
		logger.Debugf("Replica %d try to recvCommit, but it's in nego-view", pbft.id)
		return false
	}

	if !pbft.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			logger.Warningf("Replica %d ignoring commit from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			logger.Debugf("Replica %d ignoring commit for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		}
		return false
	}
	return true
}
