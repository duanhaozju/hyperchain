package hyperstate

import (
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"bytes"
)

type StorageIterator struct {
	cache      Storage
	cacheKeys  []common.Hash
	dbIter     db.Iterator
	idx        int
	addr       common.Address
	plexer     bool
}

// NewIterator returns an iterator for the latest snapshot of the
// underlying DB.
// The returned iterator is not safe for concurrent use, but it is safe to use
// multiple iterators concurrently, with each in a dedicated goroutine.
// It is also safe to use an iterator concurrently with modifying its
// underlying DB. The resultant key/value pairs are guaranteed to be
// consistent.
//
// Slice allows slicing the iterator to only contains keys in the given
// range. A nil Range.Start is treated as a key before all keys in the
// DB. And a nil Range.Limit is treated as a key after all keys in
// the DB.
//
// The iterator must be released after use, by calling Release method.
//
func NewStorageIterator(obj *StateObject, start, limit *common.Hash) *StorageIterator {
	cache := make(Storage)
	var cacheKeys []common.Hash

	var startBytes []byte
	var limitBytes []byte
	if start != nil {
		startBytes = start.Bytes()
	}

	if limit != nil {
		limitBytes = limit.Bytes()
	}

	for k, v := range obj.dirtyStorage {
		if isLarger(start, k.Bytes()) && isLess(limit, k.Bytes()) {
			cache[k] = v
			cacheKeys = append(cacheKeys, k)
		}
	}
	var dbIter db.Iterator
	if start == nil && limit == nil {
		dbIter = obj.db.db.NewIterator(GetStorageKeyPrefix(obj.address.Bytes()))
	} else {
		dbIter = obj.db.db.Scan(CompositeStorageKey(obj.address.Bytes(), startBytes), CompositeStorageKey(obj.address.Bytes(), limitBytes))
	}
	// sort.Sort(cacheKeys)
	return &StorageIterator{
		cache:     cache,
		cacheKeys: cacheKeys,
		addr:      obj.address,
		dbIter:    dbIter,
		idx:       -1,
	}
}

func (iter *StorageIterator) Next() bool {
	if iter.idx < len(iter.cacheKeys) - 1 && len(iter.cacheKeys) > 0 {
		iter.idx += 1
		return true
	}
	for {
		iter.plexer = true
		if !iter.dbIter.Next() {
			return false
		}
		tmp, b := SplitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		if b == false {
			return false
		}

		if _, exist := iter.cache[common.BytesToHash(tmp)]; exist == true {
			continue
		}
		return true
	}
}

func (iter *StorageIterator) Key() []byte {
	if !iter.plexer {
		return iter.cacheKeys[iter.idx].Bytes()
	} else {
		tmp, _ := SplitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		return tmp
	}
}

func (iter *StorageIterator) Value() []byte {
	if !iter.plexer {
		return iter.cache[iter.cacheKeys[iter.idx]]
	} else {
		return iter.dbIter.Value()
	}
}

func (iter *StorageIterator) Release() {
	iter.dbIter.Release()
	iter.idx = -1
}

func isLarger(start *common.Hash, key []byte) bool {
	if start == nil {
		return true
	}
	return bytes.Compare(key, start.Bytes()) >= 0
}

func isLess(limit *common.Hash, key []byte) bool {
	if limit == nil {
		return true
	}
	return bytes.Compare(key, limit.Bytes()) < 0
}

