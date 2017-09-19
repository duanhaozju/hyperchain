package state

import (
	"bytes"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"sort"
)

type sortable []common.Hash

func (array sortable) Len() int      { return len(array) }
func (array sortable) Swap(i, j int) { array[i], array[j] = array[j], array[i] }
func (array sortable) Less(i, j int) bool {
	return bytes.Compare(array[i].Bytes(), array[j].Bytes()) < 0
}

type MemIterator struct {
	cache     Storage
	cacheKeys sortable
	index     int
}

func NewMemIterator(kvs Storage) *MemIterator {
	cache := make(Storage)
	var key sortable
	for k, v := range kvs {
		cache[k] = v
		key = append(key, k)
	}
	sort.Sort(key)
	return &MemIterator{
		cache:     cache,
		cacheKeys: key,
		index:     -1,
	}
}

// Next moves the iterator to the next key/value pair.
// It returns whether the iterator is exhausted.
func (iter *MemIterator) Next() bool {
	iter.index += 1
	if iter.index >= len(iter.cacheKeys) {
		return false
	}
	return true
}

// Next moves the iterator to the prev key/value pair.
// It returns whether the iterator is exhausted.
func (iter *MemIterator) Prev() bool {
	iter.index -= 1
	if iter.index <= -1 {
		return false
	}
	return true
}

func (iter *MemIterator) Key() []byte {
	return iter.cacheKeys[iter.index].Bytes()
}

func (iter *MemIterator) Value() []byte {
	return iter.cache[iter.cacheKeys[iter.index]]
}

func (iter *MemIterator) Exist(key common.Hash) bool {
	_, exist := iter.cache[key]
	return exist
}

func (iter *MemIterator) Release() {
	iter.cache = nil
	iter.cacheKeys = nil
	iter.index = -1
	return
}

type StorageIterator struct {
	memIter   *MemIterator
	dbIter    db.Iterator
	addr      common.Address
	plexer    bool
	first     bool
	exhausted [2]bool
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
	var (
		cache      Storage = make(Storage)
		startBytes []byte
		limitBytes []byte
	)

	if start != nil {
		startBytes = start.Bytes()
	}

	if limit != nil {
		limitBytes = limit.Bytes()
	}

	for k, v := range obj.dirtyStorage {
		if isLarger(start, k.Bytes()) && isLess(limit, k.Bytes()) && v != nil {
			cache[k] = v
		}
	}
	memIter := NewMemIterator(cache)
	var dbIter db.Iterator
	if start == nil && limit == nil {
		dbIter = obj.db.db.NewIterator(GetStorageKeyPrefix(obj.address.Bytes()))
	} else {
		dbIter = obj.db.db.Scan(CompositeStorageKey(obj.address.Bytes(), startBytes), CompositeStorageKey(obj.address.Bytes(), limitBytes))
	}
	return &StorageIterator{
		memIter: memIter,
		addr:    obj.address,
		dbIter:  dbIter,
		first:   true,
	}
}

func (iter *StorageIterator) Next() bool {
	if !iter.exhausted[0] && !iter.memIter.Next() {
		iter.exhausted[0] = true
	}
	for !iter.exhausted[1] {
		if !iter.dbIter.Next() {
			iter.exhausted[1] = true
			break
		}
		if iter.memIter.Exist(common.BytesToRightPaddingHash(iter.dbIter.Key())) {
			continue
		}
		break
	}
	switch {
	case iter.exhausted[0] && iter.exhausted[1]:
		return false
	case !iter.exhausted[0] && iter.exhausted[1]:
		iter.plexer = false
	case iter.exhausted[0] && !iter.exhausted[1]:
		iter.plexer = true
	case !iter.exhausted[0] && !iter.exhausted[1]:
		if bytes.Compare(iter.memIter.Key(), iter.dbIter.Key()) <= 0 {
			iter.plexer = false
			iter.dbIter.Prev()
		} else {
			iter.plexer = true
			iter.memIter.Prev()
		}
	}
	return true
}

func (iter *StorageIterator) Key() []byte {
	if !iter.plexer {
		return iter.memIter.Key()
	} else {
		tmp, _ := SplitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		return tmp
	}
}

func (iter *StorageIterator) Value() []byte {
	if !iter.plexer {
		return iter.memIter.Value()
	} else {
		return iter.dbIter.Value()
	}
}

func (iter *StorageIterator) Release() {
	iter.dbIter.Release()
	iter.memIter.Release()
	iter.first = true
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
