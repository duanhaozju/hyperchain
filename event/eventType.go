// common event defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package event



//consensus event incoming from outer,peers post
type ConsensusEvent struct{Payload []byte }

type AliveEvent struct{Payload bool}

// send consensus event to outer peers for consensus module
type BroadcastConsensusEvent struct{ Payload []byte }

//receive new block event from node consensus event for consensus module
type NewBlockEvent struct{ Payload []byte
			   CommitTime int64}

//general tx local
type NewTxEvent struct{ Payload []byte  }

//transmit tx to primary node
type TxUniqueCastEvent struct{ Payload []byte
			       PeerId uint64  }

type NewBlockPoolEvent struct{ Payload []byte  }

//node receive checkpoint sync event and then,check db and send block require request to peers
type SendCheckpointSyncEvent struct{Payload []byte }

//receive event from consensus module
type StateUpdateEvent struct{Payload []byte }

// after get all required block,send this block to node
type ReceiveSyncBlockEvent struct{Payload []byte }

// after exe all txs,send the executable txs and its' hash to the pbft module
type ExeTxsEvent struct{Payload []byte }

// if the CommitStatus is true, we will commit the blocks and save the statedb
// or we will rollback the statedb
type CommitOrRollbackBlockEvent struct{ Payload []byte
				   CommitStatus bool  }






