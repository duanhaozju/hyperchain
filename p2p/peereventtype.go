package p2p

import (
	"time"
	"hyperchain/p2p/peermessage"
)

type PendingEvent struct {
	P *Peer
}

//keep alive event
type KeepAliveEvent struct {
	Interval time.Duration
}

//close peer
type CloseEvent struct {

}
//retry event
type RetryEvent struct {
	RetryTimeout time.Duration
	RetryTimes   int
}

type RecoveryEvent struct {
	addr            *peermessage.PeerAddr
	recoveryTimeout time.Duration
	recoveryTimes   int
}

// self narrate
type SelfNarrateEvent struct {
	content string
}



/*
Event life time:
Start -> KeepAlive -> Retry -> close -> End
             ^           |
             |           |
             -------------

 */