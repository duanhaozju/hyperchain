//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"testing"
	"hyperchain/common"
	hcomm "hyperchain/hyperdb/common"
	"fmt"
	"os"
	"sync/atomic"
	"sync"
	"math/rand"
	"time"
	"github.com/stretchr/testify/assert"
)

var conf *common.Config

func setupDB()  {
	conf = common.NewRawConfig()
	conf.Set(common.NAMESPACE, common.DEFAULT_NAMESPACE)
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "/tmp/hyperdb")
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
	common.SetLogLevel(common.DEFAULT_NAMESPACE, "dbmgr", "debug")
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
	err, db := ConnectToDB(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	db.Put([]byte("a"), []byte("b"))
	err, db2 := ConnectToDB(&DbName{
		name:"db2",
		namespace:common.DEFAULT_NAMESPACE,
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
			name:fmt.Sprintf("db%d", i),
			namespace:common.DEFAULT_NAMESPACE,
		}
		err, db := ConnectToDB(dbName, conf)
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
		assert.Equal(t, v, v2)
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
				name:fmt.Sprintf("db%d", id),
				namespace:common.DEFAULT_NAMESPACE,
			}
			err, db := ConnectToDB(dbName, conf)
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
				name:fmt.Sprintf("db%d", id),
				namespace:common.DEFAULT_NAMESPACE,
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
			assert.Equal(t, v, v2)
			wg.Done()
		}()
	}
	wg.Wait()
	clearDB()
}

func TestGetDB2(t *testing.T) {
	setupDB()
	err, _ := GetDB(&DbName{
		name:"sss",
		namespace:common.DEFAULT_NAMESPACE,
	})

	if err == nil {
		t.Error("no db but found db")
		return
	}
	clearDB()
}

func TestClose(t *testing.T) {//TODO: add more checks  add db root dir not database dir
	setupDB()
	err, _ := ConnectToDB(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}
	Close()

	err, _ = GetDB(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	})
	time.Sleep(1 * time.Second)
	assert.Equal(t, ErrDbNotExisted, err)

	err, _ = ConnectToDB(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	CloseByName(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	})

	err, _ = GetDB(&DbName{
		name:"db1",
		namespace:common.DEFAULT_NAMESPACE,
	})
	time.Sleep(1 * time.Second)
	assert.Equal(t, ErrDbNotExisted, err)
	clearDB()
}