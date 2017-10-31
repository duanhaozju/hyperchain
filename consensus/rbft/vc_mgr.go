//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	"github.com/op/go-logging"
)

// vcManager manages the whole process of view change
type vcManager struct {
	vcResendLimit      int           // vcResendLimit indicates a replica's view change resending upbound.
	vcResendCount      int           // vcResendCount represent times of same view change info resend
	viewChangePeriod   uint64        // period between automatic view changes. Default value is 0 means close automatic view changes
	viewChangeSeqNo    uint64        // next seqNo to perform view change
	lastNewViewTimeout time.Duration // last timeout we used during this view change
	newViewTimerReason string        // what triggered the timer

	qlist map[qidx]*Vc_PQ   //store Pre-Prepares  for view change
	plist map[uint64]*Vc_PQ //store Prepares for view change

	newViewStore    map[uint64]*NewView    // track last new-view we received or sent
	viewChangeStore map[vcidx]*ViewChange  // track view-change messages
	vcResetStore    map[FinishVcReset]bool // track vcReset message from others

	logger *logging.Logger
}

// dispatchViewChangeMsg dispatches view change consensus messages from
// other peers And push them into corresponding function
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

// newVcManager init a instance of view change manager and initialize each parameter
// according to the configuration file.
func newVcManager(config *common.Config, logger *logging.Logger) *vcManager {
	vcm := &vcManager{}
	vcm.logger = logger

	//init vcManage maps
	vcm.vcResetStore = make(map[FinishVcReset]bool)
	vcm.qlist = make(map[qidx]*Vc_PQ)
	vcm.plist = make(map[uint64]*Vc_PQ)
	vcm.newViewStore = make(map[uint64]*NewView)
	vcm.viewChangeStore = make(map[vcidx]*ViewChange)

	vcm.viewChangePeriod = uint64(config.GetInt(RBFT_VC_PERIOD))
	// automatic view changes is off by default(should be read from config)
	if vcm.viewChangePeriod > 0 {
		vcm.logger.Infof("RBFT viewChange period = %v", vcm.viewChangePeriod)
	} else {
		vcm.logger.Infof("RBFT automatic viewChange disabled")
	}

	vcm.lastNewViewTimeout = config.GetDuration(RBFT_VIEWCHANGE_TIMEOUT)
	vcm.vcResendLimit = config.GetInt(RBFT_VC_RESEND_LIMIT)
	vcm.logger.Infof("RBFT vcResendLimit = %d", vcm.vcResendLimit)

	return vcm
}

// sendViewChange sends view change message to other peers using broadcast.
// Then it sends view change message to itself and jump to recvViewChange.
func (rbft *rbftImpl) sendViewChange() consensusEvent {

	//Do some check and do some preparation
	//such as stop nullRequest timer , clean batchVdr.cacheValidatedBatch and so on.
	err := rbft.beforeSendVC()
	if err != nil {
		return nil
	}

	//create viewChange message

	vc := &ViewChange{
		Basis: &VcBasis{
			View:      rbft.view,
			H:         rbft.h,
			ReplicaId: rbft.id,
		},
	}

	cSet, pSet, qSet := rbft.gatherPQC()
	vc.Basis.Cset = append(vc.Basis.Cset, cSet...)
	vc.Basis.Pset = append(vc.Basis.Pset, pSet...)
	vc.Basis.Qset = append(vc.Basis.Qset, qSet...)

	rbft.logger.Infof("Replica %d sending viewChange, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, vc.Basis.View, vc.Basis.H, len(vc.Basis.Cset), len(vc.Basis.Pset), len(vc.Basis.Qset))

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

// recvViewChange processes ViewChange message from itself or other peers
// If the number of ViewChange message for equal view reach on
// allCorrectReplicasQuorum, return VIEW_CHANGE_QUORUM_EVENT.
// Else peers may resend vc or wait more vc message arrived.
func (rbft *rbftImpl) recvViewChange(vc *ViewChange) consensusEvent {

	rbft.logger.Debugf("Replica %d received viewChange from replica %d, v:%d, h:%d, |C|:%d, |P|:%d, |Q|:%d",
		rbft.id, vc.Basis.ReplicaId, vc.Basis.View, vc.Basis.H, len(vc.Basis.Cset), len(vc.Basis.Pset), len(vc.Basis.Qset))

	//check if inNegotiateView
	//if inNegotiateView, will return nil
	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d try to receive viewChange, but it's in negotiateVie", rbft.id)
		return nil
	}

	//check if inRecovery
	//if inRecovery, will return nil
	if rbft.in(inRecovery) {
		rbft.logger.Debugf("Replica %d try to receive viewChange, but it's in recovery", rbft.id)
		return nil
	}

	if vc.Basis.View < rbft.view {
		rbft.logger.Warningf("Replica %d found viewChange message for old view from replica %d: self view=%d, vc view=%d", rbft.id, vc.Basis.ReplicaId, rbft.view, vc.Basis.View)
		return nil
	}
	//check whether there is pqset which its view is less then vc's view and SequenceNumber more then low watermark
	//check whether there is cset which its SequenceNumber more then low watermark
	//if so ,return nil
	if !rbft.correctViewChange(vc) {
		rbft.logger.Warningf("Replica %d found viewChange message incorrect", rbft.id)
		return nil
	}

	//if vc.ReplicaId == rbft.id increase the count of vcResend
	if vc.Basis.ReplicaId == rbft.id {
		rbft.vcMgr.vcResendCount++
		rbft.logger.Infof("======== Replica %d already receive viewChange from itself for %d times", rbft.id, rbft.vcMgr.vcResendCount)
	}
	//check if this viewchange has stored in viewChangeStore
	//if so,return nil
	if old, ok := rbft.vcMgr.viewChangeStore[vcidx{vc.Basis.View, vc.Basis.ReplicaId}]; ok {
		if reflect.DeepEqual(old.Basis, vc.Basis) {
			rbft.logger.Warningf("Replica %d already has a same viewChange message"+
				" for view %d from replica %d, ignore it", rbft.id, vc.Basis.View, vc.Basis.ReplicaId)
			return nil
		}

		rbft.logger.Debugf("Replica %d already has a updated viewChange message"+
			" for view %d from replica %d, replace it", rbft.id, vc.Basis.View, vc.Basis.ReplicaId)
	}
	//check whether vcResendCount>=vcResendLimit
	//if so , reset view and stop vc and newView timer.
	//Set state to inNegotiateView and inRecovery
	//Finally, jump to initNegoView()
	if rbft.vcMgr.vcResendCount >= rbft.vcMgr.vcResendLimit {
		rbft.logger.Infof("Replica %d viewChange resend reach upper limit, try to recovery", rbft.id)
		rbft.timerMgr.stopTimer(NEW_VIEW_TIMER)
		rbft.timerMgr.stopTimer(VC_RESEND_TIMER)
		rbft.vcMgr.vcResendCount = 0
		rbft.restoreView()
		// after 10 viewchange without response from others, we will restart recovery, and set vcToRecovery to
		// true, which, after negotiate view done, we need to parse certStore
		rbft.on(inNegotiateView, inRecovery, vcToRecovery)
		rbft.off(inViewChange)
		rbft.initNegoView()
		return nil
	}

	vc.Timestamp = time.Now().UnixNano()

	//store vc to viewChangeStore
	rbft.vcMgr.viewChangeStore[vcidx{vc.Basis.View, vc.Basis.ReplicaId}] = vc

	// RBFT TOCS 4.5.1 Liveness: "if a replica receives a set of
	// f+1 valid VIEW-CHANGE messages from other replicas for
	// views greater than its current view, it sends a VIEW-CHANGE
	// message for the smallest view in the set, even if its timer
	// has not expired"
	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range rbft.vcMgr.viewChangeStore {
		if vc.Timestamp+int64(rbft.timerMgr.getTimeoutValue(CLEAN_VIEW_CHANGE_TIMER)) < time.Now().UnixNano() {
			rbft.logger.Debugf("Replica %d drop an out-of-time viewChange message from replica %d", rbft.id, vc.Basis.ReplicaId)
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
		rbft.logger.Infof("Replica %d received f+1 viewChange messages, triggering viewChange to view %d",
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
	rbft.logger.Debugf("Replica %d now has %d viewChange requests for view %d", rbft.id, quorum, rbft.view)

	//if in viewchange and vc.view=rbft.view and quorum>allCorrectReplicasQuorum
	//rbft find new view success and jump into VIEW_CHANGE_QUORUM_EVENT
	if rbft.in(inViewChange) && vc.Basis.View == rbft.view && quorum >= rbft.allCorrectReplicasQuorum() {
		//close VC_RESEND_TIMER
		rbft.timerMgr.stopTimer(VC_RESEND_TIMER)

		//start newViewTimer and increase lastNewViewTimeout.
		//if this view change failed,next view change will have more time to do it
		rbft.startNewViewTimer(rbft.vcMgr.lastNewViewTimeout, "new viewChange")
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
	if !rbft.in(inViewChange) && rbft.isPrimary(vc.Basis.ReplicaId) {
		rbft.sendViewChange()
	}

	return nil
}

// sendNewView select suitable pqc from viewChangeStore as a new view message and
// broadcast to replica peers when peer is primary and it receives
// allCorrectReplicasQuorum for new view.
// Then jump into primaryProcessNewView.
func (rbft *rbftImpl) sendNewView() consensusEvent {

	//if inNegotiateView return nil.
	if rbft.in(inNegotiateView) {
		rbft.logger.Warningf("Replica %d try to sendNewView, but it's in negotiateView", rbft.id)
		return nil
	}
	//if this new view has stored, return nil.
	if _, ok := rbft.vcMgr.newViewStore[rbft.view]; ok {
		rbft.logger.Warningf("Replica %d already has newView in store for view %d, ignore it", rbft.id, rbft.view)
		return nil
	}
	//get all viewChange message in viewChangeStore.
	vset := rbft.getViewChanges()

	//get suitable checkpoint for later recovery, replicas contains the peer id who has this checkpoint.
	//if can't find suitable checkpoint, ok return false.
	cp, ok, replicas := rbft.selectInitialCheckpoint(vset)
	if !ok {
		rbft.logger.Infof("Replica %d could not find consistent checkpoint: %+v", rbft.id, rbft.vcMgr.viewChangeStore)
		return nil
	}
	//select suitable pqcCerts for later recovery.Their sequence is greater then cp
	//if msgList is nil, must some bug happened
	msgList := rbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Replica %d could not assign sequence numbers for newView", rbft.id)
		return nil
	}
	//create new view message
	nv := &NewView{
		View:      rbft.view,
		Xset:      msgList,
		ReplicaId: rbft.id,
	}

	rbft.logger.Debugf("Replica %d is new primary, sending newView, v:%d, X:%+v",
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
	return rbft.primaryCheckNewView(cp, replicas, nv)
}

// recvNewView receives new view message and check if this node could
// process this message or not.
func (rbft *rbftImpl) recvNewView(nv *NewView) consensusEvent {

	rbft.logger.Debugf("Replica %d received newView %d", rbft.id, nv.View)

	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d try to receive NewView, but it's in negotiateView", rbft.id)
		return nil
	}

	if rbft.in(inRecovery) {
		rbft.logger.Debugf("Replica %d try to receive NewView, but it's in recovery", rbft.id)
		rbft.recoveryMgr.recvNewViewInRecovery = true
		return nil
	}

	if !(nv.View > 0 && nv.View >= rbft.view && rbft.primary(nv.View) == nv.ReplicaId && rbft.vcMgr.newViewStore[nv.View] == nil) {
		rbft.logger.Warningf("Replica %d reject invalid newView from %d, v:%d", rbft.id, nv.ReplicaId, nv.View)
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
		rbft.logger.Debugf("Replica %d has not meet viewChangedQuorum", rbft.id)
		return nil
	}

	return rbft.replicaCheckNewView()
}

// primaryCheckNewView do some prepare for change to New view
// such as check if primary need state update and fetch missed batches
func (rbft *rbftImpl) primaryCheckNewView(initialCp Vc_C, replicas []replicaInfo, nv *NewView) consensusEvent {

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(initialCp, replicas)
	if err != nil {
		return nil
	}

	//check if we have all block from low waterMark to recovery seq
	newReqBatchMissing := rbft.feedMissingReqBatchIfNeeded(nv.Xset)
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		return rbft.resetStateForNewView()
	} else if newReqBatchMissing {
		// if received all batches, jump into processReqInNewView
		rbft.fetchRequestBatches()
	}

	return nil
}

// replicaCheckNewView checkes this newView message and see if it's legal.
func (rbft *rbftImpl) replicaCheckNewView() consensusEvent {

	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Debugf("Replica %d ignore processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
		return nil
	}

	if !rbft.in(inViewChange) {
		rbft.logger.Infof("Replica %d ignore newView from %d, v:%d as we are in active view %d",
			rbft.id, nv.ReplicaId, nv.View, rbft.view)
		return nil
	}

	vset := rbft.getViewChanges()
	cp, ok, replicas := rbft.selectInitialCheckpoint(vset)
	if !ok {
		rbft.logger.Infof("Replica %d could not determine initial checkpoint: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}

	// Check if the xset sent by new primary is built correctly by the aset
	msgList := rbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	if msgList == nil {
		rbft.logger.Infof("Replica %d could not assign sequence numbers: %+v",
			rbft.id, rbft.vcMgr.viewChangeStore)
		return rbft.sendViewChange()
	}
	if !(len(msgList) == 0 && len(nv.Xset) == 0) && !reflect.DeepEqual(msgList, nv.Xset) {
		rbft.logger.Warningf("Replica %d failed to verify newView Xset: computed %+v, received %+v",
			rbft.id, msgList, nv.Xset)
		return rbft.sendViewChange()
	}

	// Check if primary need state update
	err := rbft.checkIfNeedStateUpdate(cp, replicas)
	if err != nil {
		return nil
	}

	return rbft.resetStateForNewView()
}

// resetStateForNewView reset all states for new view
func (rbft *rbftImpl) resetStateForNewView() consensusEvent {

	rbft.logger.Debugf("Replica %d accept newView to view %d", rbft.id, rbft.view)

	//if vcHandled active return nil, else set vcHandled active
	if rbft.in(vcHandled) {
		rbft.logger.Debugf("Replica %d enter processReqInNewView again, ignore it", rbft.id)
		return nil
	}
	rbft.on(vcHandled)

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
	//VcReset will exec sync and jump sendFinishVcReset after vcReset success
	//else if in stateTransfe or inVcReset
	//if it is primary, we should not sendFinishVcReset
	//else jump into sendFinishVcReset
	if !rbft.in(skipInProgress) &&
		!rbft.in(inVcReset) {
		rbft.helper.VcReset(backendVid)
		rbft.on(inVcReset)
	} else if rbft.isPrimary(rbft.id) {
		rbft.logger.Infof("New primary %d need to catch up other, wating", rbft.id)
	} else {
		rbft.logger.Infof("Replica %d cannot process local vcReset, but also send finishVcReset", rbft.id)
		rbft.sendFinishVcReset()
	}

	return nil
}

// recvFinishVcReset does some state check after receiving FinishVcReset message
func (rbft *rbftImpl) recvFinishVcReset(finish *FinishVcReset) consensusEvent {
	//Check whether we are in viewChange
	if !rbft.in(inViewChange) {
		rbft.logger.Infof("Replica %d is not in viewChange, but received finishVcReset from replica %d", rbft.id, finish.ReplicaId)
		return nil
	}

	//Check whether View from received finishVcReset is equal rbft.view
	if finish.View != rbft.view {
		rbft.logger.Warningf("Replica %d received finishVcReset from replica %d, expect view=%d, but get view=%d",
			rbft.id, finish.ReplicaId, rbft.view, finish.View)
		return nil
	}

	//Put received FinishVcReset stored in vcResetStore
	rbft.logger.Debugf("Replica %d received finishVcReset from replica %d, view=%d/h=%d",
		rbft.id, finish.ReplicaId, finish.View, finish.LowH)
	ok := rbft.vcMgr.vcResetStore[*finish]
	if ok {
		rbft.logger.Warningf("Replica %d ignore duplicate agree finishVcReset from %d", rbft.id, finish.ReplicaId)
		return nil
	}
	rbft.vcMgr.vcResetStore[*finish] = true

	return rbft.processReqInNewView()
}

// processReqInNewView checkes whether we can finish view change
// After recvFinishVcReset(), such as number of peers send finishVcReset.
// If view change success, processing will send VIEW_CHANGED_EVENT to rbft
func (rbft *rbftImpl) processReqInNewView() consensusEvent {

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
	if rbft.in(inVcReset) && !rbft.in(skipInProgress) {
		rbft.logger.Debugf("Replica %d itself has not done with vcReset and not in stateUpdate", rbft.id)
		return nil
	}

	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Warningf("Replica %d ignore processReqInNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
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

// FinishViewChange broadcasts FinishVcReset to other peers and
// send it to itself. Processing enters here after peer determined
// the new view and finished VCReset
func (rbft *rbftImpl) sendFinishVcReset() consensusEvent {

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
	rbft.logger.Debugf("Replica %d sending finishVcReset", rbft.id)

	return rbft.recvFinishVcReset(finish)
}

// recvFetchRequestBatch returns the requested batch
func (rbft *rbftImpl) recvFetchRequestBatch(fr *FetchRequestBatch) (err error) {

	//Check if inNegotiateView
	if rbft.in(inNegotiateView) {
		rbft.logger.Debugf("Replica %d received FetchRequestBatch, but it's in negotiateView", rbft.id)
		return nil
	}
	//Check if inRecovery
	if rbft.in(inRecovery) {
		rbft.logger.Debugf("Replica %d received FetchRequestBatch, but it's in recovery", rbft.id)
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

// recvReturnRequestBatch receives the RequestBatch from other peers
// If receive all request batch, processing jump to processReqInNewView
// or processReqInUpdate
func (rbft *rbftImpl) recvReturnRequestBatch(batch *ReturnRequestBatch) consensusEvent {

	//Check if in inNegotiateView
	if rbft.in(inNegotiateView) {
		rbft.logger.Infof("Replica %d try to receive returnRequestBatch, but it's in negotiateView", rbft.id)
		return nil
	}
	//Check if in inRecovery
	if rbft.in(inRecovery) {
		rbft.logger.Infof("Replica %d try to receive returnRequestBatch, but it's in recovery", rbft.id)
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
	rbft.logger.Debugf("Primary received missing request: ", digest)

	//if receive all request batch
	//if validatedBatchStore jump to processReqInNewView
	//if inUpdatingN jump to processReqInUpdate
	if len(rbft.storeMgr.missingReqBatches) == 0 {
		if rbft.in(inViewChange) {
			_, ok := rbft.vcMgr.newViewStore[rbft.view]
			if !ok {
				rbft.logger.Warningf("Replica %d ignore processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
				return nil
			}
			return rbft.resetStateForNewView()
		}
		if rbft.in(inUpdatingN) {
			update, ok := rbft.nodeMgr.updateStore[rbft.nodeMgr.updateTarget]
			if !ok {
				rbft.logger.Warningf("Replica %d ignore processUpdateN as it could not find target %v in its updateStore", rbft.id, rbft.nodeMgr.updateTarget)
				return nil
			}
			return rbft.resetStateForUpdate(update)
		}
	}
	return nil

}

//##########################################################################
//           view change auxiliary functions
//##########################################################################

// calcQSet selects Pre-prepares which satisfy the following conditions
// 1. Pre-prepares in previous qlist
// 2. Pre-prepares from certStore which is preprepared and (its view <= its idx.v or not in qlist
func (rbft *rbftImpl) calcQSet() map[qidx]*Vc_PQ {

	qset := make(map[qidx]*Vc_PQ)

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

		qset[qi] = &Vc_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    idx.d,
			View:           idx.v,
		}
	}

	return qset
}

// calcPSet selects prepares which satisfy the following conditions:
// 1. prepares in previous qlist
// 2. prepares from certStore which is prepared and (its view <= its idx.v or not in plist)
func (rbft *rbftImpl) calcPSet() map[uint64]*Vc_PQ {

	pset := make(map[uint64]*Vc_PQ)

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

		pset[idx.n] = &Vc_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    idx.d,
			View:           idx.v,
		}
	}

	return pset
}

// stopNewViewTimer stops NEW_VIEW_TIMER
func (rbft *rbftImpl) stopNewViewTimer() {

	rbft.logger.Debugf("Replica %d stop a running newView timer", rbft.id)
	rbft.off(timerActive)
	rbft.timerMgr.stopTimer(NEW_VIEW_TIMER)
}

// startNewViewTimer stops all running new view timers and start a new view timer
func (rbft *rbftImpl) startNewViewTimer(timeout time.Duration, reason string) {

	rbft.logger.Debugf("Replica %d start newView timer for %s: %s", rbft.id, timeout, reason)
	rbft.vcMgr.newViewTimerReason = reason
	rbft.on(timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	rbft.timerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, rbft.eventMux)
}

// softstartNewViewTimer starts a new view timer no matter how many existed new view timer
func (rbft *rbftImpl) softStartNewViewTimer(timeout time.Duration, reason string) {

	rbft.logger.Debugf("Replica %d soft start newView timer for %s: %s", rbft.id, timeout, reason)
	rbft.vcMgr.newViewTimerReason = reason
	rbft.on(timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}

	rbft.timerMgr.startTimerWithNewTT(NEW_VIEW_TIMER, timeout, event, rbft.eventMux)
}

// beforeSendVC operates before send view change
// 1. Check rbft.state. State should not inNegotiateView or inRecovery
// 2. Stop NewViewTimer and NULL_REQUEST_TIMER
// 3. increase the view and delete new view of old view in newViewStore
// 4. update pqlist
// 5. delete old viewChange message
func (rbft *rbftImpl) beforeSendVC() error {

	if rbft.in(inNegotiateView) {
		rbft.logger.Errorf("Replica %d try to send viewChange, but it's in negotiateView", rbft.id)
		return errors.New("node is in negotiateView now")
	}

	if rbft.in(inRecovery) {
		rbft.logger.Errorf("Replica %d try to send viewChange, but it's in recovery", rbft.id)
		return errors.New("node is in recovery now")
	}

	rbft.stopNewViewTimer()
	rbft.timerMgr.stopTimer(NULL_REQUEST_TIMER)

	delete(rbft.vcMgr.newViewStore, rbft.view)
	rbft.view++
	rbft.on(inViewChange)
	rbft.off(vcHandled)
	rbft.setAbNormal()

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

// correctViewChange checkes if view change messages correct
// 1. pqsets' view should be less then vc.View and SequenceNumber should greater then vc.H.
// 2. checkpoint's SequenceNumber should greater then vc.H
func (rbft *rbftImpl) correctViewChange(vc *ViewChange) bool {

	for _, p := range append(vc.Basis.Pset, vc.Basis.Qset...) {
		if !(p.View < vc.Basis.View && p.SequenceNumber > vc.Basis.H) {
			rbft.logger.Debugf("Replica %d find invalid p entry in viewChange: vc(v:%d h:%d) p(v:%d n:%d)",
				rbft.id, vc.Basis.View, vc.Basis.H, p.View, p.SequenceNumber)
			return false
		}
	}

	for _, c := range vc.Basis.Cset {
		if !(c.SequenceNumber >= vc.Basis.H) {
			rbft.logger.Debugf("Replica %d find invalid c entry in viewChange: vc(v:%d h:%d) c(n:%d)",
				rbft.id, vc.Basis.View, vc.Basis.H, c.SequenceNumber)
			return false
		}
	}

	return true
}

// getViewChanges returns all viewChange message from viewChangeStore
func (rbft *rbftImpl) getViewChanges() (vset []*VcBasis) {
	for _, vc := range rbft.vcMgr.viewChangeStore {
		vset = append(vset, vc.Basis)
	}
	return
}

// gatherPQC just gather all checkpoints, p entries and q entries.
func (rbft *rbftImpl) gatherPQC() (cset []*Vc_C, pset []*Vc_PQ, qset []*Vc_PQ) {
	// Gather all the checkpoints
	for n, id := range rbft.storeMgr.chkpts {
		cset = append(cset, &Vc_C{
			SequenceNumber: n,
			Id:             id,
		})
	}
	// Gather all the p entries
	for _, p := range rbft.vcMgr.plist {
		if p.SequenceNumber < rbft.h {
			rbft.logger.Errorf("Replica %d should not have anything in our pset less than h, found %+v", rbft.id, p)
			continue
		}
		pset = append(pset, p)
	}

	// Gather all the q entries
	for _, q := range rbft.vcMgr.qlist {
		if q.SequenceNumber < rbft.h {
			rbft.logger.Errorf("Replica %d should not have anything in our qset less than h, found %+v", rbft.id, q)
			continue
		}
		qset = append(qset, q)
	}

	return
}

// selectInitialCheckpoint selects checkpoint from received ViewChange message
// If find suitable checkpoint, it return a certain checkpoint and the replicas
// id list which replicas has this checkpoint.
// The checkpoint is the max checkpoint which exists in at least oneCorrectQuorum
// peers and greater then low waterMark in at least commonCaseQuorum.
func (rbft *rbftImpl) selectInitialCheckpoint(set []*VcBasis) (checkpoint Vc_C, find bool, replicas []replicaInfo) {

	// For the checkpoint as key, find the corresponding AgreeUpdateN messages
	checkpoints := make(map[Vc_C][]*VcBasis)
	for _, agree := range set {
		// Verify that we strip duplicate checkpoints from this Cset
		set := make(map[Vc_C]bool)
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

// assignSequenceNumbers finds the suitable batches for recovery to according
// to ViewChange and low waterMark.
// The selected batches match following condition: If batch is not a NullRequest
// batch, the pre-prepare of this batch is equal or greater than commonCaseQuorum
// and the prepare is equal or greater then oneCorrectQuorum.
// in this release, batch should not be NUllRequest batch
func (rbft *rbftImpl) assignSequenceNumbers(set []*VcBasis, h uint64) map[uint64]string {

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

// updateViewChangeSeqNo updates viewChangeSeqNo by viewChangePeriod
func (rbft *rbftImpl) updateViewChangeSeqNo(seqNo, K, id uint64) {

	if rbft.vcMgr.viewChangePeriod <= 0 {
		return
	}
	// Ensure the view change always occurs at a checkpoint boundary
	rbft.vcMgr.viewChangeSeqNo = seqNo - seqNo%K + rbft.vcMgr.viewChangePeriod*K
}

// feedMissingReqBatchIfNeeded feeds needed reqBatch when this node
// doesn't have all reqBatch in xset.
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
				rbft.logger.Debugf("Primary %d missing assigned, non-checkpointed request batch %s", rbft.id, d)
				if _, ok := rbft.storeMgr.missingReqBatches[d]; !ok {
					rbft.logger.Infof("Replica %v try to fetch batch %s", rbft.id, d)
					newReqBatchMissing = true
					rbft.storeMgr.missingReqBatches[d] = true
				}
			}
		}
	}
	return newReqBatchMissing
}

// primaryResendBatch validates batches which has seq > low watermark
func (rbft *rbftImpl) primaryResendBatch(xset Xset) {

	// reset validateCount before new primary validate batches.
	rbft.batchVdr.validateCount = 0
	xSetLen := len(xset)
	upper := uint64(xSetLen) + rbft.h + uint64(1)
	for i := rbft.h + uint64(1); i < upper; i++ {
		d, ok := xset[i]
		if !ok {
			rbft.logger.Errorf("viewChange Xset miss batch number %d", i)
		} else if d == "" {
			// This should not happen
			rbft.logger.Errorf("viewChange Xset has null batch, kick it out")
		} else {
			batch, ok := rbft.storeMgr.txBatchStore[d]
			if !ok {
				rbft.logger.Errorf("in Xset %s exists, but in Replica %d validatedBatchStore there is no such batch digest", d, rbft.id)
			} else if i > rbft.exec.lastExec {
				rbft.primaryValidateBatch(d, batch, i)
			}
		}
	}

}

// rebuildCertStoreForVC rebuilds cert according to xset
func (rbft *rbftImpl) rebuildCertStoreForVC() {

	//Check whether new view has stored in newViewStore
	nv, ok := rbft.vcMgr.newViewStore[rbft.view]
	if !ok {
		rbft.logger.Debugf("Replica %d ignore processNewView as it could not find view %d in its newViewStore", rbft.id, rbft.view)
		return
	}
	//do rebuild cert
	rbft.rebuildCertStore(nv.Xset)
}

// rebuildCertStore rebuilds certStore according to Xset
// Broadcast qpc for batches which has been confirmed in view change.
// So that, all correct peers will reach the seq that select in view change
func (rbft *rbftImpl) rebuildCertStore(xset Xset) {

	for n, d := range xset {
		if n <= rbft.h || n > rbft.exec.lastExec {
			continue
		}
		batch, ok := rbft.storeMgr.txBatchStore[d]
		if !ok && d != "" {
			rbft.logger.Errorf("Replica %d is missing tx batch for seqNo=%d with digest '%s' for assigned prepare", rbft.id)
			continue
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
		rbft.vcMgr.qlist = make(map[qidx]*Vc_PQ)
		rbft.vcMgr.plist = make(map[uint64]*Vc_PQ)

		id := qidx{d, n}
		pqItem := &Vc_PQ{
			SequenceNumber: n,
			BatchDigest:    d,
			View:           rbft.view,
		}
		//update vcMgr.pqlist according to Xset
		rbft.vcMgr.qlist[id] = pqItem
		rbft.vcMgr.plist[n] = pqItem

	}
}
