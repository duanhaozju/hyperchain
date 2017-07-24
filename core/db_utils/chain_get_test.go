package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"reflect"
	"hyperchain/core/test_util"
	"hyperchain/common"
)

func TestGetLatestBlockHashOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetLatestBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	InitializeChain(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	UpdateChainByBlcokNum(common.DEFAULT_NAMESPACE, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(common.DEFAULT_NAMESPACE)
	if !reflect.DeepEqual(ch.data.LatestBlockHash, []byte("blockHash")) {
		t.Errorf("TestGetLatestBlockHashOfChain fail")
	}
	deleteTestData()
}

func TestGetParentBlockHashOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetParentBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	InitializeChain(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	UpdateChainByBlcokNum(common.DEFAULT_NAMESPACE, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(common.DEFAULT_NAMESPACE)
	if !reflect.DeepEqual(ch.data.ParentBlockHash, []byte("parentHash")) {
		t.Errorf("TestGetParentBlockHashOfChain fail")
	}
	deleteTestData()
}

func TestGetHeightOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetHeightOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	InitializeChain(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	UpdateChainByBlcokNum(common.DEFAULT_NAMESPACE, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(common.DEFAULT_NAMESPACE)
	if ch.data.Height != 1 {
		t.Errorf("TestGetHeightOfChain fail")
	}
	deleteTestData()
}

func TestGetRequiredBlockNumOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetRequiredBlockNumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	putChain(db.NewBatch(), &chain, true, true)
	chain, _ := getChain(common.DEFAULT_NAMESPACE)
	if chain.RequiredBlockNum != 3 {
		t.Errorf("TestGetRequiredBlockNumOfChain fail")
	}
	deleteTestData()
}

func TestGetRequireBlockHashOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetRequireBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	putChain(db.NewBatch(), &chain, true, true)
	chain, _ := getChain(common.DEFAULT_NAMESPACE)
	if !reflect.DeepEqual(chain.RequireBlockHash, []byte("00000000000000000000000000000003")) {
		t.Errorf("TestGetRequireBlockHashOfChain fail")
	}
	deleteTestData()
}

func TestGetRecoveryNumOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetRecoveryNumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	putChain(db.NewBatch(), &chain, true, true)
	chain, _ := getChain(common.DEFAULT_NAMESPACE)
	if chain.RecoveryNum != 2 {
		t.Errorf("TestGetRecoveryNumOfChain fail")
	}
	deleteTestData()
}

func TestGetCurrentTxSumOfChain(t *testing.T) {
	t.Log("test =============> > > TestGetCurrentTxSumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	putChain(db.NewBatch(), &chain, true, true)
	chain, _ := getChain(common.DEFAULT_NAMESPACE)
	if chain.CurrentTxSum != 100 {
		t.Errorf("TestGetCurrentTxSumOfChain fail")
	}
	deleteTestData()
}