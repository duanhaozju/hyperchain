//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/event"
	"hyperchain/crypto"
	"errors"
	edb "hyperchain/core/db_utils"
	"time"
)

var (
	log   *logging.Logger // package-level logger
	EmptyPointerErr = errors.New("nil pointer")
	NoDefinedCaseErr= errors.New("no defined case")
)

func init() {
	log = logging.MustGetLogger("executor")
}


type Executor struct {
	namespace   string                // namespace tag
	commonHash  crypto.CommonHash
	encryption  crypto.Encryption
	conf        *common.Config      // block configuration
	status      ExecutorStatus
	hashUtils   ExecutorHashUtil
	cache       ExecutorCache
	helper      *Helper
	statedb     vm.Database
}

func NewExecutor(namespace string, conf *common.Config, eventMux *event.TypeMux) *Executor {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	helper := NewHelper(eventMux)
	executor := &Executor{
		namespace:       namespace,
		conf:            conf,
		commonHash:      kec256Hash,
		encryption:      encryption,
		helper:          helper,
	}
	return executor
}

// Start - start service.
func (executor *Executor) Start() {
	executor.initialize()
	log.Noticef("[Namespace = %s] ************* executor start **************", executor.namespace)
}

// Stop - stop service.
func (executor *Executor) Stop() {
	executor.setExit()
	log.Noticef("[Namespace = %s] ************* executor stop **************", executor.namespace)
}

// Status - obtain executor status.
func (executor *Executor) Status() {

}


func (executor *Executor) initialize() {
	if err := initializeExecutorStatus(executor); err != nil {
		log.Errorf("executor initiailize status failed. %s", err.Error())
	}
	if err := initializeExecutorCache(executor); err != nil {
		log.Errorf("executor initiailize cache failed. %s", err.Error())
	}
	if err := initializeExecutorStateDb(executor); err != nil {
		log.Errorf("executor initiailize state failed. %s", err.Error())
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
		log.Errorf("[Namespace = %s] executor init stateDb failed, err : %s", executor.namespace, err.Error())
		return err
	}
	executor.statedb = stateDb
	return nil
}
// NewStateDb - create a latest state.
func (executor *Executor) newStateDb() (vm.Database, error) {
	blk, err := edb.GetBlockByNumber(executor.namespace, edb.GetHeightOfChain(executor.namespace))
	if err != nil {
		log.Errorf("[Namespace = %s] can not find block #%d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
		return nil, err
	}
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Errorf("[Namespace = %s] get database failed", executor.namespace)
		return nil, err
	}
	stateDb, err := hyperstate.New(common.BytesToHash(blk.MerkleRoot), db, executor.conf, edb.GetHeightOfChain(executor.namespace))
	if err != nil {
		log.Errorf("[Namespace = %s] new stateDb failed, err : %s", executor.namespace, err.Error())
		return nil, err
	}
	return stateDb, nil
}
