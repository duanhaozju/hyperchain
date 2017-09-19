package db_utils

import (
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/core/types"
	"hyperchain/hyperdb/mdb"
	"reflect"
	"testing"
)

func TestGetTransaction(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)

	for idx, tx := range test_util.TransactionCases {
		meta := &types.TransactionMeta{
			BlockIndex: test_util.BlockCases.Number,
			Index:      int64(idx),
		}
		PersistTransactionMeta(db.NewBatch(), meta, common.BytesToHash(tx.TransactionHash), true, true)
	}

	for _, tx := range test_util.TransactionCases {
		dbTx, err := GetTransactionFunc(db, common.BytesToHash(tx.TransactionHash).Bytes())
		if err != nil {
			t.Error(err.Error())
		}
		if !CompareTransaction(tx, dbTx) {
			t.Error("expect to be same")
		}
	}
}

func TestJudgeTransactionExist(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	for idx, tx := range test_util.TransactionCases {
		meta := &types.TransactionMeta{
			BlockIndex: test_util.BlockCases.Number,
			Index:      int64(idx),
		}
		PersistTransactionMeta(db.NewBatch(), meta, common.BytesToHash(tx.TransactionHash), true, true)
	}

	for _, tx := range test_util.TransactionCases {
		if exist, _ := JudgeTransactionExistFunc(db, common.BytesToHash(tx.TransactionHash).Bytes()); !exist {
			t.Error("expect to be not exist")
		}
	}
}

// TestGetInvaildTx tests for GetDiscardTransaction
func TestGetInvaildTx(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	tx := test_util.TransactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	PersistInvalidTransactionRecord(db.NewBatch(), record, true, true)

	dbRecord, err := GetDiscardTransactionFunc(db, tx.TransactionHash)
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(record, dbRecord) {
		t.Error("expect to be same")
	}
}

func CompareTransaction(tx1 *types.Transaction, tx2 *types.Transaction) bool {
	if reflect.DeepEqual(tx1, tx2) {
		return true
	} else {
		return false
	}
}
