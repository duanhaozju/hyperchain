package manager

import (
	"math/big"

	"hyperchain-alpha/core/types"
)

const (
	// Protocol messages
	StatusMsg                   = 0x00

	TxMsg                       = 0x01

	NewBlockMsg                 = 0x02

)

type newBlockData struct {
	Block *types.Block
	TD    *big.Int
}
type errCode int

const (

	ErrDecode
	ErrInvalidMsgCode

	ErrExtraStatusMsg

)
type txPool interface {
	// AddTransactions should add the given transactions to the pool.
	AddTransactions([]*types.Transaction)

	// GetTransactions should return pending transactions.
	// The slice should be modifiable by the caller.
	GetTransactions() types.Transactions
}
