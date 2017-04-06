package fsm

import (
	"fmt"
	"github.com/looplab/fsm"
	"testing"
)

func TestSyncFSM(t *testing.T) {
	fsm := NewSyncFSM(prepareCallbacks())
	fsm.Event(SyncEvent_INITIALIZE.String())
	fsm.Event(SyncEvent_SELECT_PEER.String())
	fsm.Event(SyncEvent_SEND_REQ.String())
	fsm.Event(SyncEvent_SEND_REQ.String())
	fsm.Event(SyncEvent_RESEND_REQ.String())
	fsm.Event(SyncEvent_RESEND_REQ.String())
	fsm.Event(SyncEvent_RESELECT_PEER.String())
	fsm.Event(SyncEvent_SEND_REQ.String())
	fsm.Event(SyncEvent_EXEC.String())
	fsm.Event(SyncEvent_EXEC_DONE.String())
	if !fsm.Is(SyncFSM_Done) {
		t.Error("unexpect behavior")
	}
}

func prepareCallbacks() fsm.Callbacks {
	fns := make(fsm.Callbacks)
	fns["enter_" + SyncFSM_Prepare] = prepareCallback
	fns["enter_" + SyncFSM_Select] = selectCallback
	fns["enter_" + SyncFSM_Wait] = waitCallback
	fns["enter_" + SyncFSM_Exec] = execCallback
	fns["enter_" + SyncFSM_Done] = doneCallback
	return fns
}

func prepareCallback(event *fsm.Event) {
	fmt.Printf("prepare callback, receive event %s\n", event.Event)
}

func selectCallback(event *fsm.Event) {
	fmt.Printf("select callback, receive event %s\n", event.Event)
}

func waitCallback(event *fsm.Event) {
	fmt.Printf("wait callback, receive event %s\n", event.Event)
}

func execCallback(event *fsm.Event) {
	fmt.Printf("exec callback, receive event %s\n", event.Event)
}

func doneCallback(event *fsm.Event) {
	fmt.Printf("done callback, receive event %s\n", event.Event)
}
