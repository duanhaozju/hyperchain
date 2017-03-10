package db_utils

import (
	"testing"
	"fmt"
	"hyperchain/hyperdb"
	"reflect"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
)

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	logger.Info("test =============> > > TestGetTransaction")
	db := InitDataBase()
	err, _ := PersistTransaction(db.NewBatch(), transactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	for _, trans := range transactionCases[:1] {
		key := trans.GetHash().Bytes()
		tr, err := GetTransaction(key)
		if err != nil {
			logger.Fatal(err)
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
		if res, _ :=JudgeTransactionExist(key); res != true {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
}

func TestGetTransactionBLk(t *testing.T) {
	logger.Info("test =============> > > TestGetTransactionBLk")
	db := InitDataBase()
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	err, _ = PersistTransaction(db.NewBatch(), transactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlockByNumber(hyperdb.DefautNameSpace + hyperdb.Blockchain, 1)
	if err != nil {
		logger.Fatal(err)
	}
	batch := db.NewBatch()
	PersistTransactionMeta(batch, &transactionMeta, transactionCases[0].GetHash(), true, true)
	if len(block.Transactions) > 0 {
		fmt.Println("tx hash", block.Transactions[0].GetHash().Bytes())
		tx := block.Transactions[0]
		bn, i := GetTxWithBlock(tx.GetHash().Bytes())
		fmt.Println("block num :", bn, "tx index:", i)
		DeleteTransactionMeta(batch, tx.GetHash().Bytes(), true, true)
		a, b := GetTxWithBlock(tx.GetHash().Bytes())
		fmt.Println("block num :", a, "tx index:", b)
	}
}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransactions(t *testing.T) {
	logger.Info("test =============> > > TestGetAllTransaction")
	db := InitDataBase()
	err, _ := PersistTransaction(db.NewBatch(), transactionCases[0], true, true)
	if err != nil {
		logger.Fatal(err)
	}
	trs, err := GetAllTransaction()
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
}

// TestDeleteTransaction tests for DeleteTransaction
func TestDeleteTransaction(t *testing.T) {
	logger.Info("test =============> > > TestDeleteTransaction")
	db := InitDataBase()
	for _, trans := range transactionCases[:1] {
		err, _ := PersistTransaction(db.NewBatch(), trans, true, true)
		if err != nil {
			logger.Fatal(err)
		}
		DeleteTransaction(db.NewBatch(), trans.GetHash().Bytes(), true, true)
		_, err = GetTransaction(trans.GetHash().Bytes())
		if err.Error() != "leveldb: not found" {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", trans.GetHash().Bytes())
		}
	}
}

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	logger.Info("test =============> > > TestPutTransactions")
	db := InitDataBase()
	err := PersistTransactions(db.NewBatch(), transactionCases, "1.2", true, true)
	if err != nil {
		logger.Fatal(err)
	}
	trs, err := GetAllTransaction()
	if err != nil {
		logger.Fatal(err)
	}
	if len(trs) < 3 {
		t.Errorf("TestPutTransactions fail")
	}
}

// TestGetInvaildTx tests for GetDiscardTransaction
func TestGetInvaildTx(t *testing.T) {
	tx := transactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data, _ := proto.Marshal(record)
	// save to db
	db := InitDataBase()
	db.Put(append(InvalidTransactionPrefix, tx.TransactionHash...), data)

	result, _ := GetInvaildTxErrType(tx.TransactionHash)
	fmt.Println(result)
}

// TestGetDiscardTransaction tests for GetDiscardTransaction
func TestGetDiscardTransaction(t *testing.T) {
	tx := transactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	db := InitDataBase()
	PersistInvalidTransactionRecord(db.NewBatch(), record, true, true)

	result, _ := GetDiscardTransaction(tx.TransactionHash)
	if result.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
		t.Errorf("TestGetDiscardTransaction fail")
	}
	results, _ := GetAllDiscardTransaction()
	for _, v := range results {
		if v.ErrType != types.InvalidTransactionRecord_OUTOFBALANCE {
			t.Errorf("TestGetDiscardTransaction fail")
		}
	}
}