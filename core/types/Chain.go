package types

import "math/big"

type Chain struct {
	LastestBlockHash string
	Height big.Int
}