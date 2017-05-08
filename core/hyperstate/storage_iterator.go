package hyperstate

import (
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"bytes"
	"fmt"
)

type StorageIterator struct {
	cache      Storage
	cacheKeys  []common.Hash
	begin      common.Hash
	end        common.Hash
	dbIter     db.Iterator
	idx        int
	addr       common.Address
}

func NewStorageIterator(obj *StateObject, begin, end common.Hash) *StorageIterator {
	cache := make(Storage)
	var cacheKeys []common.Hash
	for k, v := range obj.dirtyStorage {
		if bytes.Compare(k.Bytes(), begin.Bytes()) >= 0 && bytes.Compare(k.Bytes(), end.Bytes()) <= 0 {
			cache[k] = v
			cacheKeys = append(cacheKeys, k)
		}
	}
	// sort.Sort(cacheKeys)
	return &StorageIterator{
		cache:     cache,
		cacheKeys: cacheKeys,
		begin:     begin,
		end:       end,
		// todo use range query as a replacement
		dbIter:    obj.db.db.NewIterator(GetStorageKeyPrefix(obj.Address().Bytes())),
		addr:      obj.address,
		idx:       -1,
	}
}

func (iter *StorageIterator) Next() bool {
	if iter.idx < len(iter.cacheKeys) - 1 && len(iter.cacheKeys) > 0 {
		iter.idx += 1
		return true
	}
	for {
		fmt.Println("iter in db")
		if !iter.dbIter.Next() {
			fmt.Println("iter done")
			return false
		}
		tmp, b := SplitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())

		if b == false {
			return false
		}

		if bytes.Compare(tmp, iter.end.Bytes()) > 0 {
			return false
		}
		if _, exist := iter.cache[common.BytesToHash(tmp)]; exist == true {
			continue
		}
		return true
	}
}

func (iter *StorageIterator) Key() []byte {
	if iter.idx < len(iter.cacheKeys) {
		return iter.cacheKeys[iter.idx].Bytes()
	} else {
		tmp, _ := SplitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		return tmp
	}
}

func (iter *StorageIterator) Value() []byte {
	if iter.idx < len(iter.cacheKeys) {
		return iter.cache[iter.cacheKeys[iter.idx]]
	} else {
		return iter.dbIter.Value()
	}
}

func (iter *StorageIterator) Release() {
	iter.dbIter.Release()
	iter.idx = -1
}

