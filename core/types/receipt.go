package types

import (
	"fmt"
	"math/big"
	"hyperchain/common"
)

// Receipt represents the results of a transaction.
type Receipt struct {
	PostState         []byte
	CumulativeGasUsed *big.Int
	TxHash          common.Hash
	ContractAddress common.Address
	GasUsed         *big.Int
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed *big.Int) *Receipt {
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: new(big.Int).Set(cumulativeGasUsed)}
}

// String implements the Stringer interface.
func (r *Receipt) String() string {
	return fmt.Sprintf("receipt{med=%x cgas=%v}", r.PostState, r.CumulativeGasUsed)
}

// Receipts is a wrapper around a Receipt array to implement types.DerivableList.
type Receipts []*Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
