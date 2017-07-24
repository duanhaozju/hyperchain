package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"hyperchain/core/test_util"
)


func TestGetPostStateOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetPostStateOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	deleteTestData()
}

func TestGetCumulativeGasUsedOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetCumulativeGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.CumulativeGasUsed != test_util.Receipt.CumulativeGasUsed {
		t.Errorf("TestGetCumulativeGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetTxHashOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetTxHashOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.TxHash != common.BytesToHash(test_util.Receipt.TxHash).Hex() {
		t.Errorf("TestGetTxHashOfReceipt fail")
	}
	deleteTestData()
}

func TestGetContractAddressOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetContractAddressOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.ContractAddress != common.BytesToAddress(test_util.Receipt.ContractAddress).Hex() {
		t.Errorf("TestGetContractAddressOfReceipt fail")
	}
	deleteTestData()
}

func TestGetGasUsedOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetGasUsedOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.GasUsed != test_util.Receipt.GasUsed {
		t.Errorf("TestGetGasUsedOfReceipt fail")
	}
	deleteTestData()
}

func TestGetRetOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetRetOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Ret != common.ToHex(test_util.Receipt.Ret) {
		t.Errorf("TestGetRetOfReceipt fail")
	}
	deleteTestData()
}

func TestGetStatusOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetStatusOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Status != test_util.Receipt.Status {
		t.Errorf("TestGetStatusOfReceipt fail")
	}
	deleteTestData()
}

func TestGetMessageOfReceipt(t *testing.T) {
	t.Log("test =============> > > TestGetMessageOfReceipt")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	batch := db.NewBatch()
	err, _ := PersistReceipt(batch, &test_util.Receipt, true, true)
	if err != nil {
		t.Errorf("PersistReceipt fail")
	}
	res := GetReceipt(common.DEFAULT_NAMESPACE, common.BytesToHash(test_util.Receipt.TxHash))
	if res == nil {
		t.Errorf("GetReceipt fail")
	}
	if res.Message != string(test_util.Receipt.Message) {
		t.Errorf("TestGetMessageOfReceipt fail")
	}
	deleteTestData()
}