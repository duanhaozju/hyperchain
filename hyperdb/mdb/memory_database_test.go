package mdb

import (
	"bytes"
	"hyperchain/common"
	hdb "hyperchain/hyperdb/db"
	"testing"
)

func TestMemDatabase_DoubleGet(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)
	db.Put([]byte("key"), []byte("value"))
	db.Put([]byte("key"), []byte("newvalue"))

	v, _ := db.Get([]byte("key"))
	if bytes.Compare(v, []byte("newvalue")) != 0 {
		t.Error("expect the return should be same with new value")
	}
}

func TestMemDatabase_PutGet(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)

	kvs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
		"key":  []byte("value"),
		"1":    []byte("value11"),
		"2":    []byte("value22"),
		"3":    []byte("value33"),
		"-":    []byte("value-"),
	}

	kvsCopied := make(map[string][]byte)

	for k, v := range kvs {
		kvsCopied[k] = CopyBytes(v)
	}

	for k, v := range kvs {
		db.Put([]byte(k), v)
	}

	for k, v := range kvs {
		blob, err := db.Get([]byte(k))
		if err != nil || bytes.Compare(v, blob) != 0 {
			t.Error("missing or not match")
		}
	}

	// test if modification will affect db content
	for _, v := range kvs {
		v[0] = byte(0)
	}

	for k, v := range kvsCopied {
		blob, err := db.Get([]byte(k))
		if err != nil || bytes.Compare(v, blob) != 0 {
			t.Error("missing or not match")
		}
	}

}

func TestMemDatabase_Delete(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)

	kvs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
		"key":  []byte("value"),
		"1":    []byte("value11"),
		"2":    []byte("value22"),
		"3":    []byte("value33"),
		"-":    []byte("value-"),
	}

	for k, v := range kvs {
		db.Put([]byte(k), v)
	}

	cnt := 0
	for k := range kvs {
		if cnt%2 == 0 {
			db.Delete([]byte(k))
		}
	}
	for k, v := range kvs {
		if cnt%2 == 0 {
			_, err := db.Get([]byte(k))
			if err == nil || err != hdb.DB_NOT_FOUND {
				t.Error("expect to been deleted")
			}
		} else {
			blob, err := db.Get([]byte(k))
			if err != nil || bytes.Compare(blob, v) != 0 {
				t.Error("expect to not affected by deletion")
			}
		}
	}
}

func TestMemBatch_Write(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)

	kvs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
		"key":  []byte("value"),
		"1":    []byte("value11"),
		"2":    []byte("value22"),
		"3":    []byte("value33"),
		"-":    []byte("value-"),
	}
	batch := db.NewBatch()

	for k, v := range kvs {
		batch.Put([]byte(k), v)
	}

	// check whether batch put will affect db
	for k := range kvs {
		_, err := db.Get([]byte(k))
		if err == nil || err != hdb.DB_NOT_FOUND {
			t.Error("expect to be nil")
		}
	}
	batch.Write()

	for k, v := range kvs {
		blob, err := db.Get([]byte(k))
		if err != nil || bytes.Compare(blob, v) != 0 {
			t.Error("expect to be same")
		}
	}

	cnt := 0
	for k := range kvs {
		if cnt%2 == 0 {
			batch.Delete([]byte(k))
		}
	}

	batch.Write()

	for k, v := range kvs {
		if cnt%2 == 0 {
			_, err := db.Get([]byte(k))
			if err == nil || err != hdb.DB_NOT_FOUND {
				t.Error("expect to been deleted")
			}
		} else {
			blob, err := db.Get([]byte(k))
			if err != nil || bytes.Compare(blob, v) != 0 {
				t.Error("expect to not affected by deletion")
			}
		}
	}
}

func TestMemDatabase_Iterator(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)
	for _, kv := range expect {
		db.Put([]byte(kv.key), kv.value)
	}

	iter := db.NewIterator([]byte("key"))
	if checkEqual(iter, expect[1:]) {
		t.Error("iterator with specified prefix failed")
	}

	iter = db.Scan(nil, nil)
	if checkEqual(iter, expect[:]) {
		t.Error("iterator with empty start and empty limit failed")
	}

	iter = db.Scan([]byte("key"), nil)
	if checkEqual(iter, expect[1:]) {
		t.Error("iterator with specified start and empty limit failed")
	}

	iter = db.Scan(nil, []byte("key2"))
	if checkEqual(iter, expect[:6]) {
		t.Error("iterator with empty start and specified limit failed")
	}

	iter = db.Scan([]byte("key"), []byte("key2"))
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

var expect = []KV{
	{"-", []byte("value-")},
	{"key", []byte("value")},
	{"key01", []byte("value01")},
	{"key1", []byte("value1")},
	{"key10", []byte("value10")},
	{"key11", []byte("value11")},
	{"key2", []byte("value2")},
}
