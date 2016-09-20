package types

import (
	"math/big"
	"hyperchain/common"
	"strconv"
	"fmt"
)
//
//// Receipt represents the results of a transaction.
//type Receipt struct {
//	PostState         []byte
//	CumulativeGasUsed *big.Int
//	TxHash          common.Hash
//	ContractAddress common.Address
//	GasUsed         *big.Int
//}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed *big.Int) *Receipt {
	i64,err := strconv.ParseInt(cumulativeGasUsed.String(),10,64)
	if(err != nil){
		fmt.Println("the parseInt is wrong")
	}
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: i64}
}

// Receipts is a wrapper around a Receipt array to implement types.DerivableList.
type Receipts []*Receipt

type ReceiptForStorage Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
