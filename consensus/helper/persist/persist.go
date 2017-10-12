//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package persist

import (
	"bytes"
	"errors"
	"fmt"

	"encoding/base64"
	ndb "hyperchain/core/ledger/db_utils"
	"hyperchain/core/types"
	"hyperchain/hyperdb/db"
)

type Persister interface {
	StoreState(key string, value []byte) error
	DelState(key string) error
	ReadState(key string) ([]byte, error)
	ReadStateSet(prefix string) (map[string][]byte, error)
	GetBlockchainInfo(namespace string) *types.Chain
	GetCurrentBlockInfo(namespace string) (uint64, []byte, []byte)
	GetBlockHeightAndHash(namespace string) (uint64, string)
	GetHeightOfChain(namespace string) uint64
	GetGenesisOfChain(namespace string) (uint64, error)
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

// GetBlockchainInfo waits until the executor module executes to a checkpoint then returns the blockchain info with the
// given namespace
func (persister *persisterImpl) GetBlockchainInfo(namespace string) *types.Chain {
	bcInfo := ndb.GetChainUntil(namespace)
	return bcInfo
}

// GetCurrentBlockInfo returns the current blockchain info with the given namespace immediately
func (persister *persisterImpl) GetCurrentBlockInfo(namespace string) (uint64, []byte, []byte) {
	info := ndb.GetChainCopy(namespace)
	return info.Height, info.LatestBlockHash, info.ParentBlockHash
}

// GetBlockHeightAndHash returns the current block height and hash with the given namespace immediately
func (persister *persisterImpl) GetBlockHeightAndHash(namespace string) (uint64, string) {
	bcInfo := ndb.GetChainCopy(namespace)
	hash := base64.StdEncoding.EncodeToString(bcInfo.LatestBlockHash)
	return bcInfo.Height, hash
}

// GetHeightOfChain returns the current block height with the given namespace immediately
func (persister *persisterImpl) GetHeightOfChain(namespace string) uint64 {
	return ndb.GetHeightOfChain(namespace)
}

// GetGenesisOfChain returns the genesis block info of the ledger with the given namespace
func (persister *persisterImpl) GetGenesisOfChain(namespace string) (uint64, error) {
	return ndb.GetGenesisTag(namespace)
}
