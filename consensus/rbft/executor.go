//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"sync/atomic"

	"hyperchain/consensus"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

// executor manages exec related params
type executor struct {
	lastExec    uint64
	currentExec *uint64
}

// newExecutor initializes an instance of executor
func newExecutor() *executor {
	exec := &executor{}
	return exec
}

// setLastExec sets the value of lastExec
func (e *executor) setLastExec(l uint64) {
	e.lastExec = l
}

// setCurrentExec sets the value of pointer currentExec
func (e *executor) setCurrentExec(c *uint64) {
	e.currentExec = c
}

// msgToEvent converts ConsensusMessage to the corresponding consensus event.
func (rbft *rbftImpl) msgToEvent(msg *ConsensusMessage) (interface{}, error) {
	event := eventCreators[msg.Type]().(proto.Message)
	err := proto.Unmarshal(msg.Payload, event)
	if err != nil {
		rbft.logger.Errorf("Unmarshal error, can not unmarshal %v, error: %v", msg.Type, err)
		return nil, err
	}

	return event, nil
}

// dispatchLocalEvent dispatches local Event to corresponding handles using its service type
func (rbft *rbftImpl) dispatchLocalEvent(e *LocalEvent) consensusEvent {
	switch e.Service {
	case CORE_RBFT_SERVICE:
		return rbft.handleCorePbftEvent(e)
	case VIEW_CHANGE_SERVICE:
		return rbft.handleViewChangeEvent(e)
	case NODE_MGR_SERVICE:
		return rbft.handleNodeMgrEvent(e)
	case RECOVERY_SERVICE:
		return rbft.handleRecoveryEvent(e)
	default:
		rbft.logger.Errorf("Not Supported event: %v", e)
		return nil
	}
}

// handleCorePbftEvent handles core RBFT service events
func (rbft *rbftImpl) handleCorePbftEvent(e *LocalEvent) consensusEvent {
	switch e.EventType {

	case CORE_BATCH_TIMER_EVENT:
		rbft.logger.Debugf("Primary %d batch timer expires, try to create a batch", rbft.id)
		rbft.stopBatchTimer()
		// call txPool module to generate a tx batch
		rbft.batchMgr.txPool.GenerateTxBatch()
		return nil

	case CORE_NULL_REQUEST_TIMER_EVENT:
		rbft.handleNullRequestTimerEvent()
		return nil

	case CORE_FIRST_REQUEST_TIMER_EVENT:
		rbft.logger.Noticef("Replica %d first request timer expires", rbft.id)
		return rbft.sendViewChange()

	case CORE_STATE_UPDATE_EVENT:
		rbft.recvStateUpdatedEvent(e.Event.(protos.StateUpdatedMessage))
		return nil

	case CORE_VALIDATED_TXS_EVENT:
		rbft.recvValidatedResult(e.Event.(protos.ValidatedTxs))
		return nil

	default:
		rbft.logger.Errorf("Invalid core rbft event : %v", e)
		return nil
	}
}

// handleViewChangeEvent handles view change service related events.
func (rbft *rbftImpl) handleViewChangeEvent(e *LocalEvent) consensusEvent {
	switch e.EventType {
	case VIEW_CHANGE_TIMER_EVENT:
		rbft.logger.Warningf("Replica %d view change timer expired, sending view change: %s", rbft.id, rbft.vcMgr.newViewTimerReason)
		rbft.status.inActiveState(&rbft.status.timerActive)

		// Here, we directly send viewchange with a bigger target view (which is rbft.view+1) because it is the
		// new view timer who triggered this VIEW_CHANGE_TIMER_EVENT so we send a new viewchange request
		return rbft.sendViewChange()

	case VIEW_CHANGED_EVENT:
		// set a viewChangeSeqNo if needed
		rbft.vcMgr.updateViewChangeSeqNo(rbft.seqNo, rbft.K, rbft.id)

		rbft.startTimerIfOutstandingRequests()
		rbft.vcMgr.vcResendCount = 0
		rbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
		primary := rbft.primary(rbft.view)

		// inform p2p module to reset primary peer's information in routing table as primary may have been changed
		rbft.helper.InformPrimary(primary)

		rbft.persistView(rbft.view)
		rbft.status.inActiveState(&rbft.status.inViewChange)
		rbft.status.inActiveState(&rbft.status.vcHandled)

		// set normal to 1 which indicates system comes into normal status after viewchange
		if !rbft.status.getState(&rbft.status.inUpdatingN) &&
			!rbft.status.getState(&rbft.status.inNegoView) &&
			!rbft.status.getState(&rbft.status.skipInProgress) {
			atomic.StoreUint32(&rbft.normal, 1)
		}
		rbft.batchVdr.validateCount = 0
		rbft.logger.Criticalf("======== Replica %d finished viewChange, primary=%d, view=%d/height=%d", rbft.id, primary, rbft.view, rbft.exec.lastExec)
		viewChangeResult := fmt.Sprintf("Replica %d finished viewChange, primary=%d, view=%d/height=%d", rbft.id, primary, rbft.view, rbft.exec.lastExec)

		// send viewchange result to web socket API
		rbft.helper.SendFilterEvent(consensus.FILTER_View_Change_Finish, viewChangeResult)
		if rbft.status.getState(&rbft.status.isNewNode) {
			rbft.sendReadyForN()
			return nil
		}
		// rebuild certStore using Xset
		rbft.rebuildCertStoreForVC()
		rbft.handleTransactionsAfterAbnormal()

	case VIEW_CHANGE_RESEND_TIMER_EVENT:
		if !rbft.status.getState(&rbft.status.inViewChange) {
			rbft.logger.Warningf("Replica %d had its view change resend timer expire but it's in an active view, this is benign but may indicate a bug", rbft.id)
		}
		rbft.logger.Warningf("Replica %d view change resend timer expired before view change quorum was reached, resending", rbft.id)

		// after send viewchange, if triggered the viewchange resend timeout before receive N-f
		// viewchange whose vc.view==rbft.view, we will resend viewchange with the same target view as the
		// previous viewchange request, because we cannot collect enough viewchange requests to start new view.
		// Also, if the resend count reaches the resend limit, stop this VIEW_CHANGE_RESEND_TIMER then come
		// into recovery.
		rbft.view--
		return rbft.sendViewChange()

	case VIEW_CHANGE_QUORUM_EVENT:
		rbft.logger.Debugf("Replica %d received view change quorum, processing new view", rbft.id)
		if rbft.status.getState(&rbft.status.inNegoView) {
			rbft.logger.Debugf("Replica %d try to process viewChangeQuorumEvent, but it's in nego-view", rbft.id)
			return nil
		}
		if rbft.isPrimary(rbft.id) {
			// if we are catching up, don't send new view as a primary and after a while, other nodes will
			// send a new viewchange whose seqNo=previous viewchange's seqNo + 1 because of new view timeout
			// and eventually others will finish viewchange with a new view in which primary is not in
			// skipInProgress
			if rbft.status.getState(&rbft.status.skipInProgress) {
				return nil
			}
			// primary construct and send new view message
			return rbft.sendNewView()
		}
		return rbft.replicaCheckNewView()

	case VIEW_CHANGE_VC_RESET_DONE_EVENT:
		rbft.status.inActiveState(&rbft.status.inVcReset)
		rbft.logger.Debugf("Replica %d received local VcResetDone", rbft.id)
		if rbft.status.getState(&rbft.status.inUpdatingN) {
			return rbft.sendFinishUpdate()
		}
		var seqNo uint64
		var event protos.VcResetDone
		var ok bool
		if event, ok = e.Event.(protos.VcResetDone); !ok {
			rbft.logger.Error("type assert error!")
			return nil
		}
		seqNo = event.SeqNo
		// if we start VcReset in recovery, we may encounter 2 cases such as:
		// 1. in recovery, we have executed to 25, but others only executed to 28, so our recoveryToSeqNo == 20,
		// and lastExec == 25, need to VcReset to 25, after VcResetDone quickly, we can return recovery done directly
		// 2. in recovery, we have executed to 25, but others only executed to 28, so our recoveryToSeqNo == 20,
		// and lastExec == 25, need to VcReset to 25, but during VcReset which may be a little slow, others may
		// execute to 30+ or 40+..., which triggered moveWatermarks in recvCheckpoint(), recoveryToSeqNo may have
		// been changed to 30 or 40 or bigger, in this case, after VcResetDone, we will come into
		// recvStateUpdatedEvent in which we will retryStateTransfer to the new checkpoint
		if rbft.status.getState(&rbft.status.inRecovery) && rbft.recoveryMgr.recoveryToSeqNo != nil {
			if seqNo-1 >= *rbft.recoveryMgr.recoveryToSeqNo {
				return &LocalEvent{
					Service:   RECOVERY_SERVICE,
					EventType: RECOVERY_DONE_EVENT,
				}
			} else {
				state := protos.StateUpdatedMessage{SeqNo: seqNo - 1}
				return rbft.recvStateUpdatedEvent(state)
			}
		}
		if !rbft.status.getState(&rbft.status.inViewChange) {
			rbft.logger.Warningf("Replica %d is not in viewChange, but received local VcResetDone", rbft.id)
			return nil
		}

		if seqNo != rbft.exec.lastExec+1 {
			rbft.logger.Warningf("Replica %d finds error in VcResetDone, expect=%d, but get=%d", rbft.id, rbft.exec.lastExec+1, seqNo)
			return nil
		}

		return rbft.finishViewChange()

	default:
		rbft.logger.Errorf("Invalid view change event event : %v", e)
		return nil
	}
	return nil
}

// handleNodeMgrEvent handles node management service related events.
func (rbft *rbftImpl) handleNodeMgrEvent(e *LocalEvent) consensusEvent {
	var err error
	switch e.EventType {
	case NODE_MGR_NEW_NODE_EVENT:
		err = rbft.recvLocalNewNode(e.Event.(*protos.NewNodeMessage))
	case NODE_MGR_ADD_NODE_EVENT:
		err = rbft.recvLocalAddNode(e.Event.(*protos.AddNodeMessage))
	case NODE_MGR_DEL_NODE_EVENT:
		err = rbft.recvLocalDelNode(e.Event.(*protos.DelNodeMessage))
	case NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT:
		rbft.logger.Debugf("Replica %d received agree-update-n quorum, processing update-n", rbft.id)
		if rbft.status.getState(&rbft.status.inNegoView) {
			rbft.logger.Debugf("Replica %d try to process agreeUpdateNQuorumEvent, but it's in nego-view", rbft.id)
			return nil
		}
		if rbft.isPrimary(rbft.id) {
			return rbft.sendUpdateN()
		}
		return rbft.replicaCheckUpdateN()
	case NODE_MGR_UPDATEDN_EVENT:
		rbft.startTimerIfOutstandingRequests()
		rbft.vcMgr.vcResendCount = 0
		rbft.nodeMgr.finishUpdateStore = make(map[FinishUpdate]bool)
		rbft.persistView(rbft.view)
		rbft.persistNewNode(uint64(0))
		rbft.persistDellLocalKey()
		rbft.persistN(rbft.N)
		rbft.status.inActiveState(&rbft.status.updateHandled)
		if rbft.status.getState(&rbft.status.isNewNode) {
			rbft.status.inActiveState(&rbft.status.isNewNode)
		}
		rbft.status.inActiveState(&rbft.status.inUpdatingN)
		rbft.rebuildCertStoreForUpdate()
		if !rbft.status.getState(&rbft.status.inViewChange) &&
			!rbft.status.getState(&rbft.status.inNegoView) && !rbft.status.getState(&rbft.status.skipInProgress) {
			atomic.StoreUint32(&rbft.normal, 1)
		}
		rbft.logger.Criticalf("======== Replica %d finished UpdatingN, primary=%d, n=%d/f=%d/view=%d/h=%d", rbft.id, rbft.primary(rbft.view), rbft.N, rbft.f, rbft.view, rbft.h)
		rbft.handleTransactionsAfterAbnormal()
		delete(rbft.nodeMgr.updateStore, rbft.nodeMgr.updateTarget)

	default:
		rbft.logger.Errorf("Invalid view change event event : %v", e)
		return nil
	}

	if err != nil {
		rbft.logger.Warning(err.Error())
	}

	return nil
}

// handleRecoveryEvent handles recovery services related events.
func (rbft *rbftImpl) handleRecoveryEvent(e *LocalEvent) consensusEvent {
	switch e.EventType {
	case RECOVERY_DONE_EVENT:
		rbft.status.inActiveState(&rbft.status.inRecovery)
		rbft.recoveryMgr.recoveryToSeqNo = nil
		rbft.timerMgr.stopTimer(RECOVERY_RESTART_TIMER)
		rbft.logger.Criticalf("======== Replica %d finished recovery, height: %d", rbft.id, rbft.exec.lastExec)

		// if we received new view or UpdateN during recovery, we will restart recovery after finish this round
		// of recovery, as view has been changed during recovery
		if rbft.recoveryMgr.recvNewViewInRecovery {
			rbft.logger.Noticef("#  Replica %d find itself received NewView during Recovery"+
				", will restart negotiate view", rbft.id)
			rbft.status.activeState(&rbft.status.inRecovery, &rbft.status.inNegoView)
			rbft.recoveryMgr.recvNewViewInRecovery = false
			rbft.restartNegoView()
			return nil
		}
		// after recovery, new primary need to send null request as a heartbeat, and non-primary will start a
		// first request timer which must be longer than null request timer in which non-primary must receive a
		// request from primary(null request or pre-prepare...), or this node will send viewchange
		if rbft.isPrimary(rbft.id) {
			rbft.sendNullRequest()
		} else {
			event := &LocalEvent{
				Service:   CORE_RBFT_SERVICE,
				EventType: CORE_FIRST_REQUEST_TIMER_EVENT,
			}

			rbft.timerMgr.startTimer(FIRST_REQUEST_TIMER, event, rbft.eventMux)
		}

		// if this recovery was triggered by 10 viewchange, inactive vcToRecovery
		if rbft.status.getState(&rbft.status.vcToRecovery) {
			rbft.status.inActiveState(&rbft.status.vcToRecovery)
		}
		if rbft.status.getState(&rbft.status.isNewNode) {
			rbft.sendReadyForN()
			return nil
		}

		// here, we always fetch PQC after finish recovery as we only recovery to the largest checkpoint which
		// is lower or equal to the lastExec quorum of others, which, in this way, we avoid sending prepare and
		// commit or other consensus messages during add/delete node
		rbft.fetchRecoveryPQC()

		rbft.handleTransactionsAfterAbnormal()

		// execute after recovery using the PQC information received during recovery
		// NOTICE: these PQC are not the PQC fetched above as fetched PQC are executed after recvRecoveryReturnPQC
		// these PQC are received during recovery whose seqNo may be higher than fetched PQC
		rbft.executeAfterStateUpdate()
		return nil

	case RECOVERY_NEGO_VIEW_DONE_EVENT:
		// set normal to 1 which indicates system comes into normal status after negotiate done
		if !rbft.status.getState(&rbft.status.inUpdatingN) &&
			!rbft.status.getState(&rbft.status.inViewChange) &&
			!rbft.status.getState(&rbft.status.skipInProgress) {
			atomic.StoreUint32(&rbft.normal, 1)
		}
		rbft.logger.Criticalf("======== Replica %d finished negotiating view: %d / N=%d", rbft.id, rbft.view, rbft.N)
		primary := rbft.primary(rbft.view)

		// re-construct certStore if this recovery was triggered by 10 viewchange as view may have been changed
		if rbft.status.getState(&rbft.status.vcToRecovery) {
			rbft.parseSpecifyCertStore()
		}
		// clean useless cache which may influence subsequent consensus process
		rbft.cleanAllCache()
		rbft.persistView(rbft.view)

		// inform p2p module to reset primary peer's information in routing table as primary may have been changed
		rbft.helper.InformPrimary(primary)

		rbft.initRecovery()
		return nil

	case RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT:
		if !rbft.status.getState(&rbft.status.inNegoView) {
			rbft.logger.Warningf("Replica %d had its nego-view response timer expire but it's not in nego-view, this is benign but may indicate a bug", rbft.id)
		}
		rbft.logger.Debugf("Replica %d nego-view response timer expired before N-f was reached, resending", rbft.id)
		rbft.restartNegoView()
		return nil
	case RECOVERY_RESTART_TIMER_EVENT:
		rbft.logger.Noticef("Replica %d recovery restart timer expires", rbft.id)
		rbft.restartRecovery()
		return nil
	default:
		rbft.logger.Errorf("Invalid recovery service events : %v", e)
		return nil
	}
}

// dispatchConsensusMsg dispatches consensus messages to corresponding handlers using its service type
func (rbft *rbftImpl) dispatchConsensusMsg(e consensusEvent) consensusEvent {
	rbft.logger.Debugf("start processing consensus message")
	service := rbft.dispatchMsgToService(e)
	switch service {
	case CORE_RBFT_SERVICE:
		return rbft.dispatchCorePbftMsg(e)
	case VIEW_CHANGE_SERVICE:
		return rbft.dispatchViewChangeMsg(e)
	case NODE_MGR_SERVICE:
		return rbft.dispatchNodeMgrMsg(e)
	case RECOVERY_SERVICE:
		return rbft.dispatchRecoveryMsg(e)
	default:
		rbft.logger.Errorf("Not Supported event: %v", e)
	}
	return nil
}

// dispatchMsgToService returns the service type of the given event. There exist 4 service types:
// 1. CORE_RBFT_SERVICE: including tx related events, pre-prepare, prepare, commit, checkpoint, missing txs related events
// 2. VIEW_CHANGE_SERVICE
// 3. RECOVERY_SERVICE
// 4. NODE_MGR_SERVICE
func (rbft *rbftImpl) dispatchMsgToService(e consensusEvent) int {
	switch e.(type) {
	// core RBFT service
	case *TransactionBatch:
		return CORE_RBFT_SERVICE
	case *PrePrepare:
		return CORE_RBFT_SERVICE
	case *Prepare:
		return CORE_RBFT_SERVICE
	case *Commit:
		return CORE_RBFT_SERVICE
	case *Checkpoint:
		return CORE_RBFT_SERVICE
	case *FetchMissingTransaction:
		return CORE_RBFT_SERVICE
	case *ReturnMissingTransaction:
		return CORE_RBFT_SERVICE

		// view change service
	case *ViewChange:
		return VIEW_CHANGE_SERVICE
	case *NewView:
		return VIEW_CHANGE_SERVICE
	case *FetchRequestBatch:
		return VIEW_CHANGE_SERVICE
	case *ReturnRequestBatch:
		return VIEW_CHANGE_SERVICE
	case *FinishVcReset:
		return VIEW_CHANGE_SERVICE

		// recovery service
	case *RecoveryInit:
		return RECOVERY_SERVICE
	case *NegotiateView:
		return RECOVERY_SERVICE
	case *NegotiateViewResponse:
		return RECOVERY_SERVICE
	case *RecoveryResponse:
		return RECOVERY_SERVICE
	case *RecoveryFetchPQC:
		return RECOVERY_SERVICE
	case *RecoveryReturnPQC:
		return RECOVERY_SERVICE

		// node_mgr service
	case *AddNode:
		return NODE_MGR_SERVICE
	case *DelNode:
		return NODE_MGR_SERVICE
	case *ReadyForN:
		return NODE_MGR_SERVICE
	case *UpdateN:
		return NODE_MGR_SERVICE
	case *AgreeUpdateN:
		return NODE_MGR_SERVICE
	case *FinishUpdate:
		return NODE_MGR_SERVICE
	default:
		return NOT_SUPPORT_SERVICE

	}
	return NOT_SUPPORT_SERVICE
}
