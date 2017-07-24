package db_utils

import (
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/hyperdb"
	"testing"
)

func TestGetReceipt(t *testing.T) {
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
	err = DeleteReceipt(batch, test_util.Receipt.TxHash, true, true)
	if err != nil {
		t.Errorf("DeleteReceipt fail")
	}
	deleteTestData()
}
