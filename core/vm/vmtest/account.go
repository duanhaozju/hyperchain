/**
 * Created by Meiling Hu on 11/17/16.
 */
package vmtest

import (
	"math/big"
	"hyperchain/common"
)

type Account interface {
	SubBalance(amount *big.Int)
	AddBalance(amount *big.Int)
	SetBalance(*big.Int)
	SetNonce(uint64)
	Balance() *big.Int
	Address() common.Address
	ReturnGas(*big.Int, *big.Int)
	SetCode([]byte)
	ForEachStorage(cb func(key, value common.Hash) bool)
	Value() *big.Int
}

