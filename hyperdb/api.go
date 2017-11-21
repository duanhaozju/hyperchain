//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"sync"

	"github.com/op/go-logging"

	"github.com/hyperchain/hyperchain/common"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/hyperdb/hleveldb"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
)

var (
	logPath           = ""
	measurementEnable = false
	dbType            = 0001
)

const (
	// namespace
	defaultNamespace = "global"

	// states
	closed stateDb = iota
	opened
)

// stateDb represents db state.
type stateDb int32

type DBInstance struct {
	db    db.Database
	state stateDb
}

type DbManager struct {
	dbMap  map[string]*DBInstance
	dbSync *sync.RWMutex
}

var dbMgr *DbManager

// init initializes dbMgr here.
func init() {
	dbMgr = &DbManager{
		dbMap:  make(map[string]*DBInstance),
		dbSync: new(sync.RWMutex),
	}
}

// InitDatabase initiates databases by a specific namespace
func InitDatabase(conf *common.Config, namespace string) error {
	//TODO: refactor this
	dbType = conf.GetInt(hcom.DB_TYPE)
	measurementEnable = conf.GetBool("database.leveldb.measurement_enable")
	logPath = common.GetPath(namespace, conf.GetString("database.leveldb.log_path"))

	log := getLogger(namespace)
	log.Criticalf("init db for namespace %s", namespace)

	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()

	name := getDbNameByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	_, exists := dbMgr.dbMap[name]
	if exists {
		msg := "Try to init an inited db " + namespace
		log.Notice(msg)
		return errors.New(msg)
	}

	for idx := 0; idx < hcom.DBINDEX_MAX; idx++ {
		dbname := getDbName(idx)
		db, err := NewDatabase(conf, dbname, dbType, namespace)
		if err != nil {
			msg := fmt.Sprintf("InitDatabase(%v):%v failed because it can't get a new database\n", namespace, dbname)
			log.Errorf(msg)
			return err
		}

		name := getDbNameByNamespace(namespace, dbname)
		dbMgr.dbMap[name] = &DBInstance{
			state: opened,
			db:    db,
		}
	}

	return nil
}

// NewDatabase returns a new database instance with a dbname, a dbType and a namespace.
func NewDatabase(conf *common.Config, dbname string, dbType int, namespace string) (db.Database, error) {
	switch dbType {
	case hcom.LDB_DB:
		dbpath := getDbPath(conf, namespace, dbname)
		return hleveldb.NewLDBDataBase(conf, dbpath, namespace)
	case hcom.MEMORY_DB:
		return mdb.NewMemDatabase(namespace)
	default:
		return nil, errors.New("Wrong dbType:" + strconv.Itoa(dbType))
	}
}

func getDbPath(conf *common.Config, namespace, dbname string) string {
	switch dbname {
	case hcom.DBNAME_BLOCKCHAIN:
		return path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_BLOCKCHAIN)))
	case hcom.DBNAME_CONSENSUS:
		return path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_CONSENSUS)))
	case hcom.DBNAME_ARCHIVE:
		return path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_ARCHIVE)))
	default:
		return path.Join(common.GetPath(namespace, conf.GetString(hcom.LEVEL_DB_ROOT_DIR)), dbname)
	}
}

// StartDatabase starts a database by namespace.
func StartDatabase(conf *common.Config, namespace string) error {
	return InitDatabase(conf, namespace)
}

// StopDatabase closes a database by namespace.
func StopDatabase(namespace string) error {
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()

	for idx := 0; idx < hcom.DBINDEX_MAX; idx++ {
		dbname := getDbName(idx)
		name := getDbNameByNamespace(namespace, dbname)
		if db, exists := dbMgr.dbMap[name]; exists {
			db.db.Close()
			delete(dbMgr.dbMap, name)
		}
	}

	return nil
}

// GetDBDatabase gets an ordinary database by the default namespace.
func GetDBDatabase() (db.Database, error) {
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getDbNameByNamespace(defaultNamespace, hcom.DBNAME_BLOCKCHAIN)

	if dbMgr.dbMap[name].db == nil {
		log := getLogger(defaultNamespace)
		msg := "GetDBDatabase() failed because dbMgr[GlobalBlockchain] has not been inited \n"
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[name].db, nil
}

// GetDBDatabaseByNamespace gets a database instance by namespace and dbname.
func GetDBDatabaseByNamespace(namespace, dbname string) (db.Database, error) {
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getDbNameByNamespace(namespace, dbname)

	if db, exists := dbMgr.dbMap[name]; db == nil || !exists {
		var msg string
		log := getLogger(namespace)
		if db == nil {
			msg = fmt.Sprintf("GetDBDatabaseByNamespace failed because dbMgr[%v] has not been inited \n", namespace)
		} else {
			msg = fmt.Sprintf("GetDBDatabaseByNamespace failed because database %v/%v does not exist\n", namespace, name)
		}
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[name].db, nil
}

// GetDatabaseHome returns database's storage root directory.
func GetDatabaseHome(conf *common.Config) string {
	dbType := conf.GetInt(hcom.DB_TYPE)
	switch dbType {
	case hcom.LDB_DB:
		return hleveldb.LDBDatabasePath(conf)
	default:
		return ""
	}
}

// GetDatabaseType returns database type in integer.
func GetDatabaseType(conf *common.Config) int {
	return conf.GetInt(hcom.DB_TYPE)
}

// IfLogStatus is the switch of measurement log.
func IfLogStatus() bool {
	return measurementEnable
}

// GetLogPath gets the database's log file path.
func GetLogPath() string {
	return logPath
}

// getDbName returns dbname with its corresponding index.
func getDbName(dbindex int) string {
	switch dbindex {
	case hcom.DBINDEX_BLOCKCHAIN:
		return hcom.DBNAME_BLOCKCHAIN
	case hcom.DBINDEX_CONSENSUS:
		return hcom.DBNAME_CONSENSUS
	case hcom.DBINDEX_ARCHIVE:
		return hcom.DBNAME_ARCHIVE
	default:
		return "default"
	}
}

// getDbNameByNamespace returns a dbname in a specific namespace.
func getDbNameByNamespace(namespace, dbname string) string {
	return namespace + dbname
}

// getLogger gets a Logger by namespace.
func getLogger(namespace string) *logging.Logger {
	// common.GetLogger gets a Logger with a specific namespace and a module.
	return common.GetLogger(namespace, "hyperdb")
}
