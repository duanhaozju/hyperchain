package executor

import (
	"math/big"
	"time"
	"hyperchain/core/hyperstate"
	edb "hyperchain/core/db_utils"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"strconv"
)

// CreateInitBlock - create genesis for a specific namespace.
func (executor *Executor) CreateInitBlock(config *common.Config) error {
	if edb.IsGenesisFinish(executor.namespace) {
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
	// flush state change to disk immediately
	block := types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Timestamp:  time.Now().Unix(),
		BlockHash:  common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Number:     uint64(0),
		MerkleRoot: root.Bytes(),
	}
	// flush block content to disk immediately
	batch := stateDb.FetchBatch(0)
	if err, _ := edb.PersistBlock(batch, &block, true, true); err != nil {
		return err
	}
	// flush change of chain to disk immediately
	edb.UpdateChain(executor.namespace, batch, &block, true, false, false)
	batch.Write()
	stateDb.MarkProcessFinish(0)
	return nil
}

// NewStateDb - create a empty stateDb handler.
func NewStateDb(conf *common.Config, db db.Database, namespace string) (vm.Database, error) {
	archiveDb, err := hyperdb.GetArchiveDbByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	stateDb, err := hyperstate.New(common.Hash{}, db, archiveDb, conf, 0, namespace)
	if err != nil {
		return nil, err
	}
	return stateDb, nil
}
