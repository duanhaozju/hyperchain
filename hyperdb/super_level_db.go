//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type SuperLevelDB struct {
	path  string
	db    *leveldb.DB
	index Index
}

//New new super LevelDB
func NewSLDB(filepath string) (*SuperLevelDB, error) {
	db, err := leveldb.OpenFile(filepath, nil)
	index := NewKeyIndex("defaultNS", db)

	return &SuperLevelDB{
		path: filepath,
		db:   db,
		index:   index,
	}, err
}

//Put put key value data into the database.
func (sldb  *SuperLevelDB) Put(key []byte, value []byte) error {
	err := sldb.db.Put(key, value, nil)
	if err == nil {
		sldb.index.AddIndexForKey(key)
	}
	return err
}

//Get fetch data for specify key from db.
func (sldb *SuperLevelDB) Get(key []byte) ([]byte, error) {
	var data []byte
	var err error
	if sldb.index.MayContains(key) {
		data, err = sldb.db.Get(key, nil)
	}else {
		err = leveldb.ErrNotFound
	}
	return data, err
}

// Delete deletes the value for the given key.
func (sldb *SuperLevelDB) Delete(key []byte) error {
	return sldb.db.Delete(key, nil)
}

// NewIterator returns a Iterator for traversing the database
func (sldb *SuperLevelDB) NewIterator(prefix []byte) Iterator {
	return sldb.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (sldb *SuperLevelDB) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return sldb.db.NewIterator(util.BytesPrefix(prefix), nil)
}

//Destroy, clean the whole database,
//warning: bad performance if too many data in the db
func (sldb *SuperLevelDB) Destroy() error {
	return sldb.DestroyByRange(nil, nil)
}

//DestroyByRange, clean data which key in range [start, end)
func (sldb *SuperLevelDB) DestroyByRange(start, end []byte) error {
	if bytes.Compare(start, end) > 0 {
		return errors.Errorf("start key: %v, is bigger than end key: %v", start, end)
	}
	it := sldb.db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	for it.Next() {
		err := sldb.Delete(it.Key())
		if err != nil {
			return err
		}
	}
	return nil
}

func (sldb *SuperLevelDB) Close() {
	sldb.db.Close()
}

func (sldb *SuperLevelDB) LevelDB() *leveldb.DB  {
	return sldb.db
}

func (sldb *SuperLevelDB) Index() Index {
	return sldb.index
}

func (db *SuperLevelDB) NewBatch() Batch {
	log.Debugf("new super leveldb batch")
	return &superLdbBatch{
		batch: new(leveldb.Batch),
		sldb:db,
		}
}

type superLdbBatch struct {
	batch *leveldb.Batch
	sldb  *SuperLevelDB
}

func (sb *superLdbBatch) Put(key, value []byte) error {
	sb.batch.Put(key, value)
	sb.sldb.index.AddIndexForKey(key)
	return nil
}

func (b *superLdbBatch) Delete(key []byte) error {
	b.batch.Delete(key)
	return nil
}

func (b *superLdbBatch) Write() error {
	err := b.sldb.db.Write(b.batch, nil)
	return err
}
