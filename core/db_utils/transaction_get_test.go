package db_utils

import (
	"hyperchain/hyperdb"
	"testing"
	"hyperchain/core/test_util"
	"hyperchain/common"
	"bytes"
)



func TestGetVersionOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetVersionOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.Signature) != string(test_util.TransactionCases[0].Signature) {
		t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(test_util.TransactionCases[0].Signature))
	}
	deleteTestData()
}

func TestGetFromOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetFromOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.From) != string(test_util.TransactionCases[0].From) {
		t.Error("TestGetFromOfTransaction fail")
	}
	deleteTestData()
}

func TestGetToOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetToOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.To) != string(test_util.TransactionCases[0].To) {
		t.Error("TestGetToOfTransaction fail")
	}
	deleteTestData()
}

func TestGetValueOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetValueOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.Value) != string(test_util.TransactionCases[0].Value) {
		t.Error("TestGetValueOfTransaction fail")
	}
	deleteTestData()
}

func TestGetTimestampOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetTimestampOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if tr.Timestamp != test_util.TransactionCases[0].Timestamp {
		t.Error("TestGetTimestampOfTransaction fail")
	}
	deleteTestData()
}

func TestGetSignatureOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetSignatureOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.Signature) != string(test_util.TransactionCases[0].Signature) {
		t.Error("TestGetSignatureOfTransaction fail")
	}
	deleteTestData()
}

func TestGetIdOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetIdOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if bytes.Compare(tr.Id, test_util.TransactionCases[0].Id) != 0 {
		t.Error("TestGetIdOfTransaction fail")
	}
	deleteTestData()
}

func TestGetTransactionHashOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetTransactionHashOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if string(tr.TransactionHash) != string(test_util.TransactionCases[0].TransactionHash) {
		t.Error("TestGetTransactionHashOfTransaction fail")
	}
	deleteTestData()
}

func TestGetNonceOfTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetNonceOfTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	key := test_util.TransactionCases[0].GetHash().Bytes()
	tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
	if err != nil {
		t.Error(err.Error())
	}
	if tr.Nonce != test_util.TransactionCases[0].Nonce {
		t.Error("TestGetNonceOfTransaction fail")
	}
	deleteTestData()
}