package hyperdb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)


//-- 用leveldb实现的LDBDatabase
//-- LDBDatabase实现了DataBase接口

type LDBDatabase struct {
	filepath string
	db *leveldb.DB
}

func NewLDBDataBase(filepath string) (*LDBDatabase, error) {
	DBLog("new a ldb database connect")
	db, err := leveldb.OpenFile(filepath, nil )
	return &LDBDatabase{
		filepath:filepath,
		db:db,
	}, err
}

func (self *LDBDatabase) Put(key []byte, value []byte) error {

	DBLog("put data to ldb")
	return self.db.Put(key, value, nil)
}

func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
	DBLog("get data from ldb")
	dat, err := self.db.Get(key, nil)
	return dat, err
}

func (self *LDBDatabase) Delete(key []byte) error {
	DBLog("delete data")
	return self.db.Delete(key, nil)
}

func (self *LDBDatabase) NewIterator() iterator.Iterator {
	return self.db.NewIterator(nil, nil)
}

func (self *LDBDatabase) Close() {
	DBLog("close the ldb connect")
	self.db.Close()
}

func (self *LDBDatabase) LDB() *leveldb.DB {
	return self.db
}

func (db *LDBDatabase) NewBatch() Batch {
	return &ldbBatch{db: db.db, b: new(leveldb.Batch)}
}


//-- 数据库的batch操作
type ldbBatch struct {
	db *leveldb.DB
	b  *leveldb.Batch
}

func (b *ldbBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	return nil
}

func (b *ldbBatch) Write() error {
	return b.db.Write(b.b, nil)
}