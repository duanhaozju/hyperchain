//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/consensus/events"
	"hyperchain/manager/protos"
	"sync/atomic"
)

type executor struct {
	lastExec         uint64
	currentExec      *uint64
}

func newExecutor() *executor {
	exec := &executor{}
	return exec
}

func (e *executor) setLastExec(l uint64)  {
	e.lastExec = l
}

func (e *executor) setCurrentExec(c *uint64)  {
	e.currentExec = c
}

//map: ConsensusMessage_Type, event create functions.
var eventCreators map[ConsensusMessage_Type]func() interface{}

//msgToEvent transfer ConsensusMessage to the related event.
func (pbft *pbftImpl) msgToEvent (msg *ConsensusMessage) (interface{}, error) {

	event := eventCreators[msg.Type]().(proto.Message)
	err := proto.Unmarshal(msg.Payload, event)
	if err != nil {
		pbft.logger.Errorf("Unmarshal error, can not unmarshal %v, error: %v", msg.Type, err)
		return nil, err
	}

	return event, nil
}

//dispatchMsgToService dispatch messgae to the related services.
//TODO: refactor consensus message format
func (pbft *pbftImpl) dispatchMsgToService(e events.Event) int  {
	switch e.(type) {
	//core PBFT service
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

	//view change service
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

	//recovery service
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

	//node_mgr service
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
	default:
		return NOT_SUPPORT_SERVICE

	}
	return NOT_SUPPORT_SERVICE
}

//local event process functions
//1.core pbft services
//2.viewchange services
//3.nodemgr services
//4.recovery services

//dispatchLocalEvent dispatch local Event.
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

//handleCorePbftEvent handler core PBFT service events.
func (pbft *pbftImpl) handleCorePbftEvent(e *LocalEvent) events.Event {
	switch e.EventType {
	case CORE_BATCH_TIMER_EVENT:
		pbft.logger.Debugf("Replica %d batch timer expired", pbft.id)
		if atomic.LoadUint32(&pbft.activeView) == 1 {
			return pbft.sendBatchRequest()
		}
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

//handleViewChangeEvent handle view change service events.
func (pbft *pbftImpl) handleViewChangeEvent(e *LocalEvent) events.Event {
	switch e.EventType {
	case VIEW_CHANGE_TIMER_EVENT:
		pbft.logger.Warningf("Replica %d view change timer expired, sending view change: %s", pbft.id, pbft.vcMgr.newViewTimerReason)
		pbft.status.inActiveState(&pbft.status.timerActive)
		return pbft.sendViewChange()

	case VIEW_CHANGED_EVENT:
		pbft.vcMgr.updateViewChangeSeqNo(pbft.seqNo, pbft.K, pbft.id)
		pbft.startTimerIfOutstandingRequests()
		pbft.vcMgr.vcResendCount = 0
		primary := pbft.primary(pbft.view)
		pbft.helper.InformPrimary(primary)
		pbft.persistView(pbft.view)
		atomic.StoreUint32(&pbft.activeView, 1)
		pbft.status.inActiveState(&pbft.status.vcHandled)
		pbft.logger.Criticalf("======== Replica %d finished viewChange, primary=%d, view=%d/h=%d", pbft.id, primary, pbft.view, pbft.h)
		if pbft.status.getState(&pbft.status.isNewNode){
			pbft.sendReadyForN()
			return nil
		}
		pbft.processRequestsDuringViewChange()

	case VIEW_CHANGE_RESEND_TIMER_EVENT:
		if atomic.LoadUint32(&pbft.activeView) == 1 {
			pbft.logger.Warningf("Replica %d had its view change resend timer expire but it's in an active view, this is benign but may indicate a bug", pbft.id)
		}
		pbft.logger.Warningf("Replica %d view change resend timer expired before view change quorum was reached, resending", pbft.id)
		pbft.view-- // sending the view change increments this
		return pbft.sendViewChange()

	case VIEW_CHANGE_QUORUM_EVENT:
		pbft.logger.Debugf("Replica %d received view change quorum, processing new view", pbft.id)
		if pbft.status.getState(&pbft.status.inNegoView) {
			pbft.logger.Debugf("Replica %d try to process viewChangeQuorumEvent, but it's in nego-view", pbft.id)
			return nil
		}
		if ok, _ := pbft.isPrimary(); ok {
			return pbft.sendNewView()
		}
		return pbft.processNewView()

	case VIEW_CHANGE_VC_RESET_DONE_EVENT:
		pbft.status.inActiveState(&pbft.status.inVcReset)
		pbft.logger.Debugf("Replica %d received local VcResetDone", pbft.id)
		if atomic.LoadUint32(&pbft.nodeMgr.inUpdatingN) == 1 {
			if pbft.primary(pbft.view) == pbft.id {
				return pbft.handleTailAfterUpdate()
			} else if pbft.id == uint64(pbft.N) {
				return pbft.sendFinishUpdate()
			} else {
				return &LocalEvent{
					Service:   NODE_MGR_SERVICE,
					EventType: NODE_MGR_UPDATEDN_EVENT,
				}
			}
		}
		if e.Event.(protos.VcResetDone).SeqNo != pbft.h+1 {
			pbft.logger.Warningf("Replica %d finds error in VcResetDone, expect=%d, but get=%d", pbft.id, pbft.h+1, e.Event.(protos.VcResetDone).SeqNo)
			return nil
		}
		if pbft.status.getState(&pbft.status.inRecovery) {
			state := protos.StateUpdatedMessage{SeqNo: e.Event.(protos.VcResetDone).SeqNo - 1}
			return pbft.recvStateUpdatedEvent(state)
		}
		if atomic.LoadUint32(&pbft.activeView) == 1 {
			pbft.logger.Warningf("Replica %d is not in viewChange, but received local VcResetDone", pbft.id)
			return nil
		}

		if pbft.primary(pbft.view) == pbft.id {
			return pbft.handleTailInNewView()
		}
		return pbft.finishViewChange()

	default:
		pbft.logger.Errorf("Invalid view change event event : %v", e)
		return nil
	}
	return nil
}

//handleNodeMgrEvent handle node management service events.
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
		pbft.persistView(pbft.view)
		pbft.persistNewNode(uint64(0))
		pbft.persistDellLocalKey()
		pbft.persistN(pbft.N)
		pbft.status.inActiveState(&pbft.status.updateHandled)
		if pbft.status.getState(&pbft.status.isNewNode) {
			pbft.status.inActiveState(&pbft.status.isNewNode)
		}
		atomic.StoreUint32(&pbft.nodeMgr.inUpdatingN, 0)
		pbft.processRequestsDuringUpdatingN()
		pbft.logger.Criticalf("======== Replica %d finished UpdatingN, primary=%d, n=%d/f=%d/view=%d/h=%d", pbft.id, pbft.primary(pbft.view), pbft.N, pbft.f, pbft.view, pbft.h)
	}

	if err != nil {
		pbft.logger.Warning(err.Error())
	}
	return nil
}

//handleRecoveryEvent handle recovery services events.
func (pbft *pbftImpl) handleRecoveryEvent(e *LocalEvent) events.Event {
	switch e.EventType {
	case RECOVERY_DONE_EVENT:
		pbft.logger.Criticalf("======== Replica %d finished recovery, height: %d", pbft.id, pbft.exec.lastExec)
		if pbft.recoveryMgr.recvNewViewInRecovery {
			pbft.logger.Noticef("#  Replica %d find itself received NewView during Recovery"+
				", will restart negotiate view", pbft.id)
			pbft.status.activeState(&pbft.status.inRecovery,&pbft.status.inNegoView)
			pbft.recoveryMgr.recvNewViewInRecovery = false
			pbft.restartNegoView()
			return nil
		}
		if pbft.status.getState(&pbft.status.isNewNode) {
			pbft.sendReadyForN()
			return nil
		}
		pbft.processRequestsDuringRecovery()
		return nil

	case RECOVERY_NEGO_VIEW_DONE_EVENT:
		pbft.logger.Criticalf("======== Replica %d finished negotiating view: %d", pbft.id, pbft.view)
		primary := pbft.primary(pbft.view)
		if primary == pbft.id {
			pbft.sendNullRequest()
		} else {
			event := &LocalEvent{
				Service:   CORE_PBFT_SERVICE,
				EventType: CORE_FIRST_REQUEST_TIMER_EVENT,
			}

			af := func(){
				pbft.pbftEventQueue.Push(event)
			}

			pbft.pbftTimerMgr.startTimer(FIRST_REQUEST_TIMER, af)
		}
		pbft.persistView(pbft.view)
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

//Consensus messages process functions
//1.core pbft services
//2.viewchange services
//3.nodemgr services
//4.recovery services

func (pbft *pbftImpl) dispatchConsensusMsg(e events.Event) events.Event {
	pbft.logger.Debugf("start processing consensus mesage")
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
