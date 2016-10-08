package types

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/core/vm"
	"math/big"
	"strconv"
)

//
//// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	PostState         string
	CumulativeGasUsed int64
	TxHash            string
	ContractAddress   string
	GasUsed           int64
	Ret               string
	Status            Receipt_STATUS
	Message           string
	Logs              []vm.LogTrans
}

func (receipt Receipt) ToReceiptTrans() (receiptTrans *ReceiptTrans) {
	logs, err := receipt.GetLogs()
	var logsValue []vm.LogTrans
	if err != nil {
		logsValue = nil
	} else {
		logsValue = logs.ToLogsTrans()
	}
	return &ReceiptTrans{
		GasUsed:           receipt.GasUsed,
		PostState:         common.BytesToHash(receipt.PostState).Hex(),
		ContractAddress:   common.BytesToAddress(receipt.ContractAddress).Hex(),
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		Ret:               common.ToHex(receipt.Ret),
		TxHash:            common.BytesToHash(receipt.TxHash).Hex(),
		Status:            receipt.Status,
		Message:           string(receipt.Message),
		Logs:              logsValue,
	}
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed *big.Int) *Receipt {
	i64, err := strconv.ParseInt(cumulativeGasUsed.String(), 10, 64)
	if err != nil {
		fmt.Println("the parseInt is wrong")
	}
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: i64}
}

func (r *Receipt) GetLogs() (vm.Logs, error) {
	return vm.DecodeLogs((*r).Logs)
}

func (r *Receipt) SetLogs(logs vm.Logs) error {
	buf, err := (&logs).EncodeLogs()
	if err != nil {
		return err
	}
	r.Logs = buf
	return nil
}

// Receipts is a wrapper around a Receipt array to implement types.DerivableList.
type Receipts []*Receipt

type ReceiptForStorage Receipt

// Len returns the number of receipts in this list.
func (r Receipts) Len() int { return len(r) }
