package p2p

import (
	"hyperchain/p2p/message"
	"time"
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
	addr            *message.PeerAddr
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
