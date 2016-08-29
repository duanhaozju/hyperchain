package hyperdb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

type stateldb int32

type LDBInstance struct {
	ldb    *LDBDatabase
	state  stateldb
	dbsync sync.Mutex
}

const (
	closed stateldb = iota
	opened
)

var ldbInstance = &LDBInstance{
	state: closed,
}

func getDBPath() string {
	return "cache/db"
}

// GetLDBDatabase get a single instance of LDBDatabase
// if LDBDatabase state is open, return db directly
// if LDBDatabase state id close,
func GetLDBDatabase() (*LDBDatabase, error) {
	ldbInstance.dbsync.Lock()
	defer ldbInstance.dbsync.Unlock()
	if ldbInstance.state == opened {
		return ldbInstance.ldb, nil
	}
	db, err := leveldb.OpenFile(getDBPath(), nil)
	if err != nil {
		return ldbInstance.ldb, err
	}
	ldbInstance.ldb = &LDBDatabase{
		db: db,
		path: getDBPath(),
	}
	ldbInstance.state = opened
	return ldbInstance.ldb, nil
}