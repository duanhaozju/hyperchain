//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hleveldb

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"hyperchain/common"
	hcom "hyperchain/hyperdb/common"
	"hyperchain/hyperdb/db"
	"path"
)

// the Database for LevelDB
// LDBDatabase implements the DataBase interface

type LDBDatabase struct {
	path      string
	db        *leveldb.DB
	conf      *common.Config
	namespace string
}

// NewLDBDataBase new a LDBDatabase instance
// require a data filepath
// return *LDBDataBase and  will
// return an error with type of
// ErrCorrupted if corruption detected in the DB. Corrupted
// DB can be recovered with Recover function.
// the return *LDBDatabase is goruntine-safe
// the LDBDataBase instance must be close after use, by calling Close method
func NewLDBDataBase(conf *common.Config, filepath string, namespace string) (*LDBDatabase, error) {
	db, err := leveldb.OpenFile(filepath, nil)
	t := db.NewIterator(nil, nil)
	t.Last()
	return &LDBDatabase{
		path:      filepath,
		db:        db,
		conf:      conf,
		namespace: namespace,
	}, err
}

func NewRawLDBDatabase(db *leveldb.DB, namespace string) *LDBDatabase {
	return &LDBDatabase{
		db:        db,
		namespace: namespace,
	}
}

func LDBDataBasePath(conf *common.Config) string {
	return conf.GetString(hcom.LEVEL_DB_ROOT_DIR)
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
	if err == leveldb.ErrNotFound {
		err = db.DB_NOT_FOUND
	}
	return dat, err
}

// Delete deletes the value for the given key
func (self *LDBDatabase) Delete(key []byte) error {
	return self.db.Delete(key, nil)
}

// NewIterator returns a Iterator for traversing the database
func (self *LDBDatabase) NewIterator(prefix []byte) db.Iterator {
	return self.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (self *LDBDatabase) Scan(begin, end []byte) db.Iterator {
	return self.db.NewIterator(&util.Range{Start: begin, Limit: end}, nil)
}

func (self *LDBDatabase) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return self.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (self *LDBDatabase) Namespace() string {
	return self.namespace
}

//Destroy, clean the whole database,
//warning: bad performance if to many data in the db
func (self *LDBDatabase) Destroy() error {
	return self.DestroyByRange(nil, nil)
}

//DestroyByRange, clean data which key in range [start, end)
func (self *LDBDatabase) DestroyByRange(start, end []byte) error {
	if bytes.Compare(start, end) > 0 {
		return errors.Errorf("start key: %v, is bigger than end key: %v", start, end)
	}
	it := self.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	for it.Next() {
		err := self.Delete(it.Key())
		if err != nil {
			return err
		}
	}
	return nil
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
func (db *LDBDatabase) NewBatch() db.Batch {
	return &ldbBatch{db: db.db, b: new(leveldb.Batch)}
}

func (db *LDBDatabase) MakeSnapshot(backupPath string, fields []string) error {
	p := path.Join(common.GetPath(db.namespace, LDBDataBasePath(db.conf)), backupPath)
	backupDb, err := NewLDBDataBase(db.conf, p, db.namespace)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	snapshot, err := db.db.GetSnapshot()
	if err != nil {
		return err
	}
	defer snapshot.Release()

	for _, field := range fields {
		iter := snapshot.NewIterator(util.BytesPrefix([]byte(field)), nil)
		for iter.Next() {
			if err := backupDb.Put(iter.Key(), iter.Value()); err != nil {
				iter.Release()
				return err
			}
		}
		iter.Release()
	}
	return nil
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

// Delete delete the key-value to ldbBatch
func (b *ldbBatch) Delete(key []byte) error {
	b.b.Delete(key)
	return nil
}

// Write write batch-operation to database
func (b *ldbBatch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *ldbBatch) Len() int {
	return b.b.Len()
}

type Iterator struct {
}
