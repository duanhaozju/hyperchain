//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package event

import "hyperchain/core/types"

//consensus event incoming from outer,peers post
type ConsensusEvent struct{ Payload []byte }

type AliveEvent struct{ Payload bool }

// send consensus event to outer peers for consensus module
type BroadcastConsensusEvent struct{ Payload []byte }

//receive new block event from node consensus event for consensus module
type NewBlockEvent struct {
	Payload    []byte
	CommitTime int64
}

//general tx local
type NewTxEvent struct{ Payload []byte }

type TxUniqueCastEvent struct {
	Payload []byte
	PeerId  uint64
}

type NewBlockPoolEvent struct{ Payload []byte }

//node receive checkpoint sync event and then,check db and send block require request to peers
type SendCheckpointSyncEvent struct{ Payload []byte }

//receive event from consensus module
type StateUpdateEvent struct{ Payload []byte }

// after get all required block,send this block to node
type ReceiveSyncBlockEvent struct{ Payload []byte }

//receive new block event from node consensus event for consensus module
type ExeTxsEvent struct {
	Transactions []*types.Transaction
	SeqNo        uint64
	View         uint64
	IsPrimary    bool
	Timestamp    int64
}

type ValidatedTxs struct {
	Transactions []*types.Transaction
	Hash         string
	SeqNo        uint64
	View         uint64
	Timestamp    int64
}

// if the CommitStatus is true, we will commit the blocks and save the statedb
// or we will rollback the statedb
// Flag == true, commit; Flag == false, rollback
type CommitOrRollbackBlockEvent struct {
	SeqNo      uint64
	Timestamp  int64
	CommitTime int64
	Flag       bool
	Hash       string
	IsPrimary  bool
}
//set invalid tx into db
type RespInvalidTxsEvent struct {
	Payload []byte
}
// reset blockchain to a stable checkpoint status when `viewchange` occur
type VCResetEvent struct {
	SeqNo uint64
}
//set primary in peerManager when new view and primary
type InformPrimaryEvent struct {
	Primary uint64
}

//sync all nodes status event
type ReplicaStatusEvent struct {
	Payload []byte
}


type StartJoinChainEvent  struct {
	PrimaryIp string

}
type RouteTable struct {
	Name string
}
// primary receive new peer msg from p2p
type NewPeerArriveEvent struct {
	IP         string
	ReplicaID  uint64
	RouteInfo *RouteTable
}
// broad route table to new peer,at the same time,broadcast all peer
// to send their own routeTable to their own consensus module
type BroadRouteToNewPeerEvent struct {
	RouteInfo *RouteTable
	Hash []byte
}
//after uodate n and f,post event to update routeTable,and accept new peer's connection
type UpdateRouteTableEvent struct {
	ReplicaID uint64
}
// primary notify new peer to connect all other peers,after update n and f
type ConnectChainEvent struct {

}

// new peer receive the start event and then start to connect the all peers and negociate view and recovery
type receiveStartConnectEvent struct {

}