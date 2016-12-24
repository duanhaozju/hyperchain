//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"testing"
	"os"
)

var testMap = map[string]string{
	"key1":"value1",
	"key2":"value2",
	"key3":"value3",
	"key4":"value4",
}

var db *LDBDatabase

// TestNewLDBDataBase is unit test for NewLDBDataBase
func TestNewLDBDataBase(t *testing.T) {
	dir, _ := os.Getwd()
	db, _ = NewLDBDataBase(dir + "/db")
	defer os.RemoveAll(dir+"/db")
	if db.path != dir + "/db" && db.db == nil {
		t.Error("new ldbdatabase is wrong")
	} else {
		t.Log("TestNewLDBDataBase is pass")
	}
}

// TestLDBDatabase is unit test for LDBDatabase method
// such as Put Get Detele
func TestLDBDatabase(t *testing.T) {
	// put data
	for key, value := range testMap {
		err := db.Put([]byte(key), []byte(value))
		if err != nil {
			t.Error(err)
			return
		}
	}
	// get data
	for key, value := range testMap{
		data , err := db.Get([]byte(key))
		if err != nil {
			t.Error(err)
			return
		}
		if string(data) != value {
			t.Errorf("test fail %s does not equal %s", string(data), value)
		}
	}
	// delete datas
	for key, _ := range testMap  {
		err := db.Delete([]byte(key))
		if err != nil {
			t.Error(err)
		}
	}
}
