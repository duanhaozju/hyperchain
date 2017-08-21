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
	"path"

	hcomm "hyperchain/hyperdb/common"
)

var (
	ErrInvalidConfig = errors.New("hyperdb/dbmgr: Invalid config")
	ErrInvalidDbName = errors.New("hyperdb/dbmgr: Invalid dbname or namespace")
	ErrDbExisted     = errors.New("hyperdb/dbmgr: Createdb error, db already existed")
	ErrDbNotExisted  = errors.New("hyperdb/dbmgr: Database is not existed")
)

var dbmgr *DatabaseManager
var once  sync.Once

//DatabaseManager mange all database by namespace.
type DatabaseManager struct {
	config *common.Config
	nlock  *sync.RWMutex
	ndbs   map[string]*NDB // <namespace, ndb>
	logger *logging.Logger
}

func (md *DatabaseManager) addNDB(namespace string, ndb *NDB)  {
	md.nlock.Lock()
	md.ndbs[namespace] = ndb
	md.nlock.Unlock()
}

func (md *DatabaseManager) getNDB(namespace string) *NDB  {
	md.nlock.RLock()
	defer md.nlock.RUnlock()
	return md.ndbs[namespace]
}

//for test
func (md *DatabaseManager) clearMemNDBSs()  {
	md.nlock.Lock()
	md.ndbs = make(map[string]*NDB)
	md.nlock.Unlock()
}

func (md *DatabaseManager) removeNDB(namespace string)  {
	md.nlock.Lock()
	delete(md.ndbs, namespace)
	md.nlock.Unlock()
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


//getDatabase get database instance by namespace and dbname.
func (dm *DatabaseManager) getDatabase(namespace, dbname string) (error, db.Database) {
	if len(dbname) == 0 || len(namespace) == 0 {
		return ErrInvalidDbName, nil
	}
	ndb := dm.getNDB(namespace)
	if ndb != nil {
		return ndb.getDB(dbname)
	}else {
		return ErrDbNotExisted, nil
	}
}

type DbName struct {
	Namespace string
	Name      string
}

//InitDBMgr init blockchain database manager
func InitDBMgr(conf *common.Config) error {
	if conf == nil {
		//TODO: add default
		return ErrInvalidConfig
	}
	once.Do(func() {
		dbmgr = &DatabaseManager{
			config: conf,
			ndbs: make(map[string]*NDB),
			logger: common.GetLogger(common.DEFAULT_LOG, "dbmgr"),
			nlock: &sync.RWMutex{},
		}
	})
	return nil
}

//CreateDB create database instance
func CreateDB(dbname *DbName, conf *common.Config) (error, db.Database) {

	if dbmgr == nil {
		InitDBMgr(conf)
	}

	if dbname == nil || len(dbname.Name) == 0 || len(dbname.Namespace) == 0 {
		return ErrInvalidDbName, nil
	}

	ns := conf.GetString(common.NAMESPACE)
	if len(ns) != 0 && dbname.Namespace != ns {
		dbmgr.logger.Noticef("input namespace name is not equal to namespace specified in config")
	}

	if dbmgr.contains(dbname) {
		return ErrDbExisted, nil
	}
	//create new database
	ldb, err := hleveldb.NewLDBDataBase(conf, path.Join(conf.GetString(hcomm.LEVEL_DB_ROOT_DIR), dbname.Name), dbname.Namespace)
	if err != nil {
		return err, nil
	}

	if ndb, ok := dbmgr.ndbs[dbname.Namespace]; ok {
		ndb.addDB(dbname.Name, ldb)
	}else {
		ndb = &NDB{
			namespace:dbname.Namespace,
			dbs:make(map[string]db.Database),
		}
		ndb.addDB(dbname.Name, ldb)
		dbmgr.addNDB(dbname.Namespace, ndb)
	}

	return nil, ldb
}

//GetDB get database by namespace and dbname.
func GetDB(dbname *DbName) (error, db.Database) {
	if dbmgr == nil {
		//TODO: handle situations where dbmgr is not initialized
	}
	return dbmgr.getDatabase(dbname.Namespace, dbname.Name)
}

func CloseByName(dbname *DbName) error {
	err, db := GetDB(dbname)
	if err != nil {
		return ErrDbNotExisted
	}
	db.Close()
	return nil
}

//Close close all databases
func Close()  {
	dbmgr.nlock.RLock()
	defer dbmgr.nlock.RUnlock()
	for _, ndb := range dbmgr.ndbs {
		ndb.close()
	}
}

//NDB manage databases for namespace
type NDB struct {
	namespace string	  //namespace id
	lock sync.RWMutex
	dbs map[string]db.Database // <dbname, db>
}

//close close all database during this namespace.
func (ndb *NDB) close()  {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()
	for _, db := range ndb.dbs {
		go db.Close()
	}
}

func (ndb *NDB) addDB(name string, db db.Database)  {
	ndb.lock.Lock()
	ndb.dbs[name] = db
	ndb.lock.Unlock()
}

func (ndb *NDB) getDB(name string) (error, db.Database)  {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()
	if db, ok := ndb.dbs[name]; ok {
		return nil, db
	}else {
		return ErrDbNotExisted, nil
	}
}