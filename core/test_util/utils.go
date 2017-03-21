package test_util

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"reflect"
)

const (
	DB_CONFIG_PATH = "global.dbConfig"
	BlockVersion = "1.0"
	TransactionVersion = "1.0"
	ReceiptVersion = "1.0"
)

// initial a config handler for testing.
func InitConfig(configPath string) *common.Config {
	conf := common.NewConfig(configPath)
	return conf
}

func AssertBlock(block *types.Block, expect *types.Block) bool {
	return reflect.DeepEqual(block, expect)
}


func AssertTransaction(transaction *types.Transaction, expect *types.Transaction) bool {
	return reflect.DeepEqual(transaction, expect)
}

func AssertChain(chain *types.Chain, expect *types.Chain) bool {
	return reflect.DeepEqual(chain, expect)
}

func AssertReceipt(receipt *types.Receipt, expect *types.Receipt) bool {
	return reflect.DeepEqual(receipt, expect)
}