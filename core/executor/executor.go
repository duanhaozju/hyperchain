// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/ledger/state"
	"hyperchain/core/vm"
	"hyperchain/core/vm/jcee/go"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"hyperchain/manager/event"
	"path"
)

// Executor represents a hyperchain executor implementation
type Executor struct {
	// Namespace should contain the official namespace name,
	// often a fixed size random word.
	namespace   string
	db          db.Database       // Db refers a blockchain online database handler
	archiveDb   db.Database       // Archivedb refers a offline historical database handler
	commonHash  crypto.CommonHash // Keccak256 hasher
	encryption  crypto.Encryption // ECDSA encrypter
	conf        *Config           // Conf refers a configuration reader
	context     *ExecutorContext
	hasher      Hasher
	cache       *Cache
	helper      *Helper     // Helper refers a communication mux
	statedb     vm.Database // World state handler
	logger      *logging.Logger
	jvmCli      jvm.ContractExecutor // Jvm client
	snapshotReg *SnapshotRegistry    // Snapshot registry
	archiveMgr  *ArchiveManager      // Archive manager
	nvp         NVP
}

// NewExecutor creates a new Hyperchain object (including the
// initialisation of the common hyperchain object)
func NewExecutor(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux) (*Executor, error) {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	helper := newHelper(eventMux, filterMux)

	executor := &Executor{
		namespace:  namespace,
		conf:       &Config{conf, namespace},
		commonHash: kec256Hash,
		encryption: encryption,
		helper:     helper,
		logger:     common.GetLogger(namespace, "executor"),
	}

	// Init jvm client if jvm is been permissible
	if executor.conf.isJvmEnable() {
		executor.jvmCli = jvm.NewContractExecutor(conf, namespace)
	}

	// Init several components.
	executor.snapshotReg = NewSnapshotRegistry(namespace, executor.logger, executor)
	executor.archiveMgr = NewArchiveManager(namespace, executor, executor.snapshotReg, executor.logger)
	executor.nvp = NewNVPImpl(executor)
	executor.cache = newExecutorCache()
	executor.context = newExecutorContext()

	if err := executor.initDb(); err != nil {
		executor.logger.Errorf("[Namespace = %s] executor initialize db failed. %s", executor.namespace, err.Error())
		return nil, err
	}

	return executor, nil
}

// initDb initializes the normal database handler and archive database handler.
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

// Start starts the executor's services.
func (executor *Executor) Start() error {
	if err := executor.initialize(); err != nil {
		return err
	}
	executor.logger.Noticef("[Namespace = %s]  executor start", executor.namespace)
	return nil
}

// Stop stops all the services provided by executor.
func (executor *Executor) Stop() error {
	if err := executor.finalize(); err != nil {
		return err
	}
	executor.logger.Noticef("[Namespace = %s] executor stop", executor.namespace)
	return nil
}

// initialize inits some components of executor and start the services.
func (executor *Executor) initialize() error {

	if err := executor.initExecutorContext(); err != nil {
		executor.logger.Errorf("executor initialize status failed. %s", err.Error())
		return err
	}

	if err := initStateDb(executor); err != nil {
		executor.logger.Errorf("executor initialize state failed. %s", err.Error())
		return err
	}

	if err := executor.snapshotReg.Start(); err != nil {
		executor.logger.Errorf("executor initialize snapshot registry failed. %s", err.Error())
		return err
	}

	// Start to listen for process commit event or validation event
	go executor.listenCommitEvent()
	go executor.listenValidationEvent()
	go executor.syncReplica()
	if executor.conf.isJvmEnable() {
		executor.jvmCli.Start()
	}
	return nil
}

// finalize stops all the services of executor.
func (executor *Executor) finalize() error {
	if executor.context.exit != nil {
		close(executor.context.exit)
	}
	go executor.snapshotReg.Stop()
	if executor.conf.isJvmEnable() {
		executor.jvmCli.Stop()
	}
	return nil
}

// initExecutorContext inits the ExecutorContext
func (executor *Executor) initExecutorContext() error {
	currentChain := chain.GetChainCopy(executor.namespace)
	blk, err := chain.GetBlockByNumber(executor.namespace, currentChain.Height)
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] get block #%d failed.", executor.namespace, currentChain.Height)
		return err
	}
	executor.context.initDemand(currentChain.Height + 1)
	executor.logger.Noticef("[Namespace = %s] initialize executor status success. demand block number %d, demand seqNo %d, latest state hash %s",
		executor.namespace, executor.context.demandNumber, executor.context.demandSeqNo, common.Bytes2Hex(blk.MerkleRoot))

	return nil
}

// initStateDb inits the statedb.
func initStateDb(executor *Executor) error {
	stateDb, err := executor.newStateDb()
	if err != nil {
		return err
	}
	executor.statedb = stateDb
	return nil
}

// initHistoryStateDb inits a historical database for snapshot content query.
func (executor *Executor) initHistoryStateDb(snapshotId string) (vm.Database, error, func()) {
	// never forget to close db
	if err, manifest := executor.snapshotReg.rwc.Read(snapshotId); err != nil {
		return nil, err, nil
	} else {
		blk, err := chain.GetBlockByNumber(executor.namespace, manifest.Height)
		if err != nil {
			return nil, err, nil
		}

		db, err := hyperdb.NewDatabase(executor.conf.conf, path.Join("snapshots", "SNAPSHOT_"+snapshotId), hyperdb.GetDatabaseType(executor.conf.conf), executor.namespace)
		if err != nil {
			return nil, err, nil
		}

		closeDb := func() {
			db.Close()
		}

		stateDb, err := state.New(common.BytesToHash(blk.MerkleRoot), db, nil, executor.conf.conf, manifest.Height)
		return stateDb, err, closeDb
	}
}

// newStateDb creates a latest state db.
func (executor *Executor) newStateDb() (vm.Database, error) {
	blk, err := chain.GetBlockByNumber(executor.namespace, chain.GetHeightOfChain(executor.namespace))
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] can not find block #%d", executor.namespace, chain.GetHeightOfChain(executor.namespace))
		return nil, err
	}
	stateDb, err := state.New(common.BytesToHash(blk.MerkleRoot), executor.db, executor.archiveDb, executor.conf.conf, chain.GetHeightOfChain(executor.namespace))
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] new stateDb failed, err : %s", executor.namespace, err.Error())
		return nil, err
	}
	return stateDb, nil
}

// FetchStateDb fetches the state db
func (executor *Executor) FetchStateDb() vm.Database {
	return executor.statedb
}

// GetNVP gets the nvp handler.
func (executor *Executor) GetNVP() NVP {
	return executor.nvp
}
