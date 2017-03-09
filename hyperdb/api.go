//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"strconv"
	"sync"
	"path/filepath"
	"hyperchain/hyperdb/sldb"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"hyperchain/hyperdb/hleveldb"
)
//TODO: refactor this file as soon as possible
var (
	logPath   = ""
	logStatus = false
	dbType = 001
	grpcPort             = 8001
	leveldbPath          = "./build/leveldb"
)

type stateDb int32

type DBInstance struct {
	db     db.Database
	state  stateDb
}

const (
	DB_TYPE = "dbConfig.dbType"

	// database type
	LDB_DB = 0001
	SUPER_LEVEL_DB = 0010

	// namespace
	DefautNameSpace="Global"
	Blockchain="Blockchain"
	Consensus="Consensus"

	// state
	closed stateDb = iota
	opened
)

var log *logging.Logger // package-level logger
type DbMap struct{
	dbMap map[string] *DBInstance
	dbSync  sync.Mutex
}

var  dbMap *DbMap

func init() {
	log = logging.MustGetLogger("hyperdb")
	dbMap=&DbMap{
		dbMap:make(map[string] *DBInstance),
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
	dbType = config.GetInt(DB_TYPE)

	logStatus = config.GetBool("dbConfig.logStatus")
	logPath = config.GetString("dbConfig.logPath")

	leveldbPath = config.GetString("dbConfig.leveldbPath")
	grpcPort, _ = strconv.Atoi(port)

}

func IfLogStatus() bool {
	return logStatus
}

func GetLogPath() string {
	return logPath
}

func InitDatabase(conf *common.Config, nameSpace string) error {

	dbMap.dbSync.Lock()
	defer dbMap.dbSync.Unlock()
	_,ok:=dbMap.dbMap[nameSpace+Blockchain]

	if ok{
		log.Notice("Try to init inited db "+nameSpace)
		return errors.New("Try to init inited db "+nameSpace)
	}

	db, err := NewDatabase(conf, filepath.Join(leveldbPath,nameSpace,"Blockchain"), dbType)


	if err!=nil{
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", nameSpace))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", nameSpace))
	}

	db1, err1 := NewDatabase(conf, filepath.Join(leveldbPath,nameSpace,"Consensus" ), 001)

	if err1 != nil {

		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", nameSpace))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", nameSpace))
	}

	dbMap.dbMap[nameSpace+Blockchain]=&DBInstance{
		state: opened,
		db:db,
	}

	dbMap.dbMap[nameSpace+Consensus]=&DBInstance{
		state: opened,
		db:db1,
	}

	return err
}

func GetDBDatabase() (db.Database, error) {
	dbMap.dbSync.Lock()
	defer dbMap.dbSync.Unlock()
	if dbMap.dbMap[DefautNameSpace+Blockchain].db == nil {
		log.Notice("GetDBDatabase() fail beacause dbMap[GlobalBlockchain] has not been inited \n")
		return nil, errors.New("GetDBDatabase() fail beacause dbMap[GlobalBlockchain] has not been inited \n")
	}
	return dbMap.dbMap[DefautNameSpace+Blockchain].db, nil
}

func GetDBDatabaseConsensus() (db.Database, error) {
	dbMap.dbSync.Lock()
	defer dbMap.dbSync.Unlock()
	if dbMap.dbMap[DefautNameSpace+Consensus].db == nil {
		log.Notice("GetDBDatabaseConsensus()  fail beacause dbMap[GlobalConsensus] has not been inited \n")
		return nil, errors.New("GetDBDatabaseConsensus()  fail beacause dbMap[GlobalConsensus] has not been inited \n")
	}
	return dbMap.dbMap[DefautNameSpace+Consensus].db, nil
}

func GetDBDatabaseByNamespace(namespace string)(db.Database, error){
	dbMap.dbSync.Lock()
	defer dbMap.dbSync.Unlock()

	namespace += Consensus

	if _,ok:=dbMap.dbMap[namespace];!ok{
		log.Notice(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
		return nil, errors.New(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
	}

	if dbMap.dbMap[namespace].db == nil {
		log.Notice(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
		return nil, errors.New(fmt.Sprintf("GetDBDatabaseByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
	}
	return dbMap.dbMap[namespace].db, nil
}

func GetDBConsensusByNamespcae(namespace string)(db.Database, error){
	dbMap.dbSync.Lock()
	defer dbMap.dbSync.Unlock()

	namespace+=Consensus

	if _,ok:=dbMap.dbMap[namespace];!ok{
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
	}

	if dbMap.dbMap[namespace].db == nil {
		log.Notice(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
		return nil, errors.New(fmt.Sprintf("GetDBConsensusByNamespcae fail beacause dbMap[%v] has not been inited \n",namespace))
	}
	return dbMap.dbMap[namespace].db, nil
}

func NewDatabase(conf *common.Config, path string, dbType int) (db.Database, error) {
	switch dbType {
	case LDB_DB:
		log.Notice("Use level db only")
		return hleveldb.NewLDBDataBase(path)
	case SUPER_LEVEL_DB:
		log.Notice("Use SuperLevelDB")
		return sldb.NewSLDB(conf)
	default:
		log.Errorf("Wrong dbType:" + strconv.Itoa(dbType))
		return nil, errors.New("Wrong dbType:" + strconv.Itoa(dbType))
	}
}
