//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"hyperchain/hyperdb/hleveldb"
	"hyperchain/hyperdb/sldb"
	"strconv"
	"sync"
)

var (
	logPath   = ""
	logStatus = false
	dbType    = 0001
)

type stateDb int32

type DBInstance struct {
	db    db.Database
	state stateDb
}

const (
	db_type = "dbConfig.dbType"

	// database type
	ldb_db         = 0001
	super_level_db = 0010

	// namespace
	default_namespace = "global"
	blockchain        = "blockchain"
	consensus         = "consensus"
	archive           = "archive"

	// state
	closed stateDb = iota
	opened
)

type DbManager struct {
	dbMap  map[string]*DBInstance
	dbSync *sync.RWMutex
}

var dbMgr *DbManager

func init() {
	dbMgr = &DbManager{
		dbMap:  make(map[string]*DBInstance),
		dbSync: new(sync.RWMutex),
	}
}

func SetDBConfig(dbConfig string, port string) {
	config := viper.New()
	viper.SetEnvPrefix("DBCONFIG_ENV")
	config.SetConfigFile(dbConfig)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error envPre %s reading %s", "dbConfig", err))
	}
	dbType = config.GetInt(db_type)

	logStatus = config.GetBool("dbConfig.logStatus")
	logPath = config.GetString("dbConfig.logPath")

}

func InitDatabase(conf *common.Config, namespace string) error {
	log := getLogger(namespace)
	log.Criticalf("init db for namespace %s", namespace)
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
	_, ok := dbMgr.dbMap[getDbName(namespace)]

	if ok {
		log.Notice("Try to init inited db " + namespace)
		return errors.New("Try to init inited db " + namespace)
	}

	db, err := NewDatabase(conf, "blockchain", dbType, namespace)
	if err != nil {
		log.Errorf(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
		log.Error(err.Error())
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
	}

	archieveDb, err := NewDatabase(conf, "archive", 0001, namespace)

	if err != nil {
		log.Errorf(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
		log.Error(err.Error())
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
	}

	db1, err1 := NewDatabase(conf, "Consensus", 0001, namespace)

	if err1 != nil {
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
	}

	dbMgr.dbMap[getDbName(namespace)] = &DBInstance{
		state: opened,
		db:    db,
	}

	dbMgr.dbMap[getConsensusDbName(namespace)] = &DBInstance{
		state: opened,
		db:    db1,
	}

	dbMgr.dbMap[getArchiveDbName(namespace)] = &DBInstance{
		state: opened,
		db:    archieveDb,
	}

	return err
}

//StopDatabase start the namespace by namespace.
func StartDatabase(conf *common.Config, namespace string) error {
	return InitDatabase(conf, namespace)
}

//StopDatabase close the database name by namespace.
func StopDatabase(namespace string) error {
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
	if db, ok := dbMgr.dbMap[getDbName(namespace)]; ok {
		db.db.Close()
		delete(dbMgr.dbMap, getDbName(namespace))
	}

	if db, ok := dbMgr.dbMap[getConsensusDbName(namespace)]; ok {
		db.db.Close()
		delete(dbMgr.dbMap, getConsensusDbName(namespace))
	}

	if db, ok := dbMgr.dbMap[getArchiveDbName(namespace)]; ok {
		db.db.Close()
		delete(dbMgr.dbMap, getArchiveDbName(namespace))
	}
	return nil
}

//GetDBDatabase get db database by default namespace
func GetDBDatabase() (db.Database, error) {
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()
	if dbMgr.dbMap[getDbName(default_namespace)].db == nil {
		log := getLogger(default_namespace)
		log.Notice("GetDBDatabase() fail beacause dbMgr[GlobalBlockchain] has not been inited \n")
		return nil, errors.New("GetDBDatabase() fail beacause dbMgr[GlobalBlockchain] has not been inited \n")
	}
	return dbMgr.dbMap[getDbName(default_namespace)].db, nil
}

func GetDBDatabaseByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()
	name := getDbName(namespace)
	if _, ok := dbMgr.dbMap[name]; !ok {
		log.Notice(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}

	if dbMgr.dbMap[name].db == nil {
		log.Notice(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}
	return dbMgr.dbMap[name].db, nil
}

func GetArchiveDbByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()
	name := getArchiveDbName(namespace)
	if _, ok := dbMgr.dbMap[name]; !ok {
		log.Notice(fmt.Sprintf("GetDBArchiveByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBArchiveByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
	}

	if dbMgr.dbMap[name].db == nil {
		log.Notice(fmt.Sprintf("GetDBArchiveByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBArchiveByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
	}
	return dbMgr.dbMap[name].db, nil
}

func GetDBConsensusByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.RLock()
	defer dbMgr.dbSync.RUnlock()
	name := getConsensusDbName(namespace)
	if _, ok := dbMgr.dbMap[name]; !ok {
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
	}

	if dbMgr.dbMap[name].db == nil {
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespace fail beacause dbMgr[%v] has not been inited \n", namespace))
	}
	return dbMgr.dbMap[name].db, nil
}

func GetDatabaseHome(conf *common.Config) string {
	dbType := conf.GetInt(db_type)
	switch dbType {
	case ldb_db:
		return hleveldb.LDBDataBasePath(conf)
	case super_level_db:
		return sldb.SLDBPath(conf)
	default:
		return ""
	}
}

func GetDatabaseType(conf *common.Config) int {
	return conf.GetInt(db_type)
}

func NewDatabase(conf *common.Config, path string, dbType int, namespace string) (db.Database, error) {
	switch dbType {
	case ldb_db:
		return hleveldb.NewLDBDataBase(conf, path, namespace)
	case super_level_db:
		return sldb.NewSLDB(conf, path, namespace)
	default:
		return nil, errors.New("Wrong dbType:" + strconv.Itoa(dbType))
	}
}

func IfLogStatus() bool {
	return logStatus
}

func GetLogPath() string {
	return logPath
}

//getConsensusDbName get consensus db composite name by namespace.
func getConsensusDbName(namespace string) string {
	return namespace + consensus
}

//getDbName get ordinary db composite name by namespace.
func getDbName(namespace string) string {
	return namespace + blockchain
}

//getDbName get ordinary db composite name by namespace.
func getArchiveDbName(namespace string) string {
	return namespace + archive
}

func getLogger(namespace string) *logging.Logger {
	return common.GetLogger(namespace, "hyperdb")
}
