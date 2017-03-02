//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package sldb

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
	"fmt"
	"path/filepath"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
)

var log *logging.Logger

func init() {
	log = logging.MustGetLogger("hyperdb/sldb")
}

const (
	SLDB_PATH = "dbConfig.sldb.dbpath"
)

type SuperLevelDB struct {
	path   string
	db     *leveldb.DB
	index  Index
	closed chan bool
}

func NewSLDB(conf *common.Config) (*SuperLevelDB, error) {
	dbpath := conf.GetString(SLDB_PATH)
	db, err := leveldb.OpenFile(conf.GetString(SLDB_PATH), nil)
	if err != nil {
		panic(err.Error())
	}
	index := NewKeyIndex(conf, "defaultNS", db, filepath.Join(dbpath, "index", "index.bloom.dat"))
	index.conf = conf
	sldb := &SuperLevelDB{
		path: dbpath,
		db:   db,
		index:   index,
		closed: make(chan bool),
	}
	go sldb.dumpIndexByInterval(conf.GetDuration(SLDB_INDEX_DUMP_INTERVAL))
	return sldb, err
}

//Put put key value data into the database.
func (sldb  *SuperLevelDB) Put(key []byte, value []byte) error {
	sldb.index.AddAndPersistIndexForKey(key)
	return sldb.db.Put(key, value, nil)
}

//Get fetch data for specify key from db.
func (sldb *SuperLevelDB) Get(key []byte) ([]byte, error) {
	var data []byte
	var err error
	if sldb.index.MayContains(key) {
		data, err = sldb.db.Get(key, nil)
	} else {
		err = leveldb.ErrNotFound
	}
	return data, err
}

// Delete deletes the value for the given key.
func (sldb *SuperLevelDB) Delete(key []byte) error {
	return sldb.db.Delete(key, nil)
}

// NewIterator returns a Iterator for traversing the database
func (sldb *SuperLevelDB) NewIterator(prefix []byte) db.Iterator {
	return sldb.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (sldb *SuperLevelDB) NewIteratorWithPrefix(prefix []byte) db.Iterator {
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
	sldb.closed <- true
	sldb.db.Close()
}

func (sldb *SuperLevelDB) LevelDB() *leveldb.DB {
	return sldb.db
}

func (sldb *SuperLevelDB) Index() Index {
	return sldb.index
}

func (db *SuperLevelDB) NewBatch() db.Batch {
	log.Debugf("new super leveldb batch")
	slb := &superLdbBatch{
		batch: new(leveldb.Batch),
		sldb:db,
		indexBatch: new(leveldb.Batch),
	}
	return slb
}

//dumpIndexByInterval dump indexes by interval and after close.
func (db *SuperLevelDB) dumpIndexByInterval(du time.Duration) {
	if du.Hours() == 24 {
		tm := time.Now()
		sec := (24 + 4 - tm.Hour()) * 3600 - tm.Minute() * 60 - tm.Second()// bloom persist at 4:00 AM
		d, _ := time.ParseDuration(fmt.Sprintf("%ds", sec))
		time.Sleep(d)
		db.index.Persist()// first time persist
	}
	for {
		select {
		case <-time.After(du):
			db.index.Persist()
		case <-db.closed:
			db.index.Persist()
			return
		}
	}
}

//batch related functions
type superLdbBatch struct {
	batch *leveldb.Batch
	indexBatch *leveldb.Batch
	sldb  *SuperLevelDB
}

func (sb *superLdbBatch) Put(key, value []byte) error {
	sb.sldb.index.AddIndexForKey(key, sb.indexBatch)
	sb.batch.Put(key, value)
	return nil
}

func (sb *superLdbBatch) Delete(key []byte) error {
	sb.batch.Delete(key)
	return nil
}

func (sb *superLdbBatch) Write() error {
	err := sb.sldb.db.Write(sb.indexBatch, nil)
	if err != nil {
		log.Errorf("PersistKeyBatch error, %v", err)
	}
	err = sb.sldb.db.Write(sb.batch, nil)
	return err
}

func (b *superLdbBatch) Reset() {
	b.batch.Reset()
}

func (b *superLdbBatch) Len() int {
	return b.batch.Len()
}