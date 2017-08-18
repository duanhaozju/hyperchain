//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"sync"
	"github.com/pkg/errors"
	"hyperchain/hyperdb/hleveldb"
)

var (
	ErrInvalidConfig = errors.New("hyperdb/dbmgr: Invalid config")
	ErrInvalidDbName = errors.New("hyperdb/dbmgr: Invalid dbname or namespace")
	ErrDbExisted     = errors.New("hyperdb/dbmgr: Createdb error, db already existed")
	ErrDbNotExisted  = errors.New("hyperdb/dbmgr: Database is not existed")
)

var dbmgr *DatabaseManager
var once  sync.Once

//DatabaseManager mange all database by namespace
type DatabaseManager struct {
	config *common.Config
	nlock   sync.RWMutex
	ndbs   map[string]*NDB // <namespace, ndb>
	logger *logging.Logger
}

type DbName struct {
	Namespace string
	Name      string
}

func (dm *DatabaseManager) contains(dbName *DbName) bool {
	dm.nlock.RLock()
	defer dm.nlock.RUnlock()
	if ndb, ok := dm.ndbs[dbName.Namespace]; ok {
		if _, ok = ndb.dbs[dbName.Name]; ok {
			return true
		}else {
			return false
		}
	}else {
		return false
	}
}


//InitDBMgr init blockchain database manager
func InitDBMgr(conf *common.Config) error {
	if conf == nil {
		return ErrInvalidConfig
	}
	once.Do(func() {
		dbmgr = &DatabaseManager{
			config: conf,
			ndbs: make(map[string]*NDB),
			logger: common.GetLogger(common.DEFAULT_LOG, "dbmgr"),
		}
	})
	return nil
}

//CreateDB create database instance
func CreateDB(dbname *DbName, conf *common.Config) (error, db.Database) {

	if dbname == nil || len(dbname.Name) == 0 || len(dbname.Namespace) == 0 {
		return ErrInvalidDbName, nil
	}

	ns := conf.GetString(common.NAMESPACE)
	if len(ns) != 0 && dbname.Namespace != ns {
		dbmgr.logger.Noticef("input namespace name is not equal to namespace specified in config")
	}

	if dbmgr.contains(dbname) {
		return ErrDbExisted
	}
	//create new database
	db, err := hleveldb.NewLDBDataBase(conf, dbname.Name, dbname.Namespace)
	if err != nil {
		return err, nil
	}

	if ndb, ok := dbmgr.ndbs[dbname.Namespace]; ok {
		ndb.addDB(dbname.Name, db)
	}else {
		ndb = &NDB{
			namespace:dbname.Namespace,
			dbs:make(map[string] db.Database),
		}
		ndb.addDB(dbname.Name, db)
	}

	return nil, db
}

//GetDB get database by namespace and dbname
func GetDB(dbname *DbName) (error, db.Database) {
	return dbmgr.getDatabase(dbname.Namespace, dbname.Name)
}

//NDB manage databases for namespace
type NDB struct {
	namespace string	  //namespace id
	lock sync.RWMutex
	dbs map[string] db.Database // <dbname, db>
}

func (ndb *NDB) addDB(name string, db db.Database)  {
	ndb.lock.Lock()
	ndb.dbs[name] = db
	ndb.lock.Unlock()
}

//getDatabase get database instance by namespace and dbname
func (dm *DatabaseManager) getDatabase(namespace, dbname string) (error, db.Database) {
	if dbname == nil || len(dbname) == 0 || len(namespace) == 0 {
		return ErrInvalidDbName, nil
	}

	dm.nlock.RLock()
	defer dm.nlock.RUnlock()
	if ndb, ok := dm.ndbs[namespace]; ok {
		if db, ok := ndb.dbs[dbname]; ok {
			return nil, db
		}else {
			return ErrDbNotExisted, nil
		}
	}else {
		return ErrDbNotExisted, nil
	}
}

//close close all dbs
func (dm *DatabaseManager) close() error {

	return nil
}

//close add new database instace to the database manager
func (dm *DatabaseManager) addDatabase(namespace, dbname string, db db.Database) error {

	return nil
}
