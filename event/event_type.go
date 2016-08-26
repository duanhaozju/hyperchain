// common event defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package event

import (


	"hyperchain-alpha/core/types"

)

//consensus event incoming from outer
type ConsensusEvent struct{Payload []byte }

// send consensus event to outer peers
type BroadcastConsensusEvent struct{ Payload []byte }



//receive new block event from node consensus event
type NewBlockEvent struct{ Payload []byte  }

//general tx local
type NewTxEvent struct{ Payload []byte  }




