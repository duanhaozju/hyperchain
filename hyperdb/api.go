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

//TODO: refactor this file as soon as possible
var (
	logPath   = ""
	logStatus = false
	dbType    = 001
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
	archieve          = "Archieve"

	// state
	closed stateDb = iota
	opened
)

type DbManager struct {
	dbMap  map[string]*DBInstance
	dbSync sync.Mutex
}

var dbMgr *DbManager

func init() {
	dbMgr = &DbManager{
		dbMap: make(map[string]*DBInstance),
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

	db, err := NewDatabase(conf, "blockchain", dbType)

	if err != nil {
		log.Errorf(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
		log.Error(err.Error())
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
	}

	archieveDb, err := NewDatabase(conf, "Archieve", 001)

	if err != nil {
		log.Errorf(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
		log.Error(err.Error())
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", namespace))
	}

	db1, err1 := NewDatabase(conf, "Consensus", 001)

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

	dbMgr.dbMap[getArchieveDbName(namespace)] = &DBInstance{
		state: opened,
		db:    archieveDb,
	}

	return err
}

//CloseDatabase close the database name by namespace.
func CloseDatabase(namespace string) error {
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

	return nil
}

func GetDBDatabase() (db.Database, error) {
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
	if dbMgr.dbMap[getDbName(default_namespace)].db == nil {
		log := getLogger(default_namespace)
		log.Notice("GetDBDatabase() fail beacause dbMgr[GlobalBlockchain] has not been inited \n")
		return nil, errors.New("GetDBDatabase() fail beacause dbMgr[GlobalBlockchain] has not been inited \n")
	}
	return dbMgr.dbMap[getDbName(default_namespace)].db, nil
}

func GetDBDatabaseByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
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

func GetArchieveDbByNamespace(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
	name := getArchieveDbName(namespace)
	if _, ok := dbMgr.dbMap[name]; !ok {
		log.Notice(fmt.Sprintf("GetDBArchiveByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}

	if dbMgr.dbMap[name].db == nil {
		log.Notice(fmt.Sprintf("GetDBArchiveByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}
	return dbMgr.dbMap[name].db, nil
}

func GetDBConsensusByNamespcae(namespace string) (db.Database, error) {
	log := getLogger(namespace)
	dbMgr.dbSync.Lock()
	defer dbMgr.dbSync.Unlock()
	name := getConsensusDbName(namespace)
	if _, ok := dbMgr.dbMap[name]; !ok {
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}

	if dbMgr.dbMap[name].db == nil {
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMgr[%v] has not been inited \n", namespace))
	}
	return dbMgr.dbMap[name].db, nil
}

func NewDatabase(conf *common.Config, path string, dbType int) (db.Database, error) {
	switch dbType {
	case ldb_db:
		return hleveldb.NewLDBDataBase(conf, path)
	case super_level_db:
		return sldb.NewSLDB(conf)
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
func getArchieveDbName(namespace string) string {
	return namespace + archieve
}

func getLogger(namespace string) *logging.Logger {
	return common.GetLogger(namespace, "hyperdb")
}
