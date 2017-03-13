package db_utils

import (
	"time"
	"testing"
	"hyperchain/core/types"
	"fmt"
	"hyperchain/hyperdb/db"
	"hyperchain/core"
	"hyperchain/common"
	"hyperchain/hyperdb"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
		Id:        1,
	},
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000002"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
		Id:        2,
	},
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature3"),
		Id:        3,
	},
}

var blockCases = types.Block{
	ParentHash:   []byte("parenthash"),
	BlockHash:    []byte("blockhash"),
	Transactions: transactionCases,
	Timestamp:    time.Now().UnixNano(),
	MerkleRoot:   []byte("merkeleroot"),
	Number:       1,
	WriteTime:    time.Now().UnixNano() + int64(time.Second)/2,
}

var transactionMeta = types.TransactionMeta{
	BlockIndex: 1,
	Index:      1,
}

var receipt = types.Receipt{
	Version : []byte("1.2"),
	TxHash : []byte("12345678901234567890123456789012"),
}

func InitDataBase() db.Database {
	conf := common.NewConfig("../../config/global.yaml")
	conf.MergeConfig("../../config/db.yaml")
	core.InitDB(conf, "../../config/db.yaml",conf.GetInt(common.C_NODE_ID))
	db, _ := hyperdb.GetDBDatabase()
	return db
}

// TestPutBlock tests for PutBlock
func TestPersistBlock(t *testing.T) {
	logger.Info("test =============> > > TestPersistBlock")
	db := InitDataBase()
	err,_ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace+hyperdb.Blockchain, blockCases.BlockHash)
	fmt.Println(block.Number)
}

// TestGetBlock tests for GetBlock
func TestGetBlockHash(t *testing.T) {
	logger.Info("test =============> > > TestGetBlockHash")
	db := InitDataBase()
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	blockHash, err := GetBlockHash(hyperdb.DefautNameSpace+hyperdb.Blockchain, 1)
	if err != nil {
		logger.Fatal(err)
	}
	if string(blockHash) != string(blockCases.BlockHash) {
		t.Errorf("both blockhash is not equal, %s not equal %s, TestGetBlock fail", string(blockHash), string(blockCases.BlockHash))
	}
}

// TestDeleteBlock tests for DeleteBlock
func TestDeleteBlock(t *testing.T) {
	logger.Info("test =============> > > TestDeleteBlock")
	db := InitDataBase()
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace+hyperdb.Blockchain, blockCases.BlockHash)
	fmt.Println(block.Number)
	err = DeleteBlock(hyperdb.DefautNameSpace+hyperdb.Blockchain, db.NewBatch(), blockCases.BlockHash, true, true)
	block, err = GetBlock(hyperdb.DefautNameSpace+hyperdb.Blockchain, blockCases.BlockHash)
	if err.Error() != "leveldb: not found" {
		t.Errorf("block delete fail, TestDeleteBlock fail")
	}
}