//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"path"
	"strconv"
	"sync"
	"github.com/op/go-logging"
	"errors"
	"fmt"
)

type statedb int32

type DBInstance struct {
	db    Database
	state  statedb
	dbsync sync.Mutex
}

const (
	closed statedb = iota
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



//-- --------------- about db -----------------------

func getBaseDir() string {
	//path := os.TempDir()
	return "/tmp"
}

var (
	baseDBPath = getBaseDir() + "/hyperchain/cache"
	portDBPath = "db" //different port has different db path, default "db"
)

func SetDBPath(dbpath string, port int) {
	baseDBPath = path.Join(dbpath, "hyperchain/cache/")
	portDBPath = strconv.Itoa(port)
}

func getDBPath() string {
	return baseDBPath + portDBPath
}


//Db_type 应该是一个3位的二进制数比如111 101
//第一个1表示 redis open
//第二个1表示 ssdb  open
//第三个1表示 leveldb open
//现在仅支持 redis加 ssdb 或者仅leveldb 即 110 001
func InitDatabase(Db_type int)(error){

	dbInstance.dbsync.Lock()
	defer dbInstance.dbsync.Unlock()

	if dbInstance.state!=closed{
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it has beend inited \n",Db_type))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it has beend inited \n",Db_type))
	}
	db,err:=NewDatabase(getDBPath(),portDBPath,Db_type)

	if err!=nil{
		log.Notice(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n",Db_type))
		return errors.New(fmt.Sprintf("InitDatabase(%v) fail beacause it can't get new database \n",Db_type))
	}

	dbInstance.db=db;
	dbInstance.state=opened
	log.Notice("db has been init")
	return err
}

func GetDBDatabase() (Database, error) {
	dbInstance.dbsync.Lock()
	defer dbInstance.dbsync.Unlock()
	if dbInstance.db==nil{
		log.Notice("GetDBDatabase() fail beacause it has not been inited \n")
		return nil,errors.New("GetDBDatabase() fail beacause it has not been inited \n")
	}
	return dbInstance.db,nil
}

func NewDatabase(DBPath string ,portDBPath string,Db_type int) (Database ,error){
	if Db_type==001{
		log.Notice("Use level db only")
		return NewLDBDataBase(DBPath)
	}else if Db_type==010{
		log.Notice("Use ssdb only")
		return NewSSDatabase(portDBPath,2)
	}else if Db_type==110{
		log.Notice("Use ssdb and redis")
		return NewRdSdDb(portDBPath,2)
	}else if Db_type==100{
		log.Notice("Use redis only")
		return NewRsDatabase(portDBPath)
	}else {
		fmt.Println("Wrong Db_type:"+strconv.Itoa(Db_type))
		return nil,errors.New("Wrong Db_type:"+strconv.Itoa(Db_type))
	}
}