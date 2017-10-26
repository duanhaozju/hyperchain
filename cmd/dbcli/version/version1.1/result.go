package version1_1

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperchain/hyperchain/cmd/dbcli/constant"
	"github.com/hyperchain/hyperchain/cmd/dbcli/version/version1.1/types"
	"strconv"
)

func GetBlockData(data []byte, parameter *constant.Parameter) (string, error) {
	var block version1_1.Block
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

func GetTransactionData(data []byte, parameter *constant.Parameter) (string, error) {
	var block version1_1.Block
	err := proto.Unmarshal(data, &block)
	if err != nil {
		return "", err
	} else {
		return block.EncodeTransaction(parameter.GetTxIndex()), nil
	}
}

func GetInvaildTransactionData(data []byte) (string, error) {
	var invaildTransaction version1_1.InvalidTransactionRecord
	err := proto.Unmarshal(data, &invaildTransaction)
	if err != nil {
		return "", err
	} else {
		return invaildTransaction.Encode(), nil
	}
}

func GetTransactionMetaData(data []byte) (string, error) {
	meta := &version1_1.TransactionMeta{}
	err := proto.Unmarshal(data, meta)
	if err != nil {
		return "", err
	} else {
		return meta.Encode(), nil
	}
}

func GetReceiptData(data []byte) (string, error) {
	var receipt version1_1.Receipt
	err := proto.Unmarshal(data, &receipt)
	if err != nil {
		return "", err
	} else {
		return receipt.Encode(), nil
	}
}

func GetChainData(data []byte) (string, error) {
	var chain version1_1.Chain
	err := proto.Unmarshal(data, &chain)
	if err != nil {
		return "", err
	} else {
		return chain.Encode(), nil
	}
}

func GetChainHeight(data []byte) (string, error) {
	var chain version1_1.Chain
	err := proto.Unmarshal(data, &chain)
	if err != nil {
		return "", err
	} else {
		return strconv.FormatUint(chain.GetHeight(), 10), nil
	}
}
