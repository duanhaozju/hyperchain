package mdb

import (
	"testing"
	"hyperchain/common"
	"bytes"
)

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
		if cnt % 2 == 0 {
			db.Delete([]byte(k))
		}
	}
	for k, v := range kvs {
		if cnt % 2 == 0 {
			_, err := db.Get([]byte(k))
			if err == nil {
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
		if err == nil {
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
		if cnt % 2 == 0 {
			batch.Delete([]byte(k))
		}
	}

	batch.Write()

	for k, v := range kvs {
		if cnt % 2 == 0 {
			_, err := db.Get([]byte(k))
			if err == nil {
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

func TestMemDatabase_NewIterator(t *testing.T) {
	db, _ := NewMemDatabase(common.DEFAULT_NAMESPACE)

	kvs := map[string][]byte{
		"key1":   []byte("value1"),
		"key2":   []byte("value2"),
		"key3":   []byte("value3"),
		"key100": []byte("value100"),
		"key":    []byte("value"),
		"1":      []byte("value11"),
		"2":      []byte("value22"),
		"3":      []byte("value33"),
		"-":      []byte("value-"),
		"~":      []byte("value~"),
	}

	for k, v := range kvs {
		db.Put([]byte(k), v)
	}
	// With Prefix
	iter := db.NewIterator([]byte("key"))
	iterContent := make(map[string][]byte)
	for iter.Next() {
		iterContent[string(iter.Key())] = CopyBytes(iter.Value())
	}

	for k, v := range kvs {
		if bytes.HasPrefix([]byte(k), []byte("key")) {
			if v1, ok := iterContent[k]; !ok || bytes.Compare(v, v1) != 0 {
				t.Error("expect to be same")
			}
		}
	}

	for k, v := range iterContent {
		if v1, ok := kvs[k]; !ok || bytes.Compare(v, v1) != 0 {
			t.Error("expect to be same")
		}
	}

	// With empty start and empty limit
	iterContent = make(map[string][]byte)
	iter1 := db.Scan(nil, nil)

	for iter1.Next() {
		iterContent[string(iter1.Key())] = CopyBytes(iter1.Value())
	}

	for k, v := range kvs {
		if v1, ok := iterContent[k]; !ok || bytes.Compare(v, v1) != 0 {
			t.Error("expect to be same")
		}
	}

	// With empty start but a limit
	iter2 := db.Scan(nil, []byte("key3"))
	expect := map[string][]byte{
		"key1":    []byte("value1"),
		"key2":    []byte("value2"),
		"key":     []byte("value"),
		"key100":  []byte("value100"),
		"1":       []byte("value11"),
		"2":       []byte("value22"),
		"3":       []byte("value33"),
		"-":       []byte("value-"),
	}
	for iter2.Next() {
		if v, ok := expect[string(iter2.Key())]; !ok || bytes.Compare(v, iter2.Value()) != 0 {
			t.Error("expect to be same")
		}
	}
	// With a non-empty start and a limit
	iter3 := db.Scan([]byte("key1"), []byte("key3"))
	expect = map[string][]byte{
		"key1":    []byte("value1"),
		"key2":    []byte("value2"),
		"key100":  []byte("value100"),
	}
	for iter3.Next() {
		if v, ok := expect[string(iter3.Key())]; !ok || bytes.Compare(v, iter3.Value()) != 0 {
			t.Error("expect to be same")
		}
	}
}

