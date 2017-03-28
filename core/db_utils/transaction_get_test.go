package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"hyperchain/core/test_util"
)

func TestGetVersionOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetVersionOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.Signature) != string(test_util.TransactionCases[0].Signature) {
		t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(test_util.TransactionCases[0].Signature))
	}
	deleteTestData()
}

func TestGetFromOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetFromOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.From) != string(test_util.TransactionCases[0].From) {
		t.Error("TestGetFromOfTransaction fail")
	}
	deleteTestData()
}

func TestGetToOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetToOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.To) != string(test_util.TransactionCases[0].To) {
		t.Error("TestGetToOfTransaction fail")
	}
	deleteTestData()
}

func TestGetValueOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetValueOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.Value) != string(test_util.TransactionCases[0].Value) {
		t.Error("TestGetValueOfTransaction fail")
	}
	deleteTestData()
}

func TestGetTimestampOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetTimestampOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if tr.Timestamp != test_util.TransactionCases[0].Timestamp {
		t.Error("TestGetTimestampOfTransaction fail")
	}
	deleteTestData()
}

func TestGetSignatureOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetSignatureOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.Signature) != string(test_util.TransactionCases[0].Signature) {
		t.Error("TestGetSignatureOfTransaction fail")
	}
	deleteTestData()
}

func TestGetIdOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetIdOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if tr.Id != test_util.TransactionCases[0].Id {
		t.Error("TestGetIdOfTransaction fail")
	}
	deleteTestData()
}

func TestGetTransactionHashOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetTransactionHashOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if string(tr.TransactionHash) != string(test_util.TransactionCases[0].TransactionHash) {
		t.Error("TestGetTransactionHashOfTransaction fail")
	}
	deleteTestData()
}

func TestGetNonceOfTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetNonceOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(hyperdb.defaut_namespace, key)
	if err != nil {
		logger.Fatal(err)
	}
	if tr.Nonce != test_util.TransactionCases[0].Nonce {
		t.Error("TestGetNonceOfTransaction fail")
	}
	deleteTestData()
}