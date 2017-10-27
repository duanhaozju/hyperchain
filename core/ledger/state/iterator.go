// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package state

import (
	"bytes"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"sort"
)

type sortable []common.Hash

func (array sortable) Len() int      { return len(array) }
func (array sortable) Swap(i, j int) { array[i], array[j] = array[j], array[i] }
func (array sortable) Less(i, j int) bool {
	return bytes.Compare(array[i].Bytes(), array[j].Bytes()) < 0
}

// MemIterator represents a iterator to traverse dirty kv pairs in memory.
type MemIterator struct {
	cache     Storage
	cacheKeys sortable
	deleted   Storage
	index     int
}

func NewMemIterator(kvs, deleted Storage) *MemIterator {
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
		deleted:   deleted,
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

// Key returns current key iterator points to.
func (iter *MemIterator) Key() []byte {
	return iter.cacheKeys[iter.index].Bytes()
}

// Key returns current value iterator points to.
func (iter *MemIterator) Value() []byte {
	return iter.cache[iter.cacheKeys[iter.index]]
}

// Exist checks whether the entry is existed in dirty storage by given key.
func (iter *MemIterator) Exist(key common.Hash) bool {
	_, exist := iter.cache[key]
	if !exist {
		_, exist = iter.deleted[key]
	}
	return exist
}

// Release releases system resource and clear mem iterator status.
func (iter *MemIterator) Release() {
	iter.cache = nil
	iter.cacheKeys = nil
	iter.index = -1
	return
}

// StorageIterator represents a iterator to traverse kv pairs in database and memory with order.
type StorageIterator struct {
	memIter   *MemIterator   // memory iterator
	dbIter    db.Iterator    // database raw iterator
	addr      common.Address // contract address
	plexer    bool           // indicates which iterator value should be returned.
	exhausted [2]bool        // indicates whether iterator has exhausted.
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
		mem        Storage = make(Storage)
		deleted    Storage = make(Storage)
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
			mem[k] = v
		}
		if v == nil {
			deleted[k] = []byte{0x1}
		}
	}
	memIter := NewMemIterator(mem, deleted)
	var dbIter db.Iterator
	if start == nil && limit == nil {
		dbIter = obj.db.db.NewIterator(getStorageKeyPrefix(obj.address.Bytes()))
	} else {
		var (
			dbStart []byte = compositeStorageKey(obj.address.Bytes(), startBytes)
			dbLimit []byte = compositeStorageKey(obj.address.Bytes(), limitBytes)
		)
		if limit == nil {
			dbLimit = compositeStorageKey(obj.address.Bytes(), common.FullHash().Bytes())
		}
		dbIter = obj.db.db.Scan(dbStart, dbLimit)
	}
	return &StorageIterator{
		memIter: memIter,
		addr:    obj.address,
		dbIter:  dbIter,
	}
}

// Next moves the iterator to the next key/value pair.
// It returns whether the iterator is exhausted.
func (iter *StorageIterator) Next() bool {
	if !iter.exhausted[0] && !iter.memIter.Next() {
		iter.exhausted[0] = true
	}
	for !iter.exhausted[1] {
		if !iter.dbIter.Next() {
			iter.exhausted[1] = true
			break
		}
		actualKey, match := splitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		if !match {
			panic("expect to match with contract storage key pattern")
		}
		if iter.memIter.Exist(common.BytesToRightPaddingHash(actualKey)) {
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
		actualKey, match := splitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		if !match {
			panic("expect to match with contract storage key pattern")
		}
		if bytes.Compare(iter.memIter.Key(), actualKey) <= 0 {
			// move db iterator back
			iter.plexer = false
			iter.dbIter.Prev()
		} else {
			// move mem iterator back
			iter.plexer = true
			iter.memIter.Prev()
		}
	}
	return true
}

// Key returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change on the next call to any 'seeks method'.
func (iter *StorageIterator) Key() []byte {
	if !iter.plexer {
		return iter.memIter.Key()
	} else {
		tmp, _ := splitCompositeStorageKey(iter.addr.Bytes(), iter.dbIter.Key())
		return tmp
	}
}

// Value returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change on the next call to any 'seeks method'.
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
