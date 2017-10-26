//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	hcomm "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var conf *common.Config

func setupDB() {
	conf = common.NewRawConfig()
	conf.Set(common.NAMESPACE, common.DEFAULT_NAMESPACE)
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "/tmp/hyperdb")
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
	common.SetLogLevel(common.DEFAULT_NAMESPACE, "dbmgr", "debug")
	InitDBMgr(conf)
}

func clearDB() {
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
	db, err := ConnectToDB(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	db.Put([]byte("a"), []byte("b"))
	db2, err := ConnectToDB(&DbName{
		name:      "db2",
		namespace: common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	db2.Put([]byte("a"), []byte("b"))
	clearDB()
}

func TestGetDB(t *testing.T) {
	setupDB()
	for i := 0; i < 10; i++ {
		dbName := &DbName{
			name:      fmt.Sprintf("db%d", i),
			namespace: common.DEFAULT_NAMESPACE,
		}
		db, err := ConnectToDB(dbName, conf)
		if err != nil {
			t.Errorf("create db error: %v", err)
		}
		k, v := []byte("a"), []byte("b")
		db.Put(k, v)
		db2, err := GetDB(dbName)
		if err != nil {
			t.Errorf("create db error: %v", err)
		}
		v2, err := db2.Get(k)
		assert.Equal(t, v, v2)
	}
	clearDB()
}

func TestMultiThreadDbOps(t *testing.T) {
	setupDB()
	var x int32 = -1
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			id := atomic.AddInt32(&x, 1)
			logger.Infof("create db%d......", id)
			dbName := &DbName{
				name:      fmt.Sprintf("db%d", id),
				namespace: common.DEFAULT_NAMESPACE,
			}
			db, err := ConnectToDB(dbName, conf)
			if err != nil {
				t.Errorf("create db error: %v", err)
			}
			db.Put([]byte("a"), []byte("b"))
			wg.Done()
		}()

	}
	wg.Wait()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			id := rand.Intn(10)
			dbName := &DbName{
				name:      fmt.Sprintf("db%d", id),
				namespace: common.DEFAULT_NAMESPACE,
			}
			//fmt.Printf("db name: %s \n", dbName.Name)
			db, err := GetDB(dbName)
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
	_, err := GetDB(&DbName{
		name:      "sss",
		namespace: common.DEFAULT_NAMESPACE,
	})

	if err == nil {
		t.Error("no db but found db")
		return
	}
	clearDB()
}

func TestClose(t *testing.T) { //TODO: add more checks  add db root dir not database dir
	setupDB()
	_, err := ConnectToDB(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}
	Close()

	_, err = GetDB(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	})
	time.Sleep(1 * time.Second)
	assert.Equal(t, ErrDbNotExisted, err)

	_, err = ConnectToDB(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	}, conf)

	if err != nil {
		t.Errorf("create db error: %v", err)
	}

	CloseByName(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	})

	_, err = GetDB(&DbName{
		name:      "db1",
		namespace: common.DEFAULT_NAMESPACE,
	})
	time.Sleep(1 * time.Second)
	assert.Equal(t, ErrDbNotExisted, err)
	clearDB()
}
