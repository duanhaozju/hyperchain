//author:frank
//data:2016/12/20
package hyperdb

////implement iterator
////use
//import (
//	"github.com/garyburd/redigo/redis"
//	"github.com/syndtr/goleveldb/leveldb/iterator"
//	"github.com/syndtr/goleveldb/leveldb"
//	"strconv"
////	"sync"
//	"fmt"
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
//	return &rd_Batch{ db : self.db,list:make([]string,0,1000)}
//}
//
//
//type rd_Batch struct {
//	//mutex sync.Mutex
//	db redis.Conn
//	list []string
//	num int
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
//	//batch.mutex.Lock()
//	batch.list=append(batch.list,string(key),string(value1))
//	//batch.mutex.Unlock()
//	batch.num+=2;
//	return nil
//}
//
//// Write write batch-operation to databse
////one transaction from MULTI TO EXEC
//func (batch *rd_Batch) Write() error {
//
//
//	_,err:=batch.db.Do("mset",batch.list)
//
//	batch.list=make([]string,0,1000)
//	fmt.Println("batch write. the size of batch is :"+strconv.Itoa(batch.num))
//	//fmt.Println("write end with:")
//	//fmt.Println(err)
//	return err
//}

/////////////////////////////////////////////////////////////////////


import (
	"github.com/garyburd/redigo/redis"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"sync"
	"time"
	"os"
	"fmt"
	"errors"
)

type LDBDatabase struct {
	path string
	rd_pool *redis.Pool
}


func NewLDBDatabase(portLDBPath string) (*LDBDatabase, error) {
	//fmt.Println("portLDBPath:"+portLDBPath)
	port,err:=strconv.Atoi(portLDBPath)
	if err!=nil{
		return nil,err
	}
	port+=14121 //8001 22122
	//set max pool con 4
	rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(port),redis.DialConnectTimeout(60*time.Second))},4)

	return &LDBDatabase{path:portLDBPath,rd_pool:rdP},err
}


func (self *LDBDatabase) Put(key []byte, value []byte) error {
	con:=self.rd_pool.Get()
	defer con.Close()
	_,err:=con.Do("set",key, value)
	if err!=nil{
		f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY, 0644)
		if err1 != nil {
			fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
		} else {
			// 查找文件末尾的偏移量
			n, _ := f.Seek(0, os.SEEK_END)
			// 从末尾的偏移量开始写入内容
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str:=newFormat+`con.Do("set",,key, value):`+err.Error()
			_, err = f.WriteAt([]byte(str), n)

			f.Close()
		}
	}
	return err
}


func (self *LDBDatabase) Get(key []byte) ([]byte, error) {
	con:=self.rd_pool.Get()
	defer con.Close()
	dat,err := redis.Bytes(con.Do("get",key))
	if err!=nil{
		f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY, 0644)
		if err1 != nil {
			fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
		} else {
			// 查找文件末尾的偏移量
			n, _ := f.Seek(0, os.SEEK_END)
			// 从末尾的偏移量开始写入内容
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str:=newFormat+`con.Do("get",key):`+err.Error()
			_, err = f.WriteAt([]byte(str), n)

			f.Close()
		}
	}

	if err==nil&&len(dat)==0{
		err= errors.New("not found")
	}
	return dat, err
}


func (self *LDBDatabase) Delete(key []byte) error {
	con:=self.rd_pool.Get()
	defer con.Close()
	_,err:=con.Do("DEL", key)
	return err;
}

// just for implement interface
//iterator should do in leveldb
func (self *LDBDatabase) NewIterator() iterator.Iterator {
	return nil
}

func (self *LDBDatabase) Close() {
	self.rd_pool.Close()
}

// just for implement interface
func (self *LDBDatabase) LDB() *leveldb.DB {
	return nil
}

//TODO specific the size of map
func (self *LDBDatabase) NewBatch() Batch {
	return &rd_Batch{ rd_pool : self.rd_pool,map1:make(map[string][]byte)}
}


type rd_Batch struct {
	mutex sync.Mutex
	rd_pool *redis.Pool
	map1 map[string] []byte
}

// Put put the key-value to rd_Batch
func (batch *rd_Batch) Put(key, value []byte) error {

	value1:=make([]byte,len(value))
	copy(value1,value)
	//fmt.Println("Put start:")
	//fmt.Println(key)
	//fmt.Println(string(value1))
	batch.mutex.Lock()
	batch.map1[string(key)]=value1
	batch.mutex.Unlock()
	return nil
}

// Write write batch-operation to databse
//one transaction from MULTI TO EXEC
func (batch *rd_Batch) Write() error {

	list:=make([]string,0,20)
	con:=batch.rd_pool.Get()
	defer con.Close()
	for k, v := range batch.map1 {
		list=append(list,string(k),string(v))
	}
	_,err:=con.Do("mset",list)
	if err==nil {
		batch.map1 = make(map[string][]byte)
	}else{
		f, err1 := os.OpenFile("/home/frank/1.txt", os.O_WRONLY, 0644)
		if err1 != nil {
			fmt.Println("cacheFileList.yml file create failed. err: " + err.Error())
		} else {
			// 查找文件末尾的偏移量
			n, _ := f.Seek(0, os.SEEK_END)
			// 从末尾的偏移量开始写入内容
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str:=newFormat+`con.Do("mset",list) :`+err.Error()
			_, err = f.WriteAt([]byte(str), n)

			f.Close()
		}
	}
	return err
}
