package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"reflect"
	_"fmt"
	"fmt"
)

func TestGetLatestBlockHashOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetLatestBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
	PersistBlock(db.NewBatch(), &blockCases, true, true)
	UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	if !reflect.DeepEqual(ch.data.LatestBlockHash, []byte("blockHash")) {
		t.Errorf("TestGetLatestBlockHashOfChain fail")
	}
	deleteTestData()
}

func TestGetParentBlockHashOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetParentBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
	PersistBlock(db.NewBatch(), &blockCases, true, true)
	UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	if !reflect.DeepEqual(ch.data.ParentBlockHash, []byte("parentHash")) {
		t.Errorf("TestGetParentBlockHashOfChain fail")
	}
	deleteTestData()
}

func TestGetHeightOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetHeightOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
	PersistBlock(db.NewBatch(), &blockCases, true, true)
	UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	if ch.data.Height != 1 {
		t.Errorf("TestGetHeightOfChain fail")
	}
	deleteTestData()
}

func TestGetRequiredBlockNumOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetRequiredBlockNumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	//PersistBlock(db.NewBatch(), &blockCases, true, true)
	//UpdateChainByBlcokNum(hyperdb.DefautNameSpace, db.NewBatch(), 1, true, true)
	putChain(db.NewBatch(), &chain, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	fmt.Println(ch.data.RequiredBlockNum)
	if ch.data.RequiredBlockNum != 3 {
		t.Errorf("TestGetRequiredBlockNumOfChain fail")
	}
	//deleteTestData()
}

func TestGetRequireBlockHashOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetRequireBlockHashOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	putChain(db.NewBatch(), &chain, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	fmt.Println(ch.data.RequireBlockHash)
	//if !reflect.DeepEqual(ch.data.RequireBlockHash, []byte("00000 00000 00000 00000 00000 00000 03")) {
	//	t.Errorf("TestGetRequireBlockHashOfChain fail")
	//}
	deleteTestData()
}

func TestGetRecoveryNumOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetRecoveryNumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	putChain(db.NewBatch(), &chain, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	if ch.data.RecoveryNum != 2 {
		t.Errorf("TestGetRecoveryNumOfChain fail")
	}
	deleteTestData()
}

func TestGetCurrentTxSumOfChain(t *testing.T) {
	logger.Info("test =============> > > TestGetCurrentTxSumOfChain")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	putChain(db.NewBatch(), &chain, true, true)
	ch := chains.GetChain(hyperdb.DefautNameSpace)
	if ch.data.CurrentTxSum != 100 {
		t.Errorf("TestGetCurrentTxSumOfChain fail")
	}
	deleteTestData()
}