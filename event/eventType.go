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
type NewBlockEvent struct{
	Payload []byte
	Now	uint64
	Pre	uint64
}

//general tx local
type NewTxEvent struct{ Payload []byte  }




