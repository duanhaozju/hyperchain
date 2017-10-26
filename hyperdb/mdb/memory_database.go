//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package mdb

import (
	"bytes"
	"hyperchain/common"
	hdb "hyperchain/hyperdb/db"
	"sort"
	"sync"
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

type KVs []*KV

func (kvs KVs) Len() int           { return len(kvs) }
func (kvs KVs) Swap(i, j int)      { kvs[i], kvs[j] = kvs[j], kvs[i] }
func (kvs KVs) Less(i, j int) bool { return kvs[i].key < kvs[j].key }

// MemDatabase a type of in-memory db implementation of DataBase.
type MemDatabase struct {
	kvs  KVs
	lock sync.RWMutex
	ns   string
}

func NewMemDatabase(namespace string) (*MemDatabase, error) {
	return &MemDatabase{
		kvs: nil,
		ns:  namespace,
	}, nil
}

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

func (db *MemDatabase) Set(key []byte, value []byte) {
	db.Put(key, value)
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	for _, kv := range db.kvs {
		if kv.key == common.Bytes2Hex(key) {
			return CopyBytes(kv.value), nil
		}
	}
	return nil, hdb.DB_NOT_FOUND
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for _, kv := range db.kvs {
		keys = append(keys, common.Hex2Bytes(kv.key))
	}
	return keys
}

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

func (db *MemDatabase) Namespace() string {
	return db.ns
}

func (db *MemDatabase) Close() {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.kvs = nil
}

func (db *MemDatabase) MakeSnapshot(string, []string) error {
	panic("not support")
}

type Iter struct {
	index int
	ptr   *MemDatabase
	start []byte
	limit []byte
}

func (db *MemDatabase) NewIterator(prefix []byte) hdb.Iterator {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	iter := &Iter{
		index: -1,
		ptr:   db,
		start: prefix,
		limit: limit,
	}

	return iter
}

func (db *MemDatabase) Scan(start, limit []byte) hdb.Iterator {
	iterator := &Iter{
		index: -1,
		ptr:   db,
		start: start,
		limit: limit,
	}
	return iterator
}

func (iter *Iter) Next() bool {
	for {
		iter.index += 1
		if iter.index >= len(iter.ptr.kvs) {
			iter.index -= 1
			return false
		}
		if !isLarger(iter.start, common.Hex2Bytes(iter.ptr.kvs[iter.index].key)) {
			continue
		} else if isSmaller(iter.limit, common.Hex2Bytes(iter.ptr.kvs[iter.index].key)) {
			return true
		} else {
			iter.index -= 1
			return false
		}
	}
}

func (iter *Iter) Prev() bool {
	for {
		iter.index -= 1
		if iter.index <= -1 {
			return false
		}
		if !isLarger(iter.start, common.Hex2Bytes(iter.ptr.kvs[iter.index].key)) {
			return false
		} else if isSmaller(iter.limit, common.Hex2Bytes(iter.ptr.kvs[iter.index].key)) {
			return true
		}
	}
}

func (iter *Iter) Key() []byte {
	return common.Hex2Bytes(iter.ptr.kvs[iter.index].key)
}

func (iter *Iter) Value() []byte {
	return CopyBytes(iter.ptr.kvs[iter.index].value)
}

func (iter *Iter) Release() {
	iter.index = -1
	iter.ptr = nil
}

func (iter *Iter) Error() error {
	return nil
}

func (iter *Iter) Seek(key []byte) bool {
	if iter.ptr == nil || iter.ptr.kvs == nil || len(iter.ptr.kvs) == 0 {
		return false
	}
	for i, kv := range iter.ptr.kvs {
		if isLarger(key, common.Hex2Bytes(kv.key)) {
			iter.index = i
			return true
		}
	}
	return false
}

type memBatch struct {
	db     *MemDatabase
	writes []*KV
	lock   sync.RWMutex
}

func (db *MemDatabase) NewBatch() hdb.Batch {
	return &memBatch{
		db: db,
	}
}

func (b *memBatch) Put(key, value []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.writes = append(b.writes, &KV{common.Bytes2Hex(key), CopyBytes(value)})
	return nil
}

func (b *memBatch) Delete(key []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.writes = append(b.writes, &KV{common.Bytes2Hex(key), nil})
	return nil
}

func (b *memBatch) Write() error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	b.db.lock.Lock()
	defer b.db.lock.Unlock()
	var updated bool
	for _, entry := range b.writes {
		if entry.value != nil {
			for idx, kv := range b.db.kvs {
				if kv.key == entry.key {
					b.db.kvs[idx].value = entry.value
					updated = true
					break
				}
			}
			if !updated {
				b.db.kvs = append(b.db.kvs, entry)
			}
			updated = false
		} else {
			for idx, kv := range b.db.kvs {
				if kv.key == entry.key {
					b.db.kvs = append(b.db.kvs[0:idx], b.db.kvs[idx+1:]...)
					break
				}
			}
		}
	}
	sort.Sort(b.db.kvs)
	b.writes = nil
	return nil
}
func (b *memBatch) Len() int {
	return len(b.writes)
}

func isLarger(start, elem []byte) bool {
	if start == nil || bytes.Compare(start, elem) <= 0 {
		return true
	}
	return false
}

func isSmaller(limit, elem []byte) bool {
	if limit == nil || bytes.Compare(limit, elem) > 0 {
		return true
	}
	return false
}
