//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"math"
	"time"

	"github.com/hyperchain/hyperchain/manager/protos"

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
// helper functions for RBFT
// =============================================================================

// primary returns the expected primary ID with the given view v
func (rbft *rbftImpl) primary(v uint64) uint64 {
	return (v % uint64(rbft.N)) + 1
}

// isPrimary returns if current node is primary or not
func (rbft *rbftImpl) isPrimary(id uint64) bool {
	primary := rbft.primary(rbft.view)
	return primary == id
}

// InW returns if the given seqNo is higher than h or not
func (rbft *rbftImpl) inW(n uint64) bool {
	return n > rbft.h
}

// InV returns if the given view equals the current view or not
func (rbft *rbftImpl) inV(v uint64) bool {
	return rbft.view == v
}

// InWV firstly checks if the given view is inV then checks if the given seqNo n is inW
func (rbft *rbftImpl) inWV(v uint64, n uint64) bool {
	return rbft.inV(v) && rbft.inW(n)
}

// sendInW used in findNextPrePrepareBatch checks the given seqNo is between low
// watermark and high watermark or not.
func (rbft *rbftImpl) sendInW(n uint64) bool {
	return n > rbft.h && n <= rbft.h+rbft.L
}

// getAddNodeCert returns the addnode Cert with the given addHash
func (rbft *rbftImpl) getAddNodeCert(addHash string) (cert *addNodeCert) {
	cert, ok := rbft.nodeMgr.addNodeCertStore[addHash]

	if ok {
		return
	}

	adds := make(map[AddNode]bool)
	cert = &addNodeCert{
		addNodes: adds,
	}
	rbft.nodeMgr.addNodeCertStore[addHash] = cert

	return
}

// getDelNodeCert returns the delnode Cert with the given delHash
func (rbft *rbftImpl) getDelNodeCert(delHash string) (cert *delNodeCert) {
	cert, ok := rbft.nodeMgr.delNodeCertStore[delHash]

	if ok {
		return
	}

	dels := make(map[DelNode]bool)
	cert = &delNodeCert{
		delNodes: dels,
	}
	rbft.nodeMgr.delNodeCertStore[delHash] = cert

	return
}

// getAddNV calculates the new N and view after add a new node
func (rbft *rbftImpl) getAddNV() (n int64, v uint64) {
	n = int64(rbft.N) + 1
	if rbft.view < uint64(rbft.N) {
		v = rbft.view + uint64(n)
	} else {
		v = rbft.view/uint64(rbft.N)*uint64(rbft.N+1) + rbft.view%uint64(rbft.N)
	}

	return
}

// getDelNV calculates the new N and view after delete a new node
func (rbft *rbftImpl) getDelNV(del uint64) (n int64, v uint64) {
	n = int64(rbft.N) - 1
	rbft.logger.Debugf("N: %d, view: %d, delID: %d", rbft.N, rbft.view, del)
	if rbft.primary(rbft.view) > del {
		v = rbft.view%uint64(rbft.N) - 1 + (uint64(rbft.N)-1)*(rbft.view/uint64(rbft.N)+1)
	} else {
		v = rbft.view%uint64(rbft.N) + (uint64(rbft.N)-1)*(rbft.view/uint64(rbft.N)+1)
	}

	return
}

// cleanAllCache cleans some old certs of previous view in certStore whose seqNo > lastExec
func (rbft *rbftImpl) cleanAllCache() {
	for idx := range rbft.storeMgr.certStore {
		if idx.n > rbft.exec.lastExec {
			delete(rbft.storeMgr.certStore, idx)
			rbft.persistDelQPCSet(idx.v, idx.n, idx.d)
		}
	}

	rbft.vcMgr.viewChangeStore = make(map[vcidx]*ViewChange)
	rbft.vcMgr.newViewStore = make(map[uint64]*NewView)
	rbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
	rbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)

	// we should clear the qlist & plist, since we may send viewchange after recovery
	// those previous viewchange msg will make viewchange-check incorrect
	rbft.vcMgr.qlist = make(map[qidx]*Vc_PQ)
	rbft.vcMgr.plist = make(map[uint64]*Vc_PQ)

}

// When N=3F+1, this should be 2F+1 (N-F)
// More generally, we need every two common case quorum of size X to intersect in at least F+1
// hence 2X>=N+F+1
func (rbft *rbftImpl) commonCaseQuorum() int {
	return int(math.Ceil(float64(rbft.N+rbft.f+1) / float64(2)))
}

// oneCorrectQuorum returns the number of replicas in which correct numbers must be bigger than incorrect number
func (rbft *rbftImpl) allCorrectReplicasQuorum() int {
	return rbft.N - rbft.f
}

// oneCorrectQuorum returns the number of replicas in which there must exist at least one correct replica
func (rbft *rbftImpl) oneCorrectQuorum() int {
	return rbft.f + 1
}

// allCorrectQuorum returns number if all consensus nodes
func (rbft *rbftImpl) allCorrectQuorum() int {
	return rbft.N
}

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

// prePrepared returns if there existed a pre-prepare message in certStore with the given digest,view,seqNo
func (rbft *rbftImpl) prePrepared(digest string, v uint64, n uint64) bool {
	cert := rbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest {
			return true
		}
	}

	rbft.logger.Debugf("Replica %d does not have view=%d/seqNo=%d prePrepared",
		rbft.id, v, n)

	return false
}

// prepared firstly checks if the cert with the given msgID has been prePrepared,
// then checks if this node has collected enough prepare messages for the cert with given msgID
func (rbft *rbftImpl) prepared(digest string, v uint64, n uint64) bool {

	if !rbft.prePrepared(digest, v, n) {
		return false
	}

	cert := rbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert == nil {
		return false
	}

	prepCount := len(cert.prepare)

	rbft.logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		rbft.id, v, n, prepCount)

	return prepCount >= rbft.commonCaseQuorum()-1
}

// committed firstly checks if the cert with the given msgID has been prepared,
// then checks if this node has collected enough commit messages for the cert with given msgID
func (rbft *rbftImpl) committed(digest string, v uint64, n uint64) bool {

	if !rbft.prepared(digest, v, n) {
		return false
	}

	cert := rbft.storeMgr.certStore[msgID{v, n, digest}]

	if cert == nil {
		return false
	}

	cmtCount := len(cert.commit)

	rbft.logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		rbft.id, v, n, cmtCount)

	return cmtCount >= rbft.commonCaseQuorum()
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
func (rbft *rbftImpl) getBlockchainInfo() *protos.BlockchainInfo {
	bcInfo := rbft.GetBlockchainInfo(rbft.namespace)

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
func (rbft *rbftImpl) getCurrentBlockInfo() *protos.BlockchainInfo {
	height, curHash, prevHash := rbft.GetCurrentBlockInfo(rbft.namespace)
	return &protos.BlockchainInfo{
		Height:            height,
		CurrentBlockHash:  curHash,
		PreviousBlockHash: prevHash,
	}
}

// getGenesisInfo returns the genesis block information of the current namespace
func (rbft *rbftImpl) getGenesisInfo() uint64 {
	genesis, _ := rbft.GetGenesisOfChain(rbft.namespace)
	return genesis
}

// =============================================================================
// helper functions for timer
// =============================================================================

// startTimerIfOutstandingRequests soft starts a new view timer if there exists some outstanding request batches,
// else reset the null request timer
func (rbft *rbftImpl) startTimerIfOutstandingRequests() {
	if rbft.in(skipInProgress) || rbft.exec.currentExec != nil {
		// Do not start the view change timer if we are executing or state transferring, these take arbitrarily long amounts of time
		return
	}

	if len(rbft.storeMgr.outstandingReqBatches) > 0 {
		getOutstandingDigests := func() []string {
			var digests []string
			for digest := range rbft.storeMgr.outstandingReqBatches {
				digests = append(digests, digest)
			}
			return digests
		}()
		rbft.softStartNewViewTimer(rbft.timerMgr.getTimeoutValue(REQUEST_TIMER), fmt.Sprintf("outstanding request batches num=%v, batches: %v", len(getOutstandingDigests), getOutstandingDigests))
	} else if rbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		rbft.nullReqTimerReset()
	}
}

// nullReqTimerReset reset the null request timer with a certain timeout, for different replica, null request timeout is
// different:
// 1. for primary, null request timeout is the timeout written in the config
// 2. for non-primary, null request timeout =3*(timeout written in the config)+request timeout
func (rbft *rbftImpl) nullReqTimerReset() {
	timeout := rbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER)
	if !rbft.isPrimary(rbft.id) {
		// we're waiting for the primary to deliver a null request - give it a bit more time
		timeout = 3*timeout + rbft.timerMgr.getTimeoutValue(REQUEST_TIMER)
	}

	event := &LocalEvent{
		Service:   CORE_RBFT_SERVICE,
		EventType: CORE_NULL_REQUEST_TIMER_EVENT,
	}

	rbft.timerMgr.startTimerWithNewTT(NULL_REQUEST_TIMER, timeout, event, rbft.eventMux)
}

// stopFirstRequestTimer stops the first request timer event if current node is not primary
func (rbft *rbftImpl) stopFirstRequestTimer() {
	if !rbft.isPrimary(rbft.id) {
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}
}

// =============================================================================
// helper functions for validateState
// =============================================================================

// deleteExistedTx delete batch with the given digest from cacheValidatedBatch and outstandingReqBatches
func (rbft *rbftImpl) deleteExistedTx(digest string) {
	rbft.batchVdr.updateLCVid()
	rbft.batchVdr.deleteCacheFromCVB(digest)
	delete(rbft.storeMgr.outstandingReqBatches, digest)
}

// =============================================================================
// helper functions for check the validity of consensus messages
// =============================================================================
// isPrePrepareLegal firstly checks if current status can receive pre-prepare or not, then checks pre-prepare message
// itself is legal or not
func (rbft *rbftImpl) isPrePrepareLegal(preprep *PrePrepare) bool {
	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d try to receive prePrepare, but it's in negotiateView", rbft.id)
		return false
	}

	if rbft.in(inViewChange) {
		rbft.logger.Debugf("Replica %d try to receive prePrepare, but it's in viewChange", rbft.id)
		return false
	}

	if !rbft.isPrimary(preprep.ReplicaId) {
		rbft.logger.Warningf("Replica %d received prePrepare from non-primary: got %d, should be %d",
			rbft.id, preprep.ReplicaId, rbft.primary(rbft.view))
		return false
	}

	if !rbft.inWV(preprep.View, preprep.SequenceNumber) {
		if preprep.SequenceNumber != rbft.h && !rbft.in(skipInProgress) {
			rbft.logger.Warningf("Replica %d received prePrepare with a different view or sequence "+
				"number outside watermarks: prePrep.View %d, expected.View %d, seqNo %d, low water mark %d",
				rbft.id, preprep.View, rbft.view, preprep.SequenceNumber, rbft.h)
		} else {
			// This is perfectly normal
			rbft.logger.Debugf("Replica %d received prePrepare with a different view or sequence "+
				"number outside watermarks: preprep.View %d, expected.View %d, seqNo %d, low water mark %d",
				rbft.id, preprep.View, rbft.view, preprep.SequenceNumber, rbft.h)
		}
		return false
	}
	return true
}

// isPrepareLegal firstly checks if current status can receive prepare or not, then checks prepare message itself is
// legal or not
func (rbft *rbftImpl) isPrepareLegal(prep *Prepare) bool {
	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d try to receive prepare, but it's in negotiateView", rbft.id)
		return false
	}

	// if we are not in recovery, but receive prepare from primary, which means primary behavior as a byzantine,
	// we don't send viewchange here, because in this case, replicas will eventually find primary abnormal in other
	// cases, such as inconsistent validate result or others
	if rbft.isPrimary(prep.ReplicaId) && !rbft.in(inRecovery) {
		rbft.logger.Warningf("Replica %d received prepare from primary, ignore it", rbft.id)
		return false
	}

	if !rbft.inWV(prep.View, prep.SequenceNumber) {
		if prep.SequenceNumber != rbft.h && !rbft.in(skipInProgress) {
			rbft.logger.Warningf("Replica %d ignore prepare from replica %d for view=%d/seqNo=%d: not inWv, in view: %d, h: %d",
				rbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber, rbft.view, rbft.h)
		} else {
			// This is perfectly normal
			rbft.logger.Debugf("Replica %d ignore prepare from replica %d for view=%d/seqNo=%d: not inWv, in view: %d, h: %d",
				rbft.id, prep.ReplicaId, prep.View, prep.SequenceNumber, rbft.view, rbft.h)
		}

		return false
	}
	return true
}

// isCommitLegal firstly checks if current status can receive commit or not, then checks commit message itself is legal
// or not
func (rbft *rbftImpl) isCommitLegal(commit *Commit) bool {
	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d try to receive commit, but it's in negotiateView", rbft.id)
		return false
	}

	if !rbft.inWV(commit.View, commit.SequenceNumber) {
		if commit.SequenceNumber != rbft.h && !rbft.in(skipInProgress) {
			rbft.logger.Warningf("Replica %d ignore commit from replica %d for view=%d/seqNo=%d: not inWv, in view: %d, h: %d", rbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, rbft.view, rbft.h)
		} else {
			// This is perfectly normal
			rbft.logger.Debugf("Replica %d ignore commit from replica %d for view=%d/seqNo=%d: not inWv, in view: %d, h: %d", rbft.id, commit.ReplicaId, commit.View, commit.SequenceNumber, rbft.view, rbft.h)
		}
		return false
	}
	return true
}
