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


type NewBlockPoolEvent struct{ Payload []byte  }

//receive event from consensus module
type SendCheckpointSyncEvent struct{Payload []byte }

//node receive checkpoint sync event and then,check db and send block require request to peers
type StateUpdateEvent struct{Payload []byte }

// after get all required block,send this block to node
type ReceiveSyncBlockEvent struct{Payload []byte }




