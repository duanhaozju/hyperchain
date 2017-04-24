package jcee

import "testing"

func TestConnMaintainer_Transition(t *testing.T) {
	maintainer := NewConnMaintainer(nil)
	if maintainer.fsm.Current() != conn_init {
		t.Error("invalid init status")
	}
	// initialize
	maintainer.fsm.Event(initialize)
	if maintainer.fsm.Current() != conn_health {
		t.Error("invalid init transit")
	}

	// ping success
	maintainer.fsm.Event(ping_success)
	if maintainer.pc != 0 {
		t.Error("invalid ping counter value")
	}

	// ping failed
	maintainer.fsm.Event(ping_failed)
	if maintainer.fsm.Current() != conn_sick {
		t.Error("invalid ping failed transit")
	}
	if maintainer.pc != 1 {
		t.Error("invalid ping counter value")
	}

	// ping failed second time
	maintainer.fsm.Event(ping_failed)
	if maintainer.fsm.Current() != conn_sick {
		t.Error("invalid ping failed transit")
	}
	if maintainer.pc != 2 {
		t.Error("invalid ping counter value")
	}

	// ping success
	maintainer.fsm.Event(ping_success)
	if maintainer.fsm.Current() != conn_health {
		t.Error("invalid ping success transit")
	}
	if maintainer.pc != 1 {
		t.Error("invalid ping counter value")
	}

	// ping failed 4 time
	for i := 0; i < 3; i += 1 {
		maintainer.fsm.Event(ping_failed)
		if maintainer.fsm.Current() != conn_sick {
			t.Error("invalid ping failed transit")
		}
		if maintainer.pc != 2 + int32(i) {
			t.Error("invalid ping counter value")
		}
	}
}
