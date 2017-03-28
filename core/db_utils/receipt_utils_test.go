package db_utils

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"hyperchain/core/test_util"
)

func TestGetReceipt(t *testing.T) {
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
	err = DeleteReceipt(batch, test_util.Receipt.TxHash, true, true)
	if err != nil {
		t.Errorf("DeleteReceipt fail")
	}
	deleteTestData()
}