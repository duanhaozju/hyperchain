package hleveldb

import (
	"errors"
	Lru "github.com/hashicorp/golang-lru"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("hyperdb/leveldb_lru")
}

//if key-value value is 1kb lruSize is 1GB
var lruSize = 1024 * 1024

type levelLruDatabase struct {
	leveldb  *LDBDatabase
	cache    *Lru.Cache
	dbStatus bool
}

var (
	leveldbPath = "./build/leveldb"
)

func NewLevelLruDB(conf *common.Config, filepath string, ns string) (*levelLruDatabase, error) {

	leveldb, err := NewLDBDataBase(conf, filepath, ns)
	if err != nil {
		log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n", leveldbPath, err.Error())
		return nil, err
	}

	cache, err := Lru.New(lruSize)
	if err != nil {
		log.Noticef("Lru.New(%v) fail. err is %v. \n", lruSize, err.Error())
		return nil, err
	}

	return &levelLruDatabase{leveldb: leveldb, cache: cache, dbStatus: true}, nil
}

func NewLevelLruDBWithP(conf *common.Config, filepath string, ns string) (*levelLruDatabase, error) {

	leveldb, err := NewLDBDataBase(conf, filepath, ns)
	if err != nil {
		log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n", leveldbPath, err.Error())
		return nil, err
	}

	cache, err := Lru.New(lruSize)
	if err != nil {
		log.Noticef("Lru.New(%v) fail. err is %v. \n", lruSize, err.Error())
		return nil, err
	}

	return &levelLruDatabase{leveldb: leveldb, cache: cache, dbStatus: true}, nil
}

func (lldb *levelLruDatabase) Put(key []byte, value []byte) error {

	if err := lldb.check(); err != nil {
		return err
	}

	lldb.cache.Add(string(key), value)

	go lldb.leveldb.Put(key, value)

	return nil
}

func (lldb *levelLruDatabase) Get(key []byte) ([]byte, error) {

	if err := lldb.check(); err != nil {
		return nil, err
	}

	data1, status := lldb.cache.Get(string(key))

	if status {
		return db.Bytes(data1), nil
	}

	data, err := lldb.leveldb.Get(key)

	if data != nil {
		lldb.cache.Add(string(key), data)
	}

	return data, err

}

func (lldb *levelLruDatabase) Delete(key []byte) error {

	if err := lldb.check(); err != nil {
		return err
	}

	lldb.cache.Remove(string(key))

	err := lldb.leveldb.Delete(key)

	return err
}

func (lldb *levelLruDatabase) NewIterator(prefix []byte) db.Iterator {
	return lldb.leveldb.NewIterator(prefix)
}

//关闭数据库是不安全的，因为有可能有线程在写数据库，如果做到安全要加锁
//此处仅设置状态关闭
func (lldb *levelLruDatabase) Close() {
	if err := lldb.check(); err == nil {
		lldb.dbStatus = false
	}
}

func (lldb *levelLruDatabase) check() error {
	if lldb.dbStatus == false {
		log.Notice("levelLruDB has been closed")
		return errors.New("levelLruDB has been closed")
	}
	return nil
}

type levelLruBatch struct {
	mutex       sync.Mutex
	cache       *Lru.Cache
	batchStatus bool
	b           *ldbBatch
	batchMap    map[string][]byte
}

func (lldb *levelLruDatabase) NewBatch() db.Batch {
	if err := lldb.check(); err != nil {
		log.Notice("Bad operation:try to create a new batch with closed db")
		return nil
	}
	return &levelLruBatch{
		b:           &ldbBatch{db: lldb.leveldb.db, b: new(leveldb.Batch)},
		cache:       lldb.cache,
		batchStatus: true,
		batchMap:    make(map[string][]byte),
	}
}

func (batch levelLruBatch) Put(key, value []byte) error {
	if batch.batchStatus == false {
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}

	value1 := make([]byte, len(value))
	copy(value1, value)
	batch.mutex.Lock()
	batch.batchMap[string(key)] = value1
	batch.mutex.Unlock()
	batch.b.Put(key, value1)
	return nil
}

func (batch levelLruBatch) Delete(key []byte) error {
	if batch.batchStatus == false {
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	batch.mutex.Lock()
	delete(batch.batchMap, string(key))
	batch.mutex.Unlock()

	batch.b.Delete(key)
	return nil
}

func (batch levelLruBatch) Write() error {

	if batch.batchStatus == false {
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	for k, v := range batch.batchMap {
		batch.cache.Add(k, v)
	}

	go batch.b.Write()

	return nil
}

func (batch levelLruBatch) Len() int {
	return 0
}
