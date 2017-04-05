//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package persist

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/base64"
	ndb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
)


// StoreState stores a key,value pair
func StoreState(namespace string, key string, value []byte) error {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return err
	}
	return db.Put([]byte("consensus." + key), value)
}

//DelAllState: remove all state
//func DelAllState() error {
//	db, err := hyperdb.GetLDBDatabase()
//	if err == nil {
//		db.Destroy()
//	}
//	return err
//}

// DelState removes a key,value pair
func DelState(namesapce string, key string) error {
	db, err := hyperdb.GetDBConsensusByNamespace(namesapce)
	if err != nil {
		return err
	}
	return db.Delete([]byte("consensus." + key))
}

// ReadState retrieves a value to a key
func ReadState(namespace string, key string) ([]byte, error) {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return db.Get([]byte("consensus." + key))
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix
func ReadStateSet(namespace string, prefix string) (map[string][]byte, error) {
	db, err := hyperdb.GetDBConsensusByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	prefixRaw := []byte("consensus." + prefix)

	ret := make(map[string][]byte)
	it := db.NewIterator(prefixRaw)
	if it == nil {
		err := errors.New(fmt.Sprint("Can't get Iterator"))
		return nil, err
	}
	if !it.Seek(prefixRaw) {
		err := errors.New(fmt.Sprintf("Cannot find key with %s in database", prefixRaw))
		return nil, err
	}
	for ; bytes.HasPrefix(it.Key(), prefixRaw); it.Next() {
		key := string(it.Key())
		key = key[len("consensus."):]
		ret[key] = append([]byte(nil), it.Value()...)
	}
	it.Release()
	return ret, nil
}

func GetBlockchainInfo(namespace string) *types.Chain {
	bcInfo := ndb.GetChainUntil(namespace)
	return bcInfo
}

func GetCurrentBlockInfo(namespace string) (uint64, []byte, []byte) {
	info := ndb.GetChainCopy(namespace)
	return info.Height, info.LatestBlockHash, info.ParentBlockHash
}

func GetBlockHeightAndHash(namespace string) (uint64, string) {
	bcInfo := ndb.GetChainCopy(namespace)
	hash := base64.StdEncoding.EncodeToString(bcInfo.LatestBlockHash)
	return bcInfo.Height, hash
}

func GetHeightofChain(namespace string) uint64 {
	return ndb.GetHeightOfChain(namespace)
}