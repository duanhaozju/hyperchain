//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package rbft

import (
	"errors"
	"reflect"
	"sort"
	"sync/atomic"
	"time"


	"github.com/golang/protobuf/proto"
)

//view change manager
type vcManager struct {
	vcResendLimit      int           // vcResendLimit indicates a replica's view change resending upbound.
	vcResendCount      int           // vcResendCount represent times of same view change info resend
	viewChangePeriod   uint64        // period between automatic view changes. Default value is 0 means close automatic view changes
	viewChangeSeqNo    uint64        // next seqNo to perform view change TODO: NO usage
	lastNewViewTimeout time.Duration // last timeout we used during this view change
	newViewTimerReason string        // what triggered the timer

	qlist map[qidx]*ViewChange_PQ   //store Pre-Prepares  for view change
	plist map[uint64]*ViewChange_PQ //store Prepares for view change

	newViewStore    map[uint64]*NewView    // track last new-view we received or sent
	viewChangeStore map[vcidx]*ViewChange  // track view-change messages
	vcResetStore    map[FinishVcReset]bool // track vcReset message from others
	cleanVcTimeout  time.Duration          // how long dose view-change messages keep in viewChangeStore
}

//dispatchViewChangeMsg dispatch view change consensus messages from other peers And push them into corresponding function
func (rbft *rbftImpl) dispatchViewChangeMsg(e consensusEvent) consensusEvent {
	switch et := e.(type) {
	case *ViewChange:
		return rbft.recvViewChange(et)
	case *NewView:
		return rbft.recvNewView(et)
	case *FetchRequestBatch:
		return rbft.recvFetchRequestBatch(et)
	case *ReturnRequestBatch:
		return rbft.recvReturnRequestBatch(et)
	case *FinishVcReset:
		return rbft.recvFinishVcReset(et)
	}
	return nil
}

//newVcManager init a instance of view change manager and initialize each parameter according to the configuration file.
func newVcManager(rbft *rbftImpl) *vcManager {
	vcm := &vcManager{}

	//init vcManage maps
	vcm.vcResetStore = make(map[FinishVcReset]bool)
	vcm.qlist = make(map[qidx]*ViewChange_PQ)
	vcm.plist = make(map[uint64]*ViewChange_PQ)
	vcm.newViewStore = make(map[uint64]*NewView)
	vcm.viewChangeStore = make(map[vcidx]*ViewChange)

	// clean out-of-data view change message
	var err error
	vcm.cleanVcTimeout, err = time.ParseDuration(rbft.config.GetString(RBFT_CLEAN_VIEWCHANGE_TIMEOUT))
	if err != nil {
		rbft.logger.Criticalf("Cannot parse clean out-of-data view change message timeout: %s", err)
	}
	nvTimeout := rbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER)
	//cleanVcTimeout should more then 6* viewChange time
	if vcm.cleanVcTimeout < 6*nvTimeout {
		vcm.cleanVcTimeout = 6 * nvTimeout
		rbft.logger.Criticalf("Replica %d set timeout of cleaning out-of-time view change message to %v since it's too short", rbft.id, 6*nvTimeout)
	}

	vcm.viewChangePeriod = uint64(0)
	//automatic view changes is off by default
	if vcm.viewChangePeriod > 0 {
		rbft.logger.Infof("RBFT view change period = %v", vcm.viewChangePeriod)
	} else {
		rbft.logger.Infof("RBFT automatic view change disabled")
	}
	//if Viewchange failed,lastNewViewTimeout well increase
	vcm.lastNewViewTimeout = rbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER)

	// vcResendLimit
	vcm.vcResendLimit = rbft.config.GetInt(RBFT_VC_RESEND_LIMIT)
	rbft.logger.Debugf("Replica %d set vcResendLimit %d", rbft.id, vcm.vcResendLimit)
	vcm.vcResendCount = 0

	return vcm
}

//calcQSet
//select Pre-prepares which satisfy the following conditions
//Pre-prepares in previous qlist
//Pre-prepares from certStore which is preprepared and (its view <= its idx.v or not in qlist
func (rbft *rbftImpl) calcQSet() map[qidx]*ViewChange_PQ {
	qset := make(map[qidx]*ViewChange_PQ)

	for n, q := range rbft.vcMgr.qlist {
		qset[n] = q
	}

	for idx, cert := range rbft.storeMgr.certStore {
		if cert.prePrepare == nil {
			continue
		}

		if !rbft.prePrepared(idx.d, idx.v, idx.n) {
			continue
		}

		qi := qidx{idx.d, idx.n}
		if q, ok := qset[qi]; ok && q.View > idx.v {
			continue
		}

		qset[qi] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    idx.d,
			View:           idx.v,
		}
	}

	return qset
}

//calcPSet
//select prepares which satisfy the following conditions
//prepares in previous qlist
//prepares from certStore which is prepared and (its view <= its idx.v or not in plist)
func (rbft *rbftImpl) calcPSet() map[uint64]*ViewChange_PQ {
	pset := make(map[uint64]*ViewChange_PQ)

	for n, p := range rbft.vcMgr.plist {
		pset[n] = p
	}

	for idx, cert := range rbft.storeMgr.certStore {
		if cert.prePrepare == nil {
			continue
		}

		if !rbft.prepared(idx.d, idx.v, idx.n) {
			continue
		}

		if p, ok := pset[idx.n]; ok && p.View > idx.v {
			continue
		}

		pset[idx.n] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    idx.d,
			View:           idx.v,
		}
	}

	return pset
}

//sendViewChange send view change message to other peers use broadcast
//Then it send view change message to itself and jump to recvViewChange
func (rbft *rbftImpl) sendViewChange() consensusEvent {

	//Do some check and do some preparation
	//such as stop nullRequest timer , clean batchVdr.cacheValidatedBatch and so on.
	err := rbft.beforeSendVC()
	if err != nil {
		rbft.logger.Error(err.Error())
		return nil
	}

	//create viewChange message
	vc := &ViewChange{
		View:      rbft.view,
		H:         rbft.h,
		ReplicaId: rbft.id,
	}

	cSet, pSet, qSet := rbft.gatherPQC()
	vc.Cset = append(vc.Cset, cSet...)
	vc.Pset = append(vc.Pset, pSet...)
	vc.Qset = append(vc.Qset, qSet...)

	rbft.logger.Warningf("Replica %d sending view-change, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	payload, err := proto.Marshal(vc)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_VIEW_CHANGE Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_VIEW_CHANGE,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	//Broadcast viewChange message to other peers
	rbft.helper.InnerBroadcast(msg)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_RESEND_TIMER_EVENT,
	}
	//Start VC_RESEND_TIMER. If peers can't viewChange successfully within the given time. timer well resend viewChange message
	rbft.timerMgr.startTimer(VC_RESEND_TIMER, event, rbft.eventMux)
	return rbft.recvViewChange(vc)
}

//recvViewChange process ViewChange message from itself or other peers
//if the number of ViewChange message for equal view reach on allCorrectReplicasQuorum, return VIEW_CHANGE_QUORUM_EVENT
// else peers may resend vc or wait more vc message arrived
func (rbft *rbftImpl) recvViewChange(vc *ViewChange) consensusEvent {
	rbft.logger.Warningf("Replica %d received view-change from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, vc.ReplicaId, vc.View, vc.H, len(vc.Cset), len(vc.Pset), len(vc.Qset))

	//check if inNegoView
	//if inNegoView, will return nil
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvViewChange, but it's in nego-view", rbft.id)
		return nil
	}

	//check if inRecovery
	//if inRecovery, will return nil
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Noticef("Replica %d try to recvcViewChange, but it's in recovery", rbft.id)
		return nil
	}

	if vc.View < rbft.view {
		rbft.logger.Warningf("Replica %d found view-change message for old view from replica %d: self view=%d, vc view=%d", rbft.id, vc.ReplicaId, rbft.view, vc.View)
		return nil
	}
	//check whether there is pqset which its view is less then vc's view and SequenceNumber more then low watermark
	//check whether there is cset which its SequenceNumber more then low watermark
	//if so ,return nil
	if !rbft.correctViewChange(vc) {
		rbft.logger.Warningf("Replica %d found view-change message incorrect", rbft.id)
		return nil
	}

	//if vc.ReplicaId == rbft.id increase the count of vcResend
	if vc.ReplicaId == rbft.id {
		rbft.vcMgr.vcResendCount++
		rbft.logger.Warningf("======== Replica %d already recv view change from itself for %d times", rbft.id, rbft.vcMgr.vcResendCount)
	}
	//check if this viewchange has stored in viewChangeStore
	//if so,return nil
	if old, ok := rbft.vcMgr.viewChangeStore[vcidx{vc.View, vc.ReplicaId}]; ok {
		if reflect.DeepEqual(old, vc) {
			rbft.logger.Warningf("Replica %d already has a repeated view change message"+
				" for view %d from replica %d, replcace it", rbft.id, vc.View, vc.ReplicaId)
			return nil
		}

		rbft.logger.Warningf("Replica %d already has a updated view change message"+
			" for view %d from replica %d", rbft.id, vc.View, vc.ReplicaId)
	}
	//check whether vcResendCount>=vcResendLimit
	//if so , reset view and stop vc and newView timer.
	//Set state to inNegoView and inRecovery
	//Finally, jump to initNegoView()
	if rbft.vcMgr.vcResendCount >= rbft.vcMgr.vcResendLimit {
		rbft.logger.Noticef("Replica %d view change resend reach upbound, try to recovery", rbft.id)
		rbft.timerMgr.stopTimer(NEW_VIEW_TIMER)
		rbft.timerMgr.stopTimer(VC_RESEND_TIMER)
		rbft.vcMgr.vcResendCount = 0
		rbft.restoreView()
		// after 10 viewchange without response from others, we will restart recovery, and set vcToRecovery to
		// true, which, after negotiate view done, we need to parse certStore
		rbft.status.activeState(&rbft.status.inNegoView, &rbft.status.inRecovery, &rbft.status.vcToRecovery)
		atomic.StoreUint32(&rbft.activeView, 1)
		rbft.initNegoView()
		return nil
	}

	vc.Timestamp = time.Now().UnixNano()

	//store vc to viewChangeStore
	rbft.vcMgr.viewChangeStore[vcidx{vc.View, vc.ReplicaId}] = vc

	// RBFT TOCS 4.5.1 Liveness: "if a replica receives a set of
	// f+1 valid VIEW-CHANGE messages from other replicas for
	// views greater than its current view, it sends a VIEW-CHANGE
	// message for the smallest view in the set, even if its timer
	// has not expired"
	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range rbft.vcMgr.viewChangeStore {
		if vc.Timestamp+int64(rbft.timerMgr.getTimeoutValue(CLEAN_VIEW_CHANGE_TIMER)) < time.Now().UnixNano() {
			rbft.logger.Warningf("Replica %d dropped an out-of-time view change message from replica %d", rbft.id, vc.ReplicaId)
			delete(rbft.vcMgr.viewChangeStore, idx)
			continue
		}

		if idx.v <= rbft.view {
			continue
		}

		replicas[idx.id] = true
		if minView == 0 || idx.v < minView {
			minView = idx.v
		}
	}

	// We only enter this if there are enough view change messages greater than our current view
	if len(replicas) >= rbft.oneCorrectQuorum() {
		rbft.logger.Warningf("Replica %d received f+1 view-change messages, triggering view-change to view %d",
			rbft.id, minView)
		rbft.timerMgr.stopTimer(FIRST_REQUEST_TIMER)
		// subtract one, because sendViewChange() increments
		rbft.view = minView - 1
		return rbft.sendViewChange()
	}
	//calculate how many peers has view = rbft.view
	quorum := 0
	for idx := range rbft.vcMgr.viewChangeStore {
		if idx.v == rbft.view {
			quorum++
		}
	}
	rbft.logger.Debugf("Replica %d now has %d view change requests for view %d", rbft.id, quorum, rbft.view)

	//if in viewchange and vc.view=rbft.view and quorum>allCorrectReplicasQuorum
	//rbft find new view success and jump into VIEW_CHANGE_QUORUM_EVENT
	if atomic.LoadUint32(&rbft.activeView) == 0 && vc.View == rbft.view && quorum >= rbft.allCorrectReplicasQuorum() {
		//close VC_RESEND_TIMER
		rbft.timerMgr.stopTimer(VC_RESEND_TIMER)

		//start newViewTimer and increase lastNewViewTimeout.
		//if this view change failed,next view change will have more time to do it
		rbft.startNewViewTimer(rbft.vcMgr.lastNewViewTimeout, "new view change")
		rbft.vcMgr.lastNewViewTimeout = 2 * rbft.vcMgr.lastNewViewTimeout
		if rbft.vcMgr.lastNewViewTimeout > 5*rbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER) {
			rbft.vcMgr.lastNewViewTimeout = 5 * rbft.timerMgr.getTimeoutValue(NEW_VIEW_TIMER)
		}
		//packaging  VIEW_CHANGE_QUORUM_EVENT message
		return &LocalEvent{
			Service:   VIEW_CHANGE_SERVICE,
			EventType: VIEW_CHANGE_QUORUM_EVENT,
		}
	}
	//if message from primary, peers send view change to other peers directly
	if atomic.LoadUint32(&rbft.activeView) == 1 && rbft.isPrimary(vc.ReplicaId) {
		rbft.sendViewChange()
	}

	return nil
}

//processing enter here when peer is primary and it receive allCorrectReplicasQuorum for new view.
//sendNewView  select suitable pqc from viewChangeStore as a new view message and broadcast to replica peers.
//Then jump into primaryProcessNewView.
func (rbft *rbftImpl) sendNewView() consensusEvent {

	//if inNegoView return nil.
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to sendNewView, but it's in nego-view", rbft.id)
		return nil
	}
	//if this new view has stored return nil.
	if _, ok := rbft.vcMgr.newViewStore[rbft.view]; ok {
		rbft.logger.Debugf("Replica %d already has new view in store for view %d, skipping", rbft.id, rbft.view)
		return nil
	}
	//get all viewChange message in viewChangeStore.
	vset, nset := rbft.getViewChanges()

	//get suitable checkpoint for later recovery, replicas contains the peer id who has this checkpoint.
	//if can't find suitable checkpoint, ok return false.
	cp, ok, replicas := rbft.selectInitialCheckpoint(nset)
	if !ok {
		rbft.logger.Infof("Replica %d could not find consistent checkpoint: %+v", rbft.id, rbft.vcMgr.viewChangeStore)
		return nil
	}
	//select suitable pqcCerts for later recovery.Their sequence is greater then cp
	//if msgList is nil, must some bug happened
	msgList := rbft.assignSequenceNumbers(nset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Replica %d could not assign sequence numbers for new view", rbft.id)
		return nil
	}
	//create new view message
	nv := &NewView{
		View:      rbft.view,
		Vset:      vset,
		Xset:      msgList,
		ReplicaId: rbft.id,
	}

	rbft.logger.Infof("Replica %d is new primary, sending new-view, v:%d, X:%+v",
		rbft.id, nv.View, nv.Xset)
	payload, err := proto.Marshal(nv)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_NEW_VIEW Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEW_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	//broadcast new view
	rbft.helper.InnerBroadcast(msg)
	//set new view to newViewStore
	rbft.vcMgr.newViewStore[rbft.view] = nv
	rbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
	return rbft.primaryProcessNewView(cp, replicas, nv)
}

//recvNewView
func (rbft *rbftImpl) recvNewView(nv *NewView) consensusEvent {
	rbft.logger.Infof("Replica %d received new-view %d",
		rbft.id, nv.View)

	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvNewView, but it's in nego-view", rbft.id)
		return nil
	}

	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Warningf("Replica %d try to recvNewView, but it's in recovery", rbft.id)
		rbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	if !(nv.View > 0 && nv.View >= rbft.view && rbft.primary(nv.View) == nv.ReplicaId && rbft.vcMgr.newViewStore[nv.View] == nil) {
		rbft.logger.Warningf("Replica %d rejecting invalid new-view from %d, v:%d",
			rbft.id, nv.ReplicaId, nv.View)
		return nil
	}

	rbft.vcMgr.newViewStore[nv.View] = nv

	quorum := 0
	for idx := range rbft.vcMgr.viewChangeStore {
		if idx.v == rbft.view {
			quorum++
		}
	}
	if quorum < rbft.allCorrectReplicasQuorum() {
		rbft.logger.Warningf("Replica %d has not meet ViewChangedQuorum", rbft.id)
		return nil
	}

	return rbft.processNewView()
}

//processNewView
func (rbft *rbftImpl) processNewView() consensusEvent {
	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
		return nil
	}

	if atomic.LoadUint32(&rbft.activeView) == 1 {
		rbft.logger.Infof("Replica %d ignoring new-view from %d, v:%d: we are active in view %d",
			rbft.id, nv.ReplicaId, nv.View, rbft.view)
		return nil
	}

	_, nset := rbft.getViewChanges()
	cp, ok, replicas := rbft.selectInitialCheckpoint(nset)
	if !ok {
		rbft.logger.Warningf("Replica %d could not determine initial checkpoint: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}


	// Check if the xset sent by new primary is built correctly by the aset
	msgList := rbft.assignSequenceNumbers(nset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Warningf("Replica %d could not assign sequence numbers: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}
	if !(len(msgList) == 0 && len(nv.Xset) == 0) && !reflect.DeepEqual(msgList, nv.Xset) {
		rbft.logger.Warningf("Replica %d failed to verify new-view Xset: computed %+v, received %+v",
			rbft.id, msgList, nv.Xset)
		return rbft.sendViewChange()
	}

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(cp, replicas)
	if err != nil {
		rbft.logger.Error(err.Error())
		return nil
	}

	return rbft.processReqInNewView(nv)
}

//do some prepare for change to New view
//such as get moveWatermarks to ViewChange checkpoint and fetch missed batches
func (rbft *rbftImpl) primaryProcessNewView(initialCp ViewChange_C, replicas []replicaInfo, nv *NewView) consensusEvent {

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(initialCp, replicas)
	if err != nil {
		rbft.logger.Error(err.Error())
		return nil
	}

	//check if we have all block from low waterMark to recovery seq
	newReqBatchMissing := rbft.feedMissingReqBatchIfNeeded(nv.Xset)
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		return rbft.processReqInNewView(nv)
	} else if newReqBatchMissing {
		//asynchronous
		// if received all batches, jump into processReqInNewView
		rbft.fetchRequestBatches()
	}

	return nil
}

func (vcm *vcManager) updateViewChangeSeqNo(seqNo, K, id uint64) {
	if vcm.viewChangePeriod <= 0 {
		return
	}
	// Ensure the view change always occurs at a checkpoint boundary
	vcm.viewChangeSeqNo = seqNo + vcm.viewChangePeriod*K - seqNo%K
	//logger.Debugf("Replica %d updating view change sequence number to %d", id, vcm.viewChangeSeqNo)
}

func (rbft *rbftImpl) feedMissingReqBatchIfNeeded(xset Xset) (newReqBatchMissing bool) {
	newReqBatchMissing = false
	for n, d := range xset {
		// RBFT: why should we use "h ≥ min{n | ∃d : (<n,d> ∈ X)}"?
		// "h ≥ min{n | ∃d : (<n,d> ∈ X)} ∧ ∀<n,d> ∈ X : (n ≤ h ∨ ∃m ∈ in : (D(m) = d))"
		if n <= rbft.h {
			continue
		} else {
			if d == "" {
				// NULL request; skip
				continue
			}

			if _, ok := rbft.storeMgr.txBatchStore[d]; !ok {
				rbft.logger.Warningf("Replica %d missing assigned, non-checkpointed request batch %s",
					rbft.id, d)
				if _, ok := rbft.storeMgr.missingReqBatches[d]; !ok {
					rbft.logger.Warningf("Replica %v requesting to fetch batch %s",
						rbft.id, d)
					newReqBatchMissing = true
					rbft.storeMgr.missingReqBatches[d] = true
				}
			}
		}
	}
	return newReqBatchMissing
}

func (rbft *rbftImpl) processReqInNewView(nv *NewView) consensusEvent {
	rbft.logger.Debugf("Replica %d accepting new-view to view %d", rbft.id, rbft.view)

	//if vcHandled active return nill, else set vcHandled active
	if rbft.status.getState(&rbft.status.vcHandled) {
		rbft.logger.Debugf("Replica %d repeated enter processReqInNewView, ignore it", rbft.id)
		return nil
	}
	rbft.status.activeState(&rbft.status.vcHandled)

	// empty the outstandingReqBatch, it is useless since new primary will resend pre-prepare
	rbft.storeMgr.outstandingReqBatches = make(map[string]*TransactionBatch)
	rbft.batchVdr.preparedCert = make(map[vidx]string)
	rbft.storeMgr.committedCert = make(map[msgID]string)

	//backendVid is seq to vcRest
	backendVid := rbft.exec.lastExec + 1
	rbft.seqNo = rbft.exec.lastExec
	rbft.batchVdr.setLastVid(rbft.exec.lastExec)

	//if state not in stateTransfer and not inVcReset
	//jump into VCReset
	//VcReset will exec sync and jump finishViewChange after vcReset success
	//else if in stateTransfe or inVcReset
	//if it is primary, we should not finishViewChange
	//else jump into finishViewChange
	if !rbft.status.getState(&rbft.status.skipInProgress) &&
		!rbft.status.getState(&rbft.status.inVcReset) {
		rbft.helper.VcReset(backendVid)
		rbft.status.activeState(&rbft.status.inVcReset)
	} else if rbft.isPrimary(rbft.id) {
		rbft.logger.Warningf("New primary %d need to catch up other, wating", rbft.id)
	} else {
		rbft.logger.Warningf("Replica %d cannot process local vcReset, but also send finishVcReset", rbft.id)
		rbft.finishViewChange()
	}

	return nil
}

//Processing enters here after receiving FinishVcReset message
//Do some state check
func (rbft *rbftImpl) recvFinishVcReset(finish *FinishVcReset) consensusEvent {
	//Check whether we are in viewChange
	if atomic.LoadUint32(&rbft.activeView) == 1 {
		rbft.logger.Warningf("Replica %d is not in viewChange, but received FinishVcReset from replica %d", rbft.id, finish.ReplicaId)
		return nil
	}

	//Check whether View from received finishVcReset is equal rbft.view
	if finish.View != rbft.view {
		rbft.logger.Warningf("Replica %d received finishVcReset from replica %d, expect view=%d, but get view=%d",
			rbft.id, finish.ReplicaId, rbft.view, finish.View)
		return nil
	}

	//Put received FinishVcReset stored in vcResetStore
	rbft.logger.Debugf("Replica %d received FinishVcReset from replica %d, view=%d/h=%d",
		rbft.id, finish.ReplicaId, finish.View, finish.LowH)
	ok := rbft.vcMgr.vcResetStore[*finish]
	if ok {
		rbft.logger.Warningf("Replica %d ignored duplicate agree FinishVcReset from %d", rbft.id, finish.ReplicaId)
		return nil
	}
	rbft.vcMgr.vcResetStore[*finish] = true

	return rbft.handleTailInNewView()
}

//Processing enters here after recvFinishVcReset().
//HandleTailInNewView check whether we can finish view change
//such as number of peers send finishVcReset.
//If view change success,processing will send VIEW_CHANGED_EVENT to rbft
func (rbft *rbftImpl) handleTailInNewView() consensusEvent {

	quorum := 0
	hasPrimary := false
	for finish := range rbft.vcMgr.vcResetStore {
		if finish.View == rbft.view {
			quorum++
			if rbft.isPrimary(finish.ReplicaId) {
				hasPrimary = true
			}
		}

	}
	//if the number of peers send finishVcReset not >= allCorrectReplicasQuorum or primary not sends finishVcReset
	//view change can not finish and return nil
	if quorum < rbft.allCorrectReplicasQuorum() || !hasPrimary {
		return nil
	}
	//if itself has not done with vcReset and not in stateUpdate return nil
	if rbft.status.getState(&rbft.status.inVcReset) && !rbft.status.getState(&rbft.status.skipInProgress) {
		rbft.logger.Debugf("Replica %d itself has not done with vcReset and not in stateUpdate", rbft.id)
		return nil
	}

	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Debugf("Replica %d ignoring rebuildCertStore as it could not find view %d in its newViewStore", rbft.id, rbft.view)
		return nil
	}
	//Close stopNewViewTimer
	rbft.stopNewViewTimer()

	//Update  state seqNo vid and last vid for new transaction
	rbft.seqNo = rbft.exec.lastExec
	rbft.batchVdr.setLastVid(rbft.exec.lastExec)

	rbft.putBackTxBatches(nv.Xset)

	//Primary validate batches which has seq > low watermark
	//Batch will transfer to pre-prepare
	if rbft.isPrimary(rbft.id) {
		rbft.primaryResendBatch(nv.Xset)
	}

	return &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGED_EVENT,
	}
}

// primaryResendBatch validates batches which has seq > low watermark
func (rbft *rbftImpl) primaryResendBatch(xset Xset) {

	xSetLen := len(xset)
	upper := uint64(xSetLen) + rbft.h + uint64(1)
	for i := rbft.h + uint64(1); i < upper; i++ {
		d, ok := xset[i]
		if !ok {
			rbft.logger.Critical("view change Xset miss batch number %d", i)
		} else if d == "" {
			// This should not happen
			rbft.logger.Critical("view change Xset has null batch, kick it out")
		} else {
			batch, ok := rbft.storeMgr.txBatchStore[d]
			if !ok {
				rbft.logger.Criticalf("In Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, rbft.id)
			} else if i > rbft.exec.lastExec {
				rbft.primaryValidateBatch(d, batch, i)
			}
		}
	}

}

//Rebuild Cert for Vc
func (rbft *rbftImpl) rebuildCertStoreForVC() {
	//Check whether new view has stored in newViewStore
	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
		return
	}
	//do rebuild cert
	rbft.rebuildCertStore(nv.Xset)
}

//rebuild certStore according to Xset
//Broadcast qpc for batches which has been confirmed in view change.
//So that, all correct peers will reach the seq that select in view change
func (rbft *rbftImpl) rebuildCertStore(xset Xset) {

	for n, d := range xset {
		if n <= rbft.h || n > rbft.exec.lastExec {
			continue
		}
		batch, ok := rbft.storeMgr.txBatchStore[d]
		if !ok && d != "" {
			rbft.logger.Criticalf("Replica %d is missing tx batch for seqNo=%d with digest '%s' for assigned prepare", rbft.id)
		}

		hashBatch := &HashBatch{
			List:      batch.HashList,
			Timestamp: batch.Timestamp,
		}
		cert := rbft.storeMgr.getCert(rbft.view, n, d)
		//if peer is primary ,it rebuild PrePrepare , persist it and broadcast PrePrepare
		if rbft.isPrimary(rbft.id) && ok {
			preprep := &PrePrepare{
				View:           rbft.view,
				SequenceNumber: n,
				BatchDigest:    d,
				ResultHash:     batch.ResultHash,
				HashBatch:      hashBatch,
				ReplicaId:      rbft.id,
			}
			cert.resultHash = batch.ResultHash
			cert.prePrepare = preprep
			cert.validated = true

			rbft.persistQSet(preprep)

			payload, err := proto.Marshal(preprep)
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
		} else {
			//else rebuild Prepare and broadcast Prepare
			prep := &Prepare{
				View:           rbft.view,
				SequenceNumber: n,
				BatchDigest:    d,
				ResultHash:     batch.ResultHash,
				ReplicaId:      rbft.id,
			}
			cert.prepare[*prep] = true
			cert.sentPrepare = true

			payload, err := proto.Marshal(prep)
			if err != nil {
				rbft.logger.Errorf("ConsensusMessage_PREPARE Marshal Error", err)
				rbft.batchVdr.updateLCVid()
				return
			}

			consensusMsg := &ConsensusMessage{
				Type:    ConsensusMessage_PREPARE,
				Payload: payload,
			}
			msg := cMsgToPbMsg(consensusMsg, rbft.id)
			rbft.helper.InnerBroadcast(msg)
		}
		//Broadcast commit
		cmt := &Commit{
			View:           rbft.view,
			SequenceNumber: n,
			BatchDigest:    d,
			ResultHash:     batch.ResultHash,
			ReplicaId:      rbft.id,
		}
		cert.commit[*cmt] = true
		cert.sentValidate = true
		cert.validated = true
		cert.sentCommit = true
		cert.sentExecute = true

		payload, err := proto.Marshal(cmt)
		if err != nil {
			rbft.logger.Errorf("ConsensusMessage_COMMIT Marshal Error", err)
			rbft.batchVdr.lastVid = *rbft.batchVdr.currentVid
			rbft.batchVdr.currentVid = nil
			return
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_COMMIT,
			Payload: payload,
		}
		msg := cMsgToPbMsg(consensusMsg, rbft.id)
		rbft.helper.InnerBroadcast(msg)
		// rebuild pqlist according to xset
		rbft.vcMgr.qlist = make(map[qidx]*ViewChange_PQ)
		rbft.vcMgr.plist = make(map[uint64]*ViewChange_PQ)

		id := qidx{d, n}
		pqItem := &ViewChange_PQ{
			SequenceNumber: n,
			BatchDigest:    d,
			View:           rbft.view,
		}
		//update vcMgr.pqlist according to Xset
		rbft.vcMgr.qlist[id] = pqItem
		rbft.vcMgr.plist[n] = pqItem

	}
}

//Processing enters here after peer Determined the new view and finished VCReset
//FinishViewChange Broadcast FinishVcReset to other peers and
//send it to itself.
func (rbft *rbftImpl) finishViewChange() consensusEvent {

	finish := &FinishVcReset{
		ReplicaId: rbft.id,
		View:      rbft.view,
		LowH:      rbft.h,
	}
	payload, err := proto.Marshal(finish)
	if err != nil {
		rbft.logger.Errorf("Marshal FinishVcReset Error!")
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_FINISH_VCRESET,
		Payload: payload,
	}

	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	rbft.helper.InnerBroadcast(msg)
	rbft.logger.Debugf("Replica %d broadcasting FinishVcReset", rbft.id)

	return rbft.recvFinishVcReset(finish)
}

//Return all viewChange message from viewChangeStore
func (rbft *rbftImpl) getViewChanges() (vset []*ViewChange, nset []*VCNODE) {
	for _, vc := range rbft.vcMgr.viewChangeStore {
		vset = append(vset, vc)
		nset = append(nset, &VCNODE{
			View:		vc.View,
			H:		vc.H,
			Cset:		vc.Cset,
			Pset:		vc.Pset,
			Qset:		vc.Qset,
			ReplicaId:	vc.ReplicaId,
			Genesis:	vc.Genesis,
		})
	}
	return
}

// selectInitialCheckpointselect checkpoint from received ViewChange message
// If find suitable checkpoint ,it return a certain checkpoint and the  replicas id list which replicas has this checkpoint
// The checkpoint is max checkpoint which exists in at least oneCorrectQuorum peers and greater then low waterMark
// in at least commonCaseQuorum.
func (rbft *rbftImpl) selectInitialCheckpoint(set []*VCNODE) (checkpoint ViewChange_C, find bool, replicas []replicaInfo) {
	// For the checkpoint as key, find the corresponding AgreeUpdateN messages
	checkpoints := make(map[ViewChange_C][]*VCNODE)
	for _, agree := range set {
		// Verify that we strip duplicate checkpoints from this Cset
		set := make(map[ViewChange_C]bool)
		for _, c := range agree.Cset {
			if ok := set[*c]; ok {
				continue
			}
			checkpoints[*c] = append(checkpoints[*c], agree)
			set[*c] = true
			rbft.logger.Debugf("Replica %d appending checkpoint from replica %d with seqNo=%d, h=%d, and checkpoint digest %s", rbft.id, agree.ReplicaId, agree.H, c.SequenceNumber, c.Id)
		}
	}

	// Indicate that replica cannot find any checkpoint
	if len(checkpoints) == 0 {
		rbft.logger.Debugf("Replica %d has no checkpoints to select from: %d %s",
			rbft.id, len(rbft.vcMgr.viewChangeStore), checkpoints)
		return
	}

	for idx, vcList := range checkpoints {
		// Need weak certificate for the checkpoint
		if len(vcList) < rbft.oneCorrectQuorum() { // type casting necessary to match types
			rbft.logger.Debugf("Replica %d has no weak certificate for n:%d, vcList was %d long",
				rbft.id, idx.SequenceNumber, len(vcList))
			continue
		}

		quorum := 0
		// Note, this is the whole vset (S) in the paper, not just this checkpoint set (S') (vcList)
		// We need 2f+1 low watermarks from S below this seqNo from all replicas
		// We need f+1 matching checkpoints at this seqNo (S')
		for _, vc := range set {
			if vc.H <= idx.SequenceNumber {
				quorum++
			}
		}

		if quorum < rbft.commonCaseQuorum() {
			rbft.logger.Debugf("Replica %d has no quorum for n:%d", rbft.id, idx.SequenceNumber)
			continue
		}

		// Find the highest checkpoint
		if checkpoint.SequenceNumber <= idx.SequenceNumber {
			replicas = make([]replicaInfo, len(vcList))
			for i, vc := range vcList {
				replicas[i] = replicaInfo{
					id:      vc.ReplicaId,
					height:  vc.H,
					genesis: vc.Genesis,
				}
			}

			checkpoint = idx
			find = true
		}
	}

	return
}

// Find the suitable batches for recovery to according to ViewChange and low waterMark
// The selected bathes match following condition
// If batch is not a NullRequest batch, the pre-prepare of this batch is equal or greater than commonCaseQuorum
// and the prepare is equal or greater then oneCorrectQuorum.
// In this release, batch should not be NUllRequest batch
func (rbft *rbftImpl) assignSequenceNumbers(set []*VCNODE, h uint64) map[uint64]string {
	msgList := make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
	nLoop:
	for n := h + 1; n <= h+rbft.L; n++ {
		// "∃m ∈ S..."
		for _, m := range set {
			// "...with <n,d,v> ∈ m.P"
			for _, em := range m.Pset {
				quorum := 0
				// "A1. ∃2f+1 messages m' ∈ S"
				mpLoop:
				for _, mp := range set {
					if mp.H >= n {
						continue
					}
					// "∀<n,d',v'> ∈ m'.P"
					for _, emp := range mp.Pset {
						if n != emp.SequenceNumber {
							continue
						}
						if !(emp.View < em.View || (emp.View == em.View && emp.BatchDigest == em.BatchDigest)) {
							continue mpLoop
						}
					}
					quorum++
				}

				if quorum < rbft.commonCaseQuorum() {
					continue
				}

				quorum = 0
				// "A2. ∃f+1 messages m' ∈ S"
				for _, mp := range set {
					// "∃<n,d',v'> ∈ m'.Q"
					for _, emp := range mp.Qset {
						if n != emp.SequenceNumber {
							continue
						}
						if emp.View >= em.View && emp.BatchDigest == em.BatchDigest {
							quorum++
							break
						}
					}
				}

				if quorum < rbft.oneCorrectQuorum() {
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
		for _, m := range set {
			// "m.h < n"
			if m.H >= n {
				continue
			}
			// "m.P has no entry for n"
			for _, em := range m.Pset {
				if em.SequenceNumber == n {
					continue nullLoop
				}
			}
			quorum++
		}

		if quorum >= rbft.commonCaseQuorum() {
			// "then select the null request for number n"
			msgList[n] = ""

			continue nLoop
		}

		rbft.logger.Warningf("Replica %d could not assign value to contents of seqNo %d, found only %d missing P entries", rbft.id, n, quorum)
		return nil
	}

	// prune top null requests
	for n, msg := range msgList {
		if n > maxN || msg == "" {
			delete(msgList, n)
		}
	}

	keys := make([]uint64, len(msgList))
	i := 0
	for n := range msgList {
		keys[i] = n
		i++
	}
	sort.Sort(sortableUint64Slice(keys))
	x := h + 1
	list := make(map[uint64]string)
	for _, n := range keys {
		list[x] = msgList[n]
		x++
	}

	return list
}

//Return the request of fetching missing assigned, non-checkpointed
//Return should not happen in inNegoView and inRecovery.
func (rbft *rbftImpl) recvFetchRequestBatch(fr *FetchRequestBatch) (err error) {
	//Check if inNegoView
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvFetchRequestBatch, but it's in nego-view", rbft.id)
		return nil
	}
	//Check if inRecovery
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Noticef("Replica %d try to recvFetchRequestBatch, but it's in recovery", rbft.id)
		return nil
	}

	//Check if we have requested batch
	digest := fr.BatchDigest
	if _, ok := rbft.storeMgr.txBatchStore[digest]; !ok {
		return nil // we don't have it either
	}

	reqBatch := rbft.storeMgr.txBatchStore[digest]
	batch := &ReturnRequestBatch{
		Batch:       reqBatch,
		BatchDigest: digest,
		ReplicaId:   rbft.id,
	}
	payload, err := proto.Marshal(batch)
	if err != nil {
		rbft.logger.Errorf("ConsensusMessage_RETURN_REQUEST_BATCH Marshal Error", err)
		return nil
	}
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_RETURN_REQUEST_BATCH,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)

	receiver := fr.ReplicaId
	//return requested batch
	err = rbft.helper.InnerUnicast(msg, receiver)

	return
}

//Receive the RequestBatch from other peers
//If receive all request batch,processing jump to processReqInNewView or processReqInUpdate
func (rbft *rbftImpl) recvReturnRequestBatch(batch *ReturnRequestBatch) consensusEvent {
	//Check if in inNegoView
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to recvReturnRequestBatch, but it's in nego-view", rbft.id)
		return nil
	}
	//Check if in inRecovery
	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Noticef("Replica %d try to recvReturnRequestBatch, but it's in recovery", rbft.id)
		return nil
	}

	digest := batch.BatchDigest
	if _, ok := rbft.storeMgr.missingReqBatches[digest]; !ok {
		return nil // either the wrong digest, or we got it already from someone else
	}
	//stored into validatedBatchStore
	rbft.storeMgr.txBatchStore[digest] = batch.Batch
	//delete missingReqBatches in this batch
	delete(rbft.storeMgr.missingReqBatches, digest)
	rbft.logger.Warning("Primary received missing request: ", digest)

	//if receive all request batch
	//if validatedBatchStore jump to processReqInNewView
	//if inUpdatingN jump to processReqInUpdate
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		if atomic.LoadUint32(&rbft.activeView) == 0 {
			nv, ok := rbft.vcMgr.newViewStore[rbft.view]
			if !ok {
				rbft.logger.Debugf("Replica %d ignoring processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
				return nil
			}
			return rbft.processReqInNewView(nv)
		}
		if atomic.LoadUint32(&rbft.nodeMgr.inUpdatingN) == 1 {
			update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
			if !ok {
				rbft.logger.Debugf("Replica %d ignoring processUpdateN as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
				return nil
			}
			return rbft.processReqInUpdate(update)
		}
	}
	return nil

}

//##########################################################################
//           view change auxiliary functions
//##########################################################################

//stopNewViewTimer
func (rbft *rbftImpl) stopNewViewTimer() {
	rbft.logger.Debugf("Replica %d stopping a running new view timer", rbft.id)
	rbft.status.inActiveState(&rbft.status.timerActive)
	rbft.timerMgr.stopTimer(NEW_VIEW_TIMER)
}

//startNewViewTimer stop all running new view timers and  start a new view timer
func (rbft *rbftImpl) startNewViewTimer(timeout time.Duration, reason string) {
	rbft.logger.Debugf("Replica %d starting new view timer for %s: %s", rbft.id, timeout, reason)
	rbft.vcMgr.newViewTimerReason = reason
	rbft.status.activeState(&rbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	rbft.timerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, rbft.eventMux)
}

//softstartNewViewTimer start a new view timer no matter how many existed new view timer
func (rbft *rbftImpl) softStartNewViewTimer(timeout time.Duration, reason string) {
	rbft.logger.Debugf("Replica %d soft starting new view timer for %s: %s", rbft.id, timeout, reason)
	rbft.vcMgr.newViewTimerReason = reason
	rbft.status.activeState(&rbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	rbft.timerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, rbft.eventMux)
}

//Check if View change messages correct
//pqset ' view should less then vc.View and SequenceNumber should greater then vc.H.
//checkpoint's SequenceNumber should greater then vc.H
func (rbft *rbftImpl) correctViewChange(vc *ViewChange) bool {
	for _, p := range append(vc.Pset, vc.Qset...) {
		if !(p.View < vc.View && p.SequenceNumber > vc.H) {
			rbft.logger.Debugf("Replica %d invalid p entry in view-change: vc(v:%d h:%d) p(v:%d n:%d)",
				rbft.id, vc.View, vc.H, p.View, p.SequenceNumber)
			return false
		}
	}

	for _, c := range vc.Cset {
		if !(c.SequenceNumber >= vc.H) {
			rbft.logger.Debugf("Replica %d invalid c entry in view-change: vc(v:%d h:%d) c(n:%d)",
				rbft.id, vc.View, vc.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

//beforeSendVC operations before send view change
//1 Check rbft.state. State should not inNegoView or inRecovery
//2 Stop NewViewTimer and NULL_REQUEST_TIMER
//3 increase the view and delete new view of old view in newViewStore
//4 update pqlist
//5 delete old viewChange message
func (rbft *rbftImpl) beforeSendVC() error {
	if rbft.status.getState(&rbft.status.inNegoView) {
		rbft.logger.Debugf("Replica %d try to send view change, but it's in nego-view", rbft.id)
		return errors.New("node is in nego view now!")
	}

	if rbft.status.getState(&rbft.status.inRecovery) {
		rbft.logger.Noticef("Replica %d try to send view change, but it's in recovery", rbft.id)
		return errors.New("node is in recovery now!")
	}

	rbft.stopNewViewTimer()
	rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)

	delete(rbft.vcMgr.newViewStore, rbft.view)
	rbft.view++
	atomic.StoreUint32(&rbft.activeView, 0)
	rbft.status.inActiveState(&rbft.status.vcHandled)
	atomic.StoreUint32(&rbft.normal, 0)

	rbft.vcMgr.plist = rbft.calcPSet()
	rbft.vcMgr.qlist = rbft.calcQSet()
	rbft.batchVdr.cacheValidatedBatch = make(map[string]*cacheBatch)
	// clear old messages
	for idx := range rbft.vcMgr.viewChangeStore {
		if idx.v < rbft.view {
			delete(rbft.vcMgr.viewChangeStore, idx)
		}
	}
	return nil
}

func (rbft *rbftImpl) gatherPQC() (cset []*ViewChange_C, pset []*ViewChange_PQ, qset []*ViewChange_PQ) {
	// Gather all the checkpoints
	for n, id := range rbft.storeMgr.chkpts {
		cset = append(cset, &ViewChange_C{
			SequenceNumber: n,
			Id:             id,
		})
	}
	// Gather all the p entries
	for _, p := range rbft.vcMgr.plist {
		if p.SequenceNumber < rbft.h {
			rbft.logger.Errorf("BUG! Replica %d should not have anything in our pset less than h, found %+v", rbft.id, p)
		}
		pset = append(pset, p)
	}

	// Gather all the q entries
	for _, q := range rbft.vcMgr.qlist {
		if q.SequenceNumber < rbft.h {
			rbft.logger.Errorf("BUG! Replica %d should not have anything in our qset less than h, found %+v", rbft.id, q)
		}
		qset = append(qset, q)
	}

	return
}
