package hyperdb

//import (
//	"github.com/garyburd/redigo/redis"
//	"github.com/syndtr/goleveldb/leveldb/iterator"
//	"github.com/syndtr/goleveldb/leveldb"
//	"strconv"
//	"sync"
//	"time"
//)

//type LDBDatabase struct {
//	path string
//	rd_pool *redis.Pool
//}
//
//
//func NewLDBDatabase(portLDBPath string) (*LDBDatabase, error) {
//	//fmt.Println("portLDBPath:"+portLDBPath)
//	port,err:=strconv.Atoi(portLDBPath)
//	if err!=nil{
//		return nil,err
//	}
//	port+=10
//	//set max pool con 4
//	rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(port),redis.DialConnectTimeout(60*time.Second))},4)
//
//	return &LDBDatabase{path:portLDBPath,rd_pool:rdP},err
//}
//
//
//func (self *LDBDatabase) Put(key []byte, value []byte) error {
//	con:=self.rd_pool.Get()
//	defer con.Close()
//	_,err:=con.Do("set",key, value)
//	return err
//}
//
//
//func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
//	con:=self.rd_pool.Get()
//	defer con.Close()
//	dat,err := redis.Bytes(con.Do("get",key))
//	return dat, err
//}
//
//
//func (self *LDBDatabase) Delete(key []byte) error {
//	con:=self.rd_pool.Get()
//	defer con.Close()
//	_,err:=con.Do("DEL", key)
//	return err;
//}
//
//// just for implement interface
////iterator should do in leveldb
//func (self *LDBDatabase) NewIterator() iterator.Iterator {
//	return nil
//}
//
//func (self *LDBDatabase) Close() {
//	self.rd_pool.Close()
//}
//
//// just for implement interface
//func (self *LDBDatabase) LDB() *leveldb.DB {
//	return nil
//}
//
////TODO specific the size of map
//func (self *LDBDatabase) NewBatch() Batch {
//	return &rd_Batch{ rd_pool : self.rd_pool,map1:make(map[string][]byte)}
//}
//
//
//type rd_Batch struct {
//	mutex sync.Mutex
//	rd_pool *redis.Pool
//	map1 map[string] []byte
//}
//
//// Put put the key-value to rd_Batch
//func (batch *rd_Batch) Put(key, value []byte) error {
//
//	value1:=make([]byte,len(value))
//	copy(value1,value)
//	//fmt.Println("Put start:")
//	//fmt.Println(key)
//	//fmt.Println(string(value1))
//	batch.mutex.Lock()
//	batch.map1[string(key)]=value1
//	batch.mutex.Unlock()
//	return nil
//}
//
//// Write write batch-operation to databse
////one transaction from MULTI TO EXEC
//func (batch *rd_Batch) Write() error {
//
//	con:=batch.rd_pool.Get()
//	defer con.Close()
//	con.Send("MULTI")
//	//fmt.Println("write start:")
//	batch.mutex.Lock()
//	for k, v := range batch.map1 {
//		//fmt.Println(k)
//		//fmt.Println(string(v))
//		con.Send("set",[]byte(k),v)
//	}
//	batch.mutex.Unlock()
//	_,err:=con.Do("EXEC")
//
//	if err==nil{
//		batch.map1=make(map[string][]byte)
//	}
//
//	return err
//}

//
//
//import (
//	"github.com/garyburd/redigo/redis"
//	"github.com/syndtr/goleveldb/leveldb/iterator"
//	"github.com/syndtr/goleveldb/leveldb"
//	"strconv"
////	"sync"
//)
//
//type LDBDatabase struct {
//	path string
//	db   redis.Conn
//}
//
//
//func NewLDBDatabase(portLDBPath string) (*LDBDatabase, error) {
//	//fmt.Println("portLDBPath:"+portLDBPath)
//	port,err:=strconv.Atoi(portLDBPath)
//	if err!=nil{
//		return nil,err
//	}
//	port+=10
//	db1,err:=redis.Dial("tcp", ":"+strconv.Itoa(port))
//	return &LDBDatabase{path:portLDBPath,db:db1},err
//}
//
//
//func (self *LDBDatabase) Put(key []byte, value []byte) error {
//	_,err:=self.db.Do("set",key, value)
//	return err
//}
//
//
//func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
//	dat,err := redis.Bytes(self.db.Do("get",key))
//	return dat, err
//}
//
//
//func (self *LDBDatabase) Delete(key []byte) error {
//	_,err:=self.db.Do("DEL", key)
//	return err;
//}
//
//// just for implement interface
////iterator should do in leveldb
//func (self *LDBDatabase) NewIterator() iterator.Iterator {
//	return nil
//}
//
//func (self *LDBDatabase) Close() {
//	self.db.Close()
//}
//
//// just for implement interface
//func (self *LDBDatabase) LDB() *leveldb.DB {
//	return nil
//}
//
////TODO specific the size of map
//func (self *LDBDatabase) NewBatch() Batch {
//	return &rd_Batch{ db : self.db,map1:make(map[string][]byte)}
//}
//
//
//type rd_Batch struct {
//	//mutex sync.Mutex
//	db redis.Conn
//	map1 map[string] []byte
//}
//
//// Put put the key-value to rd_Batch
//func (db *rd_Batch) Put(key, value []byte) error {
//
//	value1:=make([]byte,len(value))
//	copy(value1,value)
//	//fmt.Println("Put start:")
//	//fmt.Println(key)
//	//fmt.Println(string(value1))
//	//db.mutex.Lock()
//	db.map1[string(key)]=value1
//	//db.mutex.Unlock()
//	return nil
//}
//
//// Write write batch-operation to databse
////one transaction from MULTI TO EXEC
//func (db *rd_Batch) Write() error {
//
//	db.db.Send("MULTI")
//	//fmt.Println("write start:")
//	//db.mutex.Lock()
//	for k, v := range db.map1 {
//		//fmt.Println(k)
//		//fmt.Println(string(v))
//		db.db.Send("set",[]byte(k),v)
//	}
//	//db.mutex.Unlock()
//	_,err:=db.db.Do("EXEC")
//
//
//	db.map1=make(map[string][]byte)
//	//fmt.Println("write end with:")
//	//fmt.Println(err)
//	return err
//}
