package persist

import (
	"fmt"
	"hyperchain/hyperdb"
	"hyperchain/core"
	"bytes"
	"github.com/pkg/errors"
	"hyperchain/core/types"
)


// StoreState stores a key,value pair
func StoreState(key string, value []byte) error {
	db, _ := hyperdb.GetLDBDatabase()
	return db.Put([]byte("consensus."+key), value)
}


// DelState removes a key,value pair
func DelState(key string) error {
	db, _ := hyperdb.GetLDBDatabase()
	return db.Delete([]byte("consensus."+key))
}

// ReadState retrieves a value to a key
func ReadState(key string) ([]byte, error) {
	db, _ := hyperdb.GetLDBDatabase()
	return db.Get([]byte("consensus."+key))
}

// ReadStateSet retrieves all key-value pairs where the key starts with prefix
func ReadStateSet(prefix string) (map[string][]byte, error) {
	db, _ := hyperdb.GetLDBDatabase()
	prefixRaw := []byte("consensus." + prefix)

	ret := make(map[string][]byte)
	it := db.NewIterator()
	defer it.Release()
	if !it.Seek(prefixRaw) {
		err := errors.New(fmt.Sprintf("Cannot find key with %s in database", prefixRaw))
		return nil, err
	}
	for ; bytes.HasPrefix(it.Key(), prefixRaw); it.Next() {
		key := string(it.Key())
		key = key[len("consensus."):]
		ret[key] = append([]byte(nil), it.Value()...)
	}
	return ret, nil
}

func GetBlockchainInfo() *types.Chain {
	bcInfo := core.GetChainCopy()

	return bcInfo
}