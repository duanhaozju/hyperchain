package types

import (
	"fmt"
	"hyperchain/common"
	"math/big"
	"strconv"
)

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed *big.Int, vmType int32) *Receipt {
	i64, err := strconv.ParseInt(cumulativeGasUsed.String(), 10, 64)
	if err != nil {
		fmt.Println("the parseInt is wrong")
	}
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: i64, VmType: Receipt_VmType(vmType)}
}

// Receipts is a wrapper around a Receipt array to implement types.DerivableList.
type Receipts []*Receipt

type ReceiptForStorage Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
