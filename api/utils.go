package api

import (
	"hyperchain/core/types"
	"hyperchain/common"
	"hyperchain/core/vm/evm"
)


// ReceiptTrans are used to show in web.
type ReceiptTrans struct {
	VmType            string         `json:"vmType"`
	PostState         string         `json:"postState"`
	CumulativeGasUsed int64          `json:"cumulativeGasUsed"`
	TxHash            string         `json:"txHash"`
	ContractAddress   string         `json:"contractAddress"`
	GasUsed           int64          `json:"gasUsed"`
	Ret               string         `json:"ret"`
	Status            string         `json:"status"`
	Message           string         `json:"message"`
	Logs              []string       `json:"logs"`
}


// TransitReceipt transit origin receipt to a more user friendly format
func TransitReceipt(receipt *types.Receipt, vmType types.Receipt_VmType) ReceiptTrans {
	var logsTrans []string
	tmp, _ := RetrieveLogs(receipt, vmType)
	switch vmType {
	case types.Receipt_EVM:
		logs := tmp.(evm.Logs)
		for _, log := range logs {
			logsTrans = append(logsTrans, log.String())
		}
	case types.Receipt_JVM:

	}
	return ReceiptTrans{
		VmType:            vmType.String(),
		GasUsed:           receipt.GasUsed,
		PostState:         common.BytesToHash(receipt.PostState).Hex(),
		ContractAddress:   common.BytesToAddress(receipt.ContractAddress).Hex(),
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		Ret:               common.ToHex(receipt.Ret),
		TxHash:            common.BytesToHash(receipt.TxHash).Hex(),
		Status:            receipt.Status.String(),
		Message:           string(receipt.Message),
		Logs:              logsTrans,
	}
}

func RetrieveLogs(r *types.Receipt, vmType types.Receipt_VmType) (interface{}, error) {
	switch vmType {
	case types.Receipt_EVM:
		// EVM
		return evm.DecodeLogs((*r).Logs)
	case types.Receipt_JVM:
		// JVM
		return nil, nil
	default:
		// TODO
		return nil, nil
	}
}
