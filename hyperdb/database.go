//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/syndtr/goleveldb/leveldb/iterator"

)

// the Database for LevelDB
// LDBDatabase implements the DataBase interface

type LDBDatabase struct {
	path string
	db   *leveldb.DB
}

// NewLDBDataBase new a LDBDatabase instance
// require a data filepath
// return *LDBDataBase and  will
// return an error with type of
// ErrCorrupted if corruption detected in the DB. Corrupted
// DB can be recovered with Recover function.
// the return *LDBDatabase is goruntine-safe
// the LDBDataBase instance must be close after use, by calling Close method
func NewLDBDataBase(filepath string) (*LDBDatabase, error) {
	db, err := leveldb.OpenFile(filepath, nil)
	return &LDBDatabase{
		path: filepath,
		db: db,
	}, err
}



// Put sets value for the given key, if the key exists, it will overwrite
// the value
func (self *LDBDatabase) Put(key []byte, value []byte) error {
	return self.db.Put(key, value, nil)
}

// Get gets value for the given key, it returns ErrNotFound if
// the Database does not contains the key
func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
	dat, err := self.db.Get(key, nil)
	return dat, err
}

// Delete deletes the value for the given key
func (self *LDBDatabase) Delete(key []byte) error {
	return self.db.Delete(key, nil)
}

// NewIterator returns a Iterator for traversing the database
func (self *LDBDatabase) NewIterator() iterator.Iterator {
	return self.db.NewIterator(nil, nil)
}
func (self *LDBDatabase) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return self.db.NewIterator(util.BytesPrefix(prefix), nil)
}

// Close close the LDBDataBase
func (self *LDBDatabase) Close() {
	self.db.Close()
}

// LDB returns *leveldb.DB instance
func (self *LDBDatabase) LDB() *leveldb.DB {
	return self.db
}

// NewBatch returns a Batch instance
// it allows batch-operation
func (db *LDBDatabase) NewBatch() Batch {
	return &ldbBatch{db: db.db, b: new(leveldb.Batch)}
}


// The Batch for LevelDB
// ldbBatch implements the Batch interface
type ldbBatch struct {
	db *leveldb.DB
	b  *leveldb.Batch
}

// Put put the key-value to ldbBatch
func (b *ldbBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	return nil
}

// Write write batch-operation to databse
func (b *ldbBatch) Write() error {
	return b.db.Write(b.b, nil)
}

// Put put the key-value to ldbBatch
func (b *ldbBatch) Delete(key []byte) error {
	b.b.Delete(key)
	return nil
}
