//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package persist

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hyperchain/hyperchain/hyperdb/db"
)

type Persister interface {
	StoreState(key string, value []byte) error
	DelState(key string) error
	ReadState(key string) ([]byte, error)
	ReadStateSet(prefix string) (map[string][]byte, error)
}

type persisterImpl struct {
	db db.Database
}

func New(db db.Database) *persisterImpl {
	persister := &persisterImpl{
		db: db,
	}
	return persister
}

// StoreState stores a key,value pair to the database with the given namespace
func (persister *persisterImpl) StoreState(key string, value []byte) error {
	return persister.db.Put([]byte("consensus."+key), value)
}

// DelState removes a key,value pair from the database with the given namespace
func (persister *persisterImpl) DelState(key string) error {
	return persister.db.Delete([]byte("consensus." + key))
}

// ReadState retrieves a value to a key from the database with the given namespace
func (persister *persisterImpl) ReadState(key string) ([]byte, error) {
	return persister.db.Get([]byte("consensus." + key))
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix from the database with the given namespace
func (persister *persisterImpl) ReadStateSet(prefix string) (map[string][]byte, error) {
	prefixRaw := []byte("consensus." + prefix)

	ret := make(map[string][]byte)
	it := persister.db.NewIterator(prefixRaw)
	if it == nil {
		err := errors.New(fmt.Sprint("Can't get Iterator"))
		return nil, err
	}
	if !it.Seek(prefixRaw) {
		err := errors.New(fmt.Sprintf("Cannot find key with %s in database", prefixRaw))
		return nil, err
	}
	for bytes.HasPrefix(it.Key(), prefixRaw) {
		key := string(it.Key())
		key = key[len("consensus."):]
		ret[key] = append([]byte(nil), it.Value()...)
		if !it.Next() {
			break
		}
	}
	it.Release()
	return ret, nil
}
