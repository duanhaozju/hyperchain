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
	"math/big"
	"strconv"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/ledger/state"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
)

// CreateInitBlock creates genesis block for a specific namespace.
func (executor *Executor) CreateInitBlock(config *common.Config) error {
	if chain.IsGenesisFinish(executor.namespace) {
		executor.logger.Infof("[Namespace = %s] already genesis", executor.namespace)
		return nil
	}
	type Genesis struct {
		Timestamp  int64
		ParentHash string
		BlockHash  string
		Coinbase   string
		Number     uint64
		Alloc      map[string]int64
	}
	// create state instance with empty root hash
	stateDb, err := NewStateDb(config, executor.db, executor.namespace)
	if err != nil {
		return err
	}
	stateDb.MarkProcessStart(0)
	genMap := config.GetStringMap("genesis.alloc")
	if genMap == nil {
		panic("can not read genesis config")
	}
	for key, value := range genMap {
		object := stateDb.CreateAccount(common.HexToAddress(string(key)))
		s, ok := value.(string)
		if !ok {
			panic("invalid genesis configuration")
		}
		account, _ := strconv.ParseInt(s, 10, 64)
		object.AddBalance(big.NewInt(account))
	}

	root, err := stateDb.Commit()
	if err != nil {
		return err
	}
	// Flush state change to disk immediately
	block := types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Timestamp:  time.Now().Unix(),
		BlockHash:  common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Number:     uint64(0),
		MerkleRoot: root.Bytes(),
	}
	// Flush block content to disk immediately
	batch := stateDb.FetchBatch(0, state.BATCH_NORMAL)
	if _, err := chain.PersistBlock(batch, &block, true, true); err != nil {
		return err
	}
	// Flush change of chain to disk immediately
	chain.UpdateChain(executor.namespace, batch, &block, true, false, false)
	batch.Write()
	stateDb.MarkProcessFinish(0)
	return nil
}

// NewStateDb creates a new stateDb handler.
func NewStateDb(conf *common.Config, db db.Database, namespace string) (vm.Database, error) {
	archiveDb, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_ARCHIVE)
	if err != nil {
		return nil, err
	}
	stateDb, err := state.New(common.Hash{}, db, archiveDb, conf, 0)
	if err != nil {
		return nil, err
	}
	return stateDb, nil
}