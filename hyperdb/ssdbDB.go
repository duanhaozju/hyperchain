package hyperdb

/////////////////////////////////使用redis协议链接twemproxy，不用ssdb
//import (
//	"github.com/syndtr/goleveldb/leveldb/iterator"
//	"github.com/seefan/gossdb"
//	"sync"
//	"strconv"
//	"github.com/pkg/errors"
//)
//type LDBDatabase struct {
//	ssdb_pool *gossdb.Connectors
//}
//
//
//
//func NewLDBDatabase(portLDBPath,Ip string) (*LDBDatabase, error) {
//	//fmt.Println("portLDBPath:"+portLDBPath)
//	port,err:=strconv.Atoi(portLDBPath)
//	if err!=nil{
//		return nil,err
//	}
//	port+=10
//	//set max pool con 4
//	pool, err := gossdb.NewPool(&gossdb.Config{
//		Host:             Ip,
//		Port:             port,
//		MinPoolSize:      5,
//		MaxPoolSize:      50,
//		AcquireIncrement: 5,
//	})
//
//	return &LDBDatabase{ssdb_pool:pool},err
//}
//
//func (self *LDBDatabase) Put(key []byte, value []byte) error {
//	con,err:=self.ssdb_pool.NewClient()
//	if err!=nil{
//		return err
//	}
//	defer con.Close()
//
//	err=con.Set(string(key), value)
//
//	return err
//}
//
//
//
//func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
//	con,err:=self.ssdb_pool.NewClient()
//	if err!=nil{
//		return nil,err
//	}
//	defer con.Close()
//
//	data,err:=con.Get(string(key))
//	if data.String()==""{
//		return nil,errors.New("don't found key")
//	}
//	return []byte(data), err
//}
////
//
//func (self *LDBDatabase) Delete(key []byte) error {
//	con,err:=self.ssdb_pool.NewClient()
//	if err!=nil{
//		return err
//	}
//	defer con.Close()
//
//	err=con.Del(string(key))
//	return  err
//}
//
//
//func (self *LDBDatabase) NewIterator() iterator.Iterator {
//	return &IteratorImp{ssdb_pool:self.ssdb_pool,listkey:make([]string,50),listvalue:make([]string,50)}
//}
//
//
//
//
//func (self *LDBDatabase) Close() {
//	self.ssdb_pool.Close()
//}
//
//
//
////TODO specific the size of map
//func (self *LDBDatabase) NewBatch() Batch {
//	return &ssdb_Batch{ ssdb_pool : self.ssdb_pool,list:make([]string,0,20)}
//}
//
//
//type ssdb_Batch struct {
//	mutex sync.Mutex
//	ssdb_pool *gossdb.Connectors
//	list []string
//}
//
//// Put put the key-value to rd_Batch
//func (batch *ssdb_Batch) Put(key, value []byte) error {
//
//	value1:=make([]byte,len(value))
//	copy(value1,value)
//	//fmt.Println("Put start:")
//	//fmt.Println(key)
//	//fmt.Println(string(value1))
//	batch.mutex.Lock()
//	batch.list=append(batch.list,string(key),string(value1))
//	batch.mutex.Unlock()
//	return nil
//}
//
//// Write write batch-operation to databse
////one transaction from MULTI TO EXEC
//func (batch *ssdb_Batch) Write() error {
//
//	con,err:=batch.ssdb_pool.NewClient()
//	if err!=nil{
//		return err
//	}
//	defer con.Close()
//	batch.mutex.Lock()
//	_,err=con.Do("multi_set",batch.list)
//	bacth.list=make([]string,0,20)
//	batch.mutex.Unlock()
//	return err
//}