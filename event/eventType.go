// common event defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
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
	Transactions	[]*types.Transaction
	Digest		string
	SeqNo		uint64
	View		uint64
	IsPrimary	bool
}

type ValidatedTxs struct {
	Transactions	[]*types.Transaction
	Hash		[]byte
	Digest		string
	SeqNo		uint64
	View		uint64
}

// if the CommitStatus is true, we will commit the blocks and save the statedb
// or we will rollback the statedb
// Flag == true, commit; Flag == false, rollback
type CommitOrRollbackBlockEvent struct {
	SeqNo      	uint64
	Timestamp	int64
	CommitTime 	int64
	Flag       	bool
}
