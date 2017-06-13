package db_utils

import (
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/hyperdb"
	"testing"
)

func TestGetPostStateOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetPostStateOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.PostState != common.BytesToHash(test_util.Receipt.PostState).Hex() {
		t.Errorf("TestGetVersionOfReceipt fail")
	}
	deleteTestData()
}

func TestGetCumulativeGasUsedOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetCumulativeGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.CumulativeGasUsed != test_util.Receipt.CumulativeGasUsed {
		t.Errorf("TestGetCumulativeGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetTxHashOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetTxHashOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.TxHash != common.BytesToHash(test_util.Receipt.TxHash).Hex() {
		t.Errorf("TestGetTxHashOfReceipt fail")
	}
	deleteTestData()
}

func TestGetContractAddressOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetContractAddressOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.ContractAddress != common.BytesToAddress(test_util.Receipt.ContractAddress).Hex() {
		t.Errorf("TestGetContractAddressOfReceipt fail")
	}
	deleteTestData()
}

func TestGetGasUsedOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.GasUsed != test_util.Receipt.GasUsed {
		t.Errorf("TestGetGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetRetOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetRetOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Ret != common.ToHex(test_util.Receipt.Ret) {
		t.Errorf("TestGetRetOfReceipt fail")
	}
	deleteTestData()
}

func TestGetStatusOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetStatusOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Status != test_util.Receipt.Status {
		t.Errorf("TestGetStatusOfReceipt fail")
	}
	deleteTestData()
}

func TestGetMessageOfReceipt(t *testing.T) {
	logger.Info("test =============> > > TestGetMessageOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(hyperdb.defaut_namespace, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Message != string(test_util.Receipt.Message) {
		t.Errorf("TestGetMessageOfReceipt fail")
	}
	deleteTestData()
}
