package pbft

import (
	"time"
	"fmt"

	"hyperchain/protos"
	"hyperchain/core/types"
	"hyperchain/consensus/helper/persist"

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
// helper functions for create batch
// =============================================================================

func (pbft *pbftProtocal) postRequestEvent(event *types.Transaction) {

	pbft.muxBatch.Lock()
	defer pbft.muxBatch.Unlock()
	pbft.batchManager.Queue() <- event

}

func (pbft *pbftProtocal) postPbftEvent(event interface{}) {

	pbft.muxPbft.Lock()
	defer pbft.muxPbft.Unlock()
	pbft.pbftManager.Queue() <- event

}

// =============================================================================
// helper functions for PBFT
// =============================================================================

// Given a certain view v and replicaCount n, what is the expected primary?
func (pbft *pbftProtocal) primary(v uint64, n uint64) uint64 {
	return (v % uint64(n) + 1)
}

// Is the sequence number between watermarks?
func (pbft *pbftProtocal) inW(n uint64) bool {
	return n > pbft.h && n-pbft.h <= pbft.L
}

// Is the view right? And is the sequence number between watermarks?
func (pbft *pbftProtocal) inWV(v uint64, n uint64) bool {
	return pbft.view == v && pbft.inW(n)
}

// Given a digest/view/seq, is there an entry in the certLog?
// If so, return it. If not, create it.
func (pbft *pbftProtocal) getCert(v uint64, n uint64) (cert *msgCert) {

	idx := msgID{v, n}
	cert, ok := pbft.certStore[idx]

	if ok {
		return
	}

	prepare := make(map[Prepare]bool)
	commit := make(map[Commit]bool)
	cert = &msgCert{
		prepare:	prepare,
		commit:		commit,
	}
	pbft.certStore[idx] = cert

	return
}

// Given a seqNo/id get the checkpoint Cert
func (pbft *pbftProtocal) getChkptCert(n uint64, id string) (cert *chkptCert) {

	idx := chkptID{n, id}
	cert, ok := pbft.chkptCertStore[idx]

	if ok {
		return
	}

	chkpts := make(map[Checkpoint]bool)
	cert = &chkptCert{
		chkpts:	chkpts,
	}
	pbft.chkptCertStore[idx] = cert

	return
}

// Given a ip/digest get the addnode Cert
func (pbft *pbftProtocal) getAddNodeCert(ip string, digest string) (cert *addNodeCert) {

	idx := addNodeID{ip, digest}
	cert, ok := pbft.addNodeCertStore[idx]

	if ok {
		return
	}

	agrees := make(map[AgreeAddNode]bool)
	cert = &addNodeCert{
		agrees:	agrees,
	}
	pbft.addNodeCertStore[idx] = cert

	return
}


// =============================================================================
// prepare/commit quorum checks helper
// =============================================================================

func (pbft *pbftProtocal) preparedReplicasQuorum() int {
	return (2 * pbft.f)
}

func (pbft *pbftProtocal) committedReplicasQuorum() int {
	return (2 * pbft.f + 1)
}

// intersectionQuorum returns the number of replicas that have to
// agree to guarantee that at least one correct replica is shared by
// two intersection quora
func (pbft *pbftProtocal) intersectionQuorum() int {
	return (pbft.N + pbft.f + 2) / 2
}

func (pbft *pbftProtocal) allCorrectReplicasQuorum() int {
	return (pbft.N - pbft.f)
}

// =============================================================================
// pre-prepare/prepare/commit check helper
// =============================================================================

func (pbft *pbftProtocal) prePrepared(digest string, v uint64, n uint64) bool {

	_, mInLog := pbft.validatedBatchStore[digest]

	if digest != "" && !mInLog {
		logger.Debugf("Replica %d havan't store the reqBatch")
		return false
	}

	//if q, ok := pbft.qset[qidx{digest, n}]; ok && q.View == v {
	//	return true
	//}

	cert := pbft.certStore[msgID{v, n}]

	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.SequenceNumber == n && p.BatchDigest == digest {
			return true
		}
	}

	logger.Debugf("Replica %d does not have view=%d/seqNo=%d pre-prepared",
		pbft.id, v, n)

	return false
}

func (pbft *pbftProtocal) prepared(digest string, v uint64, n uint64) bool {

	if !pbft.prePrepared(digest, v, n) {
		return false
	}

	//if p, ok := pbft.pset[n]; ok && p.View == v && p.BatchDigest == digest {
	//	return true
	//}

	cert := pbft.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	logger.Debugf("Replica %d prepare count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cert.prepareCount)

	return cert.prepareCount >= pbft.preparedReplicasQuorum()
}

func (pbft *pbftProtocal) committed(digest string, v uint64, n uint64) bool {

	if !pbft.prepared(digest, v, n) {
		return false
	}

	cert := pbft.certStore[msgID{v, n}]

	if cert == nil {
		return false
	}

	logger.Debugf("Replica %d commit count for view=%d/seqNo=%d: %d",
		pbft.id, v, n, cert.commitCount)

	return cert.commitCount >= pbft.committedReplicasQuorum()
}

// =============================================================================
// helper functions for transfer message
// =============================================================================
// consensusMsgHelper help convert the ConsensusMessage to pb.Message
func consensusMsgHelper(msg *ConsensusMessage, id uint64) *protos.Message {

	msgPayload, err := proto.Marshal(msg)

	if err != nil {
		logger.Errorf("ConsensusMessage Marshal Error", err)
		return nil
	}

	pbMsg := &protos.Message{
		Type:		protos.Message_CONSENSUS,
		Payload:	msgPayload,
		Timestamp:	time.Now().UnixNano(),
		Id:		id,
	}

	return pbMsg
}

// nullRequestMsgHelper help convert the nullRequestMessage to pb.Message
func nullRequestMsgHelper(id uint64) *protos.Message {
	pbMsg := &protos.Message{
		Type:  		protos.Message_NULL_REQUEST,
		Payload:        nil,
		Timestamp:	time.Now().UnixNano(),
		Id:		id,
	}

	return pbMsg
}

// StateUpdateHelper help convert checkPointInfo, blockchainInfo, replicas to pb.UpdateStateMessage
func stateUpdateHelper(myId uint64, seqNo uint64, id []byte, replicaId []uint64) *protos.UpdateStateMessage {

	stateUpdateMsg := &protos.UpdateStateMessage{
		Id:		myId,
		SeqNo:		seqNo,
		TargetId:	id,
		Replicas:	replicaId,

	}
	return stateUpdateMsg
}

func getBlockchainInfo() *protos.BlockchainInfo {

	bcInfo := persist.GetBlockchainInfo()

	height := bcInfo.Height
	curBlkHash := bcInfo.LatestBlockHash
	preBlkHash := bcInfo.ParentBlockHash

	return &protos.BlockchainInfo{
		Height:			height,
		CurrentBlockHash: 	curBlkHash,
		PreviousBlockHash: 	preBlkHash,
	}
}

// =============================================================================
// helper functions for timer
// =============================================================================
func (pbft *pbftProtocal) startBatchTimer() {
	pbft.batchTimer.Reset(pbft.batchTimeout, batchTimerEvent{})
	logger.Debugf("Replica %d started the batch timer", pbft.id)
	pbft.batchTimerActive = true
}

func (pbft *pbftProtocal) stopBatchTimer() {
	pbft.batchTimer.Stop()
	logger.Debugf("Replica %d stpbftped the batch timer", pbft.id)
	pbft.batchTimerActive = false
}

func (pbft *pbftProtocal) startTimerIfOutstandingRequests() {
	if pbft.skipInProgress || pbft.currentExec != nil {
		// Do not start the view change timer if we are executing or state transferring, these take arbitrarilly long amounts of time
		return
	}

	if len(pbft.outstandingReqBatches) > 0 {
		getOutstandingDigests := func() []string {
			var digests []string
			for digest := range pbft.outstandingReqBatches {
				digests = append(digests, digest)
			}
			return digests
		}()
		//logger.Debug(getOutstandingDigests)
		pbft.softStartTimer(pbft.requestTimeout, fmt.Sprintf("outstanding request batches %v", getOutstandingDigests))
	} else if pbft.nullRequestTimeout > 0 {
		pbft.nullReqTimerReset()
	}
}

func (pbft *pbftProtocal) nullReqTimerReset(){
	timeout := pbft.nullRequestTimeout
	if pbft.primary(pbft.view, pbft.N) != pbft.id {
		// we're waiting for the primary to deliver a null request - give it a bit more time
		timeout += pbft.requestTimeout
	}
	pbft.nullRequestTimer.Reset(timeout, nullRequestEvent{})
}

func (pbft *pbftProtocal) softStartTimer(timeout time.Duration, reason string) {
	logger.Debugf("Replica %d soft starting new view timer for %s: %s", pbft.id, timeout, reason)
	pbft.newViewTimerReason = reason
	pbft.timerActive = true
	pbft.newViewTimer.SoftReset(timeout, viewChangeTimerEvent{})
}

func (pbft *pbftProtocal) startTimer(timeout time.Duration, reason string) {
	logger.Debugf("Replica %d starting new view timer for %s: %s", pbft.id, timeout, reason)
	pbft.timerActive = true
	pbft.newViewTimer.Reset(timeout, viewChangeTimerEvent{})
}

func (pbft *pbftProtocal) stopTimer() {
	logger.Debugf("Replica %d stopping a running new view timer", pbft.id)
	pbft.timerActive = false
	pbft.newViewTimer.Stop()
}

// =============================================================================
// helper functions for validateState
// =============================================================================
// invalidateState is invoked to tell us that consensus realizes the ledger is out of sync
func (pbft *pbftProtocal) invalidateState() {
	logger.Debug("Invalidating the current state")
	pbft.valid = false
}

// validateState is invoked to tell us that consensus has the ledger back in sync
func (pbft *pbftProtocal) validateState() {
	logger.Debug("Validating the current state")
	pbft.valid = true
}