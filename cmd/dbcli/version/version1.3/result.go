package version1_3

import (
	"hyperchain/cmd/dbcli/constant"
	"github.com/gogo/protobuf/proto"
	"hyperchain/cmd/dbcli/version/version1.3/types"
)

func GetBlockData(data []byte, parameter *constant.Parameter) (string, error) {
	var block version1_3.Block
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
	var transaction version1_3.Transaction
	err := proto.Unmarshal(data, &transaction)
	if err != nil {
		return "", err
	} else {
		return transaction.Encode(), nil
	}
}

func GetReceiptData(data []byte) (string, error) {
	var receipt version1_3.Receipt
	err := proto.Unmarshal(data, &receipt)
	if err != nil {
		return "", err
	} else {
		return receipt.Encode(), nil
	}
}

func GetChainData(data []byte) (string, error) {
	var chain version1_3.Chain
	err := proto.Unmarshal(data, &chain)
	if err != nil {
		return "", err
	} else {
		return chain.Encode(), nil
	}
}