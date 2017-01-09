package hyperdb

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
	"time"
	"fmt"
	"errors"
)



//ssdb database use ssdb and twemproxy
type SSDatabase struct {
	rd_pool *redis.Pool
	ssdbnum int //ssdb服务器数量 比如现在一个ssdb的twemproxy下面有两台ssdb 所以现在要遍历的台数为2
	port int
}


func NewSSDatabase(portDBPath string,ssdbnum int) (*SSDatabase, error) {
	
	port,err:=strconv.Atoi(portDBPath)
	if err!=nil{
		return nil,errors.New(fmt.Sprintf("NewSSDatabase(%v) fail because portDBPath is not a number err:%v ",portDBPath,err.Error()))
	}

	
	
	//set max pool con 10
	rdP:=redis.NewPool(func() (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(port+14121),redis.DialConnectTimeout(60*time.Second))},10)

	return &SSDatabase{rd_pool:rdP,ssdbnum:ssdbnum,port:port+10},nil
}

/*
链接时间可能会超时
现在先重发MaxConnectTimes次，如果还是超时，则抛出错误
 */
func (ssdb *SSDatabase) Put(key []byte, value []byte) error {
	//count the times of connection time out
	num:=0
	for {
		con := ssdb.rd_pool.Get()
		_, err:= con.Do("set", key, value)
		con.Close()

		if err == nil{
			return err
		}

		num++

		if IfLogStatus(){
			writeLog(`SSDB con.Do("set", key, value)`,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}

		if num>=MaxConneecTimes{
			log.Error("SSDB Set key-value to  DB failed and it may cause unexpected effects")
			return err
		}

	}
}


func (ssdb *SSDatabase) Get(key []byte) ([]byte, error) {

	num:=0
	for {
		con := ssdb.rd_pool.Get()
		dat, err:= redis.Bytes(con.Do("get", key))
		con.Close()

		if err == nil||err.Error()=="redigo: nil returned" {
			if len(dat) == 0 {
				err = errors.New("not found")
			}
			return dat,err
		}
		num++

		if IfLogStatus(){
			writeLog(`SSDB con.Do("get", key)`,num,err)
		}

		
		if num>=MaxConneecTimes{
			log.Error("SSDB Get key-value from  DB failed and it may cause unexpected effects")
			return nil,err
		}
		
		if err.Error() !="ERR Connection timed out"{
			return nil,err
		}

	}
}


func (ssdb *SSDatabase) Delete(key []byte) error {
	num:=0
	for {
		
		con := ssdb.rd_pool.Get()
		_, err := con.Do("DEL", key)
		con.Close()
		if err == nil {
			return nil
		}
		num++

		if IfLogStatus(){
			writeLog(`SSDB con.Do("DEL", key)`,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}

		if num>=MaxConneecTimes{
			log.Error("SSDB Delete key-value from  DB failed and it may cause unexpected effects")
			return err
		}
		

	}
}

type Iteratorssdb struct {
	key []byte
	ssdbnum int
	ssdbnow int
	port	int
	err 	error
	iterator *IteratorImp
}

func (it *Iteratorssdb)Key() []byte{
	return it.iterator.Key()
}

func (it *Iteratorssdb)Value() []byte{
	return it.iterator.Value()
}

func (it *Iteratorssdb)Seek(key []byte) bool{
	if it.key==nil{
		it.key=make([]byte,len(key))
		copy(it.key,key)
	}
	if !it.iterator.Seek(key) {
		if it.iterator.err!=nil{
			it.err=it.iterator.err
			return false
		}
		if it.ssdbnow<it.ssdbnum{
			it.ssdbnow++
			it.port+=10
			rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(it.port),redis.DialConnectTimeout(60*time.Second))},10)
			itertaor2:=&IteratorImp{
				Pool:rdP,
				Listkey:make([]string,Size,Size),
				Listvalue:make([]string,Size,Size),
			}
			it.iterator.Release()
			it.iterator=itertaor2
			return it.Seek(key)
		}
		return false
	}
	return true
}

func (it *Iteratorssdb)Next() bool{
	if !it.iterator.Next(){

		if it.ssdbnow<it.ssdbnum{
			it.ssdbnow++
			it.port+=10
			rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(it.port),redis.DialConnectTimeout(60*time.Second))},10)
			itertaor2:=&IteratorImp{
				Pool:rdP,
				Listkey:make([]string,Size,Size),
				Listvalue:make([]string,Size,Size),
			}
			it.iterator.Release()
			it.iterator=itertaor2
			if it.Seek(it.key){
				return it.Next()
			}
		}
		return false
	}
	return true
}

func (it *Iteratorssdb)Error() error{
	return it.err
}

func (it *Iteratorssdb)Release(){
	it.iterator.Release()
}

func (ssdb *SSDatabase) NewIterator(prefix []byte) Iterator {
	rdp:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(ssdb.port),redis.DialConnectTimeout(60*time.Second))},10)
	imp:=&IteratorImp{Pool:rdp,Listkey:make([]string,Size,Size),Listvalue:make([]string,Size,Size),}
	return &Iteratorssdb{ssdbnum:ssdb.ssdbnum,ssdbnow:1,iterator:imp,port:ssdb.port}
}

func (ssdb *SSDatabase) Close() {
	ssdb.rd_pool.Close()
}

//// just for implement interface
//func (ssdb *SSDatabase) LDB() *leveldb.DB {
//	return nil
//}

//TODO specific the size of map
func (ssdb *SSDatabase) NewBatch() Batch {
	return &sd_Batch{ rd_pool : ssdb.rd_pool,map1:make(map[string][]byte)}
}


type sd_Batch struct {
	mutex sync.Mutex
	rd_pool *redis.Pool
	map1 map[string] []byte
}


// Put put the key-value to sd_Batch
func (batch *sd_Batch) Put(key, value []byte) error {

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

func(batch *sd_Batch) Delete(key []byte) error{
	batch.mutex.Lock()
	delete(batch.map1,string(key))
	batch.mutex.Unlock()
	return nil
}

// Write write batch-operation to databse
//one transaction from MULTI TO EXEC
func (batch *sd_Batch) Write() error {

	num:=0;
	var err error

	for {
		list := make([]string, 0, 20)
		con := batch.rd_pool.Get()

		for k, v := range batch.map1 {
			list = append(list, string(k), string(v))
		}
		_, err:= con.Do("mset", list)
		con.Close()
		if err == nil {
			batch.map1 = make(map[string][]byte)
			break
		}
		num++

		if IfLogStatus(){
			writeLog(`SSDB batch write `,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}

		if num>=MaxConneecTimes{
			log.Error("SSDB write batch to  DB failed and it may cause unexpected effects")
			return err
		}
	}

	return err
}
