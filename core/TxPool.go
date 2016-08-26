package core

import "hyperchain/core/types"

type TxPool struct {
	MaxCapacity int
	Transactions []types.Transaction
}
