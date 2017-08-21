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
)

var conf *common.Config

func setupDB()  {
	conf = common.NewRawConfig()
	conf.Set(common.NAMESPACE, common.DEFAULT_NAMESPACE)
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "/tmp/hyperdb")
}

func clearDB()  {
	dbmgr.clearMemNDBSs()
	os.RemoveAll("/tmp/hyperdb")
}

func init()  {
	setupDB()
	common.InitHyperLoggerManager(conf)
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
		fmt.Printf("create db%d......\n", i)
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




	clearDB()
}