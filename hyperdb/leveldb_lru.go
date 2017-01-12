package hyperdb

import (
	Lru "github.com/hashicorp/golang-lru"
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

//if key-value value is 1kb lruSize is 1GB
var lruSize=1024*1024

type levelLruDatabase struct {
	leveldb *LDBDatabase
	cache 	*Lru.Cache
	dbStatus bool
}


func NewLevelLruDB() (*levelLruDatabase ,error){

	leveldb,err:=NewLDBDataBase(leveldbPath)
	if err!=nil{
		log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n",leveldbPath,err.Error())
		return nil,err
	}

	cache,err:=Lru.New(lruSize)
	if err!=nil{
		log.Noticef("Lru.New(%v) fail. err is %v. \n",lruSize,err.Error())
		return nil,err
	}

	return &levelLruDatabase{leveldb:leveldb,cache:cache,dbStatus:true},nil
}

func NewLevelLruDBWithP(path string ,size int) (*levelLruDatabase ,error){

	leveldb,err:=NewLDBDataBase(path)
	if err!=nil{
		log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n",leveldbPath,err.Error())
		return nil,err
	}

	cache,err:=Lru.New(size)
	if err!=nil{
		log.Noticef("Lru.New(%v) fail. err is %v. \n",lruSize,err.Error())
		return nil,err
	}

	return &levelLruDatabase{leveldb:leveldb,cache:cache,dbStatus:true},nil
}



func NewLevelLruDBInf() (Database ,error){

	leveldb,err:=NewLDBDataBase(leveldbPath)
	if err!=nil{
		log.Noticef("NewLDBDataBase(%v) fail. err is %v. \n",leveldbPath,err.Error())
		return nil,err
	}

	cache,err:=Lru.New(lruSize)
	if err!=nil{
		log.Noticef("Lru.New(%v) fail. err is %v. \n",lruSize,err.Error())
		return nil,err
	}

	return &levelLruDatabase{leveldb:leveldb,cache:cache},nil
}

func (db *levelLruDatabase)Put(key []byte, value []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}

	db.cache.Add(string(key),value)

	go db.leveldb.Put(key,value)

	return nil
}


func (db *levelLruDatabase)Get(key []byte) ([]byte, error){

	if err:=db.check(); err!=nil{
		return nil,err
	}

	data1,status:=db.cache.Get(string(key))

	if status{
		return Bytes(data1),nil
	}

	data,err:=db.leveldb.Get(key)

	if data!=nil{
		db.cache.Add(string(key),data)
	}

	return data,err

}

func (db *levelLruDatabase)Delete(key []byte) error{

	if err:=db.check(); err!=nil{
		return err
	}


	db.cache.Remove(string(key))

	err:=db.leveldb.Delete(key)

	return err
}

func (db *levelLruDatabase)NewIterator(prefix []byte) (Iterator){
	return db.leveldb.NewIterator(prefix)
}
//关闭数据库是不安全的，因为有可能有线程在写数据库，如果做到安全要加锁
//此处仅设置状态关闭
func (db *levelLruDatabase)Close(){
	if err:=db.check(); err==nil{
		db.dbStatus=false
	}
}

func (db *levelLruDatabase)check()error{
	if db.dbStatus==false{
		log.Notice("levelLruDB has been closed")
		return errors.New("levelLruDB has been closed")
	}
	return nil
}

type levelLruBatch struct {
	mutex sync.Mutex
	cache 	*Lru.Cache
	batchStatus bool
	b  *ldbBatch
	batchMap map[string] []byte
}

func (db *levelLruDatabase)NewBatch() Batch{
	if err:=db.check(); err!=nil{
		log.Notice("Bad operation:try to create a new batch with closed db")
		return nil
	}
	return &levelLruBatch{
		b : &ldbBatch{db: db.leveldb.db, b: new(leveldb.Batch)},
		cache:db.cache,
		batchStatus:true,
		batchMap:make(map[string][]byte),
	}
}

func (batch levelLruBatch) Put(key, value []byte) error{
	if batch.batchStatus==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}

	value1:=make([]byte,len(value))
	copy(value1,value)
	batch.mutex.Lock()
	batch.batchMap[string(key)]=value1
	batch.mutex.Unlock()
	batch.b.Put(key,value1)
	return nil
}

func (batch levelLruBatch)Delete(key []byte) error{
	if batch.batchStatus==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	batch.mutex.Lock()
	delete(batch.batchMap,string(key))
	batch.mutex.Unlock()

	batch.b.Delete(key)
	return nil
}

func (batch levelLruBatch)Write() error{


	if batch.batchStatus==false{
		log.Notice("batch has been closed")
		return errors.New("batch has been closed")
	}
	for k,v :=range batch.batchMap{
		batch.cache.Add(k,v)
	}

	go batch.b.Write()

	return nil
}
