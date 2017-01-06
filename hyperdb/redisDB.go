package hyperdb

//////////////////////////////////////// use pool
import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
	"time"
	"os"
	"fmt"
	"errors"
)

type RsDatabase struct {
	path    string
	rd_pool *redis.Pool
}

func NewRsDatabase(portDBPath string) (*RsDatabase, error) {
	port, err := strconv.Atoi(portDBPath)
	if err != nil {
		return nil, err
	}
	port += 100
	//set max pool con 10
	rdP := redis.NewPool(func() (redis.Conn, error) {return redis.Dial("tcp", ":"+strconv.Itoa(port), redis.DialConnectTimeout(60*time.Second))},10)
	return &RsDatabase{path: portDBPath, rd_pool: rdP}, err
}

func (self *RsDatabase) Put(key []byte, value []byte) error {

	num:=0
	var err error
	for {
		con := self.rd_pool.Get()
		_, err = con.Do("set", key, value)
		con.Close()

		if err == nil{
			return nil
		}

		num++
		f, err1 := os.OpenFile("./build/db.log", os.O_WRONLY|os.O_CREATE, 0644)
		if err1 != nil {
			fmt.Println("db.log file create failed. err: " + err.Error())
		} else {
			n, _ := f.Seek(0, os.SEEK_END)
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str := portDBPath + newFormat + `redis-con.Put("set",key,vale) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
			_, err1 = f.WriteAt([]byte(str), n)
			f.Close()
		}
		if err.Error() !="ERR Connection timed out"{
			return err
		}
	}
}

func (self *RsDatabase) Get(key []byte) ([]byte, error) {

	num:=0
	for {
		con := self.rd_pool.Get()
		dat, err := redis.Bytes(con.Do("get", key))
		con.Close()


		if err == nil {
			if len(dat) == 0 {
				err = errors.New("not found")
			}
			return dat,err
		}

		num++
		f, err1 := os.OpenFile("./build/db.log", os.O_WRONLY|os.O_CREATE, 0644)
		if err1 != nil {
			fmt.Println("db.log file create failed. err: " + err.Error())
		} else if err.Error() != "redigo: nil returned" {
			n, _ := f.Seek(0, os.SEEK_END)
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str := portDBPath + newFormat + `redis-con.DO("get",key) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
			_, err1 = f.WriteAt([]byte(str), n)
			f.Close()
		}
		if err.Error() !="ERR Connection timed out"{
			return nil,err
		}
	}
}

func (self *RsDatabase) Delete(key []byte) error {
	num := 0

	for {
		con := self.rd_pool.Get()
		_, err := con.Do("DEL", key)
		con.Close()

		if err == nil{
			return nil
		}

		num++
		f, err1 := os.OpenFile("./build/db.log", os.O_WRONLY|os.O_CREATE, 0644)
		if err1 != nil {
			fmt.Println("db.log file create failed. err: " + err.Error())
		} else {
			n, _ := f.Seek(0, os.SEEK_END)
			currentTime := time.Now().Local()
			newFormat := currentTime.Format("2006-01-02 15:04:05.000")
			str := portDBPath + newFormat + `redis-con.Delete("get",key) :` + err.Error() +" num:"+strconv.Itoa(num)+"\n"
			_, err1 = f.WriteAt([]byte(str), n)
			f.Close()
		}
		if err.Error() !="ERR Connection timed out"{
			return err
		}
	}
}
// just for implement interface
//iterator should do in leveldb
func (self *RsDatabase) NewIterator(prefix []byte) Iterator {
	return nil
}

func (self *RsDatabase) Close() {
	self.rd_pool.Close()
}

//// just for implement interface
//func (self *RsDatabase) LDB() *leveldb.DB {
//	return nil
//}

//TODO specific the size of map
func (self *RsDatabase) NewBatch() Batch {
	return &rd_Batch{rd_pool: self.rd_pool, map1: make(map[string][]byte)}
}

type rd_Batch struct {
	mutex   sync.Mutex
	rd_pool *redis.Pool
	map1    map[string][]byte
}

// Put put the key-value to rd_Batch
func (batch *rd_Batch) Put(key, value []byte) error {

	value1 := make([]byte, len(value))
	copy(value1, value)
	batch.mutex.Lock()
	batch.map1[string(key)] = value1
	batch.mutex.Unlock()
	return nil
}

func (batch *rd_Batch) Delete(key []byte) error {
	batch.mutex.Lock()
	delete(batch.map1, string(key))
	batch.mutex.Unlock()
	return nil
}

// Write write batch-operation to databse
//one transaction from MULTI TO EXEC
func (batch *rd_Batch) Write() error {

	con := batch.rd_pool.Get()
	defer con.Close()
	con.Send("MULTI")
	batch.mutex.Lock()
	for k, v := range batch.map1 {
		con.Send("set", []byte(k), v)
	}
	batch.mutex.Unlock()
	_, err := con.Do("EXEC")

	if err == nil {
		batch.map1 = make(map[string][]byte)
	}

	return err
}
