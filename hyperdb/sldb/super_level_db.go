//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package sldb

import (
	"bytes"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	pa "path/filepath"
	"time"
)

const (
	sldb_path = "dbConfig.sldb.dbpath"
	module    = "hyperdb/sldb"
)

type SuperLevelDB struct {
	path      string
	db        *leveldb.DB
	index     Index
	closed    chan bool
	logger    *logging.Logger
	conf      *common.Config
	namespace string
}

func NewSLDB(conf *common.Config, path string, namespace string) (*SuperLevelDB, error) {
	var filepath = path
	if conf != nil {
		if conf != nil {
			filepath = pa.Join(conf.GetString(sldb_path), path)
		}
	}

	db, err := leveldb.OpenFile(filepath, nil)
	log := common.GetLogger(conf.GetString(common.NAMESPACE), module)
	if err != nil {
		//panic(err.Error())
		log.Error(err.Error())
		return nil, err

	}
	index := NewKeyIndex(conf, "defaultNS", db, pa.Join(filepath, "index"))
	index.logger = log
	index.Init()
	index.conf = conf
	sldb := &SuperLevelDB{
		path:      filepath,
		db:        db,
		index:     index,
		closed:    make(chan bool),
		logger:    log,
		conf:      conf,
		namespace: namespace,
	}
	go sldb.dumpIndexByInterval(conf.GetDuration(sldb_index_dump_interval))
	return sldb, err
}

func SLDBPath(conf *common.Config) string {
	return conf.GetString(sldb_path)
}

//Put put key value data into the database.
func (sldb *SuperLevelDB) Put(key []byte, value []byte) error {
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
		err = db.DB_NOT_FOUND
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

func (sldb *SuperLevelDB) Scan(begin, end []byte) db.Iterator {
	return sldb.db.NewIterator(&util.Range{Start: begin, Limit: end}, nil)
}

//Destroy, clean the whole database,
//warning: bad performance if too many data in the db
func (sldb *SuperLevelDB) Destroy() error {
	return sldb.DestroyByRange(nil, nil)
}

func (sldb *SuperLevelDB) Namespace() string {
	return sldb.namespace
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
	sldb.index.Persist()
	sldb.db.Close()
	sldb.closed <- true
}

func (sldb *SuperLevelDB) LevelDB() *leveldb.DB {
	return sldb.db
}

func (sldb *SuperLevelDB) Index() Index {
	return sldb.index
}

func (sldb *SuperLevelDB) NewBatch() db.Batch {
	sldb.logger.Debugf("new super leveldb batch")
	slb := &superLdbBatch{
		batch:      new(leveldb.Batch),
		sldb:       sldb,
		indexBatch: new(leveldb.Batch),
	}
	return slb
}

//dumpIndexByInterval dump indexes by interval and after close.
func (sldb *SuperLevelDB) dumpIndexByInterval(du time.Duration) {
	if du.Hours() == 24 {
		tm := time.Now()
		sec := (24+4-tm.Hour())*3600 - tm.Minute()*60 - tm.Second() // bloom persist at 4:00 AM
		d, _ := time.ParseDuration(fmt.Sprintf("%ds", sec))
		select {
		case <-time.After(d):
			sldb.index.Persist()
		case <-sldb.closed:
			sldb.index.Persist()
			sldb.db.Close()
			return
		}
	}
	for {
		select {
		case <-time.After(du):
			sldb.index.Persist()
		case <-sldb.closed:
			sldb.index.Persist()
			sldb.db.Close()
			return
		}
	}
}

func (sb *SuperLevelDB) MakeSnapshot(backupPath string, fields []string) error {
	backupDb, err := NewSLDB(sb.conf, backupPath, sb.namespace)
	if err != nil {
		return err
	}
	defer backupDb.Close()

	snapshot, err := sb.db.GetSnapshot()
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

//batch related functions
type superLdbBatch struct {
	batch      *leveldb.Batch
	indexBatch *leveldb.Batch
	sldb       *SuperLevelDB
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
		sb.sldb.logger.Errorf("PersistKeyBatch error, %v", err)
	}
	err = sb.sldb.db.Write(sb.batch, nil)
	return err
}

func (sb *superLdbBatch) Reset() {
	sb.batch.Reset()
}

func (sb *superLdbBatch) Len() int {
	return sb.batch.Len()
}
