package hyperstate

import (
	"testing"
	"hyperchain/hyperdb/mdb"
	"hyperchain/common"
	tutil "hyperchain/core/test_util"
	"hyperchain/core/vm"
	"bytes"
)

func TestStorageIteratorWithPrefix(t *testing.T) {
	stateDb := InitTestState()
	expect := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key01")) : []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key3")) : []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) : []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key6")) : []byte("value6"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
	}

	iter, _ := stateDb.NewIterator(common.BytesToAddress([]byte("address001")), vm.BytesPrefix([]byte("key")))
	for iter.Next() {
		if v1, ok := expect[common.BytesToRightPaddingHash(iter.Key())]; !ok || bytes.Compare(iter.Value(), v1) != 0 {
			t.Error("expect to be same")
		}
	}
}

func TestStorageIteratorWithRange(t *testing.T) {
	stateDb := InitTestState()
	// Checking
	// 1. dump all
	expect := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key01")) : []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key3")) : []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) : []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key6")) : []byte("value6"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
		common.BytesToRightPaddingHash([]byte("-"))    : []byte("value-"),
	}
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), nil, nil, expect) {
		t.Error("nil start and nil limit, failed")
	}
	// 2. with start and limit
	expect2 := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key3")) : []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) : []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
	}
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), []byte("key6"), expect2) {
		t.Error("nil start and nil limit, failed")
	}
	// 3. with nil start and limit
	expect3 := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("-"))    : []byte("value-"),
		common.BytesToRightPaddingHash([]byte("key01")) : []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key3")) : []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) : []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
	}
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), nil, []byte("key6"), expect3) {
		t.Error("nil start and nil limit, failed")
	}
	// 4. with start and nil limit
	expect4 := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key3")) : []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) : []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key6")) : []byte("value6"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
	}
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), nil, expect4) {
		t.Error("nil start and nil limit, failed")
	}
}

func InitTestState() *StateDB {
	configPath := "../../configuration/namespaces/global/config/namespace.toml"
	conf := tutil.InitConfig(configPath)
	common.InitHyperLoggerManager(conf)
	common.InitRawHyperLogger(common.DEFAULT_NAMESPACE)
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)


	stateDb, _ := New(common.Hash{}, db, db, conf, 0, common.DEFAULT_NAMESPACE)
	stateDb.MarkProcessStart(1)
	stateDb.CreateAccount(common.BytesToAddress([]byte("address001")))

	kvs := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key01")) : []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")) :  []byte("value1"),
		common.BytesToRightPaddingHash([]byte("key2")) :  []byte("value2"),
		common.BytesToRightPaddingHash([]byte("key3")) :  []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key4")) :  []byte("value4"),
		common.BytesToRightPaddingHash([]byte("key10")) : []byte("value10"),
		common.BytesToRightPaddingHash([]byte("key"))  :  []byte("value"),
		common.BytesToRightPaddingHash([]byte("-"))    :  []byte("value-"),
	}

	for k, v := range kvs {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}

	stateDb.Commit()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	kvs2 := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key1")) : []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")) : []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key5")) : []byte("value5"),
		common.BytesToRightPaddingHash([]byte("key6")) : []byte("value6"),
		common.BytesToRightPaddingHash([]byte("key"))  : nil,
	}


	stateDb.MarkProcessStart(2)

	for k, v := range kvs2 {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}
	return stateDb

}

func CheckIteratorResult(stateDb *StateDB, addr common.Address, start, limit []byte, expect map[common.Hash][]byte) bool {
	var (
		startH, limitH *common.Hash
		cnt            int
	)
	if start != nil {
		tmp := common.BytesToRightPaddingHash(start)
		startH = &tmp
	}
	if limit != nil {
		tmp := common.BytesToRightPaddingHash(limit)
		limitH = &tmp
	}
	iter, err := stateDb.NewIterator(addr, &vm.IterRange{startH, limitH})
	if err != nil {
		return false
	}
	defer iter.Release()
	for iter.Next() {
		if v1, ok := expect[common.BytesToRightPaddingHash(iter.Key())]; !ok || bytes.Compare(iter.Value(), v1) != 0 {
			return false
		}
		cnt += 1
	}
	if cnt != len(expect) {
		return false
	}
	return true
}
