//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hleveldb

import (
	"os"
	"testing"
	"github.com/hyperchain/hyperchain/common"
)

var testMap = map[string]string{
	"key1": "value1",
	"key2": "value2",
	"key3": "value3",
	"key4": "value4",
}

var testdb *LDBDatabase

// TestNewLDBDataBase is unit test for NewLDBDataBase
func TestNewLDBDataBase(t *testing.T) {
	dir, _ := os.Getwd()
	testdb, _ = NewLDBDataBase(common.NewRawConfig(), dir + "/db", common.DEFAULT_NAMESPACE)
	defer os.RemoveAll(dir + "/db")
	if testdb.path != dir+"/db" && testdb.db == nil {
		t.Error("new ldbdatabase is wrong")
	} else {
		t.Log("TestNewLDBDataBase is pass")
	}
}

// TestLDBDatabase is unit test for LDBDatabase method
// such as Put Get Delete
func TestLDBDatabase(t *testing.T) {
	// put data
	for key, value := range testMap {
		err := testdb.Put([]byte(key), []byte(value))
		if err != nil {
			t.Error(err)
			return
		}
	}
	// get data
	for key, value := range testMap {
		data, err := testdb.Get([]byte(key))
		if err != nil {
			t.Error(err)
			return
		}
		if string(data) != value {
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
