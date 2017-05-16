package event

import "hyperchain/core/types"

type FilterNewBlockEvent struct {
	Block *types.Block
}

