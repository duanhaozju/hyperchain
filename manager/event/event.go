//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package event

import (
	"github.com/hyperchain/hyperchain/core/types"
)

type AliveEvent struct{ Payload bool }

// send consensus event to outer peers for consensus module
type BroadcastConsensusEvent struct{ Payload []byte }

//general tx local
type NewTxEvent struct {
	Transaction *types.Transaction
	Simulate    bool
	SnapshotId  string
	Ch          chan bool
}

type TxUniqueCastEvent struct {
	Payload []byte
	PeerId  uint64
}

type SessionEvent struct {
	Message []byte
}

//TODO:transfer all this event to proto defined

/* Peer Maintain Event */
//a new peer past ca validation
type NewPeerEvent struct {
	Payload []byte //ip port marshaled
}

//broadcast local ca validation result for new peer to all replicas
//payload is a consenus message after encoding
type BroadcastNewPeerEvent struct {
	Payload []byte
}

//a new peer's join chain request has been accept, update routing table
// type: true for add, false for delete
type UpdateRoutingTableEvent struct {
	Payload []byte
	Type    bool
}

// update routing table finished
type AlreadyInChainEvent struct {
}

// delete VP peer event
type DelVPEvent struct {
	Payload []byte
}

// delete NVP peer event
type DelNVPEvent struct {
	Payload []byte
}

type BroadcastDelPeerEvent struct {
	Payload []byte
}

/*
	Non verified peer events definition
*/

type VerifiedBlock struct {
	Payload []byte
}

type ReceiveVerifiedBlock struct {
	Payload []byte
}

type CommitedBlockEvent struct {
	Payload []byte
}

/*
	Admin events
*/

type DeleteSnapshotEvent struct {
	FilterId string
	Cont     chan error //TODO: need to fix
}

type ArchiveEvent struct {
	FilterId string
	Cont     chan error //TODO: need to fix
	Sync     bool
}

type ArchiveRestoreEvent struct {
	FilterId string
	Ack      chan error
	Sync     bool
	All      bool
}

// receive tx from a nvp
type NvpRelayTxEvent struct {
	Payload []byte
}

type BloomEvent struct {
	Namespace string
	Cont     chan error
	Transactions []*types.Transaction
}
