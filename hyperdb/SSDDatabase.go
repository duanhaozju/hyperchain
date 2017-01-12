package hyperdb

import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
	"time"
	"errors"
)



//ssdb database use ssdb and twemproxy
type SSDatabase struct {
	rdPool *redis.Pool
	ssdbNum int //ssdb服务器数量 比如现在一个ssdb的twemproxy下面有两台ssdb 所以现在要遍历的台数为2
	port int
}


func NewSSDatabase() (*SSDatabase, error) {


	rdP:=redis.NewPool(func() (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(ssdbProxyPort),redis.DialConnectTimeout(time.Duration(ssdbTimeout)*time.Second))},ssdbPoolSize)

	return &SSDatabase{rdPool:rdP,ssdbNum:ssdbServerNumber,port:ssdbFirstPort},nil
}

/*
链接时间可能会超时
现在先重发MaxConnectTimes次，如果还是超时，则抛出错误
 */
func (ssdb *SSDatabase) Put(key []byte, value []byte) error {
	//count the times of connection time out
	num:=0
	for {
		con := ssdb.rdPool.Get()
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

		if num>=ssdbMaxConnectTimes{
			log.Error("SSDB Set key-value to  DB failed and it may cause unexpected effects")
			return err
		}

	}
}


func (ssdb *SSDatabase) Get(key []byte) ([]byte, error) {

	num:=0
	for {
		con := ssdb.rdPool.Get()
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

		
		if num>=ssdbMaxConnectTimes{
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
		
		con := ssdb.rdPool.Get()
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

		if num>=ssdbMaxConnectTimes{
			log.Error("SSDB Delete key-value from  DB failed and it may cause unexpected effects")
			return err
		}
		

	}
}

type IteratorSsdb struct {
	key []byte
	ssdbNum int
	ssdbNow int
	port	int
	err 	error
	iterator *IteratorImp
}

func (it *IteratorSsdb)Key() []byte{
	return it.iterator.Key()
}

func (it *IteratorSsdb)Value() []byte{
	return it.iterator.Value()
}

func (it *IteratorSsdb)Seek(key []byte) bool{
	if it.key==nil{
		it.key=make([]byte,len(key))
		copy(it.key,key)
	}
	if !it.iterator.Seek(key) {
		if it.iterator.err!=nil{
			it.err=it.iterator.err
			return false
		}
		if it.ssdbNow<it.ssdbNum{
			it.ssdbNow++
			it.port+=ssdbGap
			rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(it.port),redis.DialConnectTimeout(time.Duration(ssdbTimeout)*time.Second))},ssdbPoolSize)
			itertaor2:=&IteratorImp{
				Pool:rdP,
				ListKey:make([]string,ssdbIteratorSize,ssdbIteratorSize),
				ListValue:make([]string,ssdbIteratorSize,ssdbIteratorSize),
			}
			it.iterator.Release()
			it.iterator=itertaor2
			return it.Seek(key)
		}
		return false
	}
	return true
}

func (it *IteratorSsdb)Next() bool{
	if !it.iterator.Next(){

		if it.ssdbNow<it.ssdbNum{
			it.ssdbNow++
			it.port+=10
			rdP:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(it.port),redis.DialConnectTimeout(60*time.Second))},10)
			itertaor2:=&IteratorImp{
				Pool:rdP,
				ListKey:make([]string,ssdbIteratorSize,ssdbIteratorSize),
				ListValue:make([]string,ssdbIteratorSize,ssdbIteratorSize),
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

func (it *IteratorSsdb)Error() error{
	return it.err
}

func (it *IteratorSsdb)Release(){
	it.iterator.Release()
}

func (ssdb *SSDatabase) NewIterator(prefix []byte) Iterator {
	rdp:=redis.NewPool(func () (redis.Conn, error) { return redis.Dial("tcp", ":"+strconv.Itoa(ssdb.port),redis.DialConnectTimeout(time.Duration(ssdbTimeout)*time.Second))},ssdbPoolSize)
	imp:=&IteratorImp{Pool:rdp,ListKey:make([]string,ssdbIteratorSize,ssdbIteratorSize),ListValue:make([]string,ssdbIteratorSize,ssdbIteratorSize),}
	return &IteratorSsdb{ssdbNum:ssdbServerNumber,ssdbNow:1,iterator:imp,port:ssdb.port}
}

func (ssdb *SSDatabase) Close() {
	ssdb.rdPool.Close()
}

//// just for implement interface
//func (ssdb *SSDatabase) LDB() *leveldb.DB {
//	return nil
//}

//TODO specific the size of map
func (ssdb *SSDatabase) NewBatch() Batch {
	return &sd_Batch{ rdPool : ssdb.rdPool,map1:make(map[string][]byte)}
}


type sd_Batch struct {
	mutex sync.Mutex
	rdPool *redis.Pool
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

// Write write batch-operation to database
//one transaction from MULTI TO EXEC
func (batch *sd_Batch) Write() error {

	num:=0;
	var err error

	for {
		list := make([]string, 0, 20)
		con := batch.rdPool.Get()

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

		if num>=ssdbMaxConnectTimes{
			log.Error("SSDB write batch to  DB failed and it may cause unexpected effects")
			return err
		}
	}

	return err
}
