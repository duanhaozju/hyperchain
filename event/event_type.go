package event

import (


	"hyperchain-alpha/core/types"

)

//consensus event incoming from outer
type ConsensusEvent struct{ Msg *types.Msg }

// send consensus event to outer peers
type BroadcastConsensusEvent struct{ Msg *types.Msg }



//receive new block event from node consensus event
type NewBlockEvent struct{ Block *types.Block }



