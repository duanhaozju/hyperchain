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
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
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

func TestHandleCorePbftEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
	atomic.StoreUint32(&rbft.activeView, 1)

	event := &LocalEvent{
		Service:   CORE_RBFT_SERVICE,
		EventType: CORE_BATCH_TIMER_EVENT,
	}
	rbft.startBatchTimer()
	ast.Equal(true, rbft.batchMgr.isBatchTimerActive(), "startBatchTimer failed")
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.batchMgr.isBatchTimerActive(), "handle CORE_BATCH_TIMER_EVENT failed")

	atomic.StoreUint32(&rbft.activeView, 1)
	event.EventType = CORE_NULL_REQUEST_TIMER_EVENT
	rbft.dispatchLocalEvent(event)
	ast.Equal(uint32(0), atomic.LoadUint32(&rbft.activeView), "handle CORE_NULL_REQUEST_TIMER_EVENT failed")

	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
	atomic.StoreUint32(&rbft.activeView, 1)
	atomic.StoreUint32(&rbft.normal, 1)
	event.EventType = CORE_FIRST_REQUEST_TIMER_EVENT
	rbft.dispatchLocalEvent(event)
	ast.Equal(uint32(0), atomic.LoadUint32(&rbft.activeView), "handle CORE_FIRST_REQUEST_TIMER_EVENT failed")

	rbft.status.activeState(&rbft.status.stateTransferring)
	event.EventType = CORE_STATE_UPDATE_EVENT
	event.Event = protos.StateUpdatedMessage{SeqNo: uint64(10)}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.status.getState(&rbft.status.stateTransferring))

	atomic.StoreUint32(&rbft.activeView, 1)
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 0)
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
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.status.activeState(&rbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_TIMER_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.status.getState(&rbft.status.timerActive), "handle CORE_BATCH_TIMER_EVENT failed")
}

func TestViewChangedEventAndViewChangeResendTimerEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.status.inActiveState(&rbft.status.skipInProgress)
	atomic.StoreUint32(&rbft.activeView, 0)
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 0)
	rbft.status.activeState(&rbft.status.timerActive)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGED_EVENT,
	}
	rbft.dispatchLocalEvent(event)

	ast.Equal(0, rbft.vcMgr.vcResendCount, "handle VIEW_CHANGED_EVENT, set vcResendCount failed")
	ast.Equal(uint32(1), atomic.LoadUint32(&rbft.activeView), "handle VIEW_CHANGED_EVENT failed")
	ast.Equal(uint32(1), atomic.LoadUint32(&rbft.normal), "handle VIEW_CHANGED_EVENT failed")

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
	rbft.status.activeState(&rbft.status.inVcReset)

	event := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGE_VC_RESET_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.status.getState(&rbft.status.inVcReset), "handle VIEW_CHANGE_VC_RESET_DONE_EVENT failed")

	event.Event = protos.VcResetDone{
		SeqNo: uint64(100),
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.status.getState(&rbft.status.inVcReset), "handle VIEW_CHANGE_VC_RESET_DONE_EVENT failed")
}

func TestHandleNodeMgrEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.activeState(&rbft.status.updateHandled)

	event := &LocalEvent{
		Service:   NODE_MGR_SERVICE,
		EventType: NODE_MGR_UPDATEDN_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(0, rbft.vcMgr.vcResendCount, "handle NODE_MGR_UPDATEDN_EVENT, set vcResendCount failed")
	ast.Equal(false, rbft.status.getState(&rbft.status.updateHandled), "handle NODE_MGR_UPDATEDN_EVENT, inactive updateHandled failed")

	event.EventType = NODE_MGR_AGREE_UPDATEN_QUORUM_EVENT
	rbft.dispatchLocalEvent(event)
}

func TestRecoveryDoneEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.status.activeState(&rbft.status.vcToRecovery)
	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(false, rbft.status.getState(&rbft.status.inRecovery), "handle RECOVERY_DONE_EVENT, inactive inRecovery failed")
	ast.Equal(false, rbft.status.getState(&rbft.status.vcToRecovery), "handle RECOVERY_DONE_EVENT, inactive vcToRecovery failed")
}

func TestRecoveryNegoViewDoneEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	atomic.StoreUint32(&rbft.nodeMgr.inUpdatingN, 0)
	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.status.inActiveState(&rbft.status.skipInProgress)

	event := &LocalEvent{
		Service:   RECOVERY_SERVICE,
		EventType: RECOVERY_NEGO_VIEW_DONE_EVENT,
	}
	rbft.dispatchLocalEvent(event)
	ast.Equal(uint32(1), atomic.LoadUint32(&rbft.normal), "handle RECOVERY_NEGO_VIEW_DONE_EVENT failed")
}
