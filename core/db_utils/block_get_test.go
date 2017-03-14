package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"reflect"
	"os"
	"fmt"
	"hyperchain/common"
	"strconv"
	"hyperchain/core/test_util"
)

func TestGetVersionOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetVersionOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.Version) != BlockVersion {
		t.Error("TestGetVersionOfBlock fail")
	}
	deleteTestData()
}

func TestGetParentHashOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetParentHashOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.ParentHash) != "parentHash" {
		t.Error("TestGetParentHashOfBlock fail")
	}
	deleteTestData()
}

func TestGetBlockHashOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetBlockHashOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.BlockHash) != "blockHash" {
		t.Error("TestGetBlockHashOfBlock fail")
	}
	deleteTestData()
}

func TestGetTransactionsOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetTransactionsOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if !reflect.DeepEqual(block.Transactions, test_util.TransactionCases) {
		t.Error("TestGetTransactionsOfBlock fail")
	}
	deleteTestData()
}

func TestGetTimestampOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetTimestampOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if block.Timestamp != 1489387222 {
		t.Error("TestGetTimestampOfBlock fail")
	}
	deleteTestData()
}

func TestGetMerkleRootOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetMerkleRootOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.MerkleRoot) != "merkleRoot" {
		t.Error("TestGetMerkleRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetTxRootOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetTxRootOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.TxRoot) != "txRoot" {
		t.Error("TestGetTxRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetReceiptRootOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetReceiptRootOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if string(block.ReceiptRoot) != "receiptRoot" {
		t.Error("TestGetReceiptRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetNumberOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetNumberOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if block.Number != 1 {
		t.Error("TestGetNumberOfBlock fail")
	}
	deleteTestData()
}

func TestGetWriteTimeOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetWriteTimeOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if block.WriteTime != 1489387223 {
		t.Error("TestGetWriteTimeOfBlock fail")
	}
	deleteTestData()
}

func TestGetCommitTimeOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetCommitTimeOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if block.CommitTime != 1489387224 {
		t.Error("TestGetCommitTimeOfBlock fail")
	}
	deleteTestData()
}

func TestGetEvmTimeOfBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetEvmTimeOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace, 1)
	if block.EvmTime != 1489387225 {
		t.Error("TestGetEvmTimeOfBlock fail")
	}
	deleteTestData()
}




func InitDataBase() {
	conf := common.NewConfig("../../config/global.yaml")
	conf.MergeConfig("../../config/db.yaml")
	hyperdb.SetDBConfig("../../config/db.yaml", strconv.Itoa(conf.GetInt(common.C_NODE_ID)))
	hyperdb.InitDatabase(conf, hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
}

func deleteTestData() {
	childPath := "/build"
	current_dir, _ := os.Getwd()
	path := current_dir + childPath
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println("remove test data file fail!", err)
	} else {
		fmt.Println("remove test data file success!")
	}
}