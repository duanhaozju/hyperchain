//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package mdb

import (
	"errors"
	"sync"
	"hyperchain/hyperdb/db"
)

//-- 拷贝并返回一个[]byte
func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)
	return
}

//-- 用内存模拟实现一个mem db
//-- 实现了DataBase接口
type MemDatabase struct {
	key []string
	value [][]byte
	lock sync.RWMutex
}

func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{
		key: nil,
		value: nil,
	}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.key = append(db.key, string(key))
	db.value = append(db.value, value)
	return nil
}

func (db *MemDatabase) Set(key []byte, value []byte) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.Put(key, value)
}

func (db *MemDatabase) Get(key []byte) ([]byte, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	for k, v := range db.key {
		if v == string(key) {
			return db.value[k], nil
		}
	}
	return nil, errors.New("not found")
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
		if v == string(key) {
			db.key = append(db.key[0:k], db.key[k+1: len(db.key)]...)
			db.value = append(db.value[0:k], db.value[k+1: len(db.value)]...)
		}
	}
	return nil
}

func (db *MemDatabase) Close() {}

type Iter struct {
	index int
	ptr *MemDatabase
	str string
}

func (db *MemDatabase) NewIterator(str []byte) db.Iterator {
	var iter Iter
	iter.index = -1
	iter.ptr = db
	iter.str = string(str)
	return &iter
}

func (iter *Iter) Next() bool{
	for {
		iter.index += 1
		if iter.index >= len(iter.ptr.key) {
			iter.index -= 1
			return false
		}
		if iter.str == iter.ptr.key[iter.index][:len(iter.str)] {
			break;
		}
	}
	return true
}

func (iter *Iter) Key() []byte {
	return []byte(iter.ptr.key[iter.index])
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
	return true
}

//-- mem db的batch操作
type kv struct{ k, v []byte }

type memBatch struct {
	db     *MemDatabase
	writes []kv
	lock   sync.RWMutex
}

func (db *MemDatabase) NewBatch() db.Batch {
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

	for _, kv := range b.writes {
		if kv.v != nil {
			b.db.key = append(b.db.key, string(kv.k))
			b.db.value = append(b.db.value, kv.v)
		} else {
			for k, v := range b.db.key {
				if v == string(kv.k) {
					b.db.key = append(b.db.key[0:k], b.db.key[k+1: len(b.db.key)]...)
					b.db.value = append(b.db.value[0:k], b.db.value[k+1: len(b.db.value)]...)
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
