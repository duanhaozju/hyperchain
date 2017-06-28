package types

import (
	"hyperchain/common"
)

//// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	PostState         string         `json:"postState"`
	CumulativeGasUsed int64          `json:"cumulativeGasUsed"`
	TxHash            string         `json:"txHash"`
	ContractAddress   string         `json:"contractAddress"`
	GasUsed           int64          `json:"gasUsed"`
	Ret               string         `json:"ret"`
	Status            Receipt_STATUS `json:"status"`
	Message           string         `json:"message"`
	Logs              []LogTrans     `json:"logs"`
	VmType            string         `json:"vmType"`
}

func (receipt Receipt) ToReceiptTrans() (receiptTrans *ReceiptTrans) {
	logs, err := receipt.RetrieveLogs()
	var logsValue []LogTrans
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
		VmType:            receipt.VmType.String(),
	}
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, vmType int32) *Receipt {
	return &Receipt{PostState: common.CopyBytes(root), VmType: Receipt_VmType(vmType)}
}

func (r *Receipt) RetrieveLogs() (Logs, error) {
	return DecodeLogs((*r).Logs)
}

func (r *Receipt) SetLogs(logs Logs) error {
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
