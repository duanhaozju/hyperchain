package db_utils

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/core/test_util"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"reflect"
	"testing"
)

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	for _, trans := range test_util.TransactionCases[:1] {
		key := trans.GetHash().Bytes()
		tr, err := GetTransaction(hyperdb.defaut_namespace, key)
		if err != nil {
			logger.Fatal(err)
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
		if res, _ := JudgeTransactionExist(hyperdb.defaut_namespace, key); res != true {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
	deleteTestData()
}

func TestGetTransactionBLk(t *testing.T) {
	logger.Info("test =============> > > TestGetTransactionBLk")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	err, _ = PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.defaut_namespace, 1)
	if err != nil {
		logger.Fatal(err)
	}
	batch := db.NewBatch()
	PersistTransactionMeta(batch, &test_util.TransactionMeta, test_util.TransactionCases[0].GetHash(), true, true)
	if len(block.Transactions) > 0 {
		tx := block.Transactions[0]
		bn, i := GetTxWithBlock(hyperdb.defaut_namespace, tx.GetHash().Bytes())
		if bn != 1 || i != 1 {
			t.Errorf("TestGetTransactionBLk fail")
		}
		DeleteTransactionMeta(batch, tx.GetHash().Bytes(), true, true)
		a, b := GetTxWithBlock(hyperdb.defaut_namespace, tx.GetHash().Bytes())
		if a != 0 || b != 0 {
			t.Errorf("TestGetTransactionBLk fail")
		}
	}
	deleteTestData()
}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransactions(t *testing.T) {
	logger.Info("test =============> > > TestGetAllTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err, _ := PersistTransaction(db.NewBatch(), test_util.TransactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	trs, err := GetAllTransaction(hyperdb.defaut_namespace)
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
	logger.Info("test =============> > > TestDeleteTransaction")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	for _, trans := range test_util.TransactionCases[:1] {
		err, _ := PersistTransaction(db.NewBatch(), trans, true, true)
		if err != nil {
			logger.Fatal(err)
		}
		DeleteTransaction(db.NewBatch(), trans.GetHash().Bytes(), true, true)
		_, err = GetTransaction(hyperdb.defaut_namespace, trans.GetHash().Bytes())
		if err.Error() != "leveldb: not found" {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", trans.GetHash().Bytes())
		}
	}
	deleteTestData()
}

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	logger.Info("test =============> > > TestPutTransactions")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	err := PersistTransactions(db.NewBatch(), test_util.TransactionCases, TransactionVersion, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	trs, err := GetAllTransaction(hyperdb.defaut_namespace)
	if err != nil {
		logger.Fatal(err)
	}
	if len(trs) < 3 {
		t.Errorf("TestPutTransactions fail")
	}
	deleteTestData()
}

// TestGetInvaildTx tests for GetDiscardTransaction
func TestGetInvaildTx(t *testing.T) {
	tx := test_util.TransactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data, _ := proto.Marshal(record)
	// save to db
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	db.Put(append(InvalidTransactionPrefix, tx.TransactionHash...), data)

	result, _ := GetInvaildTxErrType(hyperdb.defaut_namespace, tx.TransactionHash)
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
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.defaut_namespace)
	PersistInvalidTransactionRecord(db.NewBatch(), record, true, true)

	result, _ := GetDiscardTransaction(hyperdb.defaut_namespace, tx.TransactionHash)
	if result.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
		t.Errorf("TestGetDiscardTransaction fail")
	}
	results, _ := GetAllDiscardTransaction(hyperdb.defaut_namespace)
	for _, v := range results {
		if v.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
			t.Errorf("TestGetDiscardTransaction fail")
		}
	}
	deleteTestData()
}
