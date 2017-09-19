package state

import (
	"bytes"
	"hyperchain/common"
	tutil "hyperchain/core/test_util"
	"hyperchain/core/vm"
	"hyperchain/hyperdb/mdb"
	"testing"
)

type KV struct {
	Key   common.Hash
	Value []byte
}

func TestMemIterator_Iter(t *testing.T) {
	testValue := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key")):   []byte("value"),
		common.BytesToRightPaddingHash([]byte("key01")): []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")):  []byte("value1"),
		common.BytesToRightPaddingHash([]byte("key11")): []byte("value11"),
		common.BytesToRightPaddingHash([]byte("key2")):  []byte("value2"),
	}
	expect := []KV{
		{
			Key:   common.BytesToRightPaddingHash([]byte("key")),
			Value: []byte("value"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key01")),
			Value: []byte("value01"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key1")),
			Value: []byte("value1"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key11")),
			Value: []byte("value11"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key2")),
			Value: []byte("value2"),
		},
	}
	cnt := 0
	iter := NewMemIterator(Storage(testValue))
	defer iter.Release()
	for iter.Next() {
		if bytes.Compare(iter.Key(), expect[cnt].Key.Bytes()) != 0 {
			t.Error("expect to be same key")
		}
		if bytes.Compare(iter.Value(), expect[cnt].Value) != 0 {
			t.Error("expect to be same value")
		}
		cnt += 1
	}
	if cnt != len(expect) {
		t.Error("except to be same length")
	}
}

func TestStorageIteratorWithPrefix(t *testing.T) {
	stateDb := InitTestState()
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")),
		vm.BytesPrefix([]byte("key")).Start.Bytes(), vm.BytesPrefix([]byte("key")).Limit.Bytes(), ExpectResult()[1:]) {
		t.Error("iter with prefix failed")
	}
}

func TestStorageIteratorWithRange(t *testing.T) {
	stateDb := InitTestState()
	// Checking
	// 1. dump all
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), nil, nil, ExpectResult()) {
		t.Error("nil start and nil limit, failed")
	}
	// 2. with start and limit
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), []byte("key2"), ExpectResult()[2:4]) {
		t.Error("start and limit, failed")
	}
	// 3. with nil start and limit
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), nil, []byte("key2"), ExpectResult()[:4]) {
		t.Error("nil start and limit, failed")
	}
	// 4. with start and nil limit
	if !CheckIteratorResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), nil, ExpectResult()[2:]) {
		t.Error("start and nil limit, failed")
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

	for k, v := range InitData1() {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}

	stateDb.Commit()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	stateDb.MarkProcessStart(2)

	for k, v := range InitData2() {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}
	return stateDb

}

func CheckIteratorResult(stateDb *StateDB, addr common.Address, start, limit []byte, expect []KV) bool {
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

	validity := true
	for iter.Next() {
		if bytes.Compare(iter.Key(), expect[cnt].Key.Bytes()) != 0 {
			validity = false
		}
		if bytes.Compare(iter.Value(), expect[cnt].Value) != 0 {
			validity = false
		}
		cnt += 1
	}
	if cnt != len(expect) {
		validity = false
	}
	return validity
}

func InitData1() map[common.Hash][]byte {
	return map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key01")): []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")):  []byte("value1"),
		common.BytesToRightPaddingHash([]byte("key10")): []byte("value10"),
		common.BytesToRightPaddingHash([]byte("key")):   []byte("value"),
		common.BytesToRightPaddingHash([]byte("-")):     []byte("value-"),
	}
}

func InitData2() map[common.Hash][]byte {
	return map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key1")): []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")): []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key")):  nil,
	}
}

func ExpectResult() []KV {
	expect := []KV{
		{
			Key:   common.BytesToRightPaddingHash([]byte("-")),
			Value: []byte("value-"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key01")),
			Value: []byte("value01"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key1")),
			Value: []byte("newvalue1"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key10")),
			Value: []byte("value10"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key2")),
			Value: []byte("newvalue2"),
		},
	}
	return expect
}
