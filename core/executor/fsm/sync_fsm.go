package fsm

import "github.com/looplab/fsm"

func NewSyncFSM(callbacks map[string]func(e *fsm.Event)) *fsm.FSM {
	return fsm.NewFSM(
		SyncStatus_INIT,
		fsm.Events{
			{Name: SyncEvent_SELECT_PEER, Src: []string{SyncStatus_INIT}, Dst: SyncStatus_SELECT},
			{Name: SyncEvent_SEND_REQ, Src: []string{SyncStatus_SELECT}, Dst: SyncStatus_WAIT},
			{Name: SyncEvent_SEND_REQ, Src: []string{SyncStatus_WAIT}, Dst: SyncStatus_WAIT},
			{Name: SyncEvent_RESEND_REQ, Src: []string{SyncStatus_WAIT}, Dst: SyncStatus_WAIT},
			{Name: SyncEvent_RESELECT_PEER, Src: []string{SyncStatus_WAIT}, Dst: SyncStatus_SELECT},
			{Name: SyncEvent_EXEC, Src: []string{SyncStatus_WAIT}, Dst: SyncStatus_EXEC},
			{Name: SyncEvent_EXEC_DONE, Src: []string{SyncStatus_EXEC}, Dst: SyncStatus_DONE},
		},
		fsm.Callbacks{
			"enter_" + SyncStatus_INIT:   callbacks[SyncStatus_INIT],
			"enter_" + SyncStatus_SELECT: callbacks[SyncStatus_SELECT],
			"enter_" + SyncStatus_WAIT:   callbacks[SyncStatus_WAIT],
			"enter_" + SyncStatus_EXEC:   callbacks[SyncStatus_EXEC],
			"enter_" + SyncStatus_DONE:   callbacks[SyncStatus_DONE],
			"enter_state": TraceEnter,
			"leave_state": TraceLeave,
		},
	)
}

func TraceEnter(event *fsm.Event) {

}
func TraceLeave(event *fsm.Event) {

}
