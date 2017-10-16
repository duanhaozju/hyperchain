package v1_2

import (
	"errors"

	v1_2_types "hyperchain/core/ledger/codec/v1_2/types"
	"hyperchain/core/types"

	"github.com/golang/protobuf/proto"
)

// V1_2Encoder is the encoder for transaction v1.2
type V1_2Encoder struct{}

// EncodeTransaction encodes the transaction of v1.2 to bytes.
func (encoder *V1_2Encoder) EncodeTransaction(tx *types.Transaction) ([]byte, error) {
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

// EncodeReceipt encodes the receipt of v1.2 to bytes.
func (encoder *V1_2Encoder) EncodeReceipt(receipt *types.Receipt) ([]byte, error) {
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
