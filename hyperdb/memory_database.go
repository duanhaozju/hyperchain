//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"sync"
	"errors"
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
	db   map[string][]byte
	lock sync.RWMutex
}

func NewMemDatabase() (*MemDatabase, error) {
	return &MemDatabase{
		db: make(map[string][]byte),
	}, nil
}

func (db *MemDatabase) Put(key []byte, value []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.db[string(key)] = CopyBytes(value)
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

	if entry, ok := db.db[string(key)]; ok {
		return entry, nil
	}
	return nil, errors.New("not found")
}

func (db *MemDatabase) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := [][]byte{}
	for key, _ := range db.db {
		keys = append(keys, []byte(key))
	}
	return keys
}


func (db *MemDatabase) Delete(key []byte) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.db, string(key))
	return nil
}


func (db *MemDatabase) Close() {}

func (db *MemDatabase) NewBatch() Batch {
	return &memBatch{db: db}
}


//-- mem db的batch操作
type kv struct{ k, v []byte }

type memBatch struct {
	db     *MemDatabase
	writes []kv
	lock   sync.RWMutex
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
		b.db.db[string(kv.k)] = kv.v
	}
	return nil
}

