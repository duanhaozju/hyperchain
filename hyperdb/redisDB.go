package hyperdb

//////////////////////////////////////// use pool
import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"sync"
	"time"
)

type RsDatabase struct {
	rdPool *redis.Pool
}

func NewRsDatabase() (*RsDatabase, error) {
	rdP := redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", ":"+strconv.Itoa(redisPort), redis.DialConnectTimeout(time.Duration(redisTimeout)*time.Second))
	}, redisPoolSize)
	return &RsDatabase{rdPool: rdP}, nil
}

func (self *RsDatabase) Put(key []byte, value []byte) error {

	num := 0
	for {
		con := self.rdPool.Get()
		_, err := con.Do("set", key, value)
		con.Close()

		if err == nil {
			return nil
		}

		num++
		if IfLogStatus() {
			writeLog(`Redis con.Do("set", key, value)`, num, err)
		}

		if err.Error() != "ERR Connection timed out" {
			return err
		}
		if num >= redisMaxConnectTimes {
			log.Error("Redis Put key-value to DB failed and it may cause unexpected effects")
			return err
		}
	}
}

func (self *RsDatabase) Get(key []byte) ([]byte, error) {

	num := 0
	for {
		con := self.rdPool.Get()
		dat, err := redis.Bytes(con.Do("get", key))
		con.Close()

		if err == nil || err.Error() == "redigo: nil returned" {
			if len(dat) == 0 {
				err = errors.New("not found")
			}
			return dat, err
		}

		num++

		if IfLogStatus() {
			writeLog(`Redis con.Do("get", key)`, num, err)
		}

		if err.Error() != "ERR Connection timed out" {
			return nil, err
		}

		if num >= redisMaxConnectTimes {
			log.Error("Redis Get key-value from  DB failed and it may cause unexpected effects")
			return nil, err
		}
	}
}

func (self *RsDatabase) Delete(key []byte) error {
	num := 0

	for {
		con := self.rdPool.Get()
		_, err := con.Do("DEL", key)
		con.Close()

		if err == nil {
			return nil
		}

		num++

		if IfLogStatus() {
			writeLog(`Redis con.Do("DEL", key)`, num, err)
		}

		if err.Error() != "ERR Connection timed out" {
			return err
		}

		if num >= redisMaxConnectTimes {
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
	self.rdPool.Close()
}

//// just for implement interface
//func (self *RsDatabase) LDB() *leveldb.DB {
//	return nil
//}

//TODO specific the size of map
func (self *RsDatabase) NewBatch() Batch {
	return &rdBatch{rdPool: self.rdPool, map1: make(map[string][]byte)}
}

type rdBatch struct {
	mutex  sync.Mutex
	rdPool *redis.Pool
	map1   map[string][]byte
}

// Put put the key-value to rdBatch
func (batch *rdBatch) Put(key, value []byte) error {

	value1 := make([]byte, len(value))
	copy(value1, value)
	batch.mutex.Lock()
	batch.map1[string(key)] = value1
	batch.mutex.Unlock()
	return nil
}

func (batch *rdBatch) Delete(key []byte) error {
	batch.mutex.Lock()
	delete(batch.map1, string(key))
	batch.mutex.Unlock()
	return nil
}

// Write write batch-operation to databse
//one transaction from MULTI TO EXEC
func (batch *rdBatch) Write() error {
	num := 0
	for {
		con := batch.rdPool.Get()
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

		if IfLogStatus() {
			writeLog(`Redis batch write `, num, err)
		}

		if err.Error() != "ERR Connection timed out" {
			return err
		}

		if num >= redisMaxConnectTimes {
			log.Error("Redis write batch to  DB failed and it may cause unexpected effects")
			return err
		}
	}

}
