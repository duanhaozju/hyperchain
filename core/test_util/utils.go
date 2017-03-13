package test_util

import (
	"hyperchain/common"
	"hyperchain/core/types"
	edb "hyperchain/core/db_utils"
	"bytes"
)

const (
	DB_CONFIG_PATH = "global.dbConfig"
)

// initial a config handler for testing.
func InitConfig(configPath string, dbConfigPath string) *common.Config {
	conf := common.NewConfig(configPath)
	conf.MergeConfig(dbConfigPath)
	return conf
}

func AssertBlock(block *types.Block, expect *types.Block) bool {
	// assert via hash
	bHash := block.Hash()
	eHash := expect.Hash()
	if bHash.Hex() != eHash.Hex() {
		return false
	}
	if bytes.Compare(block.BlockHash, expect.BlockHash) != 0 {
		return false
	}
	// assert other field
	if string(block.Version) != edb.BlockVersion {
		return false
	}
	if len(block.Transactions) != len(expect.Transactions) {
		return false
	}
	for idx := range block.Transactions {
		if !AssertTransaction(block.Transactions[idx], block.Transactions[idx]) {
			return false
		}
	}
	return true
}


func AssertTransaction(transaction *types.Transaction, expect *types.Transaction) bool {
	// assert via hash
	tHash := transaction.GetHash()
	eHash := expect.GetHash()
	if tHash.Hex() != eHash.Hex() {
		return false
	}
	if bytes.Compare(transaction.TransactionHash, expect.TransactionHash) != 0 {
		return false
	}
	// assert other field
	if string(transaction.Version) != edb.TransactionVersion {
		return false
	}
	if transaction.Id != expect.Id {
		return false
	}
	return true
}
