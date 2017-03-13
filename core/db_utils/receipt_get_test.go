package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/common"
)

func TestGetPostStateOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetPostStateOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.PostState != common.BytesToHash(receipt.PostState).Hex() {
		t.Errorf("TestGetVersionOfReceipt fail")
	}
	deleteTestData()
}

func TestGetCumulativeGasUsedOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetCumulativeGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.CumulativeGasUsed != receipt.CumulativeGasUsed {
		t.Errorf("TestGetCumulativeGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetTxHashOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetTxHashOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.TxHash != common.BytesToHash(receipt.TxHash).Hex() {
		t.Errorf("TestGetTxHashOfReceipt fail")
	}
	deleteTestData()
}

func TestGetContractAddressOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetContractAddressOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.ContractAddress != common.BytesToAddress(receipt.ContractAddress).Hex() {
		t.Errorf("TestGetContractAddressOfReceipt fail")
	}
	deleteTestData()
}

func TestGetGasUsedOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.GasUsed != receipt.GasUsed {
		t.Errorf("TestGetGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetRetOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetRetOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Ret != common.ToHex(receipt.Ret) {
		t.Errorf("TestGetRetOfReceipt fail")
	}
	deleteTestData()
}

func TestGetStatusOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetStatusOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Status != receipt.Status {
		t.Errorf("TestGetStatusOfReceipt fail")
	}
	deleteTestData()
}

func TestGetMessageOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetMessageOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.DefautNameSpace, common.BytesToHash(receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Message != string(receipt.Message) {
		t.Errorf("TestGetMessageOfReceipt fail")
	}
	deleteTestData()
}