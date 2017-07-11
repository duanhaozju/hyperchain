package version1_2

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/cmd/dbcli/version/version1.2/types"
	"hyperchain/cmd/dbcli/constant"
)

func GetBlockData(data []byte, parameter *constant.Parameter) (string, error) {
	var block version1_2.Block
	err := proto.Unmarshal(data, &block)
	if err != nil {
		return "", err
	} else {
		if parameter.GetVerbose() {
			return block.EncodeVerbose(), nil
		} else {
			return block.Encode(), nil
		}
	}
}

func GetTransactionData(data []byte) (string, error) {
	var transaction version1_2.Transaction
	err := proto.Unmarshal(data, &transaction)
	if err != nil {
		return "", err
	} else {
		return transaction.Encode(), nil
	}
}

func GetInvaildTransactionData(data []byte) (string, error) {
	var invaildTransaction version1_2.InvalidTransactionRecord
	err := proto.Unmarshal(data, &invaildTransaction)
	if err != nil {
		return "", err
	} else {
		return invaildTransaction.Encode(), nil
	}
}

func GetTransactionMetaData(data []byte) (string, error) {
	meta := &version1_2.TransactionMeta{}
	err := proto.Unmarshal(data, meta)
	if err != nil {
		return "", err
	} else {
		return meta.Encode(), nil
	}
}

func GetReceiptData(data []byte) (string, error) {
	var receipt version1_2.Receipt
	err := proto.Unmarshal(data, &receipt)
	if err != nil {
		return "", err
	} else {
		return receipt.Encode(), nil
	}
}

func GetChainData(data []byte) (string, error) {
	var chain version1_2.Chain
	err := proto.Unmarshal(data, &chain)
	if err != nil {
		return "", err
	} else {
		return chain.Encode(), nil
	}
}
