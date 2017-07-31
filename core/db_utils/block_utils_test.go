package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/core/test_util"
	"hyperchain/common"
)

// TestPutBlock tests for PutBlock
func TestPersistBlock(t *testing.T) {
	t.Log("test =============> > > TestPersistBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err,_ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlock(common.DEFAULT_NAMESPACE, test_util.BlockCases.BlockHash)
	if block.Number != 1 {
		t.Error("TestPersistBlock fail")
	}
	deleteTestData()
}

// TestGetBlock tests for GetBlock
func TestGetBlockHash(t *testing.T) {
	t.Log("test =============> > > TestGetBlockHash")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	blockHash, err := GetBlockHash(common.DEFAULT_NAMESPACE, 1)
	if err != nil {
		t.Error(err.Error())
	}
	if string(blockHash) != string(test_util.BlockCases.BlockHash) {
		t.Errorf("both blockhash is not equal, %s not equal %s, TestGetBlock fail", string(blockHash), string(test_util.BlockCases.BlockHash))
	}
	deleteTestData()
}

// TestDeleteBlock tests for DeleteBlock
func TestDeleteBlock(t *testing.T) {
	t.Log("test =============> > > TestDeleteBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlock(common.DEFAULT_NAMESPACE, test_util.BlockCases.BlockHash)
	if block.Number != 1 {
		t.Error("GetBlock fail")
	}
	err = DeleteBlock(common.DEFAULT_NAMESPACE, db.NewBatch(), test_util.BlockCases.BlockHash, true, true)
	block, err = GetBlock(common.DEFAULT_NAMESPACE, test_util.BlockCases.BlockHash)
	if err.Error() != "leveldb: not found" {
		t.Errorf("block delete fail, TestDeleteBlock fail")
	}
	deleteTestData()
}

