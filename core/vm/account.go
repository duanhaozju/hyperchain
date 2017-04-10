package vm

import (
	"math/big"
	"hyperchain/common"
)

// Account represents a contract or basic ethereum account.
type Account interface {
	SubBalance(amount *big.Int)
	AddBalance(amount *big.Int)
	SetBalance(*big.Int)
	SetNonce(uint64)
	Balance() *big.Int
	Address() common.Address
	ReturnGas(*big.Int, *big.Int)
	SetCode(common.Hash, []byte)
	ForEachStorage(cb func(key common.Hash, value []byte) bool) map[common.Hash][]byte
	Value() *big.Int
}
