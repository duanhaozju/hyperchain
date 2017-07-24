package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"reflect"
	"os"
	"hyperchain/common"
	"strconv"
	"hyperchain/core/test_util"
)

func init() {
	conf := common.NewConfig("../../configuration/namespaces/global/config/global.yaml")
	conf.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
}

func TestGetVersionOfBlock(t *testing.T) {
	t.Log("test =============> > > TestGetVersionOfBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.Version) != BlockVersion {
		t.Error("TestGetVersionOfBlock fail")
	}
	deleteTestData()
}

func TestGetParentHashOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.ParentHash) != "parentHash" {
		t.Error("TestGetParentHashOfBlock fail")
	}
	deleteTestData()
}

func TestGetBlockHashOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.BlockHash) != "blockHash" {
		t.Error("TestGetBlockHashOfBlock fail")
	}
	deleteTestData()
}

func TestGetTransactionsOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if !reflect.DeepEqual(block.Transactions, test_util.TransactionCases) {
		t.Error("TestGetTransactionsOfBlock fail")
	}
	deleteTestData()
}

func TestGetTimestampOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if block.Timestamp != 1489387222 {
		t.Error("TestGetTimestampOfBlock fail")
	}
	deleteTestData()
}

func TestGetMerkleRootOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.MerkleRoot) != "merkleRoot" {
		t.Error("TestGetMerkleRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetTxRootOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.TxRoot) != "txRoot" {
		t.Error("TestGetTxRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetReceiptRootOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if string(block.ReceiptRoot) != "receiptRoot" {
		t.Error("TestGetReceiptRootOfBlock fail")
	}
	deleteTestData()
}

func TestGetNumberOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if block.Number != 1 {
		t.Error("TestGetNumberOfBlock fail")
	}
	deleteTestData()
}

func TestGetWriteTimeOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if block.WriteTime != 1489387223 {
		t.Error("TestGetWriteTimeOfBlock fail")
	}
	deleteTestData()
}

func TestGetCommitTimeOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if block.CommitTime != 1489387224 {
		t.Error("TestGetCommitTimeOfBlock fail")
	}
	deleteTestData()
}

func TestGetEvmTimeOfBlock(t *testing.T) {
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if block.EvmTime != 1489387225 {
		t.Error("TestGetEvmTimeOfBlock fail")
	}
	deleteTestData()
}




func InitDataBase() {
	conf := common.NewConfig("../../configuration/namespaces/global/config/global.yaml")
	conf.MergeConfig("../../configuration/namespaces/global/config/db.yaml")
	hyperdb.SetDBConfig("../../configuration/namespaces/global/config/db.yaml", strconv.Itoa(conf.GetInt(common.C_NODE_ID)))
	hyperdb.InitDatabase(conf, common.DEFAULT_NAMESPACE)
	InitializeChain(common.DEFAULT_NAMESPACE)
}

func deleteTestData() {
	childPath := "/namespaces"
	current_dir, _ := os.Getwd()
	path := current_dir + childPath
	os.RemoveAll(path)
}