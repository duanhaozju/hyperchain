//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"hyperchain/core/vm/jcee/go"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"hyperchain/manager/event"
	"path"
)

// Executor represents a hyperchain exector implementation
type Executor struct {
	// Namespace should contain the official namespace name,
	// often a fixed size random word.
	namespace   string
	db          db.Database       // Db refers a blockchain online database handler
	archiveDb   db.Database       // Archivedb refers a offline historical database handler
	commonHash  crypto.CommonHash // Keccak256 hasher
	encryption  crypto.Encryption // ECDSA encrypter
	conf        *common.Config    // Conf refers a configuration reader
	context     ExecutorContext
	hasher      Hasher
	cache       Cache
	helper      *Helper     // Helper refers a communication mux
	statedb     vm.Database // world state handler
	logger      *logging.Logger
	jvmCli      jvm.ContractExecutor // jvm client
	snapshotReg *SnapshotRegistry    // snapshot registry
	archiveMgr  *ArchiveManager      // archive manager
}

// NewExecutor creates a new Hyperchain object (including the
// initialisation of the common hyperchain object)
func NewExecutor(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux) (*Executor, error) {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	helper := NewHelper(eventMux, filterMux)

	executor := &Executor{
		namespace:  namespace,
		conf:       conf,
		commonHash: kec256Hash,
		encryption: encryption,
		helper:     helper,
		jvmCli:     jvm.NewContractExecutor(conf, namespace),
	}

	// initialise jvm client if jvm is been permissible
	if executor.isJvmEnable() {
		executor.jvmCli = jvm.NewContractExecutor(conf, namespace)
	}

	executor.logger = common.GetLogger(namespace, "executor")
	executor.snapshotReg = NewSnapshotRegistry(namespace, executor.logger, executor)
	executor.archiveMgr = NewArchiveManager(namespace, executor, executor.snapshotReg, executor.logger)
	// TODO doesn't know why to add this statement here.
	// TODO ask @Rongjialei to fix this.
	if err := executor.initDb(); err != nil {
		return nil, err
	}
	return executor, nil
}

// initDb make the initilisation of the database handler.
func (executor *Executor) initDb() error {
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		return err
	}
	executor.db = db
	archiveDb, err := hyperdb.GetArchiveDbByNamespace(executor.namespace)
	if err != nil {
		return err
	}
	executor.archiveDb = archiveDb
	return nil
}

// Start - start service.
func (executor *Executor) Start() error {
	if err := executor.initialize(); err != nil {
		return err
	}
	executor.logger.Noticef("[Namespace = %s]  executor start", executor.namespace)
	return nil
}

// Stop - stop service.
func (executor *Executor) Stop() error {
	if err := executor.finailize(); err != nil {
		return err
	}
	executor.logger.Noticef("[Namespace = %s] executor stop", executor.namespace)
	return nil
}

// Status - obtain executor status.
func (executor *Executor) Status() {

}

func (executor *Executor) initialize() error {
	if err := executor.initDb(); err != nil {
		executor.logger.Errorf("executor initiailize db failed. %s", err.Error())
		return err
	}
	if err := initializeExecutorContext(executor); err != nil {
		executor.logger.Errorf("executor initiailize status failed. %s", err.Error())
		return err
	}
	if err := initializeExecutorCache(executor); err != nil {
		executor.logger.Errorf("executor initiailize cache failed. %s", err.Error())
		return err
	}
	if err := initializeExecutorStateDb(executor); err != nil {
		executor.logger.Errorf("executor initiailize state failed. %s", err.Error())
		return err
	}
	if err := executor.snapshotReg.Start(); err != nil {
		executor.logger.Errorf("executor initiailize snapshot registry failed. %s", err.Error())
		return err
	}
	// start to listen for process commit event or validation event
	go executor.listenCommitEvent()
	go executor.listenValidationEvent()
	go executor.syncReplica()
	if executor.isJvmEnable() {
		executor.jvmCli.Start()
	}
	return nil
}

// setExit - notify all backend to exit.
func (executor *Executor) finailize() error {
	go executor.setValidationExit()
	go executor.setCommitExit()
	go executor.setReplicaSyncExit()
	go executor.snapshotReg.Stop()
	if executor.isJvmEnable() {
		executor.jvmCli.Stop()
	}
	return nil
}

// initializeExecutorStateDb - initialize statedb.
func initializeExecutorStateDb(executor *Executor) error {
	stateDb, err := executor.newStateDb()
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] executor init stateDb failed, err : %s", executor.namespace, err.Error())
		return err
	}
	executor.statedb = stateDb
	return nil
}

// initHistoryStateDb - open a historical database for snapshot content query.
func (executor *Executor) initHistoryStateDb(snapshotId string) (vm.Database, error, func()) {
	// never forget to close db
	if err, manifest := executor.snapshotReg.rwc.Read(snapshotId); err != nil {
		return nil, err, nil
	} else {
		blk, err := edb.GetBlockByNumber(executor.namespace, manifest.Height)
		if err != nil {
			return nil, err, nil
		}

		db, err := hyperdb.NewDatabase(executor.conf, path.Join("snapshots", "SNAPSHOT_"+snapshotId), hyperdb.GetDatabaseType(executor.conf), executor.namespace)
		if err != nil {
			return nil, err, nil
		}

		closeDb := func() {
			db.Close()
		}

		stateDb, err := hyperstate.New(common.BytesToHash(blk.MerkleRoot), db, nil, executor.conf, manifest.Height, executor.namespace)
		return stateDb, err, closeDb
	}
}

// NewStateDb - create a latest state.
func (executor *Executor) newStateDb() (vm.Database, error) {
	blk, err := edb.GetBlockByNumber(executor.namespace, edb.GetHeightOfChain(executor.namespace))
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] can not find block #%d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
		return nil, err
	}
	stateDb, err := hyperstate.New(common.BytesToHash(blk.MerkleRoot), executor.db, executor.archiveDb, executor.conf, edb.GetHeightOfChain(executor.namespace), executor.namespace)
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] new stateDb failed, err : %s", executor.namespace, err.Error())
		return nil, err
	}
	return stateDb, nil
}

// FetchStateDb - fetch state db
func (executor *Executor) FetchStateDb() vm.Database {
	return executor.statedb
}
