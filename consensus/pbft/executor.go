//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"fmt"
	"sync/atomic"

	"hyperchain/consensus"
	"hyperchain/consensus/events"
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
func (pbft *pbftImpl) msgToEvent(msg *ConsensusMessage) (interface{}, error) {
	event := eventCreators[msg.Type]().(proto.Message)
	err := proto.Unmarshal(msg.Payload, event)
	if err != nil {
		pbft.logger.Errorf("Unmarshal error, can not unmarshal %v, error: %v", msg.Type, err)
		return nil, err
	}

	return event, nil
}

// dispatchLocalEvent dispatches local Event to corresponding handles using its service type
func (pbft *pbftImpl) dispatchLocalEvent(e *LocalEvent) events.Event {
	switch e.Service {
	case CORE_PBFT_SERVICE:
		return pbft.handleCorePbftEvent(e)
	case VIEW_CHANGE_SERVICE:
		return pbft.handleViewChangeEvent(e)
	case NODE_MGR_SERVICE:
		return pbft.handleNodeMgrEvent(e)
	case RECOVERY_SERVICE:
		return pbft.handleRecoveryEvent(e)
	default:
		pbft.logger.Errorf("Not Supported event: %v", e)
		return nil
	}
}

// handleCorePbftEvent handles core PBFT service events
func (pbft *pbftImpl) handleCorePbftEvent(e *LocalEvent) events.Event {
	switch e.EventType {

	case CORE_BATCH_TIMER_EVENT:
		pbft.logger.Debugf("Primary %d batch timer expires, try to create a batch", pbft.id)
		pbft.stopBatchTimer()
		// call txPool module to generate a tx batch
		pbft.batchMgr.txPool.GenerateTxBatch()
		return nil

	case CORE_NULL_REQUEST_TIMER_EVENT:
		pbft.handleNullRequestTimerEvent()
		return nil

	case CORE_FIRST_REQUEST_TIMER_EVENT:
		pbft.logger.Noticef("Replica %d first request timer expires", pbft.id)
		return pbft.sendViewChange()

	case CORE_STATE_UPDATE_EVENT:
		pbft.recvStateUpdatedEvent(e.Event.(protos.StateUpdatedMessage))
		return nil

	case CORE_VALIDATED_TXS_EVENT:
		pbft.recvValidatedResult(e.Event.(protos.ValidatedTxs))
		return nil

	default:
		pbft.logger.Errorf("Invalid core pbft event : %v", e)
		return nil
	}
}

// handleViewChangeEvent handles view change service related events.
func (pbft *pbftImpl) handleViewChangeEvent(e *LocalEvent) events.Event {
	switch e.EventType {
	case VIEW_CHANGE_TIMER_EVENT:
		pbft.logger.Warningf("Replica %d view change timer expired, sending view change: %s", pbft.id, pbft.vcMgr.newViewTimerReason)
		pbft.status.inActiveState(&pbft.status.timerActive)

		// Here, we directly send viewchange with a bigger target view (which is pbft.view+1) because it is the
		// new view timer who triggered this VIEW_CHANGE_TIMER_EVENT so we send a new viewchange request
		return pbft.sendViewChange()

	case VIEW_CHANGED_EVENT:
		// set a viewChangeSeqNo if needed
		pbft.vcMgr.updateViewChangeSeqNo(pbft.seqNo, pbft.K, pbft.id)

		pbft.startTimerIfOutstandingRequests()
		pbft.vcMgr.vcResendCount = 0
		pbft.vcMgr.vcResetStore = make(map[FinishVcReset]bool)
		primary := pbft.primary(pbft.view)

		// inform p2p module to reset primary peer's information in routing table as primary may have been changed
		pbft.helper.InformPrimary(primary)

		pbft.persistView(pbft.view)
		atomic.StoreUint32(&pbft.activeView, 1)
		pbft.status.inActiveState(&pbft.status.vcHandled)

		// set normal to 1 which indicates system comes into normal status after viewchange
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 &&
			!pbft.status.getState(&pbft.status.inNegoView) && !pbft.status.getState(&pbft.status.skipInProgress) {
			atomic.StoreUint32(&pbft.normal, 1)
		}
		pbft.logger.Criticalf("======== Replica %d finished viewChange, primary=%d, view=%d/height=%d", pbft.id, primary, pbft.view, pbft.exec.lastExec)
		viewChangeResult := fmt.Sprintf("Replica %d finished viewChange, primary=%d, view=%d/height=%d", pbft.id, primary, pbft.view, pbft.exec.lastExec)

		// send viewchange result to web socket API
		pbft.helper.SendFilterEvent(consensus.FILTER_View_Change_Finish, viewChangeResult)
		if pbft.status.getState(&pbft.status.isNewNode) {
			pbft.sendReadyForN()
			return nil
		}
		// rebuild certStore using Xset
		pbft.rebuildCertStoreForVC()
		pbft.handleTransactionsAfterAbnormal()

	case VIEW_CHANGE_RESEND_TIMER_EVENT:
		if atomic.LoadUint32(&pbft.activeView) == 1 {
			pbft.logger.Warningf("Replica %d had its view change resend timer expire but it's in an active view, this is benign but may indicate a bug", pbft.id)
		}
		pbft.logger.Warningf("Replica %d view change resend timer expired before view change quorum was reached, resending", pbft.id)

		// after send viewchange, if triggered the viewchange resend timeout before receive N-f
		// viewchange whose vc.view==pbft.view, we will resend viewchange with the same target view as the
		// previous viewchange request, because we cannot collect enough viewchange requests to start new view.
		// Also, if the resend count reaches the resend limit, stop this VIEW_CHANGE_RESEND_TIMER then come
		// into recovery.
		pbft.view--
		return pbft.sendViewChange()

	case VIEW_CHANGE_QUORUM_EVENT:
		pbft.logger.Debugf("Replica %d received view change quorum, processing new view", pbft.id)
		if pbft.status.getState(&pbft.status.inNegoView) {
			pbft.logger.Debugf("Replica %d try to process viewChangeQuorumEvent, but it's in nego-view", pbft.id)
			return nil
		}
		if ok := pbft.isPrimary(); ok {
			// if we are catching up, don't send new view as a primary and after a while, other nodes will
			// send a new viewchange whose seqNo=previous viewchange's seqNo + 1 because of new view timeout
			// and eventually others will finish viewchange with a new view in which primary is not in
			// skipInProgress
			if pbft.status.getState(&pbft.status.skipInProgress) {
				return nil
			}
			// primary construct and send new view message
			return pbft.sendNewView()
		}
		return pbft.processNewView()

	case VIEW_CHANGE_VC_RESET_DONE_EVENT:
		pbft.status.inActiveState(&pbft.status.inVcReset)
		pbft.logger.Debugf("Replica %d received local VcResetDone", pbft.id)
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
			return pbft.sendFinishUpdate()
		}
		var seqNo uint64
		var event protos.VcResetDone
		var ok bool
		if event, ok = e.Event.(protos.VcResetDone); !ok {
			pbft.logger.Error("type assert error!")
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
		if pbft.status.getState(&pbft.status.inRecovery) && pbft.recoveryMgr.recoveryToSeqNo != nil {
			if seqNo-1 >= *pbft.recoveryMgr.recoveryToSeqNo {
				return &LocalEvent{
					Service:   RECOVERY_SERVICE,
					EventType: RECOVERY_DONE_EVENT,
				}
			} else {
				state := protos.StateUpdatedMessage{SeqNo: seqNo - 1}
				return pbft.recvStateUpdatedEvent(state)
			}
		}
		if atomic.LoadUint32(&pbft.activeView) == 1 {
			pbft.logger.Warningf("Replica %d is not in viewChange, but received local VcResetDone", pbft.id)
			return nil
		}

		if seqNo != pbft.exec.lastExec+1 {
			pbft.logger.Warningf("Replica %d finds error in VcResetDone, expect=%d, but get=%d", pbft.id, pbft.exec.lastExec+1, seqNo)
			return nil
		}

		return pbft.finishViewChange()

	default:
		pbft.logger.Errorf("Invalid view change event event : %v", e)
		return nil
	}
	return nil
}

// handleNodeMgrEvent handles node management service related events.
func (pbft *pbftImpl) handleNodeMgrEvent(e *LocalEvent) events.Event {
	var err error
	switch e.EventType {
	case NODE_MGR_NEW_NODE_EVENT:
		err = pbft.recvLocalNewNode(e.Event.(*protos.NewNodeMessage))
	case NODE_MGR_ADD_NODE_EVENT:
		err = pbft.recvLocalAddNode(e.Event.(*protos.AddNodeMessage))
	case NODE_MGR_DEL_NODE_EVENT:
		err = pbft.recvLocalDelNode(e.Event.(*protos.DelNodeMessage))
	case NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT:
		pbft.logger.Debugf("Replica %d received agree-update-n quorum, processing update-n", pbft.id)
		if pbft.status.getState(&pbft.status.inNegoView) {
			pbft.logger.Debugf("Replica %d try to process agreeUpdateNQuorumEvent, but it's in nego-view", pbft.id)
			return nil
		}
		if pbft.primary(pbft.view) == pbft.id {
			return pbft.sendUpdateN()
		}
		return pbft.processUpdateN()
	case NODE_MGR_UPDATEDN_EVENT:
		delete(pbft.nodeMgr.updateStore, pbft.nodeMgr.updateTarget)
		pbft.startTimerIfOutstandingRequests()
		pbft.vcMgr.vcResendCount = 0
		pbft.nodeMgr.finishUpdateStore = make(map[FinishUpdate]bool)
		pbft.persistView(pbft.view)
		pbft.persistNewNode(uint64(0))
		pbft.persistDellLocalKey()
		pbft.persistN(pbft.N)
		pbft.status.inActiveState(&pbft.status.updateHandled)
		if pbft.status.getState(&pbft.status.isNewNode) {
			pbft.status.inActiveState(&pbft.status.isNewNode)
		}
		atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 0)
		pbft.rebuildCertStoreForUpdate()
		if atomic.LoadUint32(&pbft.activeView) == 1 &&
			!pbft.status.getState(&pbft.status.inNegoView) && !pbft.status.getState(&pbft.status.skipInProgress) {
			atomic.StoreUint32(&pbft.normal, 1)
		}
		pbft.logger.Criticalf("======== Replica %d finished UpdatingN, primary=%d, n=%d/f=%d/view=%d/h=%d", pbft.id, pbft.primary(pbft.view), pbft.N, pbft.f, pbft.view, pbft.h)
		pbft.handleTransactionsAfterAbnormal()

	default:
		pbft.logger.Errorf("Invalid view change event event : %v", e)
		return nil
	}

	if err != nil {
		pbft.logger.Warning(err.Error())
	}

	return nil
}

// handleRecoveryEvent handles recovery services related events.
func (pbft *pbftImpl) handleRecoveryEvent(e *LocalEvent) events.Event {
	switch e.EventType {
	case RECOVERY_DONE_EVENT:
		pbft.status.inActiveState(&pbft.status.inRecovery)
		pbft.recoveryMgr.recoveryToSeqNo = nil
		pbft.timerMgr.stopTimer(RECOVERY_RESTART_TIMER)
		pbft.logger.Criticalf("======== Replica %d finished recovery, height: %d", pbft.id, pbft.exec.lastExec)

		// if we received new view or UpdateN during recovery, we will restart recovery after finish this round
		// of recovery, as view has been changed during recovery
		if pbft.recoveryMgr.recvNewViewInRecovery {
			pbft.logger.Noticef("#  Replica %d find itself received NewView during Recovery"+
				", will restart negotiate view", pbft.id)
			pbft.status.activeState(&pbft.status.inRecovery, &pbft.status.inNegoView)
			pbft.recoveryMgr.recvNewViewInRecovery = false
			pbft.restartNegoView()
			return nil
		}
		// after recovery, new primary need to send null request as a heartbeat, and non-primary will start a
		// first request timer which must be longer than null request timer in which non-primary must receive a
		// request from primary(null request or pre-prepare...), or this node will send viewchange
		if pbft.primary(pbft.view) == pbft.id {
			pbft.sendNullRequest()
		} else {
			event := &LocalEvent{
				Service:   CORE_PBFT_SERVICE,
				EventType: CORE_FIRST_REQUEST_TIMER_EVENT,
			}

			pbft.timerMgr.startTimer(FIRST_REQUEST_TIMER, event, pbft.pbftEventQueue)
		}

		// if this recovery was triggered by 10 viewchange, inactive vcToRecovery
		if pbft.status.getState(&pbft.status.vcToRecovery) {
			pbft.status.inActiveState(&pbft.status.vcToRecovery)
		}
		if pbft.status.getState(&pbft.status.isNewNode) {
			pbft.sendReadyForN()
			return nil
		}

		// here, we always fetch PQC after finish recovery as we only recovery to the largest checkpoint which
		// is lower or equal to the lastExec quorum of others, which, in this way, we avoid sending prepare and
		// commit or other consensus messages during add/delete node
		pbft.fetchRecoveryPQC()

		pbft.handleTransactionsAfterAbnormal()

		// execute after recovery using the PQC information received during recovery
		// NOTICE: these PQC are not the PQC fetched above as fetched PQC are executed after recvRecoveryReturnPQC
		// these PQC are received during recovery whose seqNo may be higher than fetched PQC
		pbft.executeAfterStateUpdate()
		return nil

	case RECOVERY_NEGO_VIEW_DONE_EVENT:
		// set normal to 1 which indicates system comes into normal status after negotiate done
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 0 &&
			atomic.LoadUint32(&pbft.activeView) == 1 && !pbft.status.getState(&pbft.status.skipInProgress) {
			atomic.StoreUint32(&pbft.normal, 1)
		}
		pbft.logger.Criticalf("======== Replica %d finished negotiating view: %d / N=%d", pbft.id, pbft.view, pbft.N)
		primary := pbft.primary(pbft.view)

		// re-construct certStore if this recovery was triggered by 10 viewchange as view may have been changed
		if pbft.status.getState(&pbft.status.vcToRecovery) {
			pbft.parseSpecifyCertStore()
		}
		// clean useless cache which may influence subsequent consensus process
		pbft.cleanAllCache()
		pbft.persistView(pbft.view)

		// inform p2p module to reset primary peer's information in routing table as primary may have been changed
		pbft.helper.InformPrimary(primary)

		pbft.initRecovery()
		return nil

	case RECOVERY_NEGO_VIEW_RSP_TIMER_EVENT:
		if !pbft.status.getState(&pbft.status.inNegoView) {
			pbft.logger.Warningf("Replica %d had its nego-view response timer expire but it's not in nego-view, this is benign but may indicate a bug", pbft.id)
		}
		pbft.logger.Debugf("Replica %d nego-view response timer expired before N-f was reached, resending", pbft.id)
		pbft.restartNegoView()
		return nil
	case RECOVERY_RESTART_TIMER_EVENT:
		pbft.logger.Noticef("Replica %d recovery restart timer expires", pbft.id)
		pbft.restartRecovery()
		return nil
	default:
		pbft.logger.Errorf("Invalid recovery service events : %v", e)
		return nil
	}
}

// dispatchConsensusMsg dispatches consensus messages to corresponding handlers using its service type
func (pbft *pbftImpl) dispatchConsensusMsg(e events.Event) events.Event {
	pbft.logger.Debugf("start processing consensus message")
	service := pbft.dispatchMsgToService(e)
	switch service {
	case CORE_PBFT_SERVICE:
		return pbft.dispatchCorePbftMsg(e)
	case VIEW_CHANGE_SERVICE:
		return pbft.dispatchViewChangeMsg(e)
	case NODE_MGR_SERVICE:
		return pbft.dispatchNodeMgrMsg(e)
	case RECOVERY_SERVICE:
		return pbft.dispatchRecoveryMsg(e)
	default:
		pbft.logger.Errorf("Not Supported event: %v", e)
	}
	return nil
}

// dispatchMsgToService returns the service type of the given event. There exist 4 service types:
// 1. CORE_PBFT_SERVICE: including tx related events, pre-prepare, prepare, commit, checkpoint, missing txs related events
// 2. VIEW_CHANGE_SERVICE
// 3. RECOVERY_SERVICE
// 4. NODE_MGR_SERVICE
func (pbft *pbftImpl) dispatchMsgToService(e events.Event) int {
	switch e.(type) {
	// core PBFT service
	case *TransactionBatch:
		return CORE_PBFT_SERVICE
	case *PrePrepare:
		return CORE_PBFT_SERVICE
	case *Prepare:
		return CORE_PBFT_SERVICE
	case *Commit:
		return CORE_PBFT_SERVICE
	case *Checkpoint:
		return CORE_PBFT_SERVICE
	case *FetchMissingTransaction:
		return CORE_PBFT_SERVICE
	case *ReturnMissingTransaction:
		return CORE_PBFT_SERVICE

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
