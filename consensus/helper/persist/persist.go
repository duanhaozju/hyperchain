//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package persist

import (
	"fmt"
	"errors"
	"bytes"

	"hyperchain/hyperdb"
	"hyperchain/core"
	"hyperchain/core/types"
	"encoding/base64"
)


// StoreState stores a key,value pair
func StoreState(key string, value []byte) error {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return db.Put([]byte("consensus."+key), value)
}

//DelAllState: remove all state
func DelAllState() error {
	db, err := hyperdb.GetLDBDatabase()
	if err == nil {
		db.Destroy()
	}
	return err
}

// DelState removes a key,value pair
func DelState(key string) error {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return db.Delete([]byte("consensus."+key))
}

// ReadState retrieves a value to a key
func ReadState(key string) ([]byte, error) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil, err
	}
	return db.Get([]byte("consensus."+key))
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix
func ReadStateSet(prefix string) (map[string][]byte, error) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil, err
	}
	prefixRaw := []byte("consensus." + prefix)

	ret := make(map[string][]byte)
	it := db.NewIterator()
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

func GetBlockchainInfo() *types.Chain {
	bcInfo := core.GetChainUntil()
	return bcInfo
}

func GetBlockHeightAndHash() (uint64, string) {
	bcInfo := core.GetChainCopy()
	hash := base64.StdEncoding.EncodeToString(bcInfo.LatestBlockHash)
	return bcInfo.Height, hash
}

func GetHeightofChain() uint64 {
	return core.GetHeightOfChain()
}