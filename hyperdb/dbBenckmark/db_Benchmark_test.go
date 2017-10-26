package dbBenckmark

import (
	"fmt"
	"github.com/facebookgo/ensure"
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/common"
	"hyperchain/hyperdb/cassandra"
	hcom "hyperchain/hyperdb/common"
	dbUtils "hyperchain/hyperdb/db"
	"testing"
	"time"
)

func TestGetDB2(t *testing.T) {
	conf := common.NewRawConfig()
	conf.Set(hcom.NODE_NUM, 3)
	conf.Set(hcom.NODE_prefix+"1", "172.16.0.11")
	conf.Set(hcom.NODE_prefix+"2", "172.16.0.12")
	conf.Set(hcom.NODE_prefix+"3", "172.16.0.14")
	DB, err := cassandra.NewCassandra(conf, "global")
	ensure.Nil(t, err)

	err = DB.Put([]byte("Key1"), []byte("Value1"))
	//fmt.Println(time.Since(start).Seconds())
	ensure.Nil(t, err)
	data, err := DB.Get([]byte("Key1"))
	ensure.DeepEqual(t, data, []byte("Value1"))

	err = DB.Delete([]byte("Key1"))
	ensure.Nil(t, err)

	_, err = DB.Get([]byte("Key1"))
	ensure.DeepEqual(t, err, dbUtils.DB_NOT_FOUND)

	putTimes := 1000
	start := time.Now()
	for i := 0; i < putTimes; i++ {
		err = DB.Put([]byte("Put_key"+fmt.Sprintf("%09d", i)), []byte("Put_Value"+fmt.Sprintf("%09d", i)))
		ensure.Nil(t, err)
	}
	fmt.Printf("Put %v k-v spend time : %v \n", putTimes, time.Since(start).Seconds())

	start = time.Now()
	for i := 0; i < putTimes; i++ {
		data, err = DB.Get([]byte("Put_key" + fmt.Sprintf("%09d", i)))
		ensure.Nil(t, err)
		ensure.DeepEqual(t, data, []byte("Put_Value"+fmt.Sprintf("%09d", i)))
	}
	fmt.Printf("Get %v k-v spend time : %v \n", putTimes, time.Since(start).Seconds())

	batchTimes := 100000
	cassandraBatch := DB.NewBatch()

	start = time.Now()
	for i := 0; i < batchTimes; i++ {
		key := []byte("disbatch_key" + fmt.Sprintf("%09d", i))
		value := []byte("disbatch_value" + fmt.Sprintf("%09d", i))
		err = cassandraBatch.Put(key, value)
		ensure.Nil(t, err)
		if i%1000 == 0 {
			err = cassandraBatch.Write()
			ensure.Nil(t, err)
			cassandraBatch = DB.NewBatch()
		}
	}
	err = cassandraBatch.Write()
	ensure.Nil(t, err)
	fmt.Printf("batch put %v  k-v use batch spend time  : %v \n", batchTimes, time.Since(start).Seconds())

	for i := 0; i < 10000; i++ {
		data, err = DB.Get([]byte("disbatch_key" + fmt.Sprintf("%09d", i)))
		ensure.Nil(t, err)
		ensure.DeepEqual(t, data, []byte("disbatch_value"+fmt.Sprintf("%09d", i)))
	}

	start = time.Now()
	for i := 0; i < putTimes; i++ {
		_, err = DB.Get([]byte("Nil_key" + fmt.Sprintf("%09d", i)))
		ensure.DeepEqual(t, err, dbUtils.DB_NOT_FOUND)
	}
	fmt.Printf("Get %v nil k-v spend time : %v \n", putTimes, time.Since(start).Seconds())

	levelDB, err := leveldb.OpenFile("./build/testdb", nil)

	start = time.Now()
	for i := 0; i < putTimes; i++ {
		err = levelDB.Put([]byte("Put_key"+fmt.Sprintf("%09d", i)), []byte("Put_Value"+fmt.Sprintf("%09d", i)), nil)
		ensure.Nil(t, err)
	}
	fmt.Printf("leveldb Put %v k-v spend time : %v \n", putTimes, time.Since(start).Seconds())

	start = time.Now()
	for i := 0; i < putTimes; i++ {
		data, err = levelDB.Get([]byte("Put_key"+fmt.Sprintf("%09d", i)), nil)
		ensure.Nil(t, err)
		ensure.DeepEqual(t, data, []byte("Put_Value"+fmt.Sprintf("%09d", i)))
	}
	fmt.Printf("leveldb Get %v k-v spend time : %v \n", putTimes, time.Since(start).Seconds())

	ldbbatch := new(leveldb.Batch)

	start = time.Now()
	for i := 0; i < batchTimes; i++ {
		key := []byte("disbatch_key" + fmt.Sprintf("%09d", i))
		value := []byte("disbatch_value" + fmt.Sprintf("%09d", i))
		ldbbatch.Put(key, value)
	}
	levelDB.Write(ldbbatch, nil)
	fmt.Printf("leveldb batch put %v  k-v spend time  : %v \n", batchTimes, time.Since(start).Seconds())

	iterMap := make(map[string][]byte)
	cassandraBatch = DB.NewBatch()
	for i := 0; i < putTimes; i++ {
		key := []byte("disIter_key" + fmt.Sprintf("%09d", i))
		value := []byte("disIter_value" + fmt.Sprintf("%09d", i))
		err = cassandraBatch.Put(key, value)
		ensure.Nil(t, err)
		iterMap[string(key)] = value
	}
	err = cassandraBatch.Write()
	ensure.Nil(t, err)

	time.Sleep(1 * time.Second)

	start = time.Now()
	iterator := DB.NewIterator([]byte("disIter_key"))
	for iterator.Next() {
		iterator.Key()
		iterator.Value()
	}
	fmt.Printf("iterator  %v  k-v spend time  : %v \n", putTimes, time.Since(start).Seconds())
	iterator.Release()

	iterator = DB.NewIterator([]byte("disIter_key"))

	for iterator.Next() {
		ensure.DeepEqual(t, iterMap[string(iterator.Key())], iterator.Value())
		delete(iterMap, string(iterator.Key()))
	}
	iterator.Release()
	ensure.DeepEqual(t, len(iterMap), 0)

}
