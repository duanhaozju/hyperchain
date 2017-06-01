//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"hyperchain/consensus/events"
	"github.com/golang/protobuf/proto"
	"time"

	"hyperchain/common"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"sync/atomic"
)

//view change manager
type vcManager struct {
	vcResendLimit      int                       // vcResendLimit indicates a replica's view change resending upbound.
	vcResendCount      int                       // vcResendCount represent times of same view change info resend
	viewChangePeriod   uint64                    // period between automatic view changes
	viewChangeSeqNo    uint64                    // next seqNo to perform view change TODO: NO usage
	lastNewViewTimeout time.Duration             // last timeout we used during this view change
	newViewTimerReason string                    // what triggered the timer

	qlist		   map[qidx]*ViewChange_PQ
	plist		   map[uint64]*ViewChange_PQ

	newViewStore       map[uint64]*NewView       	// track last new-view we received or sent
	viewChangeStore    map[vcidx]*ViewChange     	// track view-change messages
	vcResetStore       map[FinishVcReset]bool 	// track vcReset message from others

}

//dispatchViewChangeMsg dispatch view change consensus messages from other peers.
func (pbft *pbftImpl) dispatchViewChangeMsg(e events.Event) events.Event {
	switch et := e.(type) {
	case *ViewChange:
		return pbft.recvViewChange(et)
	case *NewView:
		return pbft.recvNewView(et)
	case *FetchRequestBatch:
		return pbft.recvFetchRequestBatch(et)
	case *ReturnRequestBatch:
		return pbft.recvReturnRequestBatch(et)
	case *FinishVcReset:
		if atomic.LoadUint32(&pbft.activeView) == 0 {
			return pbft.recvFinishVcReset(et)
		}
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1  {
			return pbft.recvFinishUpdate(et)
		}
		return nil
	}
	return nil
}

//newVcManager init a instance of view change manager.
func newVcManager(pbftTm *timerManager, pbft *pbftImpl, conf *common.Config) *vcManager {
	vcm := &vcManager{}
	//vcm.pset = make(map[uint64]*ViewChange_PQ)
	//vcm.qset = make(map[qidx]*ViewChange_PQ)
	vcm.vcResetStore = make(map[FinishVcReset]bool)

	vcm.qlist = make(map[qidx]*ViewChange_PQ)
	vcm.plist = make(map[uint64]*ViewChange_PQ)
	vcm.newViewStore = make(map[uint64]*NewView)
	vcm.viewChangeStore = make(map[vcidx]*ViewChange)

	vcm.viewChangePeriod = uint64(0)

	if vcm.viewChangePeriod > 0 {
		pbft.logger.Infof("PBFT view change period = %v", vcm.viewChangePeriod)
	} else {
		pbft.logger.Infof("PBFT automatic view change disabled")
	}

	vcm.lastNewViewTimeout = pbftTm.getTimeoutValue(NEW_VIEW_TIMER)

	// vcResendLimit
	vcm.vcResendLimit = conf.GetInt(PBFT_VC_RESEND_LIMIT)
	pbft.logger.Debugf("Replica %d set vcResendLimit %d", pbft.id, vcm.vcResendLimit)
	vcm.vcResendCount = 0

	return vcm
}

//calcQSet
func (pbft *pbftImpl) calcQSet() map[qidx]*ViewChange_PQ {
	qset := make(map[qidx]*ViewChange_PQ)

	for n, q := range pbft.vcMgr.qlist {
		qset[n] = q
	}

	for idx, cert := range pbft.storeMgr.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prePrepared(digest, idx.v, idx.n) {
			continue
		}

		qi := qidx{digest, idx.n}
		if q, ok := qset[qi]; ok && q.View > idx.v {
			continue
		}

		qset[qi] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return qset
}

//calcQSet
func (pbft *pbftImpl) calcPSet() map[uint64]*ViewChange_PQ {
	pset := make(map[uint64]*ViewChange_PQ)

	for n, p := range pbft.vcMgr.plist {
		pset[n] = p
	}

	for idx, cert := range pbft.storeMgr.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prepared(digest, idx.v, idx.n) {
			continue
		}

		if p, ok := pset[idx.n]; ok && p.View > idx.v {
			continue
		}

		pset[idx.n] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return pset
}

//sendViewChange send view change msg
func (pbft *pbftImpl) sendViewChange() events.Event {

	err := pbft.beforeSendVC()
	if err != nil {
		pbft.logger.Error(err.Error())
		return nil
	}

	vc := &ViewChange{
		View:      pbft.view,
		H:         pbft.h,
		ReplicaId: pbft.id,
	}

	for n, id := range pbft.storeMgr.chkpts {
		vc.Cset = append(vc.Cset, &ViewChange_C{
			SequenceNumber: n,
			Id:             id,
		})
	}

	for _, p := range pbft.vcMgr.plist {
		if p.SequenceNumber < pbft.h {
			pbft.logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", pbft.id, p)
		}
		vc.Pset = append(vc.Pset, p)
	}

	for _, q := range pbft.vcMgr.qlist {
		if q.SequenceNumber < pbft.h {
			pbft.logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", pbft.id, q)
		}
		vc.Qset = append(vc.Qset, q)
	}

	pbft.logger.Debugf("Replica %d sending view-change, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	payload, err := proto.Marshal(vc)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_VIEW_CHANGE Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_VIEW_CHANGE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_RESEND_TIMER_EVENT,
	}

	pbft.pbftTimerMgr.startTimer(VC_RESEND_TIMER, event, pbft.pbftEventQueue)
	return pbft.recvViewChange(vc)
}

//recvViewChange
func (pbft *pbftImpl) recvViewChange(vc *ViewChange) events.Event {
	pbft.logger.Infof("Replica %d received view-change from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		pbft.id, vc.ReplicaId, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvViewChange, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to recvcViewChange, but it's in recovery", pbft.id)
		return nil
	}

	if vc.View < pbft.view {
		pbft.logger.Warningf("Replica %d found view-change message for old view from replica %d", pbft.id, vc.ReplicaId)
		return nil
	}

	if !pbft.correctViewChange(vc) {
		pbft.logger.Warningf("Replica %d found view-change message incorrect", pbft.id)
		return nil
	}

	// record same vc from self times
	if vc.ReplicaId == pbft.id {
		pbft.vcMgr.vcResendCount++
		pbft.logger.Warningf("======== Replica %d already recv view change from itself for %d times", pbft.id, pbft.vcMgr.vcResendCount)
	}

	if _, ok := pbft.vcMgr.viewChangeStore[vcidx{vc.View, vc.ReplicaId}]; ok {
		pbft.logger.Warningf("Replica %d already has a view change message"+
			" for view %d from replica %d", pbft.id, vc.View, vc.ReplicaId)

		if pbft.vcMgr.vcResendCount >= pbft.vcMgr.vcResendLimit {
			pbft.logger.Noticef("Replica %d view change resend reach upbound, try to recovery", pbft.id)
			pbft.pbftTimerMgr.stopTimer(VC_RESEND_TIMER)
			pbft.vcMgr.vcResendCount = 0
			pbft.status.activeState(&pbft.status.inNegoView, &pbft.status.inRecovery)
			atomic.StoreUint32(&pbft.activeView, 1)
			pbft.processNegotiateView()
		}

		return nil
	}

	pbft.vcMgr.viewChangeStore[vcidx{vc.View, vc.ReplicaId}] = vc

	// PBFT TOCS 4.5.1 Liveness: "if a replica receives a set of
	// f+1 valid VIEW-CHANGE messages from other replicas for
	// views greater than its current view, it sends a VIEW-CHANGE
	// message for the smallest view in the set, even if its timer
	// has not expired"
	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range pbft.vcMgr.viewChangeStore {
		if idx.v <= pbft.view {
			continue
		}

		replicas[idx.id] = true
		if minView == 0 || idx.v < minView {
			minView = idx.v
		}
	}

	// We only enter this if there are enough view change messages greater than our current view
	if len(replicas) >= pbft.f + 1 {
		pbft.logger.Warningf("Replica %d received f+1 view-change messages, triggering view-change to view %d",
			pbft.id, minView)
		pbft.pbftTimerMgr.stopTimer(FIRST_REQUEST_TIMER)
		// subtract one, because sendViewChange() increments
		pbft.view = minView - 1
		return pbft.sendViewChange()
	}

	quorum := 0
	for idx := range pbft.vcMgr.viewChangeStore {
		if idx.v == pbft.view {
			quorum++
		}
	}
	pbft.logger.Debugf("Replica %d now has %d view change requests for view %d", pbft.id, quorum, pbft.view)

	if atomic.LoadUint32(&pbft.activeView) == 0 && vc.View == pbft.view && quorum >= pbft.allCorrectReplicasQuorum() {
		pbft.pbftTimerMgr.stopTimer(VC_RESEND_TIMER)
		pbft.startNewViewTimer(pbft.vcMgr.lastNewViewTimeout, "new view change")
		pbft.vcMgr.lastNewViewTimeout = 2 * pbft.vcMgr.lastNewViewTimeout
		return &LocalEvent{
			Service:VIEW_CHANGE_SERVICE,
			EventType:VIEW_CHANGE_QUORUM_EVENT,
		}
	}

	return nil
}

//sendNewView
func (pbft *pbftImpl) sendNewView() events.Event {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to sendNewView, but it's in nego-view", pbft.id)
		return nil
	}

	if _, ok := pbft.vcMgr.newViewStore[pbft.view]; ok {
		pbft.logger.Debugf("Replica %d already has new view in store for view %d, skipping", pbft.id, pbft.view)
		return nil
	}

	vset := pbft.getViewChanges()

	cp, ok, replicas := pbft.selectInitialCheckpoint(vset)

	if !ok {
		pbft.logger.Infof("Replica %d could not find consistent checkpoint: %+v", pbft.id, pbft.vcMgr.viewChangeStore)
		return nil
	}

	msgList := pbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	if msgList == nil {
		pbft.logger.Infof("Replica %d could not assign sequence numbers for new view", pbft.id)
		return nil
	}

	nv := &NewView{
		View:      pbft.view,
		Vset:      vset,
		Xset:      msgList,
		ReplicaId: pbft.id,
	}

	pbft.logger.Infof("Replica %d is new primary, sending new-view, v:%d, X:%+v",
		pbft.id, nv.View, nv.Xset)
	payload, err := proto.Marshal(nv)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_NEW_VIEW Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEW_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	pbft.helper.InnerBroadcast(msg)
	pbft.vcMgr.newViewStore[pbft.view] = nv
	pbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
	return pbft.primaryProcessNewView(cp, replicas, nv)
}

//recvNewView
func (pbft *pbftImpl) recvNewView(nv *NewView) events.Event {
	pbft.logger.Infof("Replica %d received new-view %d",
		pbft.id, nv.View)

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvNewView, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to recvNewView, but it's in recovery", pbft.id)
		pbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	if !(nv.View > 0 && nv.View >= pbft.view && pbft.primary(nv.View) == nv.ReplicaId && pbft.vcMgr.newViewStore[nv.View] == nil) {
		pbft.logger.Infof("Replica %d rejecting invalid new-view from %d, v:%d",
			pbft.id, nv.ReplicaId, nv.View)
		return nil
	}

	pbft.vcMgr.newViewStore[nv.View] = nv

	quorum := 0
	for idx := range pbft.vcMgr.viewChangeStore {
		if idx.v == pbft.view {
			quorum++
		}
	}
	if quorum < pbft.allCorrectReplicasQuorum() {
		pbft.logger.Warningf("Replica %d has not meet ViewChangedQuorum", pbft.id)
		return nil
	}

	return pbft.processNewView()
}

//processNewView
func (pbft *pbftImpl) processNewView() events.Event {
	nv, ok := pbft.vcMgr.newViewStore[pbft.view]
	if !ok {
		pbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", pbft.id, pbft.view)
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 1 {
		pbft.logger.Infof("Replica %d ignoring new-view from %d, v:%d: we are active in view %d",
			pbft.id, nv.ReplicaId, nv.View, pbft.view)
		return nil
	}

	cp, ok, replicas := pbft.selectInitialCheckpoint(nv.Vset)
	if !ok {
		pbft.logger.Warningf("Replica %d could not determine initial checkpoint: %+v",
			pbft.id, pbft.vcMgr.viewChangeStore)
		return pbft.sendViewChange()
	}
	// 以上 primary 不必做
	speculativeLastExec := pbft.exec.lastExec
	if pbft.exec.currentExec != nil {
		speculativeLastExec = *pbft.exec.currentExec
	}

	// If we have not reached the sequence number, check to see if we can reach it without state transfer
	// In general, executions are better than state transfer
	if speculativeLastExec < cp.SequenceNumber {
		if pbft.canExecuteToTarget(speculativeLastExec, cp) {
			return nil
		}
	}
	// --
	msgList := pbft.assignSequenceNumbers(nv.Vset, cp.SequenceNumber)

	if msgList == nil {
		pbft.logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			pbft.id, pbft.vcMgr.viewChangeStore)
		return pbft.sendViewChange()
	}

	if !(len(msgList) == 0 && len(nv.Xset) == 0) && !reflect.DeepEqual(msgList, nv.Xset) {
		pbft.logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			pbft.id, msgList, nv.Xset)
		return pbft.sendViewChange()
	}
	// -- primary 不必做
	if pbft.h < cp.SequenceNumber {
		pbft.moveWatermarks(cp.SequenceNumber)
	}

	if speculativeLastExec < cp.SequenceNumber {
		pbft.logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, cp.SequenceNumber, cp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(cp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, cp.Id)
			pbft.logger.Error(err.Error())
			return nil
		}

		target := &stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: cp.SequenceNumber,
				id:    snapshotID,
			},
			replicas: replicas,
		}

		pbft.updateHighStateTarget(target)
		pbft.stateTransfer(target)
	}

	return pbft.processReqInNewView(nv)
}

//primaryProcessNewView
func (pbft *pbftImpl) primaryProcessNewView(initialCp ViewChange_C, replicas []replicaInfo, nv *NewView) events.Event {
	var newReqBatchMissing bool

	speculativeLastExec := pbft.exec.lastExec
	if pbft.exec.currentExec != nil {
		speculativeLastExec = *pbft.exec.currentExec
	}
	// If we have not reached the sequence number, check to see if we can reach it without state transfer
	// In general, executions are better than state transfer
	if speculativeLastExec < initialCp.SequenceNumber {
		if pbft.canExecuteToTarget(speculativeLastExec, initialCp) {
			return nil
		}
	}

	if pbft.h < initialCp.SequenceNumber {
		pbft.moveWatermarks(initialCp.SequenceNumber)
	}

	// true means we can not execToTarget need state transfer
	if speculativeLastExec < initialCp.SequenceNumber {
		pbft.logger.Warningf("Replica %d missing base checkpoint %d (%s), our most recent execution %d", pbft.id, initialCp.SequenceNumber, initialCp.Id, speculativeLastExec)

		snapshotID, err := base64.StdEncoding.DecodeString(initialCp.Id)
		if nil != err {
			err = fmt.Errorf("Replica %d received a view change whose hash could not be decoded (%s)", pbft.id, initialCp.Id)
			pbft.logger.Error(err.Error())
			return nil
		}

		target := &stateUpdateTarget{
			checkpointMessage: checkpointMessage{
				seqNo: initialCp.SequenceNumber,
				id:    snapshotID,
			},
			replicas: replicas,
		}

		pbft.updateHighStateTarget(target)
		pbft.stateTransfer(target)
	}

	newReqBatchMissing = pbft.feedMissingReqBatchIfNeeded(nv.Xset)

	if len(pbft.storeMgr.missingReqBatches) == 0 {
		return pbft.processReqInNewView(nv)
	} else if newReqBatchMissing {
		pbft.fetchRequestBatches()
	}

	return nil
}

func (vcm *vcManager) updateViewChangeSeqNo(seqNo, K, id uint64) {
	if vcm.viewChangePeriod <= 0 {
		return
	}
	// Ensure the view change always occurs at a checkpoint boundary
	vcm.viewChangeSeqNo = seqNo + vcm.viewChangePeriod * K - seqNo % K
	//logger.Debugf("Replica %d updating view change sequence number to %d", id, vcm.viewChangeSeqNo)
}

func (pbft *pbftImpl) canExecuteToTarget(specLastExec uint64, initialCp ViewChange_C) bool {

	canExecuteToTarget := true
outer:
	for seqNo := specLastExec + 1; seqNo <= initialCp.SequenceNumber; seqNo++ {
		found := false
		for idx, cert := range pbft.storeMgr.certStore {
			if idx.n != seqNo {
				continue
			}

			quorum := 0
			for p := range cert.commit {
				// Was this committed in the previous view
				if p.View == idx.v && p.SequenceNumber == seqNo {
					quorum++
				}
			}

			if quorum < pbft.intersectionQuorum() {
				pbft.logger.Debugf("Replica %d missing quorum of commit certificate for seqNo=%d, only has %d of %d", pbft.id, quorum, pbft.intersectionQuorum())
				continue
			}

			found = true
			break
		}

		if !found {
			canExecuteToTarget = false
			pbft.logger.Debugf("Replica %d missing commit certificate for seqNo=%d", pbft.id, seqNo)
			break outer
		}

	}

	if canExecuteToTarget {
		pbft.nvInitialSeqNo = initialCp.SequenceNumber
		pbft.logger.Debugf("Replica %d needs to process a new view, but can execute to the checkpoint seqNo %d, delaying processing of new view", pbft.id, initialCp.SequenceNumber)
	} else {
		pbft.nvInitialSeqNo = 0
		pbft.logger.Infof("Replica %d cannot execute to the view change checkpoint with seqNo %d", pbft.id, initialCp.SequenceNumber)
	}
	return canExecuteToTarget
}

func (pbft *pbftImpl) feedMissingReqBatchIfNeeded(xset Xset) (newReqBatchMissing bool) {
	newReqBatchMissing = false
	for n, d := range xset {
		// PBFT: why should we use "h ≥ min{n | ∃d : (<n,d> ∈ X)}"?
		// "h ≥ min{n | ∃d : (<n,d> ∈ X)} ∧ ∀<n,d> ∈ X : (n ≤ h ∨ ∃m ∈ in : (D(m) = d))"
		if n <= pbft.h {
			continue
		} else {
			if d == "" {
				// NULL request; skip
				continue
			}

			if !pbft.batchVdr.containsInVBS(d) {
				pbft.logger.Warningf("Replica %d missing assigned, non-checkpointed request batch %s",
					pbft.id, d)
				if _, ok := pbft.storeMgr.missingReqBatches[d]; !ok {
					pbft.logger.Warningf("Replica %v requesting to fetch batch %s",
						pbft.id, d)
					newReqBatchMissing = true
					pbft.storeMgr.missingReqBatches[d] = true
				}
			}
		}
	}
	return newReqBatchMissing
}

func (pbft *pbftImpl) processReqInNewView(nv *NewView) events.Event {
	pbft.logger.Debugf("Replica %d accepting new-view to view %d", pbft.id, pbft.view)

	if pbft.status.getState(&pbft.status.vcHandled) {
		pbft.logger.Debugf("Replica %d repeated enter processReqInNewView, ignore it")
		return nil
	}
	pbft.status.activeState(&pbft.status.vcHandled)

	pbft.stopNewViewTimer()
	pbft.pbftTimerMgr.stopTimer(NULL_REQUEST_TIMER)

	delete(pbft.vcMgr.newViewStore, pbft.view - 1)
	// empty the outstandingReqBatch, it is useless since new primary will resend pre-prepare
	pbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
	pbft.exec.lastExec = pbft.h
	pbft.seqNo = pbft.h
	prevPrimary := pbft.primary(pbft.view - 1)
	if prevPrimary == pbft.id {
		pbft.rebuildDuplicator()
		if len(pbft.batchMgr.batchStore) > 0 {
			for _, tx := range pbft.batchMgr.batchStore {
				go pbft.reqEventQueue.Push(tx)
			}
			pbft.batchMgr.batchStore = nil
		}
	} else {
		pbft.clearDuplicator()
	}
	pbft.batchVdr.setVid(pbft.h)
	pbft.batchVdr.setLastVid(pbft.h)

	if !pbft.status.getState(&pbft.status.skipInProgress) &&
		!pbft.status.getState(&pbft.status.inVcReset) {
		backendVid := uint64(pbft.batchVdr.vid + 1)
		pbft.helper.VcReset(backendVid)
		pbft.status.activeState(&pbft.status.inVcReset)
		return nil
	}

	return &LocalEvent{
		Service:VIEW_CHANGE_SERVICE,
		EventType:VIEW_CHANGED_EVENT,
	}
}

func (pbft *pbftImpl) recvFinishVcReset(finish *FinishVcReset) events.Event {

	if pbft.primary(pbft.view) != pbft.id {
		pbft.logger.Warningf("Replica %d is not primary, but received others FinishVcReset", pbft.id)
		return nil
	}

	if atomic.LoadUint32(&pbft.activeView) == 1 {
		pbft.logger.Warningf("Replica %d is not in viewChange, but received others FinishVcReset", pbft.id)
		return nil
	}

	if finish.View != pbft.view || finish.LowH != pbft.h {
		pbft.logger.Warningf("Replica %d received finishVcReset from replica %d, expect view=%d/h=%d, but get view=%d/h=%d",
			pbft.id, finish.ReplicaId, pbft.view, pbft.h, finish.View, finish.LowH)
		return nil
	}

	pbft.logger.Debugf("Primary %d received FinishVcReset from replica %d, view=%d/h=%d",
		pbft.id, finish.ReplicaId, finish.View, finish.LowH)
	ok := pbft.vcMgr.vcResetStore[*finish]
	if ok {
		pbft.logger.Warningf("Replica %d ignored duplicate agree FinishVcReset from %d", pbft.id, finish.ReplicaId)
		return nil
	}
	pbft.vcMgr.vcResetStore[*finish] = true

	return pbft.handleTailInNewView()
}

func (pbft *pbftImpl) handleTailInNewView() events.Event {

	if atomic.LoadUint32(&pbft.activeView) == 1 {
		pbft.logger.Debugf("Replica %d in active view, ignore handleTail request", pbft.id)
		return nil
	}

	if len(pbft.vcMgr.vcResetStore) < pbft.allCorrectReplicasQuorum()-1 {
		return nil
	}

	if pbft.status.getState(&pbft.status.inVcReset) {
		pbft.logger.Debugf("Primary %d itself has not done with vcReset", pbft.id)
		return nil
	}

	nv, ok := pbft.vcMgr.newViewStore[pbft.view]
	if !ok {
		pbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", pbft.id, pbft.view)
		return nil
	}

	atomic.StoreUint32(&pbft.activeView, 1)

	xSetLen := len(nv.Xset)
	upper := uint64(xSetLen) + pbft.h + uint64(1)
	for i := pbft.h + uint64(1); i < upper; i++ {
		d, ok := nv.Xset[i]
		if !ok {
			pbft.logger.Critical("view change Xset miss batch number %d", i)
		} else if d == "" {
			// This should not happen
			pbft.logger.Critical("view change Xset has null batch, kick it out")
		} else {
			if !pbft.batchVdr.containsInVBS(d) {
				pbft.logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, pbft.id)
			} else {
				pbft.logger.Critical("send validate")
				pbft.primaryValidateBatch(pbft.batchVdr.getTxBatchFromVBS(d), i)
			}
		}
	}

	return &LocalEvent{
		Service:VIEW_CHANGE_SERVICE,
		EventType:VIEW_CHANGED_EVENT,
	}
}

func (pbft *pbftImpl) finishViewChange() events.Event {

	finish := &FinishVcReset{
		ReplicaId: pbft.id,
		View:      pbft.view,
		LowH:      pbft.h,
	}
	payload, err := proto.Marshal(finish)
	if err != nil {
		pbft.logger.Errorf("Marshal FinishVcReset Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_FINISH_VCRESET,
		Payload: payload,
	}

	msg := cMsgToPbMsg(consensusMsg, pbft.id)
	primary := pbft.primary(pbft.view)
	pbft.helper.InnerUnicast(msg, primary)
	pbft.logger.Debugf("Replica %d send FinishVcReset to primary %d", pbft.id, primary)
	return &LocalEvent{
		Service:VIEW_CHANGE_SERVICE,
		EventType:VIEW_CHANGED_EVENT,
	}
}

func (pbft *pbftImpl) getViewChanges() (vset []*ViewChange) {
	for _, vc := range pbft.vcMgr.viewChangeStore {
		vset = append(vset, vc)
	}
	return
}

func (pbft *pbftImpl) selectInitialCheckpoint(vset []*ViewChange) (checkpoint ViewChange_C, ok bool, replicas []replicaInfo) {
	checkpoints := make(map[ViewChange_C][]*ViewChange)
	for _, vc := range vset {
		for _, c := range vc.Cset {
			// TODO, verify that we strip duplicate checkpoints from this set
			checkpoints[*c] = append(checkpoints[*c], vc)
			pbft.logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s",
				pbft.id, vc.ReplicaId, vc.H, c.SequenceNumber, c.Id)
		}
	}

	if len(checkpoints) == 0 {
		pbft.logger.Debugf("Replica %d has no checkpoints to select from: %d %s",
			pbft.id, len(pbft.vcMgr.viewChangeStore), checkpoints)
		return
	}

	for idx, vcList := range checkpoints {
		// need weak certificate for the checkpoint
		if len(vcList) <= pbft.f {
			// type casting necessary to match types
			pbft.logger.Debugf("Replica %d has no weak certificate for n:%d, vcList was %d long",
				pbft.id, idx.SequenceNumber, len(vcList))
			continue
		}

		quorum := 0
		// Note, this is the whole vset (S) in the paper, not just this checkpoint set (S') (vcList)
		// We need 2f+1 low watermarks from S below this seqNo from all replicas
		// We need f+1 matching checkpoints at this seqNo (S')
		for _, vc := range vset {
			if vc.H <= idx.SequenceNumber {
				quorum++
			}
		}

		if quorum < pbft.intersectionQuorum() {
			pbft.logger.Debugf("Replica %d has no quorum for n:%d", pbft.id, idx.SequenceNumber)
			continue
		}

		if checkpoint.SequenceNumber <= idx.SequenceNumber {
			replicas = make([]replicaInfo, len(vcList))
			for i, vc := range vcList {
				replicas[i] = replicaInfo{
					id: vc.ReplicaId,
					height: vc.H,
					genesis: vc.Genesis,
				}
			}

			checkpoint = idx
			ok = true
		}
	}

	return
}

func (pbft *pbftImpl) assignSequenceNumbers(vset []*ViewChange, h uint64) (msgList map[uint64]string) {
	msgList = make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
nLoop:
	for n := h + 1; n <= h + pbft.L; n++ {
		// "∃m ∈ S..."
		for _, m := range vset {
			// "...with <n,d,v> ∈ m.P"
			for _, em := range m.Pset {
				quorum := 0
				// "A1. ∃2f+1 messages m' ∈ S"
			mpLoop:
				for _, mp := range vset {
					if mp.H >= n {
						continue
					}
					// "∀<n,d',v'> ∈ m'.P"
					for _, emp := range mp.Pset {
						if n == emp.SequenceNumber && !(emp.View < em.View || (emp.View == em.View && emp.BatchDigest == em.BatchDigest)) {
							continue mpLoop
						}
					}
					quorum++
				}

				if quorum < pbft.intersectionQuorum() {
					continue
				}

				quorum = 0
				// "A2. ∃f+1 messages m' ∈ S"
				for _, mp := range vset {
					// "∃<n,d',v'> ∈ m'.Q"
					for _, emp := range mp.Qset {
						if n == emp.SequenceNumber && emp.View >= em.View && emp.BatchDigest == em.BatchDigest {
							quorum++
						}
					}
				}

				if quorum < pbft.f + 1 {
					continue
				}

				// "then select the request with digest d for number n"
				msgList[n] = em.BatchDigest
				maxN = n

				continue nLoop
			}
		}

		quorum := 0
		// "else if ∃2f+1 messages m ∈ S"
	nullLoop:
		for _, m := range vset {
			// "m.P has no entry"
			for _, em := range m.Pset {
				if em.SequenceNumber == n {
					continue nullLoop
				}
			}
			quorum++
		}

		if quorum >= pbft.intersectionQuorum() {
			// "then select the null request for number n"
			msgList[n] = ""

			continue nLoop
		}

		pbft.logger.Warningf("Replica %d could not assign value to contents of seqNo %d, found only %d missing P entries", pbft.id, n, quorum)
		return nil
	}

	// prune top null requests
	for n, msg := range msgList {
		if n > maxN || msg == "" {
			delete(msgList, n)
		}
	}
	return
}

func (pbft *pbftImpl) recvFetchRequestBatch(fr *FetchRequestBatch) (err error) {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvFetchRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}

	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to recvFetchRequestBatch, but it's in recovery", pbft.id)
		return nil
	}

	digest := fr.BatchDigest
	if !pbft.batchVdr.containsInVBS(digest) {
		return nil // we don't have it either
	}

	reqBatch := pbft.batchVdr.getTxBatchFromVBS(digest)
	batch := &ReturnRequestBatch{
		Batch:  reqBatch,
		Digest: digest,
	}
	payload, err := proto.Marshal(batch)
	if err != nil {
		pbft.logger.Errorf("ConsensusMessage_RETURN_REQUEST_BATCH Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RETURN_REQUEST_BATCH,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, pbft.id)

	receiver := fr.ReplicaId
	err = pbft.helper.InnerUnicast(msg, receiver)

	return
}

func (pbft *pbftImpl) recvReturnRequestBatch(batch *ReturnRequestBatch) events.Event {

	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to recvReturnRequestBatch, but it's in nego-view", pbft.id)
		return nil
	}
	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to recvReturnRequestBatch, but it's in recovery", pbft.id)
		return nil
	}

	digest := batch.Digest
	if _, ok := pbft.storeMgr.missingReqBatches[digest]; !ok {
		return nil // either the wrong digest, or we got it already from someone else
	}
	pbft.batchVdr.saveToVBS(digest, batch.Batch)
	delete(pbft.storeMgr.missingReqBatches, digest)
	pbft.logger.Warning("Primary received missing request: ", digest)

	if len(pbft.storeMgr.missingReqBatches) == 0 {
		if atomic.LoadUint32(&pbft.activeView) == 0 {
			nv, ok := pbft.vcMgr.newViewStore[pbft.view]
			if !ok {
				pbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", pbft.id, pbft.view)
				return nil
			}
			return pbft.processReqInNewView(nv)
		}
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
			update, ok := pbft.nodeMgr.updateStore[pbft.nodeMgr.updateTarget]
			if !ok {
				pbft.logger.Debugf("Replica %d ignoring processUpdateN as it could not find target %v in its updateStore", pbft.id, pbft.nodeMgr.updateTarget)
				return nil
			}
			return pbft.processReqInUpdate(update)
		}
	}
	return nil

}

//##########################################################################
//           view change auxiliary functions
//##########################################################################

//stopNewViewTimer
func (pbft *pbftImpl) stopNewViewTimer() {
	pbft.logger.Debugf("Replica %d stopping a running new view timer", pbft.id)
	pbft.status.inActiveState(&pbft.status.timerActive)
	pbft.pbftTimerMgr.stopTimer(NEW_VIEW_TIMER)
}

//startNewViewTimer stop all running new view timers and  start a new view timer
func (pbft *pbftImpl) startNewViewTimer(timeout time.Duration, reason string) {
	pbft.logger.Debugf("Replica %d starting new view timer for %s: %s", pbft.id, timeout, reason)
	pbft.vcMgr.newViewTimerReason = reason
	pbft.status.activeState(&pbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	pbft.pbftTimerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, pbft.pbftEventQueue)
}

//softstartNewViewTimer start a new view timer no matter how many existed new view timer
func (pbft *pbftImpl) softStartNewViewTimer(timeout time.Duration, reason string) {
	pbft.logger.Debugf("Replica %d soft starting new view timer for %s: %s", pbft.id, timeout, reason)
	pbft.vcMgr.newViewTimerReason = reason
	pbft.status.activeState(&pbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	pbft.pbftTimerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, pbft.pbftEventQueue)
}

//correctViewChange
func (pbft *pbftImpl) correctViewChange(vc *ViewChange) bool {
	for _, p := range append(vc.Pset, vc.Qset...) {
		if !(p.View < vc.View && p.SequenceNumber > vc.H && p.SequenceNumber <= vc.H + pbft.L) {
			pbft.logger.Debugf("Replica %d invalid p entry in view-change: vc(v:%d h:%d) p(v:%d n:%d)",
				pbft.id, vc.View, vc.H, p.View, p.SequenceNumber)
			return false
		}
	}

	for _, c := range vc.Cset {
		if !(c.SequenceNumber >= vc.H && c.SequenceNumber <= vc.H + pbft.L) {
			pbft.logger.Debugf("Replica %d invalid c entry in view-change: vc(v:%d h:%d) c(n:%d)",
				pbft.id, vc.View, vc.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

//beforeSendVC operations before send view change
func (pbft *pbftImpl) beforeSendVC() error{
	if pbft.status.getState(&pbft.status.inNegoView) {
		pbft.logger.Debugf("Replica %d try to send view change, but it's in nego-view", pbft.id)
		return errors.New("node is in nego view now!")
	}

	if pbft.status.getState(&pbft.status.inRecovery) {
		pbft.logger.Noticef("Replica %d try to send view change, but it's in recovery", pbft.id)
		return errors.New("node is in recovery now!")
	}

	pbft.stopNewViewTimer()

	delete(pbft.vcMgr.newViewStore, pbft.view)
	pbft.view ++
	atomic.StoreUint32(&pbft.activeView, 0)

	pbft.vcMgr.plist = pbft.calcPSet()
	pbft.vcMgr.qlist = pbft.calcQSet()

	// clear old messages
	for idx := range pbft.storeMgr.certStore {
		if idx.v < pbft.view {
			delete(pbft.storeMgr.certStore, idx)
			pbft.persistDelQPCSet(idx.v, idx.n)
		}
	}
	for idx := range pbft.vcMgr.viewChangeStore {
		if idx.v < pbft.view {
			delete(pbft.vcMgr.viewChangeStore, idx)
		}
	}
	return nil
}
