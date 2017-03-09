//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/crypto"
	"errors"
	edb "hyperchain/core/db_utils"
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
	consenter   consensus.Consenter // consensus module handler
	peerManager p2p.PeerManager
	commonHash  crypto.CommonHash
	encryption  crypto.Encryption
	conf        *common.Config      // block configuration
	status      ExecutorStatus
	hashUtils   ExecutorHashUtil
	cache       ExecutorCache
	helper      *Helper
	statedb     vm.Database
}

func NewExecutor(namespace string, consenter consensus.Consenter, peerManager p2p.PeerManager, conf *common.Config, commonHash crypto.CommonHash, encryption crypto.Encryption, eventMux *event.TypeMux) *Executor {
	helper := NewHelper(eventMux)
	executor := &Executor{
		namespace:       namespace,
		consenter:       consenter,
		peerManager:     peerManager,
		conf:            conf,
		commonHash:      commonHash,
		encryption:      encryption,
		helper:          helper,
	}
	return executor
}


// initializeExecutorStateDb - initialize statedb.
func initializeExecutorStateDb(executor *Executor) error {
	stateDb, err := executor.NewStateDb()
	if err != nil {
		log.Errorf("[Namespace = %s] executor init stateDb failed, err : %s", executor.namespace, err.Error())
		return err
	}
	executor.statedb = stateDb
	return nil
}


func (executor *Executor) Initialize() {
	if err := initializeExecutorStatus(executor); err != nil {

	}
	if err := initializeExecutorCache(executor); err != nil {

	}
	if err := initializeExecutorStateDb(executor); err != nil {

	}
	// start to listen for process commit event or validation event
	go executor.ListenCommitEvent()
	go executor.listenValidationEvent()
}

// NewStateDb - create a latest state.
func (executor *Executor) NewStateDb() (vm.Database, error) {
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
