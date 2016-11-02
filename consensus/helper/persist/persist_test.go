// author: Xiaoyi Wang
// email: wangxiaoyi@hyperchain.cn
// date: 16/11/1
// last modified: 16/11/1
// last Modified Author: Xiaoyi Wang
// change log: 1.new unit test for persist

package persist

import (
	"testing"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"reflect"
)

func TestDaoOnState(t *testing.T) {

	k := "k"
	v1 := []byte("v1")
	v2 := []byte("v2")

	var err = StoreState(k, v1)
	if err != nil {
		t.Errorf(`error StoreState("k", "v1")`, err) //todo: try to format a common error style
	}
	value, err := ReadState(k)
	if bytes.Compare(value, v1) != 0 || err != nil {
		t.Errorf(`error ReadState("k") != "v1"`)
	}

	err = StoreState(k, v2)
	if err != nil {
		t.Errorf(`error StoreState("k", "v2")`)
	}

	value2, err := ReadState(k)
	if bytes.Compare(value2, v2) != 0 || err != nil {
		t.Errorf(`error ReadState("k") != "v2"`)
	}

	nk := "no_exists_key"
	err = DelState(nk)
	if err != nil {
		t.Errorf(`error DelState("no_exists_key")`)
	}

	DelState(k)
	_, err = ReadState(k)
	if (err != nil && err != errors.ErrNotFound) {
		t.Error(`error ReadState("k"), no value for k`)
	}
}

func TestReadStateSet(t *testing.T) {
	kvs := map[string][]byte{
		"key1" : []byte("hello1"),
		"key2" : []byte("hello2"),
		"sssdddss" : []byte("hello3"),
	}
	for k, v := range kvs {
		StoreState(k, v)
	}
	var v, err = ReadStateSet("key1")
	var target = map[string][]byte{
		"key1":[]byte("hello1"),
	}
	if err != nil || !reflect.DeepEqual(target, v) {
		t.Error(`"error ReadStateSet("key1") not found "hello1"`)
	}

	v, err = ReadStateSet("k")
	target = map[string][]byte{
		"key1" : []byte("hello1"),
		"key2" : []byte("hello2"),
	}

	if err != nil || !reflect.DeepEqual(target, v) {
		t.Error(`"error ReadStateSet("k")`)
	}

	v, err = ReadStateSet("")
	target = map[string][]byte{
		"key1" : []byte("hello1"),
		"key2" : []byte("hello2"),
		"sssdddss" : []byte("hello3"),
	}
	if err != nil || !reflect.DeepEqual(target, v) {
		t.Error(`"error ReadStateSet("k")`)
	}

	v, err = ReadStateSet("no key")
	target = map[string][]byte{}

	if err != nil || !reflect.DeepEqual(target, v) {
		t.Error(`"error ReadStateSet("k")`)
	}
	for k := range kvs {// clear the test data
		DelState(k)
	}

}
