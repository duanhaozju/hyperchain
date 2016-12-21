//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"path"
	"strconv"
	"sync"
	//"github.com/syndtr/goleveldb/leveldb"
	"github.com/op/go-logging"

)

type stateldb int32

type LDBInstance struct {
	ldb    *LDBDatabase
	state  stateldb
	dbsync sync.Mutex
}

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("hyperdb")
}

const (
	closed stateldb = iota
	opened
)

var ldbInstance = &LDBInstance{
	state: closed,
}

//-- --------------- about ldb -----------------------

func getBaseDir() string {
	//path := os.TempDir()
	return "/tmp"
}

var (
	baseLDBPath = getBaseDir() + "/hyperchain/cache"
	portLDBPath = "db" //different port has different db path, default "db"
)

func SetLDBPath(dbpath string, port int) {
	baseLDBPath = path.Join(dbpath, "hyperchain/cache/")
	portLDBPath = strconv.Itoa(port)
}

func getDBPath() string {
	return baseLDBPath + portLDBPath
}

//-- ------------------ ldb end ----------------------

 //GetLDBDatabase get a single instance of LDBDatabase
 //if LDBDatabase state is open, return db directly
 //if LDBDatabase state id close,
//func GetLDBDatabase() (*LDBDatabase, error) {
//	ldbInstance.dbsync.Lock()
//	defer ldbInstance.dbsync.Unlock()
//	if ldbInstance.state == opened {
//		return ldbInstance.ldb, nil
//	}
//	db, err := leveldb.OpenFile(getDBPath(), nil)
//	if err != nil {
//		return ldbInstance.ldb, err
//	}
//	ldbInstance.ldb = &LDBDatabase{
//		db:   db,
//		path: getDBPath(),
//	}
//	ldbInstance.state = opened
//	return ldbInstance.ldb, nil
//}


/////////////////////////////////////////////////////redis/
func GetLDBDatabase() (*LDBDatabase, error) {
	ldbInstance.dbsync.Lock()
	defer ldbInstance.dbsync.Unlock()
	if ldbInstance.state == opened {
		return ldbInstance.ldb, nil
	}

	db,err:= NewLDBDatabase(portLDBPath)
	if err!=nil{
		return nil, err
	}
	ldbInstance.ldb=db
	ldbInstance.state = opened
	return db, nil
}


//func GetLDBDatabase() (*LDBDatabase, error) {
//	return NewLDBDatabase(portLDBPath)
//}

//////////////////////////////////////////ssdb
//func GetLDBDatabase() (*LDBDatabase, error) {
//	ldbInstance.dbsync.Lock()
//	defer ldbInstance.dbsync.Unlock()
//	if ldbInstance.state == opened {
//		return ldbInstance.ldb, nil
//	}
//
//	db,err:= NewLDBDatabase(portLDBPath,"127.0.0.1")
//	if err!=nil{
//		return nil, err
//	}
//	ldbInstance.ldb=db
//	ldbInstance.state = opened
//	return db, nil
//}