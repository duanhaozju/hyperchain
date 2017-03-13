package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/common"
)

func TestGetReceipt(t *testing.T) {
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
	err = DeleteReceipt(batch, receipt.TxHash, true, true)
	if err != nil {
		t.Errorf("DeleteReceipt fail")
	}
	deleteTestData()
}