//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"path"
	"sync"

	"github.com/op/go-logging"
	"github.com/pkg/errors"

	"github.com/hyperchain/hyperchain/common"
	hcomm "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/hyperdb/hleveldb"
)

var (
	ErrInvalidConfig = errors.New("hyperdb/dbmgr: Invalid config")
	ErrInvalidDbName = errors.New("hyperdb/dbmgr: Invalid dbname or namespace")
	ErrDbExisted     = errors.New("hyperdb/dbmgr: Createdb error, db already existed")
	ErrDbNotExisted  = errors.New("hyperdb/dbmgr: Database is not existed")
	ErrDbMgrIsNil    = errors.New("hyperdb/dbmgr: Database manager is nil")
)

var (
	dbmgr  *DatabaseManager
	once   sync.Once
	logger *logging.Logger
)

// DatabaseManager manges all databases by namespace.
type DatabaseManager struct {
	config *common.Config
	nlock  *sync.RWMutex
	ndbs   map[string]*NDB // <namespace, ndb>
	logger *logging.Logger
}

// addNDB adds namespace databases to dbmgr.
func (dm *DatabaseManager) addNDB(namespace string, ndb *NDB) {
	dm.nlock.Lock()
	defer dm.nlock.Unlock()

	dm.ndbs[namespace] = ndb
}

// getNDB gets namespace databases from dbmgr.
func (dm *DatabaseManager) getNDB(namespace string) *NDB {
	dm.nlock.RLock()
	defer dm.nlock.RUnlock()

	return dm.ndbs[namespace]
}

// clearMemNDBSs is for test.
func (dm *DatabaseManager) clearMemNDBSs() {
	dm.nlock.Lock()
	defer dm.nlock.Unlock()

	dm.ndbs = make(map[string]*NDB)
}

// removeNDB removes a database from dbmgr by namespace.
func (dm *DatabaseManager) removeNDB(namespace string) {
	dm.nlock.Lock()
	defer dm.nlock.Unlock()

	delete(dm.ndbs, namespace)
}

// contains returns whether a database is contained in dbmgr or not.
func (dm *DatabaseManager) contains(dbName *DbName) bool {
	dm.nlock.RLock()
	defer dm.nlock.RUnlock()

	if ndb, exists := dm.ndbs[dbName.namespace]; exists {
		if _, exists = ndb.dbs[dbName.name]; exists {
			return true
		}
	}
	return false
}

// getDatabase gets database instance by namespace and dbname.
func (dm *DatabaseManager) getDatabase(namespace, dbname string) (db.Database, error) {
	if len(dbname) == 0 || len(namespace) == 0 {
		return nil, ErrInvalidDbName
	}

	if ndb := dm.getNDB(namespace); ndb != nil {
		return ndb.getDB(dbname)
	} else {
		return nil, ErrDbNotExisted
	}
}

type DbName struct {
	namespace string
	name      string
}

// NewDbName generates a new DbName pointer with a namespace and dbname.
func NewDbName(dbname, namespace string) *DbName {
	if len(dbname) == 0 {
		logger.Error(ErrInvalidDbName)
		return nil
	}
	if len(namespace) == 0 {
		namespace = common.DEFAULT_NAMESPACE
	}

	dn := &DbName{
		namespace: namespace,
		name:      dbname,
	}
	return dn
}

// defaultConfig generates default configurations.
func defaultConfig() *common.Config {
	conf := common.NewRawConfig()
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "./tmp")
	conf.Set(hcomm.DB_TYPE, hcomm.LDB_DB)
	return conf
}

//InitDBMgr inits blockchain database manager.
func InitDBMgr(conf *common.Config) error {
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
	if conf == nil {
		conf = defaultConfig()
		logger.Noticef("%v, use default db config, %v", ErrInvalidConfig, conf)
	}
	logger = common.GetLogger(common.DEFAULT_LOG, "dbmgr")
	once.Do(func() {
		dbmgr = &DatabaseManager{
			config: conf,
			ndbs:   make(map[string]*NDB),
			logger: logger,
			nlock:  &sync.RWMutex{},
		}
	})
	return nil
}

//ConnectToDB is for test, it connects to a specified database.
func ConnectToDB(dbname *DbName, conf *common.Config) (db.Database, error) {
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
	if conf == nil || len(conf.GetString(hcomm.LEVEL_DB_ROOT_DIR)) == 0 {
		return nil, ErrInvalidConfig
	}

	if dbmgr == nil {
		InitDBMgr(conf)
	}

	if dbname == nil || len(dbname.name) == 0 || len(dbname.namespace) == 0 {
		return nil, ErrInvalidDbName
	}

	ns := conf.GetString(common.NAMESPACE)
	if len(ns) != 0 && dbname.namespace != ns {
		dbmgr.logger.Noticef("input namespace name is not equal to namespace specified in config")
	}

	if dbmgr.contains(dbname) {
		return nil, ErrDbExisted
	}
	//create new database
	ldb, err := hleveldb.NewLDBDataBase(conf,
		path.Join(conf.GetString(hcomm.LEVEL_DB_ROOT_DIR), dbname.name),
		dbname.namespace)
	if err != nil {
		return nil, err
	}

	if ndb, exists := dbmgr.ndbs[dbname.namespace]; exists {
		ndb.addDB(dbname.name, ldb)
	} else {
		ndb = &NDB{
			namespace: dbname.namespace,
			dbs:       make(map[string]db.Database),
		}
		ndb.addDB(dbname.name, ldb)
		dbmgr.addNDB(dbname.namespace, ndb)
	}

	return ldb, nil
}

// GetDB is for test, it gets database by namespace and dbname.
func GetDB(dbname *DbName) (db.Database, error) {
	if dbmgr == nil {
		//TODO: handle situations where dbmgr is not initialized
		// TODO: never happen?
		return nil, ErrDbMgrIsNil
	}
	return dbmgr.getDatabase(dbname.namespace, dbname.name)
}

// CloseByName is for test, it closes connection to the database.
func CloseByName(dbname *DbName) error {
	db, err := GetDB(dbname)
	if err != nil {
		return err
	}
	db.Close()
	ndb := dbmgr.getNDB(dbname.namespace)
	// TODO: why remove?
	ndb.removeDB(dbname.name)
	return nil
}

// Close close all databases, and remove all connected db instances.
func Close() {
	dbmgr.nlock.RLock()
	defer dbmgr.nlock.RUnlock()

	for _, ndb := range dbmgr.ndbs {
		ndb.close()
	}
	dbmgr.ndbs = make(map[string]*NDB)
}

//NDB manages databases for the namespace.
type NDB struct {
	namespace string //namespace id
	lock      sync.RWMutex
	dbs       map[string]db.Database // <dbname, db>
}

// close close all database during this namespace.
func (ndb *NDB) close() {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()

	for name, db := range ndb.dbs {
		logger.Infof("try to close db %s", name)
		go db.Close()
	}
}

// addDB adds a database to its namespace by dbname.
func (ndb *NDB) addDB(dbname string, db db.Database) {
	ndb.lock.Lock()
	defer ndb.lock.Unlock()

	ndb.dbs[dbname] = db
}

// getDB gets a database from its namespace by dbname.
func (ndb *NDB) getDB(dbname string) (db.Database, error) {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()

	if db, exists := ndb.dbs[dbname]; exists {
		return db, nil
	} else {
		return nil, ErrDbNotExisted
	}
}

// removeDB removes a database from its namespace by dbname.
func (ndb *NDB) removeDB(dbname string) {
	ndb.lock.Lock()
	defer ndb.lock.Unlock()

	delete(ndb.dbs, dbname)
}
