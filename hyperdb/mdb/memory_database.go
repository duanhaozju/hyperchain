//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package mdb

import (
	"bytes"
	"sort"
	"sync"

	"github.com/hyperchain/hyperchain/common"
	hdb "github.com/hyperchain/hyperchain/hyperdb/db"
)

// CopyBytes Copy and return []byte.
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

type KV struct {
	key   string
	value []byte
}

// KVS implements sort functionality.
type KVs []*KV

func (kvs KVs) Len() int           { return len(kvs) }
func (kvs KVs) Swap(i, j int)      { kvs[i], kvs[j] = kvs[j], kvs[i] }
func (kvs KVs) Less(i, j int) bool { return kvs[i].key < kvs[j].key }

// MemDatabase is a type of in-memory db implementation for storage.
type MemDatabase struct {
	kvs  KVs
	lock sync.RWMutex
	ns   string
}

// NewMemDatabase creates a new memory db by namespace.
func NewMemDatabase(namespace string) (*MemDatabase, error) {
	return &MemDatabase{
		kvs: nil,
		ns:  namespace,
	}, nil
}

// Put inserts a K/V pair to database.
func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// check existence
	for _, kv := range db.kvs {
		if common.Bytes2Hex(key) == kv.key {
			kv.value = CopyBytes(value)
			return nil
		}
	}

	db.kvs = append(db.kvs, &KV{
		key:   common.Bytes2Hex(key),
		value: CopyBytes(value),
	})
	sort.Sort(db.kvs)
	return nil
}

// Get gets a key's value from the database.
func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	for _, kv := range db.kvs {
		if kv.key == common.Bytes2Hex(key) {
			return CopyBytes(kv.value), nil
		}
	}
	return nil, hdb.ErrDbNotFound
}

// Delete removes a K/V pair from the database.
func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	for idx, kv := range db.kvs {
		if kv.key == common.Bytes2Hex(key) {
			db.kvs = append(db.kvs[0:idx], db.kvs[idx+1:]...)
			sort.Sort(db.kvs)
		}
	}
	return nil
}

// Close cleans the whole in-memory db.
func (db *MemDatabase) Close() {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.kvs = nil
}

// MakeSnapshot creates a new snapshot for the database.
func (db *MemDatabase) MakeSnapshot(string, []string) error {
	// TODO: returns an error instead?
	//panic("not support")

	return hdb.ErrNotSupport
}

// Namespace returns database's namespace literal.
func (db *MemDatabase) Namespace() string {
	return db.ns
}

// Set inserts a K/V pair to database.
func (db *MemDatabase) Set(key []byte, value []byte) {
	db.Put(key, value)
}

// Keys returns all keys stored in the database.
func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for _, kv := range db.kvs {
		keys = append(keys, common.Hex2Bytes(kv.key))
	}
	return keys
}

// memBatch implementes the Batch interface.
type memBatch struct {
	db     *MemDatabase
	writes KVs
	lock   sync.RWMutex
}

// NewBatch returns a Batch instance.
func (db *MemDatabase) NewBatch() hdb.Batch {
	return &memBatch{
		db:     db,
		writes: nil,
	}
}

// Put appends 'put operation' of the given K/V pair to the batch.
func (b *memBatch) Put(key, value []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.writes = append(b.writes, &KV{common.Bytes2Hex(key), CopyBytes(value)})
	return nil
}

// Delete appends 'delete operation' of the given key to the batch.
func (b *memBatch) Delete(key []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.writes = append(b.writes, &KV{common.Bytes2Hex(key), nil})
	return nil
}

// Write apply the given batch to the DB.
func (b *memBatch) Write() error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	// TODO: SIEG
	// Sort b.writes first,
	// and record a cursor of ranged b.db.kvs,
	// then range from the cursor position to avoid redundant range.
	var updated bool
	var cursor int = 0

	sort.Sort(b.writes)

	for _, entry := range b.writes {
		if entry.value != nil {
			for idx, kv := range b.db.kvs[cursor:] {
				if kv.key == entry.key {
					b.db.kvs[cursor+idx].value = entry.value
					updated = true
					// update cursor to the next position
					cursor = cursor + idx + 1
					break
				}
			}
			if !updated {
				b.db.kvs = append(b.db.kvs, entry)
				// reset cursor if key not found
				cursor = 0
			}
			updated = false
		} else {
			for idx, kv := range b.db.kvs[cursor:] {
				if kv.key == entry.key {
					b.db.kvs = append(b.db.kvs[0:cursor+idx], b.db.kvs[cursor+idx+1:]...)
					// update cursor to the next position
					cursor = cursor + idx + 1
					break
				}
			}
		}
	}
	sort.Sort(b.db.kvs)
	b.writes = nil
	return nil
}

// Len returns number of records in the batch.
func (b *memBatch) Len() int {
	return len(b.writes)
}

// Iterator implements the Iterator interface.
type Iterator struct {
	index int
	ptr   *MemDatabase
	begin []byte
	end   []byte
}

// NewIterator returns a Iterator for traversing the database.
func (db *MemDatabase) NewIterator(prefix []byte) hdb.Iterator {
	var end []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, prefix)
			end[i] = c + 1
			break
		}
	}
	iter := &Iterator{
		index: -1,
		ptr:   db,
		begin: prefix,
		end:   end,
	}

	return iter
}

// Scan is MemDatabase's Scan method which scans objects in range [begin, end).
func (db *MemDatabase) Scan(begin, end []byte) hdb.Iterator {
	iterator := &Iterator{
		index: -1,
		ptr:   db,
		begin: begin,
		end:   end,
	}
	return iterator
}

func (iter *Iterator) Key() []byte {
	return common.Hex2Bytes(iter.ptr.kvs[iter.index].key)
}

func (iter *Iterator) Value() []byte {
	return CopyBytes(iter.ptr.kvs[iter.index].value)
}

// Seek moves the iterator to the first key/value pair whose key is greater
// than or equal to the given key.
// It returns whether such pair exist.
func (iter *Iterator) Seek(key []byte) bool {
	if iter.ptr == nil || iter.ptr.kvs == nil || len(iter.ptr.kvs) == 0 {
		return false
	}
	for i, kv := range iter.ptr.kvs {
		// return true if kv.key is greater than or equal to the given key
		if isGE(common.Hex2Bytes(kv.key), key) {
			iter.index = i
			return true
		}
	}
	return false
}

// Next moves the iterator to the next key/value pair.
// It returns whether the iterator is exhausted.
func (iter *Iterator) Next() bool {
	for {
		iter.index += 1
		if iter.index >= len(iter.ptr.kvs) {
			iter.index -= 1
			return false
		}
		if isLT(common.Hex2Bytes(iter.ptr.kvs[iter.index].key), iter.begin) {
			// jump to the next loop if the key is less than the iter.begin
			continue
		} else if isLT(common.Hex2Bytes(iter.ptr.kvs[iter.index].key), iter.end) {
			// return true if a key is less than the iter.end
			return true
		} else {
			// reset iterator's index and return false if the iterator is exhausted
			iter.index -= 1
			return false
		}
	}
}

// Prev moves the iterator to the previous key/value pair.
// It returns whether the iterator is exhausted.
func (iter *Iterator) Prev() bool {
	for {
		iter.index -= 1
		if iter.index <= -1 {
			return false
		}
		if isLT(common.Hex2Bytes(iter.ptr.kvs[iter.index].key), iter.begin) {
			// return false if a key is less than the iter.begin
			return false
		} else if isLT(common.Hex2Bytes(iter.ptr.kvs[iter.index].key), iter.end) {
			// return true if a key is less than the iter.end
			return true
		}
	}
}

func (iter *Iterator) Error() error {
	return nil
}

func (iter *Iterator) Release() {
	iter.index = -1
	iter.ptr = nil
	iter.begin = nil
	iter.end = nil
}

// isGE returns whether elem is greater than or equal to the given key.
func isGE(elem, key []byte) bool {
	if key == nil || bytes.Compare(elem, key) >= 0 {
		return true
	}
	return false
}

// isLT returns whether elem is less than the given key.
func isLT(elem, key []byte) bool {
	if key == nil || bytes.Compare(elem, key) < 0 {
		return true
	}
	return false
}
