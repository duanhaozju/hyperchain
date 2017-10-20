//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"sync/atomic"
	"testing"

	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestExecutor(t *testing.T) {
	ast := assert.New(t)
	exec := newExecutor()

	lastExec := uint64(10)
	currentExec := uint64(11)
	exec.setCurrentExec(&currentExec)
	exec.setLastExec(lastExec)

	ast.Equal(lastExec, exec.lastExec, "set lastExec failed")
	ast.Equal(currentExec, *exec.currentExec, "set currentExec failed")
}

func TestMsgToEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.initMsgEventMap()

	pp := &PrePrepare{
		View:           uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:    "1",
		ResultHash:     "1",
		HashBatch:      nil,
		ReplicaId:      uint64(0),
	}
	payload, err := proto.Marshal(pp)
	msg := &ConsensusMessage{
		Type:    ConsensusMessage_PRE_PREPARE,
		Payload: payload,
	}
	event, err := rbft.msgToEvent(msg)
	ast.EqualValues(pp, event, "msgToEvent failed")
	ast.Nil(err, "msgToEvent failed")
}

func TestHandleCoreRbftEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.off(inViewChange)

	event := &LocalEvent{
		Service:   CORE_RBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}
	rbft.startBatchTimer()
	ast.Equal(true, rbft.batchMgr.isBatchTimerActive(), "startBatchTimer failed")
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.batchMgr.isBatchTimerActive(), "handle CORE_BATCH_TIMER_EVENT failed")

	rbft.off(inViewChange)
	event.EventType = CORE_NULL_REQUEST_TIMER_EVENT
	rbft.dispatchLocalEvent(event)
	ast.Equal(true, rbft.in(inViewChange), "handle CORE_NULL_REQUEST_TIMER_EVENT failed")

	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.off(inViewChange)
	rbft.setNormal()
	event.EventType = CORE_FIRST_REQUEST_TIMER_EVENT
	rbft.dispatchLocalEvent(event)
	ast.Equal(true, rbft.in(inViewChange), "handle CORE_FIRST_REQUEST_TIMER_EVENT failed")

	rbft.on(stateTransferring)
	event.EventType = CORE_STATE_UPDATE_EVENT
	event.Event = protos.StateUpdatedMessage{SeqNo: uint64(10)}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.in(stateTransferring))

	rbft.off(inViewChange)
	rbft.off(inUpdatingN)
	event.EventType = CORE_VALIDATED_TXS_EVENT
	event.Event = protos.ValidatedTxs{}
	rbft.dispatchLocalEvent(event)

	event.EventType = -100
	rbft.dispatchLocalEvent(event)
}

func TestViewChangeTimerEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.off(inViewChange)
	rbft.on(timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.in(timerActive), "handle CORE_BATCH_TIMER_EVENT failed")
}

func TestViewChangedEventAndViewChangeResendTimerEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.off(skipInProgress)
	rbft.on(inViewChange)
	rbft.off(inUpdatingN)
	rbft.on(timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGED_EVENT,
	}
	rbft.dispatchLocalEvent(event)

	ast.Equal(0, rbft.vcMgr.vcResendCount, "handle VIEW_CHANGED_EVENT, set vcResendCount failed")
	ast.Equal(false, rbft.in(inViewChange), "handle VIEW_CHANGED_EVENT failed")
	ast.Equal(true, rbft.isNormal(), "handle VIEW_CHANGED_EVENT failed")

	rbft.view = uint64(10)
	event.EventType = VIEW_CHANGE_RESEND_TIMER_EVENT
	rbft.dispatchLocalEvent(event)
	ast.Equal(uint64(10), rbft.view, "handle VIEW_CHANGE_RESEND_TIMER_EVENT failed")

	event.EventType = -100
	rbft.dispatchLocalEvent(event)
}

func TestViewChangeVCResetDoneEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.on(inVcReset)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_VC_RESET_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.in(inVcReset), "handle VIEW_CHANGE_VC_RESET_DONE_EVENT failed")

	event.Event = protos.VcResetDone{
		SeqNo: uint64(100),
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.in(inVcReset), "handle VIEW_CHANGE_VC_RESET_DONE_EVENT failed")
}

func TestHandleNodeMgrEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.on(updateHandled)

	event := &LocalEvent{
		Service:   NODE_MGR_SERVICE,
		EventType: NODE_MGR_UPDATEDN_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(0, rbft.vcMgr.vcResendCount, "handle NODE_MGR_UPDATEDN_EVENT, set vcResendCount failed")
	ast.Equal(false, rbft.in(updateHandled), "handle NODE_MGR_UPDATEDN_EVENT, inactive updateHandled failed")

	event.EventType = NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT
	rbft.dispatchLocalEvent(event)
}

func TestRecoveryDoneEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.on(vcToRecovery)
	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.in(inRecovery), "handle RECOVERY_DONE_EVENT, inactive inRecovery failed")
	ast.Equal(false, rbft.in(vcToRecovery), "handle RECOVERY_DONE_EVENT, inactive vcToRecovery failed")
}

func TestRecoveryNegoViewDoneEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.off(inUpdatingN)
	rbft.off(inViewChange)
	rbft.off(skipInProgress)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_NEGO_VIEW_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(true, rbft.isNormal(), "handle RECOVERY_NEGO_VIEW_DONE_EVENT failed")
}
