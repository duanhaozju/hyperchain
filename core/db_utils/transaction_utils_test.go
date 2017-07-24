package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"reflect"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/test_util"
	"hyperchain/common"
)


// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	t.Log("test =============> > > TestGetTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	for _, trans := range test_util.TransactionCases[:1] {
		key := trans.GetHash().Bytes()
		tr, err := GetTransaction(common.DEFAULT_NAMESPACE, key)
		if err != nil {
			t.Error(err.Error())
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
		if res, _ :=JudgeTransactionExist(common.DEFAULT_NAMESPACE, key); res != true {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
	deleteTestData()
}

func TestGetTransactionBLk(t *testing.T) {
	t.Log("test =============> > > TestGetTransactionBLk")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	err, _ = PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	block, err := GetBlockByNumber(common.DEFAULT_NAMESPACE, 1)
	if err != nil {
		t.Error(err.Error())
	}
	batch := db.NewBatch()
	PersistTransactionMeta(batch, &test_util.TransactionMeta, test_util.TransactionCases[0].GetHash(), true, true)
	if len(block.Transactions) > 0 {
		tx := block.Transactions[0]
		bn, i := GetTxWithBlock(common.DEFAULT_NAMESPACE, tx.GetHash().Bytes())
		if bn != 1 || i != 1 {
			t.Errorf("TestGetTransactionBLk fail")
		}
		DeleteTransactionMeta(batch, tx.GetHash().Bytes(), true, true)
		a, b := GetTxWithBlock(common.DEFAULT_NAMESPACE, tx.GetHash().Bytes())
		if a != 0 || b != 0 {
			t.Errorf("TestGetTransactionBLk fail")
		}
	}
	deleteTestData()
}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransactions(t *testing.T) {
	t.Log("test =============> > > TestGetAllTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		t.Error(err.Error())
	}
	trs, err := GetAllTransaction(common.DEFAULT_NAMESPACE)
	var zero = types.Transaction{}
	for _, trans := range trs {
		if !reflect.DeepEqual(*trans, zero) {
			isPass := false
			if string(trans.Signature) == "signature1" ||
				string(trans.Signature) == "signature2" ||
				string(trans.Signature) == "signature3" {
				isPass = true
			}
			if !isPass {
				t.Errorf("%s not exist", string(trans.Signature))
			}
		}
	}
	deleteTestData()
}

// TestDeleteTransaction tests for DeleteTransaction
func TestDeleteTransaction(t *testing.T) {
	t.Log("test =============> > > TestDeleteTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	for _, trans := range test_util.TransactionCases[:1] {
		err, _ := PersistTransaction(db.NewBatch(), trans, true, true)
		if err != nil {
			t.Error(err.Error())
		}
		DeleteTransaction(db.NewBatch(), trans.GetHash().Bytes(), true, true)
		_, err = GetTransaction(common.DEFAULT_NAMESPACE, trans.GetHash().Bytes())
		if err.Error() != "leveldb: not found" {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", trans.GetHash().Bytes())
		}
	}
	deleteTestData()
}

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	t.Log("test =============> > > TestPutTransactions")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	err := PersistTransactions(db.NewBatch(), test_util.TransactionCases, TransactionVersion, true, true)
	if err != nil {
		t.Error(err.Error())
	}
	trs, err := GetAllTransaction(common.DEFAULT_NAMESPACE)
	if err != nil {
		t.Error(err.Error())
	}
	if len(trs) < 3 {
		t.Errorf("TestPutTransactions fail")
	}
	deleteTestData()
}

// TestGetInvaildTx tests for GetDiscardTransaction
func TestGetInvaildTx(t *testing.T) {
	t.Skip("a")
	tx := test_util.TransactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data, _ := proto.Marshal(record)
	// save to db
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	db.Put(append(InvalidTransactionPrefix, tx.TransactionHash...), data)

	result, _ := GetInvaildTxErrType(common.DEFAULT_NAMESPACE, tx.TransactionHash)
	if result != types.InvalidTransactionRecord_OUTOFBALANCE {
		t.Error("TestGetInvaildTx fail")
	}
	deleteTestData()
}

// TestGetDiscardTransaction tests for GetDiscardTransaction
func TestGetDiscardTransaction(t *testing.T) {
	tx := test_util.TransactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	PersistInvalidTransactionRecord(db.NewBatch(), record, true, true)

	result, _ := GetDiscardTransaction(common.DEFAULT_NAMESPACE, tx.TransactionHash)
	if result.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
		t.Errorf("TestGetDiscardTransaction fail")
	}
	results, _ := GetAllDiscardTransaction(common.DEFAULT_NAMESPACE)
	for _, v := range results {
		if v.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
			t.Errorf("TestGetDiscardTransaction fail")
		}
	}
	deleteTestData()
}