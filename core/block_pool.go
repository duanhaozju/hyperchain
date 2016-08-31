package core

import "hyperchain/core/types"

type BlockPool struct {
	demandNumber uint64
	pending      []*types.Block
}