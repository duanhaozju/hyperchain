//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"testing"
	"hyperchain/common"
	hcomm "hyperchain/hyperdb/common"
	"fmt"
	"reflect"
	"os"
	"sync/atomic"
	"sync"
	"math/rand"
)

var conf *common.Config

func setupDB()  {
	conf = common.NewRawConfig()
	conf.Set(common.NAMESPACE, common.DEFAULT_NAMESPACE)
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "/tmp/hyperdb")
	common.InitHyperLoggerManager(conf)
	InitDBMgr(conf)
}

func clearDB()  {
	dbmgr.clearMemNDBSs()
	os.RemoveAll("/tmp/hyperdb")
}

func TestInitDBMgr(t *testing.T) {
	setupDB()
	conf := common.NewRawConfig()
	InitDBMgr(conf)
}

func TestCreateDB(t *testing.T) {
	setupDB()
	err, db := CreateDB(&DbName{
		Name:"db1",
		Namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	db.Put([]byte("a"), []byte("b"))
	err, db2 := CreateDB(&DbName{
		Name:"db2",
		Namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	db2.Put([]byte("a"), []byte("b"))
	clearDB()
}

func TestGetDB(t *testing.T) {
	setupDB()
	for i := 0; i < 10; i ++ {
		dbName := &DbName{
			Name:fmt.Sprintf("db%d", i),
			Namespace:common.DEFAULT_NAMESPACE,
		}
		err, db := CreateDB(dbName, conf)
		if err != nil {
			t.Errorf("create db error: %v", err)
		}
		k, v := []byte("a"), []byte("b")
		db.Put(k, v)
		err, db2 := GetDB(dbName)
		if err != nil {
			t.Errorf("create db error: %v", err)
		}
		v2, err := db2.Get(k)
		if reflect.DeepEqual(v, v2) != true {
			t.Errorf("put and get data not equal")
		}
	}
	clearDB()
}

func TestMultiThreadDbOps(t *testing.T)  {
	setupDB()
	var x int32 = -1
	var wg sync.WaitGroup
	for i := 0; i < 10; i ++ {
		wg.Add(1)
		go func() {
			id := atomic.AddInt32(&x, 1)
			fmt.Printf("create db%d......\n", id)
			dbName := &DbName{
				Name:fmt.Sprintf("db%d", id),
				Namespace:common.DEFAULT_NAMESPACE,
			}
			err, db := CreateDB(dbName, conf)
			if err != nil {
				t.Errorf("create db error: %v", err)
			}
			db.Put([]byte("a"), []byte("b"))
			wg.Done()
		}()

	}
	wg.Wait()

	for i := 0; i < 100; i ++ {
		wg.Add(1)
		go func() {
			id := rand.Intn(10)
			dbName := &DbName{
				Name:fmt.Sprintf("db%d", id),
				Namespace:common.DEFAULT_NAMESPACE,
			}
			//fmt.Printf("db name: %s \n", dbName.Name)
			err, db := GetDB(dbName)
			if err != nil {
				t.Errorf("get db error: %v", err)
			}
			k, v := []byte("a"), []byte("b")
			db.Put(k, v)
			v2, err := db.Get(k)
			if err != nil {
				t.Errorf("get db error: %v", err)
			}
			if reflect.DeepEqual(v, v2) != true {
				t.Errorf("put and get data not equal")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	clearDB()
}

func TestGetDB2(t *testing.T) {
	setupDB()
	err, _ := GetDB(&DbName{
		Name:"sss",
		Namespace:common.DEFAULT_NAMESPACE,
	})

	if err == nil {
		t.Error("no db but found db")
		return
	}
	clearDB()
}

func TestClose(t *testing.T) {//TODO: add more checks
	setupDB()
	err, _ := CreateDB(&DbName{
		Name:"db1",
		Namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}
	Close()

	err, db := GetDB(&DbName{
		Name:"db1",
		Namespace:common.DEFAULT_NAMESPACE,
	})

	db.Put([]byte("11"), []byte("11"))
	t.Error(err)
	clearDB()
}