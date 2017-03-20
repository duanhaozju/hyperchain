package executor

import (
	"io/ioutil"
	"github.com/buger/jsonparser"
	"math/big"
	"time"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/core/hyperstate"
	edb "hyperchain/core/db_utils"
	"hyperchain/common"
	"strconv"
	"hyperchain/core/types"
)

const (
	genesisPath  = "global.configs.genesis"
)

// CreateInitBlock - create genesis for a specific namespace.
func (executor *Executor) CreateInitBlock(config *common.Config) error {
	if edb.IsGenesisFinish(executor.namespace) {
		log.Infof("[Namespace = %s] already genesis", executor.namespace)
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

	bytes, err := ioutil.ReadFile(getGenesisPath(config))
	if err != nil {
		return err
	}
	// create state instance with empty root hash
	stateDb, err := NewStateDb(executor.namespace, config)
	if err != nil {
		return err
	}
	stateDb.MarkProcessStart(0)
	jsonparser.ObjectEach(bytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		object := stateDb.CreateAccount(common.HexToAddress(string(key)))
		account, _ := strconv.ParseInt(string(value), 10, 64)
		object.AddBalance(big.NewInt(account))
		return nil
	}, "genesis", "alloc")
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
func NewStateDb(namespace string, conf *common.Config) (vm.Database, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		log.Errorf("[Namespace = %s] get database failed", namespace)
		return nil, err
	}
	stateDb, err := hyperstate.New(common.Hash{}, db, conf, 0)
	if err != nil {
		log.Errorf("[Namespace = %s] new stateDb failed, err : %s", namespace, err.Error())
		return nil, err
	}
	return stateDb, nil
}

// getGenesisPath - load genesis file path from config.
func getGenesisPath(conf *common.Config) string {
	return conf.GetString(genesisPath)
}