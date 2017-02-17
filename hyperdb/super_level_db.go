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
	filter := bloom.New(10000 * 10000, 3) //TODO: make this configurable

	//c := cache.New(5*time.Minute, 30 * time.Second)

	return &SuperLevelDB{
		path: filepath,
		db:   db,
		bf:   filter,
		//cache:c,
	}, err
}
//var y int = 0
func (sldb  *SuperLevelDB) Put(key []byte, value []byte) error {
	//t1 := time.Now()
	err := sldb.db.Put(key, value, nil)
	if err == nil {
		sldb.bf.Add(key)
		//sldb.cache.Add(string(key), value, 5 * time.Minute)
	}
	//y ++
	//log.Warningf("%d put used %v", y, time.Now().Sub(t1))
	return err
}

var getCount int = 0
func (sldb *SuperLevelDB) Get(key []byte) ([]byte, error) {
	var data []byte
	var err error
	//t1 := time.Now()
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
	//getCount ++
	//log.Warningf("%d get used %v", getCount, time.Now().Sub(t1))
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

func (db *SuperLevelDB) NewBatch() Batch {
	log.Debugf("new super level db batch")

	return &superLdbBatch{
		batch: new(leveldb.Batch),
		sldb:db,
		//batchMap:make(map[string]interface{}),
		}
}

type superLdbBatch struct {
	batch *leveldb.Batch
	sldb  *SuperLevelDB
	//batchMap    map[string]interface{}
	//bl sync.Mutex
}

func (b *superLdbBatch) Put(key, value []byte) error {
	b.batch.Put(key, value)
	b.sldb.bf.Add(key)
	//b.bl.Lock()
	//defer b.bl.Unlock()
	//value1 := make([]byte, len(value))
	//copy(value1, value)
	//b.batchMap[string(key)]=value1
	return nil
}

func (b *superLdbBatch) Delete(key []byte) error {
	b.batch.Delete(key)
	//b.sldb.cache.Delete(string(key))
	return nil
}
//var x int = 0
func (b *superLdbBatch) Write() error {
	//b.bl.Lock()
	//for k, v := range b.batchMap{
	//	b.sldb.cache.Add(k, v, 5 * time.Minute)
	//}
	//b.bl.Unlock()
	//x ++
	//t1 := time.Now()
	err := b.sldb.db.Write(b.batch, nil)
	//log.Warningf("%d batch write time used: %v", x,  time.Now().Sub(t1))
	return err
}
