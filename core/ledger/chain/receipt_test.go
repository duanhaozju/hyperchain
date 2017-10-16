package chain

import (
	"bytes"
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/hyperdb/mdb"
	"reflect"
	"testing"
)

func TestGetReceipt(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := test_util.Receipt
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true)
	dbReceipt := GetReceiptFunc(db, common.BytesToHash(receipt.TxHash))
	if dbReceipt == nil || !reflect.DeepEqual(&receipt, dbReceipt) {
		t.Error("not exist in db or not deep equal exactly")
	}
}

func TestPersistReceiptWithVersion(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := test_util.Receipt
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true, "1.4")
	dbReceipt := GetReceiptFunc(db, common.BytesToHash(receipt.TxHash))
	if dbReceipt == nil || bytes.Compare(dbReceipt.Version, []byte("1.4")) != 0 {
		t.Error("not exist in db or not deep equal exactly")
	}
}

func TestDeleteReceipt(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	receipt := test_util.Receipt
	receipt.TxHash = common.BytesToHash(receipt.TxHash).Bytes()
	PersistReceipt(db.NewBatch(), &receipt, true, true, "1.4")
	DeleteReceipt(db.NewBatch(), receipt.TxHash, true, true)

	if dbReceipt := GetReceiptFunc(db, common.BytesToHash(receipt.TxHash)); dbReceipt != nil {
		t.Error("expect to be nil")
	}
}
