//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"hyperchain/consensus/helper/persist"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
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

// primary returns the expected primary ID with the given view v
func (pbft *pbftImpl) primary(v uint64) uint64 {
	return (v % uint64(pbft.N)) + 1
}

// isPrimary returns if current node is primary or not
func (pbft *pbftImpl) isPrimary(id uint64) bool {
	primary := pbft.primary(pbft.view)
	return primary == id
}

// InW returns if the given seqNo is higher than h or not
func (pbft *pbftImpl) inW(n uint64) bool {
	return n > pbft.h
}

// InV returns if the given view equals the current view or not
func (pbft *pbftImpl) inV(v uint64) bool {
	return pbft.view == v
}

// InWV firstly checks if the given view is inV then checks if the given seqNo n is inW
func (pbft *pbftImpl) inWV(v uint64, n uint64) bool {
	return pbft.inV(v) && pbft.inW(n)
}

// sendInWV used in findNextPrePrepareBatch firstly check the given view equals the current view or not and then check
// the given seqNo is between low watermark and high watermark or not
func (pbft *pbftImpl) sendInWV(v uint64, n uint64) bool {
	return pbft.view == v && n > pbft.h && n <= pbft.h+pbft.L
}

// getAddNodeCert returns the addnode Cert with the given addHash
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

// getDelNodeCert returns the delnode Cert with the given delHash
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

// getAddNV calculates the new N and view after add a new node
func (pbft *pbftImpl) getAddNV() (n int64, v uint64) {
	n = int64(pbft.N) + 1
	if pbft.view < uint64(pbft.N) {
		v = pbft.view + uint64(n)
	} else {
		v = pbft.view/uint64(pbft.N)*uint64(pbft.N+1) + pbft.view%uint64(pbft.N)
	}

	return
}

// getDelNV calculates the new N and view after delete a new node
func (pbft *pbftImpl) getDelNV(del uint64) (n int64, v uint64) {
	n = int64(pbft.N) - 1
	if pbft.primary(pbft.view) > del {
		v = pbft.view%uint64(pbft.N) - 1 + (uint64(pbft.N)-1)*(pbft.view/uint64(pbft.N)+1)
	} else {
		pbft.logger.Debug("N: ", pbft.N, " view: ", pbft.view, " del: ", del)
		v = pbft.view%uint64(pbft.N) + (uint64(pbft.N)-1)*(pbft.view/uint64(pbft.N)+1)
	}

	return
}

// cleanAllCache cleans some old certs of previous view in certStore whose seqNo > lastExec
func (pbft *pbftImpl) cleanAllCache() {
	for idx := range pbft.storeMgr.certStore {
		if idx.n > pbft.exec.lastExec {
			delete(pbft.storeMgr.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n, idx.d)
		}
	}

	pbft.vcMgr.viewChangeStore = make(map[vcidx]*ViewChange)
	pbft.vcMgr.newViewStore = make(map[uint64]*NewView)
	pbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
	pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)

	// we should clear the qlist & plist, since we may send viewchange after recovery
	// those previous viewchange msg will make viewchange-check incorrect
	pbft.vcMgr.qlist = make(map[qidx]*ViewChange_PQ)
	pbft.vcMgr.plist = make(map[uint64]*ViewChange_PQ)

}

// When N=3F+1, this should be 2F+1 (N-F)
// More generally, we need every two common case quorum of size X to intersect in at least F+1
// hence 2X>=N+F+1
func (pbft *pbftImpl) commonCaseQuorum() int {
	return int(math.Ceil(float64(pbft.N+pbft.f+1) / float64(2)))
}

// oneCorrectQuorum returns the number of replicas in which correct numbers must be bigger than incorrect number
func (pbft *pbftImpl) allCorrectReplicasQuorum() int {
	return (pbft.N - pbft.f)
}

// oneCorrectQuorum returns the number of replicas in which there must exist at least one correct replica
func (pbft *pbftImpl) oneCorrectQuorum() int {
	return pbft.f + 1
}

// allCorrectQuorum returns number if all consensus nodes
func (pbft *pbftImpl) allCorrectQuorum() int {
	return pbft.N
}
// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

// prePrepared returns if there existed a pre-prepare message in certStore with the given digest,view,seqNo
func (pbft *pbftImpl) prePrepared(digest string, v uint64, n uint64) bool {
	cert := pbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest {
			return true
		}
	}

	pbft.logger.Debugf("Replica %d does not have view=%d/seqNo=%d pre-prepared",
		pbft.id, v, n)

	return false
}

// prepared firstly checks if the cert with the given msgID has been prePrepared,
// then checks if this node has collected enough prepare messages for the cert with given msgID
func (pbft *pbftImpl) prepared(digest string, v uint64, n uint64) bool {

	if !pbft.prePrepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert == nil {
		return false
	}

	prepCount := len(cert.prepare)

	pbft.logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, prepCount)

	return prepCount >= pbft.commonCaseQuorum()-1
}

// committed firstly checks if the cert with the given msgID has been prepared,
// then checks if this node has collected enough commit messages for the cert with given msgID
func (pbft *pbftImpl) committed(digest string, v uint64, n uint64) bool {

	if !pbft.prepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert == nil {
		return false
	}

	cmtCount := len(cert.commit)

	pbft.logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cmtCount)

	return cmtCount >= pbft.commonCaseQuorum()
}

// =============================================================================
// helper functions for transfer message
// =============================================================================
// consensusMsgHelper helps convert the ConsensusMessage to pb.Message
func cMsgToPbMsg(msg *ConsensusMessage, id uint64) *protos.Message {
	msgPayload, err := proto.Marshal(msg)

	if err != nil {
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

// nullRequestMsgToPbMsg helps convert the nullRequestMessage to pb.Message.
func nullRequestMsgToPbMsg(id uint64) *protos.Message {
	pbMsg := &protos.Message{
		Type:      protos.Message_NULL_REQUEST,
		Payload:   nil,
		Timestamp: time.Now().UnixNano(),
		Id:        id,
	}

	return pbMsg
}

// getBlockchainInfo used after commit gets the current blockchain information from database when our lastExec reached
// the checkpoint, so here we wait until the executor module executes to a checkpoint to return the current BlockchainInfo
func (pbft *pbftImpl) getBlockchainInfo() *protos.BlockchainInfo {
	bcInfo := persist.GetBlockchainInfo(pbft.namespace)

	height := bcInfo.Height
	curBlkHash := bcInfo.LatestBlockHash
	preBlkHash := bcInfo.ParentBlockHash

	return &protos.BlockchainInfo{
		Height:            height,
		CurrentBlockHash:  curBlkHash,
		PreviousBlockHash: preBlkHash,
	}
}

// getCurrentBlockInfo used after recvStateUpdatedEvent gets the current BlockchainInfo immediately rather than waiting
// waiting until the executor module executes to a checkpoint as the current height must be a checkpoint if we reach the
// checkpoint after state update
func (pbft *pbftImpl) getCurrentBlockInfo() *protos.BlockchainInfo {
	height, curHash, prevHash := persist.GetCurrentBlockInfo(pbft.namespace)
	return &protos.BlockchainInfo{
		Height:            height,
		CurrentBlockHash:  curHash,
		PreviousBlockHash: prevHash,
	}
}

// getGenesisInfo returns the genesis block information of the current namespace
func (pbft *pbftImpl) getGenesisInfo() uint64 {
	_, genesis := persist.GetGenesisOfChain(pbft.namespace)
	return genesis
}

// =============================================================================
// helper functions for timer
// =============================================================================

// startTimerIfOutstandingRequests soft starts a new view timer if there exists some outstanding request batches,
// else reset the null request timer
func (pbft *pbftImpl) startTimerIfOutstandingRequests() {
	if pbft.status.getState(&pbft.status.skipInProgress) || pbft.exec.currentExec != nil {
		// Do not start the view change timer if we are executing or state transferring, these take arbitrarily long amounts of time
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
		pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout, fmt.Sprintf("outstanding request batches num=%v, batches: %v", len(getOutstandingDigests), getOutstandingDigests))
	} else if pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		pbft.nullReqTimerReset()
	}
}

// nullReqTimerReset reset the null request timer with a certain timeout, for different replica, null request timeout is
// different:
// 1. for primary, null request timeout is the timeout written in the config
// 2. for non-primary, null request timeout =3*(timeout written in the config)+request timeout
func (pbft *pbftImpl) nullReqTimerReset() {
	timeout := pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER)
	if !pbft.isPrimary(pbft.id) {
		// we're waiting for the primary to deliver a null request - give it a bit more time
		timeout = 3 * timeout + pbft.timerMgr.requestTimeout
	}

	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_NULL_REQUEST_TIMER_EVENT,
	}

	pbft.timerMgr.startTimerWithNewTT(NULL_REQUEST_TIMER, timeout, event, pbft.pbftEventQueue)
}

// stopFirstRequestTimer stops the first request timer event if current node is not primary
func (pbft *pbftImpl) stopFirstRequestTimer() {
	if !pbft.isPrimary(pbft.id) {
		pbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}
}

// =============================================================================
// helper functions for validateState
// =============================================================================
// invalidateState is invoked to tell us that consensus module has realized the ledger is out of sync
func (pbft *pbftImpl) invalidateState() {
	pbft.logger.Debug("Invalidating the current state")
	pbft.status.inActiveState(&pbft.status.valid)
}

// validateState is invoked to tell us that consensus module has realized the ledger is back in sync
func (pbft *pbftImpl) validateState() {
	pbft.logger.Debug("Validating the current state")
	pbft.status.activeState(&pbft.status.valid)
}

// deleteExistedTx delete batch with the given digest from cacheValidatedBatch and outstandingReqBatches
func (pbft *pbftImpl) deleteExistedTx(digest string) {
	pbft.batchVdr.updateLCVid()
	pbft.batchVdr.deleteCacheFromCVB(digest)
	delete(pbft.storeMgr.outstandingReqBatches, digest)
}

// =============================================================================
// helper functions for check the validity of consensus messages
// =============================================================================
// isPrePrepareLegal firstly checks if current status can receive pre-prepare or not, then checks pre-prepare message
// itself is legal or not
func (pbft *pbftImpl) isPrePrepareLegal(preprep *PrePrepare) bool {
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try recvPrePrepare, but it's in nego-view", pbft.id)
		return false
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Debugf("Replica %d ignoring pre-prepare as we are in view change", pbft.id)
		return false
	}

	if !pbft.isPrimary(preprep.ReplicaId) {
		pbft.logger.Warningf("Pre-prepare from other than primary: got %d, should be %d",
			preprep.ReplicaId, pbft.primary(pbft.view))
		return false
	}

	if !pbft.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			pbft.logger.Warningf("Replica %d pre-prepare view different, or sequence number outside "+
				"watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d",
				pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		} else {
			// This is perfectly normal
			pbft.logger.Debugf("Replica %d pre-prepare view different, or sequence number outside " +
				"watermarks: preprep.View %d, expected.View %d, seqNo %d, low-mark %d",
				pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		}
		return false
	}
	return true
}

// isPrepareLegal firstly checks if current status can receive prepare or not, then checks prepare message itself is
// legal or not
func (pbft *pbftImpl) isPrepareLegal(prep *Prepare) bool {
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvPrepare, but it's in nego-view", pbft.id)
		return false
	}

	// if we are not in recovery, but receive prepare from primary, which means primary behavior as a byzantine,
	// we don't send viewchange here, because in this case, replicas will eventually find primary abnormal in other
	// cases, such as inconsistent validate result or others
	if pbft.isPrimary(prep.ReplicaId) && !pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Warningf("Replica %d received prepare from primary, ignoring", pbft.id)
		return false
	}

	if !pbft.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			pbft.logger.Warningf("Replica %d ignoring prepare from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			pbft.logger.Debugf("Replica %d ignoring prepare from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d",
				pbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber, pbft.view, pbft.h)
		}

		return false
	}
	return true
}

// isCommitLegal firstly checks if current status can receive commit or not, then checks commit message itself is legal
// or not
func (pbft *pbftImpl) isCommitLegal(commit *Commit) bool {
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvCommit, but it's in nego-view", pbft.id)
		return false
	}

	if !pbft.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != pbft.h && !pbft.status.getState(&pbft.status.skipInProgress) {
			pbft.logger.Warningf("Replica %d ignoring commit from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		} else {
			// This is perfectly normal
			pbft.logger.Debugf("Replica %d ignoring commit from replica %d for view=%d/seqNo=%d: not in-wv, in view %d, low water mark %d", pbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, pbft.view, pbft.h)
		}
		return false
	}
	return true
}
