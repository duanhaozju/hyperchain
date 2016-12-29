package hyperdb

//////////////////////////////////////// use pool
import (
	"github.com/garyburd/redigo/redis"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"strconv"
	"sync"
	"time"
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
	con := self.rd_pool.Get()
	defer con.Close()
	_, err := con.Do("set", key, value)
	return err
}

func (self *RsDatabase) Get(key []byte) ([]byte, error) {
	con := self.rd_pool.Get()
	defer con.Close()
	dat, err := redis.Bytes(con.Do("get", key))
	return dat, err
}

func (self *RsDatabase) Delete(key []byte) error {
	con := self.rd_pool.Get()
	defer con.Close()
	_, err := con.Do("DEL", key)
	return err
}

// just for implement interface
//iterator should do in leveldb
func (self *RsDatabase) NewIterator() iterator.Iterator {
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
