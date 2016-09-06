package persist

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
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
	for ok := it.Seek(prefixRaw); ok; ok = it.Next() {
		key := it.Key()
		if len(key) > len(prefixRaw) && string(key[0 : len(prefixRaw)]) == string(prefixRaw) {
			ret[string(key)] = it.Value()
		} else {
			break
		}
	}
	it.Release()
	return ret, nil
}

func GetBlockchainInfo() *types.Chain {
	bcInfo := core.GetChainCopy()
	return bcInfo
}