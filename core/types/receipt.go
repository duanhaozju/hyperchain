package types

import (
	"math/big"
	"hyperchain/common"
	"strconv"
	"fmt"
)
//
//// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	PostState         []byte
	CumulativeGasUsed int64
	TxHash            string
	ContractAddress   string
	GasUsed           int64
	Ret               string
}

func (receipt Receipt)ToReceiptTrans() (receiptTrans *ReceiptTrans){
	return &ReceiptTrans{GasUsed:receipt.GasUsed,PostState:receipt.PostState,
		ContractAddress:common.Bytes2Hex(receipt.ContractAddress),
		CumulativeGasUsed:receipt.CumulativeGasUsed,Ret:common.ToHex(receipt.Ret),
		TxHash:common.ToHex(receipt.TxHash)}
}

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
