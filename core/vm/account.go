package vm

import (
	"github.com/hyperchain/hyperchain/common"
	"math/big"
)

// Account represents a contract or basic hyperchain account.
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

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	ReturnGas(*big.Int, *big.Int)
	Address() common.Address
	Value() *big.Int
	SetCode(common.Hash, []byte)
	ForEachStorage(callback func(key common.Hash, value []byte) bool) map[common.Hash][]byte
}
