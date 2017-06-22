//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package persist

import (
	"bytes"
	"errors"
	"fmt"
	"hyperchain/hyperdb"
)

// StoreState stores a key,value pair
func PutData(key string, value []byte, namespace string) error {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return err
	}
	return db.Put([]byte("p2p."+key), value)
}

func PutBool(key string, value bool, namespace string) error {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return err
	}
	var persistValue []byte

	if value {
		persistValue = []byte("true")
	} else {
		persistValue = []byte("false")
	}

	return db.Put([]byte("p2p."+key), persistValue)
}

func GetBool(key string, namespace string) (bool, error) {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return false, err
	}
	persistKey, err := db.Get([]byte("p2p." + key))
	if string(persistKey) == "true" {
		return true, err
	}
	return false, err
}

//DelAllState: remove all state
//func DelAllState() error {
//	db, err := hyperdb.GetDBDatabase()
//	if err == nil {
//		db.Destroy()
//	}
//	return err
//}

// DelState removes a key,value pair
func DelData(key string, namespace string) error {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return err
	}
	return db.Delete([]byte("p2p." + key))
}

// ReadState retrieves a value to a key
func GetData(key string, namespace string) ([]byte, error) {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return db.Get([]byte("p2p." + key))
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix
func GetDataSet(prefix string, namespace string) (map[string][]byte, error) {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	prefixRaw := []byte("p2p." + prefix)

	ret := make(map[string][]byte)
	it := db.NewIterator(prefixRaw)
	if it == nil {
		err := errors.New(fmt.Sprintf("Can't get Iterator"))
		return nil, err
	}
	if !it.Seek(prefixRaw) {
		err := errors.New(fmt.Sprintf("Cannot find key with %s in database", prefixRaw))
		return nil, err
	}
	for ; bytes.HasPrefix(it.Key(), prefixRaw); it.Next() {
		key := string(it.Key())
		key = key[len("p2p."):]
		ret[key] = append([]byte(nil), it.Value()...)
	}
	it.Release()
	return ret, nil
}
