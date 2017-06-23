package v1_2

import (
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"errors"
)


func EncodeTransaction(tx *types.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, errors.New("empty pointer")
	}
	// process
	tx.Version = []byte(TransactionVersion)
	data, err := proto.Marshal(tx)
	if err != nil {
		return nil, err
	}
	wrapper := &types.TransactionWrapper{
		TransactionVersion: []byte(TransactionVersion),
		Transaction:        data,
	}
	return proto.Marshal(wrapper)
}

func EncodeReceipt(receipt *types.Receipt) ([]byte, error) {
	if receipt == nil {
		return nil, errors.New("empty pointer")
	}
	receipt.Version = []byte(ReceiptVersion)
	data, err := proto.Marshal(receipt)
	if err != nil {
		return nil, err
	}
	wrapper := &types.ReceiptWrapper{
		ReceiptVersion: []byte(ReceiptVersion),
		Receipt:        data,
	}
	return proto.Marshal(wrapper)
}
