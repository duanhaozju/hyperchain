package core

import "hyperchain-alpha/core/types"

type TxPool struct {
	MaxCapacity int
	Transactions []types.Transaction
}
