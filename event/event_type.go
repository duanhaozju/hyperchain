package event

import (


	"hyperchain-alpha/core/types"

)


type ConsensusEvent struct{ Msg *types.Msg }

type BroadcastConsensusEvent struct{ Msg *types.Msg }





// NewBlockEvent is posted when a block has been imported.
type NewBlockEvent struct{ Block *types.Block }



