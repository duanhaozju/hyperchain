package version1_1

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"math/big"
	"strconv"
)

//
//// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	PostState         string           `json:"postState"`
	CumulativeGasUsed int64            `json:"cumulativeGasUsed"`
	TxHash            string           `json:"txHash"`
	ContractAddress   string           `json:"contractAddress"`
	GasUsed           int64            `json:"gasUsed"`
	Ret               string           `json:"ret"`
	Status            Receipt_STATUS   `json:"status"`
	Message           string           `json:"message"`
	Logs              []types.LogTrans `json:"logs"`
}

func (receipt Receipt) ToReceiptTrans() (receiptTrans *ReceiptTrans) {
	logs, err := receipt.GetLogs()
	var logsValue []types.LogTrans
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

func (r *Receipt) GetLogs() (types.Logs, error) {
	return types.DecodeLogs((*r).Logs)
}

func (r *Receipt) SetLogs(logs types.Logs) error {
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

func (self *Receipt) Encode() string {
	status := "SUCCESS"
	if self.Status == 1 {
		status = "FAILED"
	}
	receiptView := &ReceiptView{
		Version:           string(self.Version),
		PostState:         common.Bytes2Hex(self.PostState),
		CumulativeGasUsed: self.CumulativeGasUsed,
		TxHash:            common.Bytes2Hex(self.TxHash),
		ContractAddress:   common.Bytes2Hex(self.ContractAddress),
		GasUsed:           self.GasUsed,
		Ret:               common.Bytes2Hex(self.Ret),
		Logs:              string(self.Logs),
		Status:            status,
		Message:           common.Bytes2Hex(self.Message),
	}
	res, err := json.MarshalIndent(receiptView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}
