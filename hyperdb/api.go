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
)

var (
	logPath   = ""
	logStatus = false

	//dbType 应该是一个3位的二进制数比如111 101
	//第一个1表示 redis open
	//第二个1表示 ssdb  open
	//第三个1表示 leveldb open
	//现在仅支持 redis加 ssdb 或者仅leveldb 即 110 001
	dbType = 001

	ssdbProxyPort       = 22122
	ssdbFirstPort       = 8011
	ssdbGap             = 10
	ssdbServerNumber    = 2
	ssdbPoolSize        = 10
	ssdbTimeout         = 60
	ssdbMaxConnectTimes = 15
	ssdbIteratorSize    = 40

	redisPort            = 8101
	redisPoolSize        = 10
	redisTimeout         = 60
	redisMaxConnectTimes = 15
	grpcPort             = 8001
	leveldbPath          = "./build/leveldb"
)

type stateDb int32

type DBInstance struct {
	db     Database
	state  stateDb
	dbSync sync.Mutex
}

const (
	closed stateDb = iota
	opened
)

var log *logging.Logger // package-level logger
//dbInstance include the instance of Database interface
var dbInstance = &DBInstance{
	state: closed,
}

func init() {
	log = logging.MustGetLogger("hyperdb")
}

func setDBConfig(dbConfig string, port string) {

	config := viper.New()
	viper.SetEnvPrefix("DBCONFIG_ENV")
	config.SetConfigFile(dbConfig)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error envPre %s reading %s", "dbConfig", err))
	}

	dbType = config.GetInt("dbConfig.dbType")

	ssdbFirstPort = config.GetInt("dbConfig.ssdb.firstPort")
	ssdbGap = config.GetInt("dbConfig.ssdb.gap")
	ssdbIteratorSize = config.GetInt("dbConfig.ssdb.iteratorSize")
	ssdbMaxConnectTimes = config.GetInt("dbConfig.ssdb.maxConnectTimes")
	ssdbServerNumber = config.GetInt("dbConfig.ssdb.serverNumber")
	ssdbPoolSize = config.GetInt("dbConfig.ssdb.poolSize")
	ssdbTimeout = config.GetInt("dbConfig.ssdb.timeout")

	redisMaxConnectTimes = config.GetInt("dbConfig.redis.maxConnectTimes")
	redisPoolSize = config.GetInt("dbConfig.redis.poolSize")
	redisPort = config.GetInt("dbConfig.redis.port")
	redisTimeout = config.GetInt("dbConfig.redis.timeout")

	logStatus = config.GetBool("dbConfig.logStatus")
	logPath = config.GetString("dbConfig.logPath")

	leveldbPath = config.GetString("dbConfig.leveldbPath")
	grpcPort, _ = strconv.Atoi(port)
	leveldbPath += port

}

func IfLogStatus() bool {
	return logStatus
}

func GetLogPath() string {
	return logPath
}

func InitDatabase(dbConfig string, port string) error {

	setDBConfig(dbConfig, port)

	dbInstance.dbSync.Lock()
	defer dbInstance.dbSync.Unlock()

	if dbInstance.state != closed {
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it has beend inited \n", dbType))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it has beend inited \n", dbType))
	}
	db, err := NewDatabase()

	if err != nil {
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", dbType))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n", dbType))
	}

	dbInstance.db = db
	dbInstance.state = opened
	log.Notice("db has been init")
	return err
}

func GetDBDatabase() (Database, error) {
	dbInstance.dbSync.Lock()
	defer dbInstance.dbSync.Unlock()
	if dbInstance.db == nil {
		log.Notice("GetDBDatabase() fail beacause it has not been inited \n")
		return nil, errors.New("GetDBDatabase() fail beacause it has not been inited \n")
	}
	return dbInstance.db, nil
}

func NewDatabase() (Database, error) {

	if dbType == 001 {
		//log.Notice("Use level db only")
		//return NewLDBDataBase(leveldbPath)
		log.Notice("Use SuperLevelDB")
		return NewSLDB(leveldbPath)
	} else if dbType == 010 {
		log.Notice("Use ssdb only")
		return NewSSDatabase()
	} else if dbType == 110 {
		log.Notice("Use ssdb and redis")
		return NewRdSdDb()
	} else if dbType == 100 {
		log.Notice("Use redis only")
		return NewRsDatabase()
	} else {
		log.Notice("Wrong dbType:" + strconv.Itoa(dbType))
		return nil, errors.New("Wrong dbType:" + strconv.Itoa(dbType))
	}
}
