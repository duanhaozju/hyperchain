//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package hyperdb

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/hyperdb/hleveldb"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"path"
	"sync"

	"fmt"
	hcomm "github.com/hyperchain/hyperchain/hyperdb/common"
)

//func init() {
//	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
//}

var (
	ErrInvalidConfig = errors.New("hyperdb/dbmgr: Invalid config")
	ErrInvalidDbName = errors.New("hyperdb/dbmgr: Invalid dbname or namespace")
	ErrDbExisted     = errors.New("hyperdb/dbmgr: Createdb error, db already existed")
	ErrDbNotExisted  = errors.New("hyperdb/dbmgr: Database is not existed")
)

var (
	dbmgr  *DatabaseManager
	once   sync.Once
	logger *logging.Logger
)

//DatabaseManager mange all database by namespace.
type DatabaseManager struct {
	config *common.Config
	nlock  *sync.RWMutex
	ndbs   map[string]*NDB // <namespace, ndb>
	logger *logging.Logger
}

func (md *DatabaseManager) addNDB(namespace string, ndb *NDB) {
	md.nlock.Lock()
	md.ndbs[namespace] = ndb
	md.nlock.Unlock()
}

func (md *DatabaseManager) getNDB(namespace string) *NDB {
	md.nlock.RLock()
	defer md.nlock.RUnlock()
	return md.ndbs[namespace]
}

//for test
func (md *DatabaseManager) clearMemNDBSs() {
	md.nlock.Lock()
	md.ndbs = make(map[string]*NDB)
	md.nlock.Unlock()
}

func (md *DatabaseManager) removeNDB(namespace string) {
	md.nlock.Lock()
	delete(md.ndbs, namespace)
	md.nlock.Unlock()
}

func (dm *DatabaseManager) contains(dbName *DbName) bool {
	dm.nlock.RLock()
	defer dm.nlock.RUnlock()
	if ndb, ok := dm.ndbs[dbName.namespace]; ok {
		if _, ok = ndb.dbs[dbName.name]; ok {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

//getDatabase get database instance by namespace and dbname.
func (dm *DatabaseManager) getDatabase(namespace, dbname string) (db.Database, error) {
	if len(dbname) == 0 || len(namespace) == 0 {
		return nil, ErrInvalidDbName
	}
	ndb := dm.getNDB(namespace)
	if ndb != nil {
		return ndb.getDB(dbname)
	} else {
		return nil, ErrDbNotExisted
	}
}

type DbName struct {
	namespace string
	name      string
}

func NewDbName(name, namesapce string) *DbName {
	if len(name) == 0 {
		logger.Error(ErrInvalidDbName)
		return nil
	}
	dn := &DbName{}

	if len(namesapce) == 0 {
		dn.namespace = common.DEFAULT_NAMESPACE
	}
	dn.name = name
	return dn
}

func defaultConfig() *common.Config {
	conf := common.NewRawConfig()
	conf.Set(hcomm.LEVEL_DB_ROOT_DIR, "./tmp")
	conf.Set(hcomm.DB_TYPE, hcomm.LDB_DB)
	return conf
}

//InitDBMgr init blockchain database manager
func InitDBMgr(conf *common.Config) error {
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
	if conf == nil {
		conf = defaultConfig()
		logger.Noticef("%v, use default db config, %v", ErrInvalidConfig, conf)
	}
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
	once.Do(func() {
		dbmgr = &DatabaseManager{
			config: conf,
			ndbs:   make(map[string]*NDB),
			logger: common.GetLogger(common.DEFAULT_LOG, "dbmgr"),
			nlock:  &sync.RWMutex{},
		}
	})
	return nil
}

//ConnectToDB connect to a specified database
func ConnectToDB(dbname *DbName, conf *common.Config) (db.Database, error) {
	logger = common.GetLogger(common.DEFAULT_NAMESPACE, "dbmgr")
	if conf == nil || len(conf.GetString(hcomm.LEVEL_DB_ROOT_DIR)) == 0 {
		return nil, fmt.Errorf("hyperdb/dbmgr: invalid conf, no db configs found")
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
	ldb, err := hleveldb.NewLDBDataBase(conf, path.Join(conf.GetString(hcomm.LEVEL_DB_ROOT_DIR), dbname.name), dbname.namespace)
	if err != nil {
		return nil, err
	}

	if ndb, ok := dbmgr.ndbs[dbname.namespace]; ok {
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

//GetDB get database by namespace and dbname.
func GetDB(dbname *DbName) (db.Database, error) {
	if dbmgr == nil {
		//TODO: handle situations where dbmgr is not initialized
	}
	return dbmgr.getDatabase(dbname.namespace, dbname.name)
}

//CloseByName close connection to the database.
func CloseByName(dbname *DbName) error {
	db, err := GetDB(dbname)
	if err != nil {
		return ErrDbNotExisted
	}
	db.Close()
	ndb := dbmgr.getNDB(dbname.namespace)
	ndb.removeDB(dbname.name)
	return nil
}

//Close close all databases, and remove all connected db instances.
func Close() {
	dbmgr.nlock.RLock()
	defer dbmgr.nlock.RUnlock()
	for _, ndb := range dbmgr.ndbs {
		ndb.close()
	}
	dbmgr.ndbs = make(map[string]*NDB)
}

//NDB manage databases for namespace
type NDB struct {
	namespace string //namespace id
	lock      sync.RWMutex
	dbs       map[string]db.Database // <dbname, db>
}

//close close all database during this namespace.
func (ndb *NDB) close() {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()
	for name, db := range ndb.dbs {
		logger.Infof("try to close db %s", name)
		go db.Close()
	}
}

func (ndb *NDB) addDB(name string, db db.Database) {
	ndb.lock.Lock()
	ndb.dbs[name] = db
	ndb.lock.Unlock()
}

func (ndb *NDB) getDB(name string) (db.Database, error) {
	ndb.lock.RLock()
	defer ndb.lock.RUnlock()
	if db, ok := ndb.dbs[name]; ok {
		return db, nil
	} else {
		return nil, ErrDbNotExisted
	}
}

func (ndb *NDB) removeDB(name string) {
	ndb.lock.Lock()
	defer ndb.lock.Unlock()
	delete(ndb.dbs, name)
}
