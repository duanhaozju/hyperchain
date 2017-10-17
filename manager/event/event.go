//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package event

import (
	"hyperchain/core/types"
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

//node receive checkpoint sync event and then,check db and send block require request to peers
type SyncReplica struct {
	Id      uint64
	Height  uint64
	Genesis uint64
}

type ChainSyncReqEvent struct {
	Id              uint64
	TargetHeight    uint64
	TargetBlockHash []byte
	Replicas        []SyncReplica
}

type SessionEvent struct {
	Message []byte
}

////set primary in peerManager when new view and primary
//type InformPrimaryEvent struct {
//	Primary uint64
//}
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
	Executor events
*/
type ExecutorToConsensusEvent struct {
	Payload interface{}
	Type    int
}

type ExecutorToP2PEvent struct {
	Payload   []byte
	Type      int
	Peers     []uint64
	PeersHash []string
}

/*
	Admin events
*/

type SnapshotEvent struct {
	FilterId    string
	BlockNumber uint64
}

type DeleteSnapshotEvent struct {
	FilterId string
	Cont     chan error
}

type ArchiveEvent struct {
	FilterId string
	Cont     chan error
	Sync     bool
}

// receive tx from a nvp
type NvpRelayTxEvent struct {
	Payload []byte
}
