//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/willf/bloom"

	//"time"
	"github.com/go-cache"
	////"github.com/syndtr/goleveldb/leveldb/cache"
	//"github.com/syndtr/goleveldb/leveldb/cache"
	"time"
)

type SuperLevelDB struct {
	path  string
	db    *leveldb.DB
	bf    *bloom.BloomFilter //TODO: we may need index for keys stored in the db
	//cache *cache.Cache
}

//New new super LevelDB
func NewSLDB(filepath string) (*SuperLevelDB, error) {
	db, err := leveldb.OpenFile(filepath, nil)
	filter := bloom.New(1000 * 10000, 5) //TODO: make this configurable
	c := cache.New(5 * time.Minute, 30 * time.Second)
	return &SuperLevelDB{
		path: filepath,
		db:   db,
		bf:   filter,
		//cache:c,
	}, err
}

func (sldb  *SuperLevelDB) Put(key []byte, value []byte) error {
	err := sldb.db.Put(key, value, nil)
	if err == nil {
		sldb.bf.Add(key)
		//sldb.cache.Add(string(key), value, 5 * time.Minute)
	}
	return err
}

func (sldb *SuperLevelDB) Get(key []byte) ([]byte, error) {
	var data []byte
	var err error
	if sldb.bf.Test(key) {
		//if val, ok := sldb.cache.Get(string(key)); ok {
		//	if dat, ok := (val).([]byte); ok {
		//		return dat, nil
		//	}
		//}
		data, err = sldb.db.Get(key, nil)
	}else {
		err = leveldb.ErrNotFound
	}
	return data, err
}

// Delete deletes the value for the given key
func (sldb *SuperLevelDB) Delete(key []byte) error {
	//sldb.cache.Delete(string(key))
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
//warning: bad performance if to many data in the db
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

func (sldb *SuperLevelDB) LDB() *leveldb.DB {
	return sldb.db
}


//////////////////////////////////////////////////////////////
/////// ba batch operation////////////////////////////////////

func (db *SuperLevelDB) NewBatch() Batch {
	log.Warning("new super level db batch")

	return &superLdbBatch{
		batch: new(leveldb.Batch),
		sldb:db,
		data:make(map[string][]byte, 20000),//todo: fix 2000
	}
}

type superLdbBatch struct {
	batch *leveldb.Batch
	sldb  *SuperLevelDB
	data  map[string][]byte
}

func (b *superLdbBatch) Put(key, value []byte) error {
	b.batch.Put(key, value)
	b.sldb.bf.Add(key)
	//b.data[string(key)] = value
	return nil
}

func (b *superLdbBatch) Delete(key []byte) error {
	b.batch.Delete(key)
	//delete(b.data, string(key))
	return nil
}

func (b *superLdbBatch) Write() error {
	//for k, v := range b.data {
	//	//log.Warningf("k: %v, v: %v", k, len(v))
	//	b.sldb.cache.Add(k, v, 5 * time.Minute) //TODO: fix this
	//}
	return b.sldb.db.Write(b.batch, nil)
}
