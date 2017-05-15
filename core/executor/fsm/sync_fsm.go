package fsm

import "github.com/looplab/fsm"

const (
	SyncFSM_Init    = "init"
	SyncFSM_Prepare = "prepare"
	SyncFSM_Select  = "select"
	SyncFSM_Wait    = "wait"
	SyncFSM_Exec    = "exec"
	SyncFSM_Done    = "done"
)

func NewSyncFSM(callbacks fsm.Callbacks) *fsm.FSM {
	return fsm.NewFSM(
		SyncFSM_Init,
		fsm.Events{
			{Name: SyncEvent_INITIALIZE.String(), Src: []string{SyncFSM_Init}, Dst: SyncFSM_Prepare},
			{Name: SyncEvent_SELECT_PEER.String(), Src: []string{SyncFSM_Prepare}, Dst: SyncFSM_Select},
			{Name: SyncEvent_SEND_REQ.String(), Src: []string{SyncFSM_Select}, Dst: SyncFSM_Wait},
			{Name: SyncEvent_SEND_REQ.String(), Src: []string{SyncFSM_Wait}, Dst: SyncFSM_Wait},
			{Name: SyncEvent_RESEND_REQ.String(), Src: []string{SyncFSM_Wait}, Dst: SyncFSM_Wait},
			{Name: SyncEvent_RESELECT_PEER.String(), Src: []string{SyncFSM_Wait}, Dst: SyncFSM_Select},
			{Name: SyncEvent_EXEC.String(), Src: []string{SyncFSM_Wait}, Dst: SyncFSM_Exec},
			{Name: SyncEvent_EXEC_DONE.String(), Src: []string{SyncFSM_Exec}, Dst: SyncFSM_Done},
		},
		callbacks,
	)
}
