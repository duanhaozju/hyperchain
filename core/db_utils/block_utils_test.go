package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/core/test_util"
)

// TestPutBlock tests for PutBlock
func TestPersistBlock(t *testing.T) {
	logger.Info("test =============> > > TestPersistBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err,_ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, test_util.BlockCases.BlockHash)
	if block.Number != 1 {
		t.Error("TestPersistBlock fail")
	}
	deleteTestData()
}

// TestGetBlock tests for GetBlock
func TestGetBlockHash(t *testing.T) {
	logger.Info("test =============> > > TestGetBlockHash")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	blockHash, err := GetBlockHash(hyperdb.DefautNameSpace, 1)
	if err != nil {
		logger.Fatal(err)
	}
	if string(blockHash) != string(test_util.BlockCases.BlockHash) {
		t.Errorf("both blockhash is not equal, %s not equal %s, TestGetBlock fail", string(blockHash), string(test_util.BlockCases.BlockHash))
	}
	deleteTestData()
}

// TestDeleteBlock tests for DeleteBlock
func TestDeleteBlock(t *testing.T) {
	logger.Info("test =============> > > TestDeleteBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, test_util.BlockCases.BlockHash)
	if block.Number != 1 {
		t.Error("GetBlock fail")
	}
	err = DeleteBlock(hyperdb.DefautNameSpace, db.NewBatch(), test_util.BlockCases.BlockHash, true, true)
	block, err = GetBlock(hyperdb.DefautNameSpace, test_util.BlockCases.BlockHash)
	if err.Error() != "leveldb: not found" {
		t.Errorf("block delete fail, TestDeleteBlock fail")
	}
	deleteTestData()
}

func TestGetMarshalBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetMarshalBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, test_util.BlockCases.BlockHash)
	if block.Number != 1 {
		t.Error("GetBlock fail")
	}
	err, _ = GetMarshalBlock(block)
	if err != nil {
		t.Errorf("TestGetMarshalBlock fail")
	}
	deleteTestData()
}