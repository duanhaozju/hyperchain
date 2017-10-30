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
	logPath   = ""
	logStatus = false
	dbType    = 0001
)

var (
	dbNameBlockchain = "blockchain"
	dbNameConsensus  = "consensus"
	dbNameArchive    = "archive"
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
	dbType = conf.GetInt(hcom.DB_TYPE)
	logStatus = conf.GetBool("database.leveldb.log_status")
	logPath = common.GetPath(namespace, conf.GetString("database.leveldb.log_path"))

	dbNameBlockchain = conf.GetString(hcom.DBNAME_BLOCKCHAIN)
	dbNameConsensus = conf.GetString(hcom.DBNAME_CONSENSUS)
	dbNameArchive = conf.GetString(hcom.DBNAME_ARCHIVE)

	log := getLogger(namespace)
	log.Criticalf("init db for namespace %s", namespace)

	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()

	_, exists := dbMgr.dbMap[getDbNameNew(namespace, dbNameBlockchain)]
	if exists {
		msg := "Try to init an inited db " + namespace
		log.Notice(msg)
		return errors.New(msg)
	}

	dbnames := []string{dbNameBlockchain, dbNameConsensus, dbNameArchive}

	for _, dbname := range dbnames {
		db, err := NewDatabase(conf, dbname, dbType, namespace)
		if err != nil {
			msg := fmt.Sprintf("InitDatabase(%v):%v failed because it can't get a new database\n", namespace, dbname)
			log.Errorf(msg)
			return err
		}

		dbMgr.dbMap[getDbNameNew(namespace, dbname)] = &DBInstance{
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
		var dbpath string
		if dbname == dbNameArchive {
			dbpath = path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_ARCHIVE)))
		} else if dbname == dbNameBlockchain {
			dbpath = path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_BLOCKCHAIN)))
		} else if dbname == dbNameConsensus {
			dbpath = path.Join(common.GetPath(namespace, conf.GetString(hcom.DBPATH_CONSENSUS)))
		} else {
			dbpath = path.Join(common.GetPath(namespace, conf.GetString(hcom.LEVEL_DB_ROOT_DIR)), dbname)
		}
		return hleveldb.NewLDBDataBase(conf, dbpath, namespace)
	case hcom.MEMORY_DB:
		return mdb.NewMemDatabase(namespace)
	default:
		return nil, errors.New("Wrong dbType:" + strconv.Itoa(dbType))
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

	dbnames := []string{
		getDbNameNew(namespace, dbNameBlockchain),
		getDbNameNew(namespace, dbNameConsensus),
		getDbNameNew(namespace, dbNameArchive)}

	for _, dbname := range dbnames {
		if db, exists := dbMgr.dbMap[dbname]; exists {
			db.db.Close()
			delete(dbMgr.dbMap, dbname)
		}
	}

	return nil
}

// GetDBDatabase gets an ordinary database by the default namespace.
func GetDBDatabase() (db.Database, error) {
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	if dbMgr.dbMap[getDbName(defaultNamespace)].db == nil {
		log := getLogger(defaultNamespace)
		msg := "GetDBDatabase() failed because dbMgr[GlobalBlockchain] has not been inited \n"
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[getDbName(defaultNamespace)].db, nil
}

// GetDBDatabaseByNamespace gets an ordinary database by namespace.
func GetDBDatabaseByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)

	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getDbName(namespace)

	if db, exists := dbMgr.dbMap[name]; db == nil || !exists {
		msg := fmt.Sprintf("GetDBDatabaseByNamespace failed because dbMgr[%v] has not been inited \n", namespace)
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[name].db, nil
}

// GetDBConsensusByNamespace gets an consensus database by namespace.
func GetDBConsensusByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)

	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getConsensusDbName(namespace)

	if db, exists := dbMgr.dbMap[name]; db == nil || !exists {
		msg := fmt.Sprintf("GetDBConsensusByNamespace failed because dbMgr[%v] has not been inited \n", namespace)
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[name].db, nil
}

// GetArchiveDbByNamespace gets an archived database by namespace.
func GetArchiveDbByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)

	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getArchiveDbName(namespace)

	if db, exists := dbMgr.dbMap[name]; db == nil || !exists {
		msg := fmt.Sprintf("GetArchiveDbByNamespace failed because dbMgr[%v] has not been inited \n", namespace)
		log.Notice(msg)
		return nil, errors.New(msg)
	}

	return dbMgr.dbMap[name].db, nil
}

// GetDBDatabaseByNamespaceNew gets a database instance by namespace and dbname.
func GetDBDatabaseByNamespaceNew(namespace, dbname string) (db.Database, error) {
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()

	name := getDbNameNew(namespace, dbname)

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
	return logStatus
}

// GetLogPath gets the database's log file path.
func GetLogPath() string {
	return logPath
}

// getDbName gets ordinary database name by namespace.
func getDbName(namespace string) string {
	return namespace + dbNameBlockchain
}

// getConsensusDbName gets consensus database name by namespace.
func getConsensusDbName(namespace string) string {
	return namespace + dbNameConsensus
}

// getArchiveDbName gets archived database name by namespace.
func getArchiveDbName(namespace string) string {
	return namespace + dbNameArchive
}

// getDbNameNew gets databases' name by namespace and dbname.
func getDbNameNew(namespace, dbname string) string {
	return namespace + dbname
}

// getLogger gets a Logger by namespace.
func getLogger(namespace string) *logging.Logger {
	// common.GetLogger gets a Logger with a specific namespace and a module.
	return common.GetLogger(namespace, "hyperdb")
}
