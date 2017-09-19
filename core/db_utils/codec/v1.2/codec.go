package v1_2

import (
	"errors"
	"github.com/golang/protobuf/proto"
	v1_2_types "hyperchain/core/db_utils/codec/v1.2/types"
	"hyperchain/core/types"
)

func EncodeTransaction(tx *types.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, errors.New("empty pointer")
	}
	localTransaction := &v1_2_types.TransactionV1_2{
		Version:         []byte(TransactionVersion),
		From:            tx.From,
		To:              tx.To,
		Value:           tx.Value,
		Timestamp:       tx.Timestamp,
		Signature:       tx.Signature,
		Id:              tx.Id,
		TransactionHash: tx.TransactionHash,
		Nonce:           tx.Nonce,
	}
	data, err := proto.Marshal(localTransaction)
	if err != nil {
		return nil, err
	}
	wrapper := &types.TransactionWrapper{
		TransactionVersion: []byte(ReceiptVersion),
		Transaction:        data,
	}
	return proto.Marshal(wrapper)
}

func EncodeReceipt(receipt *types.Receipt) ([]byte, error) {
	if receipt == nil {
		return nil, errors.New("empty pointer")
	}

	localReceipt := &v1_2_types.ReceiptV1_2{
		Version:           []byte(ReceiptVersion),
		PostState:         nil,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		TxHash:            receipt.TxHash,
		ContractAddress:   receipt.ContractAddress,
		GasUsed:           100000, // For backward compatibility
		Ret:               receipt.Ret,
		Logs:              receipt.Logs,
		Status:            v1_2_types.ReceiptV1_2_STATUS(receipt.Status),
		Message:           receipt.Message,
	}
	data, err := proto.Marshal(localReceipt)
	if err != nil {
		return nil, err
	}
	wrapper := &types.ReceiptWrapper{
		ReceiptVersion: []byte(ReceiptVersion),
		Receipt:        data,
	}
	return proto.Marshal(wrapper)
}