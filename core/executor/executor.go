//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm/evm"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"hyperchain/manager/event"
)

type Executor struct {
	namespace  string // namespace tag
	db         db.Database
	archieveDb db.Database
	commonHash crypto.CommonHash
	encryption crypto.Encryption
	conf       *common.Config // block configuration
	status     ExecutorStatus
	hashUtils  ExecutorHashUtil
	cache      ExecutorCache
	helper     *Helper
	statedb    evm.Database
	logger     *logging.Logger
	exception  ExceptionHandler
}

func NewExecutor(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux) *Executor {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	helper := NewHelper(eventMux, filterMux)

	executor := &Executor{
		namespace:  namespace,
		conf:       conf,
		commonHash: kec256Hash,
		encryption: encryption,
		helper:     helper,
		exception:  NewExceptionHandler(helper),
	}
	executor.logger = common.GetLogger(namespace, "executor")
	executor.initDb()
	return executor
}

func (executor *Executor) initDb() {
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		//return nil
	}
	executor.db = db
	archieveDb, err := hyperdb.GetArchieveDbByNamespace(executor.namespace)
	if err != nil {
		//return nil
	}
	executor.archieveDb = archieveDb
}

// Start - start service.
func (executor *Executor) Start() {
	executor.initialize()
	executor.logger.Noticef("[Namespace = %s]  executor start", executor.namespace)
}

// Stop - stop service.
func (executor *Executor) Stop() {
	executor.setExit()
	executor.logger.Noticef("[Namespace = %s] executor stop", executor.namespace)
}

// Status - obtain executor status.
func (executor *Executor) Status() {

}

func (executor *Executor) initialize() {
	executor.initDb()
	if err := initializeExecutorStatus(executor); err != nil {
		executor.logger.Errorf("executor initiailize status failed. %s", err.Error())
	}
	if err := initializeExecutorCache(executor); err != nil {
		executor.logger.Errorf("executor initiailize cache failed. %s", err.Error())
	}
	if err := initializeExecutorStateDb(executor); err != nil {
		executor.logger.Errorf("executor initiailize state failed. %s", err.Error())
	}
	// start to listen for process commit event or validation event
	go executor.listenCommitEvent()
	go executor.listenValidationEvent()
	go executor.syncReplica()
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

// NewStateDb - create a latest state.
func (executor *Executor) newStateDb() (evm.Database, error) {
	blk, err := edb.GetBlockByNumber(executor.namespace, edb.GetHeightOfChain(executor.namespace))
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] can not find block #%d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
		return nil, err
	}
	stateDb, err := hyperstate.New(common.BytesToHash(blk.MerkleRoot), executor.db, executor.archieveDb, executor.conf, edb.GetHeightOfChain(executor.namespace), executor.namespace)
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] new stateDb failed, err : %s", executor.namespace, err.Error())
		return nil, err
	}
	return stateDb, nil
}
