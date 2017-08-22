//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"time"
	"fmt"

	"hyperchain/manager/protos"
	"hyperchain/consensus/helper/persist"

	"github.com/golang/protobuf/proto"
	"sync/atomic"
	"math"
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
func (pbft *pbftImpl) isPrimary() (bool, uint64) {
	primary := pbft.primary(pbft.view)
	return primary == pbft.id, primary
}

// Is the sequence number between watermarks?
func (pbft *pbftImpl) inW(n uint64) bool {
	return n > pbft.h
}

// Is the view right?
func (pbft *pbftImpl) inV(v uint64) bool {
	return pbft.view == v
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

// Is the view right? And is the sequence number between low watermark and high watermark?
func (pbft *pbftImpl) sendInWV(v uint64, n uint64) bool {
	return pbft.view == v && n > pbft.h && n <= pbft.h + pbft.L
}


func (pbft *pbftImpl) allCorrectQuorum() int {
	return pbft.N
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
		v = pbft.view/uint64(pbft.N)*uint64(pbft.N+1) + pbft.view%uint64(pbft.N)
	}

	return
}

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

func (pbft *pbftImpl) handleCachedTxs(cache map[uint64]*transactionStore) {
	for vid, batch := range cache {
		if vid < pbft.h {
			continue
		}
		for batch.Len() != 0 {
			temp := batch.order.Front().Value
			txc, ok := interface{}(temp).(transactionContainer)
			if !ok {
				pbft.logger.Error("type assert error:", temp)
				return
			}
			tx := txc.tx
			if tx != nil {
				pbft.reqStore.storeOutstanding(tx)
			}
			batch.remove(tx)
		}
	}
}

func (pbft *pbftImpl) cleanAllCache() {

	for idx := range pbft.storeMgr.certStore {
		if idx.n > pbft.exec.lastExec{
			delete(pbft.storeMgr.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n)
		}
	}
	pbft.dupLock.Lock()
	for n := range pbft.duplicator {
		if n > pbft.exec.lastExec {
			delete(pbft.duplicator, n)
		}
	}
	pbft.dupLock.Unlock()

	pbft.vcMgr.viewChangeStore = make(map[vcidx]*ViewChange)
	pbft.vcMgr.newViewStore = make(map[uint64]*NewView)
	pbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
	pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)

}

// When N=3F+1, this should be 2F+1 (N-F)
// More generally, we need every two common case quorum of size X to intersect in at least F+1
// hence 2X>=N+F+1
func (pbft *pbftImpl) commonCaseQuorum() int {
	return int(math.Ceil(float64(pbft.N+pbft.f+1)/float64(2)))
}

func (pbft *pbftImpl) allCorrectReplicasQuorum() int {
	return (pbft.N - pbft.f)
}

func (pbft *pbftImpl) oneCorrectQuorum() int {
	return pbft.f + 1
}

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

func (pbft *pbftImpl) prePrepared(digest string, v uint64, n uint64) bool {

	if digest != "" && !pbft.batchVdr.containsInVBS(digest) {
		pbft.logger.Debugf("Replica %d havan't store the reqBatch", pbft.id)
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest && cert.digest == digest {
			return true
		}
	}

	pbft.logger.Debugf("Replica %d does not have view=%d/seqNo=%d pre-prepared",
		pbft.id, v, n)

	return false
}

//prepared judge whether collect enough prepare messages.
func (pbft *pbftImpl) prepared(digest string, v uint64, n uint64) bool {

	if !pbft.prePrepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	prepCount := len(cert.prepare)

	pbft.logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, prepCount)

	return prepCount >= pbft.commonCaseQuorum()-1
}

func (pbft *pbftImpl) onlyPrepared(digest string, v uint64, n uint64) bool {

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	prepCount := len(cert.prepare)

	return prepCount >= pbft.commonCaseQuorum()-1
}

func (pbft *pbftImpl) committed(digest string, v uint64, n uint64) bool {

	if !pbft.prepared(digest, v, n) {
		return false
	}

	cert := pbft.storeMgr.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	cmtCount := len(cert.commit)

	pbft.logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cmtCount)

	return cmtCount >= pbft.commonCaseQuorum()
}

func (pbft *pbftImpl) onlyCommitted(digest string, v uint64, n uint64) bool {

	cert := pbft.storeMgr.certStore[msgID{v, n}]

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
// consensusMsgHelper help convert the ConsensusMessage to pb.Message
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

func (pbft *pbftImpl) getCurrentBlockInfo() *protos.BlockchainInfo {
	height, curHash, prevHash := persist.GetCurrentBlockInfo(pbft.namespace)
	return &protos.BlockchainInfo{
		Height:            height,
		CurrentBlockHash:  curHash,
		PreviousBlockHash: prevHash,
	}
}

func (pbft *pbftImpl) getGenesisInfo() uint64 {
	_, genesis := persist.GetGenesisOfChain(pbft.namespace)
	return genesis
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
		pbft.softStartNewViewTimer(pbft.timerMgr.requestTimeout, fmt.Sprintf("outstanding request batches num=%v", len(getOutstandingDigests)))
	} else if pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER) > 0 {
		pbft.nullReqTimerReset()
	}
}

func (pbft *pbftImpl) nullReqTimerReset() {
	timeout := pbft.timerMgr.getTimeoutValue(NULL_REQUEST_TIMER)
	if pbft.primary(pbft.view) != pbft.id {
		// we're waiting for the primary to deliver a null request - give it a bit more time
		timeout = 3*timeout + pbft.timerMgr.requestTimeout
	}

	event := &LocalEvent{
		Service:   CORE_PBFT_SERVICE,
		EventType: CORE_NULL_REQUEST_TIMER_EVENT,
	}

	//pbft.logger.Errorf("replica: %d, primary: %d, reset null request timeout to %v", pbft.id, pbft.primary(pbft.view), timeout)

	pbft.timerMgr.startTimerWithNewTT(NULL_REQUEST_TIMER, timeout, event, pbft.pbftEventQueue)
}

//stopFirstRequestTimer
func (pbft *pbftImpl) stopFirstRequestTimer() {
	if ok, _ := pbft.isPrimary(); !ok {
		pbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
	}
}

// =============================================================================
// helper functions for validateState
// =============================================================================
// invalidateState is invoked to tell us that consensus realizes the ledger is out of sync
func (pbft *pbftImpl) invalidateState() {
	pbft.logger.Debug("Invalidating the current state")
	pbft.status.inActiveState(&pbft.status.valid)
}

// validateState is invoked to tell us that consensus has the ledger back in sync
func (pbft *pbftImpl) validateState() {
	pbft.logger.Debug("Validating the current state")
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
		pbft.logger.Debugf("Replica %d try recvPrePrepare, but it's in nego-view", pbft.id)
		return false
	}

	if atomic.LoadUint32(&pbft.activeView) == 0 {
		pbft.logger.Debugf("Replica %d ignoring pre-prepare as we are in view change", pbft.id)
		return false
	}

	if pbft.primary(pbft.view) != preprep.ReplicaId {
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
			pbft.logger.Debugf("Replica %d pre-prepare view different, or sequence number outside watermarks: "+
				"preprep.View %d, expected.View %d, seqNo %d, low-mark %d",
				pbft.id, preprep.View, pbft.view, preprep.SequenceNumber, pbft.h)
		}
		return false
	}
	return true
}

//isPrepareLegal both PBFT state Prepare message are legal.
func (pbft *pbftImpl) isPrepareLegal(prep *Prepare) bool {
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvPrepare, but it's in nego-view", pbft.id)
		return false
	}
	//TODO: need !pbft.status[IN_RECOVERY] ?
	if pbft.primary(prep.View) == prep.ReplicaId && !pbft.status.getState(&pbft.status.inRecovery) {
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

//isPrepareLegal both PBFT state Commit message are legal.
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

func (pbft *pbftImpl) parseCertStore() {
	tmpStore := make(map[msgID]*msgCert)
	for idx := range pbft.storeMgr.certStore {
		cert := pbft.storeMgr.getCert(idx.v, idx.n)
		if cert.prePrepare != nil {
			cert.prePrepare.View = pbft.view
		}
		for prep := range cert.prepare {
			prep.View = pbft.view
		}
		for cmt := range cert.commit {
			cmt.View = pbft.view
		}
		idx.v = pbft.view
		tmpStore[msgID{pbft.view, idx.n}] = cert
	}
	pbft.storeMgr.certStore = tmpStore
}
