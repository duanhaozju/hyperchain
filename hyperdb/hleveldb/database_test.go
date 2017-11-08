//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hleveldb

import (
	"bytes"
	"os"
	"testing"

	"github.com/hyperchain/hyperchain/common"
	hdb "github.com/hyperchain/hyperchain/hyperdb/db"
)

var testMap = map[string][]byte{
	"key1": []byte("value1"),
	"key2": []byte("value2"),
	"key3": []byte("value3"),
	"key":  []byte("value"),
	"1":    []byte("value11"),
	"2":    []byte("value22"),
	"3":    []byte("value33"),
	"-":    []byte("value-"),
}

type KV struct {
	key   string
	value []byte
}

var expect = []KV{
	{"-", []byte("value-")},
	{"key", []byte("value")},
	{"key01", []byte("value01")},
	{"key1", []byte("value1")},
	{"key10", []byte("value10")},
	{"key11", []byte("value11")},
	{"key2", []byte("value2")},
}

// TestNewLDBDataBase is unit test for NewLDBDataBase
func TestNewLDBDataBase(t *testing.T) {
	dir, _ := os.Getwd()
	testdb, _ := NewLDBDataBase(common.NewRawConfig(), dir+"/db", common.DEFAULT_NAMESPACE)
	defer func() {
		testdb.Close()
		os.RemoveAll(dir + "/db")
	}()
	if testdb.path != dir+"/db" && testdb.db == nil {
		t.Error("new ldbdatabase is wrong")
	}
}

// TestLDBDatabasePutGetDel is unit test for LDBDatabase method
// such as Put Get Delete
func TestLDBDatabase_PutGetDel(t *testing.T) {
	dir, _ := os.Getwd()
	testdb, _ := NewLDBDataBase(common.NewRawConfig(), dir+"/db", common.DEFAULT_NAMESPACE)
	defer func() {
		testdb.Close()
		os.RemoveAll(dir + "/db")
	}()

	// put data
	for key, value := range testMap {
		err := testdb.Put([]byte(key), value)
		if err != nil {
			t.Error(err)
		}
	}

	// get data
	for key, value := range testMap {
		data, err := testdb.Get([]byte(key))
		if err != nil {
			t.Error(err)
		}
		if string(data) != string(value) {
			t.Errorf("test fail %s does not equal %s", string(data), value)
		}
	}

	// delete datas
	for key, _ := range testMap {
		err := testdb.Delete([]byte(key))
		if err != nil {
			t.Error(err)
		}
	}
}

func TestLDBDatabase_Iterator(t *testing.T) {
	dir, _ := os.Getwd()
	testdb, _ := NewLDBDataBase(common.NewRawConfig(), dir+"/db", "global")
	defer func() {
		testdb.Close()
		os.RemoveAll(dir + "/db")
	}()

	for _, kv := range expect {
		err := testdb.Put([]byte(kv.key), kv.value)
		if err != nil {
			t.Error(err)
		}
	}

	iter := testdb.NewIterator([]byte("key"))
	if checkEqual(iter, expect[1:]) {
		t.Error("iterator with specified prefix failed")
	}

	iter = testdb.Scan(nil, nil)
	if checkEqual(iter, expect[:]) {
		t.Error("iterator with empty start and empty limit failed")
	}

	iter = testdb.Scan([]byte("key"), nil)
	if checkEqual(iter, expect[1:]) {
		t.Error("iterator with specified start and empty limit failed")
	}

	iter = testdb.Scan(nil, []byte("key2"))
	if checkEqual(iter, expect[:6]) {
		t.Error("iterator with empty start and specified limit failed")
	}

	iter = testdb.Scan([]byte("key"), []byte("key2"))
	if checkEqual(iter, expect[1:6]) {
		t.Error("iterator with specified start and limit failed")
	}
}

func checkEqual(iter hdb.Iterator, expect []KV) (equal bool) {
	defer func() {
		if r := recover(); r != nil {
			equal = false
		}
	}()
	var index int
	for iter.Next() {
		if string(iter.Key()) != expect[index].key {
			equal = false
			return
		}
		if bytes.Compare(iter.Value(), expect[index].value) != 0 {
			equal = false
			return
		}
	}
	return true
}
