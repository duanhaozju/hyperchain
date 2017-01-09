package hyperdb

//////////////////////////////////////// use pool
import (
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
	"time"
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
	for {
		con := self.rd_pool.Get()
		_, err := con.Do("set", key, value)
		con.Close()

		if err == nil{
			return nil
		}

		num++
		if IfLogStatus(){
			writeLog(`Redis con.Do("set", key, value)`,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}
		if num>=MaxConneecTimes{
			log.Error("Redis Put key-value to DB failed and it may cause unexpected effects")
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


		if err == nil ||err.Error()=="redigo: nil returned"{
			if len(dat) == 0 {
				err = errors.New("not found")
			}
			return dat,err
		}

		num++

		if IfLogStatus(){
			writeLog(`Redis con.Do("get", key)`,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return nil,err
		}

		if num>=MaxConneecTimes{
			log.Error("Redis Get key-value from  DB failed and it may cause unexpected effects")
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

		if IfLogStatus(){
			writeLog(`Redis con.Do("DEL", key)`,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}

		if num>=MaxConneecTimes{
			log.Error("Redis Delete key-value from  DB failed and it may cause unexpected effects")
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
	num:=0
	for {
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
			return err
		}

		num++

		if IfLogStatus(){
			writeLog(`Redis batch write `,num,err)
		}

		if err.Error() !="ERR Connection timed out"{
			return err
		}

		if num>=MaxConneecTimes{
			log.Error("Redis write batch to  DB failed and it may cause unexpected effects")
			return err
		}
	}

}
