package hyperdb

import (
	"bytes"
	"fmt"
	Lru "github.com/hashicorp/golang-lru"
	"strconv"
	"testing"
)

var lrDb *levelLruDatabase
var level_db *LDBDatabase
var err error

func init() {
	lrDb, err = NewLevelLruDB()
	if err != nil {
		fmt.Println("NewLevelLruDB() faild")
	}
	level_db, err = NewLDBDataBase("./build/leveldbForTestLevelDb")
	if err != nil {
		fmt.Println("NewLDBDataBase faild")
	}
}

func Test1(t *testing.T) {
	cache, _ := Lru.New(10)
	status := cache.Add("key111111111111", []byte("value1"))
	fmt.Println(status)
	data, _ := cache.Get("key111111111111")
	data1 := Bytes(data)
	fmt.Println(string(data1))

	data, status = cache.Get("key2")

	fmt.Println(status)
	fmt.Println(data)
}

func TestLevelLruBatch(t *testing.T) {
	Db, err := NewLevelLruDBWithP("./build/leveltest", 20000)

	if err != nil {
		t.Error("NewLevelLruDB() faild")
	}

	times := 20000
	for i := 0; i < times; i++ {
		Db.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	for i := 0; i < times; i++ {
		value, err := Db.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			t.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}

	for i := 0; i < times; i++ {
		value, err := Db.leveldb.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			t.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}

}

//因为要等leveldb go 结束 所以时间长
func BenchmarkLevelLruDatabase_Put(b *testing.B) {

	for i := 0; i < b.N; i++ {
		lrDb.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	for i := 0; i < b.N; i++ {
		value, err := lrDb.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			b.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}

	for i := 0; i < b.N; i++ {
		value, err := lrDb.leveldb.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			b.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}

	for i := 0; i < b.N; i++ {
		value1, status := lrDb.cache.Get((strconv.Itoa(i)))
		if !status {
			b.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		value := Bytes(value1)
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			b.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}
}

//put 300000	      		5103 ns/op
//get 200000	      		5205 ns/op
//put get 200000	        7411 ns/op
//因为put get 一起以后是在内存中的命中率高
func BenchmarkPut(b *testing.B) {

	for i := 0; i < b.N; i++ {
		level_db.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	for i := 0; i < b.N; i++ {
		value, err := level_db.Get([]byte(strconv.Itoa(i)))
		if err != nil {
			b.Error("db.get fail with K :" + strconv.Itoa(i) + " error: " + err.Error())
		}
		if !bytes.Equal(value, []byte(strconv.Itoa(i))) {
			b.Error("the value from db is not correct. the suppose is " + strconv.Itoa(i) + " and the return is " + string(value))
		}
	}
}
