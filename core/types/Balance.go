package types

import (
	"math/big"

	"hyperchain-alpha/common"
)

type Balance struct {
	AccountPublicKeyHash common.Hash
	Value *big.Int
}