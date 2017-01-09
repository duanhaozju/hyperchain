package hyperdb

import (
	"testing"

	"strconv"
	"fmt"
	"bytes"
)

var Db Database
var ReDb Database
var SSDB Database
func init(){
	var err error
	Db, err= NewRdSdDb("8001", 2)
	if err != nil {
		fmt.Println("NewRdSdDb fail")
	}

	SSDB, err= NewSSDatabase("8001", 2)
	if err != nil {
		fmt.Println("NewSSDatabase fail")
	}

	ReDb,err=NewRsDatabase("8001")
	if err != nil {
		fmt.Println("NewRsDatabase fail")
	}
	logPath="./db.log"
	logStatus=true
}
func TestBatchWrite(t *testing.T) {

	batch := Db.NewBatch()
	times:=2000
	for i := 0; i < times; i++ {
		batch.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	if err := batch.Write(); err != nil {
		fmt.Println("batch.Write fail with " + err.Error())
		t.Error("batch.Write fail with " + err.Error())
	}

	for i := 0; i < times; i++ {
		value,err:=Db.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}

	for i := 0; i < times; i++ {
		value,err:=Db.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}

	for i := 0; i < times; i++ {
		value,err:=ReDb.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}

}

func TestDBPut(t *testing.T){
	times:=200
	for i := 0; i < times; i++ {
		Db.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	for i := 0; i < times; i++ {
		value,err:=Db.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}


	for i := 0; i < times; i++ {
		value,err:=SSDB.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}

	for i := 0; i < times; i++ {
		value,err:=ReDb.Get([]byte(strconv.Itoa(i)))
		if err!=nil{
			t.Error("db.get fail with K :"+strconv.Itoa(i)+" error: "+err.Error())
		}
		if  !bytes.Equal(value,[]byte(strconv.Itoa(i))) {
			t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(i)+" and the return is "+string(value))
		}
	}
}


func TestDBDelete(t *testing.T){
	value,err:=Db.Get([]byte{'1'})
	if err!=nil{
		t.Error("db.get fail with K :"+strconv.Itoa(1)+" error: "+err.Error())
	}
	if  !bytes.Equal(value,[]byte(strconv.Itoa(1))) {
		t.Error("the value from db is not correct. the suppose is "+strconv.Itoa(1)+" and the return is "+string(value))
	}
	err=Db.Delete([]byte{'1'})

	if err!=nil{
		t.Error("db.delete fail with K :"+strconv.Itoa(1)+" error: "+err.Error())
	}

	value,err=Db.Get([]byte{'1'})
	if err==nil||err.Error()!="not found"{
		fmt.Println(err)
		fmt.Println(value)
		t.Error("db.delete fail with K :"+strconv.Itoa(1))
	}
}

func TestIterator(t *testing.T) {
	batch := Db.NewBatch()
	map1:=make(map[string][]byte)
	for i := 0; i < 200; i++ {
		key:=append([]byte("test"),strconv.Itoa(i)...)
		value:=append([]byte("testvalue"),strconv.Itoa(i)...)
		map1[string(key)]=value
		batch.Put(key, value)
	}

	if err := batch.Write(); err != nil {
		fmt.Println("batch.Write fail with " + err.Error())
		t.Error("batch.Write fail with " + err.Error())
	}

	iterator:=Db.NewIterator([]byte("test"))

	iterator.Seek([]byte("test"))
	for iterator.Next(){
		key:=iterator.Key()
		value:=iterator.Value()
		if ! bytes.Equal(map1[string(key)],value){
			t.Errorf("failed with key %d value %d map1[%d] %d \n", key,value,key,map1[string(key)])
		}
		delete(map1,string(key))
	}
	if iterator.Error()!=nil{
		t.Error("iterator.error is not nil with :"+iterator.Error().Error())
	}else{
		if len(map1)!=0{
			t.Error("len(map1) is :"+strconv.Itoa(len(map1)))
			t.Error("did not iterator all the elements which was  put in to the db ")
		}
	}

}

