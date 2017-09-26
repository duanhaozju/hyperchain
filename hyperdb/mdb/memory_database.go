//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package mdb

import (
	"bytes"
	"hyperchain/common"
	hdb "hyperchain/hyperdb/db"
	"sync"
)

//CopyBytes Copy and return []byte.
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

//MemDatabase a type of in-memory db implementation of DataBase.
type MemDatabase struct {
	key   []string
	value [][]byte
	lock  sync.RWMutex
	ns    string
}

func NewMemDatabase(namespace string) (*MemDatabase, error) {
	return &MemDatabase{
		key:   nil,
		value: nil,
		ns:    namespace,
	}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.key = append(db.key, common.Bytes2Hex(key))
	db.value = append(db.value, CopyBytes(value))
	return nil
}

func (db *MemDatabase) Set(key []byte, value []byte) {
	db.Put(key, value)
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	for idx, v := range db.key {
		if v == common.Bytes2Hex(key) {
			return CopyBytes(db.value[idx]), nil
		}
	}
	return nil, hdb.DB_NOT_FOUND
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for _, key := range db.key {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	for k, v := range db.key {
		if v == common.Bytes2Hex(key) {
			db.key = append(db.key[0:k], db.key[k+1:]...)
			db.value = append(db.value[0:k], db.value[k+1:]...)
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
	db.key = nil
	db.value = nil
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
		if iter.index >= len(iter.ptr.key) {
			iter.index -= 1
			return false
		}

		if InRange(iter.start, iter.limit, common.Hex2Bytes(iter.ptr.key[iter.index])) {
			break
		}
	}
	return true
}

func (iter *Iter) Prev() bool {
	for {
		iter.index -= 1
		if iter.index <= -1 {
			return false
		}
		if InRange(iter.start, iter.limit, common.Hex2Bytes(iter.ptr.key[iter.index])) {
			break
		}
	}
	return true
}

func (iter *Iter) Key() []byte {
	return common.Hex2Bytes(iter.ptr.key[iter.index])
}

func (iter *Iter) Value() []byte {
	return iter.ptr.value[iter.index]
}

func (iter *Iter) Release() {
	iter.index = -1
	iter.ptr = nil
}

func (iter *Iter) Error() error {
	return nil
}

func (iter *Iter) Seek(key []byte) bool {
	panic("not support")
}

//-- mem db的batch操作
type kv struct{ k, v []byte }

type memBatch struct {
	db     *MemDatabase
	writes []kv
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

	b.writes = append(b.writes, kv{CopyBytes(key), CopyBytes(value)})
	return nil
}

func (b *memBatch) Delete(key []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.writes = append(b.writes, kv{CopyBytes(key), nil})
	return nil
}

func (b *memBatch) Write() error {
	b.lock.RLock()
	defer b.lock.RUnlock()

	b.db.lock.Lock()
	defer b.db.lock.Unlock()
	var isUpdate bool
	for _, kv := range b.writes {
		if kv.v != nil {
			for idx, k := range b.db.key {
				if k == common.Bytes2Hex(kv.k) {
					b.db.value[idx] = kv.v
					isUpdate = true
					break
				}
			}
			if !isUpdate {
				b.db.key = append(b.db.key, common.Bytes2Hex(kv.k))
				b.db.value = append(b.db.value, kv.v)
			}
			isUpdate = false
		} else {
			for idx, k := range b.db.key {
				if k == common.Bytes2Hex(kv.k) {
					b.db.key = append(b.db.key[0:idx], b.db.key[idx+1:len(b.db.key)]...)
					b.db.value = append(b.db.value[0:idx], b.db.value[idx+1:len(b.db.value)]...)
					break
				}
			}
		}
	}
	b.writes = nil
	return nil
}
func (b *memBatch) Len() int {
	return len(b.writes)
}

func InRange(start []byte, limit []byte, elem []byte) bool {
	if (start == nil || bytes.Compare(start, elem) <= 0) && (limit == nil || bytes.Compare(limit, elem) > 0) {
		return true
	}
	return false
}
