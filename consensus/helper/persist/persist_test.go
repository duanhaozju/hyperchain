//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package persist

import (
	"bytes"
	"testing"

	"hyperchain/common"
	mdb "hyperchain/hyperdb/mdb"

	"github.com/stretchr/testify/assert"
)

func TestDaoOnState(t *testing.T) {
	k := "k"
	v1 := []byte("v1")
	v2 := []byte("v2")

	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	persister := New(db)

	err = persister.StoreState(k, v1)
	if err != nil {
		t.Errorf(`error type %v : StoreState(%q, %v)`, err, k, v1)
	}
	value, err := persister.ReadState(k)
	if err != nil || bytes.Compare(value, v1) != 0 {
		t.Errorf(`error type %v : ReadState(%q) = %v, actual: %v`, err, k, value, v1)
	}

	err = persister.StoreState(k, v2)
	if err != nil {
		t.Errorf(`error type %v : StoreState(%q, %v)`, err, k, v2)
	}

	value2, err := persister.ReadState(k)
	if err != nil || bytes.Compare(value2, v2) != 0 {
		t.Errorf(`error type %v : ReadState(%q) = %v, actual: %v`, err, k, value2, v2)
	}

	nk := "no_exists_key"
	err = persister.DelState(nk)
	if err != nil {
		t.Errorf(`error type %v : DelState(%q)`, err, nk)
	}

	persister.DelState(k)
	_, err = persister.ReadState(k)
	if err != nil {
		assert.EqualError(t, err, "db not found", "should not read a deleted item")
	}
}

func TestReadStateSet(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	persister := New(db)
	ast := assert.New(t)

	kvs := map[string][]byte{
		"key1":     []byte("hello1"),
		"key2":     []byte("hello2"),
		"sssdddss": []byte("hello3"),
	}

	for k, v := range kvs {
		persister.StoreState(k, v)
	}

	v, err := persister.ReadStateSet("key1")
	var target = map[string][]byte{
		"key1": []byte("hello1"),
	}
	ast.Nil(err, "error ReadStateSet('key1') not found 'hello1'")
	ast.Equal(target, v, "error ReadStateSet('key1') not found 'hello1'")

	v, err = persister.ReadStateSet("k")
	target = map[string][]byte{
		"key1": []byte("hello1"),
		"key2": []byte("hello2"),
	}
	ast.Nil(err, "error ReadStateSet('k')")
	ast.Equal(target, v, "error ReadStateSet('k')")

	v, err = persister.ReadStateSet("")
	target = map[string][]byte{
		"key1":     []byte("hello1"),
		"key2":     []byte("hello2"),
		"sssdddss": []byte("hello3"),
	}
	ast.Nil(err, "error ReadStateSet('')")
	ast.Equal(target, v, "error ReadStateSet('')")

	v, err = persister.ReadStateSet("no key")
	ast.Equal(0, len(v), "error ReadStateSet('no key')")

	for k := range kvs { // clear the test data
		persister.DelState(k)
	}
}
