package version1_3

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/willf/bloom"
	"math/big"
)

//// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	Version           string         `json:"version"`
	Bloom             string         `json:"bloom"`
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
		Version:           string(receipt.Version),
		GasUsed:           receipt.GasUsed,
		Bloom:             common.BytesToHash(receipt.Bloom).Hex(),
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
func NewReceipt(cumulativeGasUsed *big.Int, vmType int32) *Receipt {
	return &Receipt{CumulativeGasUsed: cumulativeGasUsed.Int64(), VmType: Receipt_VmType(vmType)}
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

func CreateBloom(receipts []*Receipt) ([]byte, error) {
	bloom := bloom.New(256*8, 3)

	for _, r := range receipts {
		logs, err := r.RetrieveLogs()
		if err != nil {
			return nil, err
		}
		for _, log := range logs {
			bloom.Add(log.Address.Bytes())
			for _, topic := range log.Topics {
				bloom.Add(topic.Bytes())
			}
		}
	}
	return bloom.GobEncode()
}

func (r *Receipt) BloomFilter() (error, *bloom.BloomFilter) {
	bloom := bloom.New(256*8, 3)
	if err := bloom.GobDecode(r.GetBloom()); err != nil {
		return err, nil
	}
	return nil, bloom
}

func BloomLookup(bloom *bloom.BloomFilter, content []byte) bool {
	return bloom.Test(content)
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
		Bloom:             common.Bytes2Hex(self.Bloom),
		CumulativeGasUsed: self.CumulativeGasUsed,
		TxHash:            common.Bytes2Hex(self.TxHash),
		ContractAddress:   common.Bytes2Hex(self.ContractAddress),
		GasUsed:           self.GasUsed,
		Ret:               common.Bytes2Hex(self.Ret),
		Logs:              string(self.Logs),
		Status:            status,
		Message:           common.Bytes2Hex(self.Message),
		VmType:            self.encodeVmType(),
	}
	res, err := json.MarshalIndent(receiptView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (self *Receipt) encodeVmType() string {
	switch self.VmType {
	case Receipt_EVM:
		return "evm"
	case Receipt_JVM:
		return "jvm"
	default:
		return ""
	}
}
