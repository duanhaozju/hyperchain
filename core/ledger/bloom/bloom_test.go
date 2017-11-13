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

package bloom

import (
	"crypto/rand"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/chain"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	checker "gopkg.in/check.v1"
	"os"
	"testing"
	"time"
)

type BloomSuite struct{}

func TestBloom(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&BloomSuite{})

var (
	config *common.Config
	ns     string = common.DEFAULT_NAMESPACE
)

// Run once when the suite starts running.
func (suite *BloomSuite) SetUpSuite(c *checker.C) {
	// init conf
	config = common.NewRawConfig()
	config.Set(RebuildTime, 3)
	config.Set(ActiveTime, "24h")
	config.Set(RebuildInterval, 24)
	config.Set(BloomBit, 10000)
	config.Set(hcom.DB_TYPE, hcom.LDB_DB)
	// init logger
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, config)
	common.GetLogger(common.DEFAULT_NAMESPACE, "bloom")
	common.SetLogLevel(common.DEFAULT_NAMESPACE, "bloom", "NOTICE")
	// init db
	hyperdb.StartDatabase(config, ns)
	// init chan
	chain.InitializeChain(ns)
}

// Run before each test or benchmark starts running.
func (suite *BloomSuite) SetUpTest(c *checker.C) {
	db, _ := hyperdb.GetDBDatabaseByNamespace(ns, hcom.DBNAME_BLOCKCHAIN)
	chain.PutChain(db.NewBatch(), &types.Chain{}, true, true)
}

// Run after each test or benchmark runs.
func (suite *BloomSuite) TearDownTest(c *checker.C) {
	os.RemoveAll("namespaces")
}

// Run once after all tests or benchmarks have finished running.
func (suite *BloomSuite) TearDownSuite(c *checker.C) {}

func (suite *BloomSuite) TestLook(c *checker.C) {
	var (
		err     error
		txs     []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
	)
	cache := NewBloomFilterCache(config)
	cache.Start()
	defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	WriteTxBloomFilter(ns, txs)
	if !checkExist(txs, another) {
		c.Error("bloom filter check failed")
	}
}

func (suite *BloomSuite) TestBuildInTimeScope(c *checker.C) {
	var (
		err     error
		txs     []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
		db      db.Database
	)
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns, hcom.DBNAME_BLOCKCHAIN)
	cache := NewBloomFilterCache(config)
	cache.Start()
	defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	block := RandomBlock()
	block.Number = 1
	block.Transactions = txs
	block.Timestamp = time.Now().UnixNano()

	chain.PersistBlock(db.NewBatch(), block, true, true)
	chain.UpdateChain(ns, db.NewBatch(), block, false, true, true)

	cache.rebuild(nil)

	if !checkExist(txs, another) {
		c.Error("bloom filter check failed")
	}
}

func (suite *BloomSuite) TestBuildOutTimeScope(c *checker.C) {
	var (
		err     error
		txs     []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
		db      db.Database
	)
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns, hcom.DBNAME_BLOCKCHAIN)
	cache := NewBloomFilterCache(config)
	cache.Start()
	defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	block := RandomBlock()
	block.Number = 1
	block.Transactions = txs
	block.Timestamp = time.Now().UnixNano() - 2*24*time.Hour.Nanoseconds()

	chain.PersistBlock(db.NewBatch(), block, true, true)
	chain.UpdateChain(ns, db.NewBatch(), block, false, true, true)

	cache.rebuild(nil)

	if !checkExist(nil, append(txs, another...)) {
		c.Error("bloom filter check failed")
	}

}

func (suite *BloomSuite) TestBuildConcurrently(c *checker.C) {
	var (
		err      error
		txs      []*types.Transaction
		checkTxs []*types.Transaction
		another  []*types.Transaction = RandomTxs()
		db       db.Database
	)
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns, hcom.DBNAME_BLOCKCHAIN)
	cache := NewBloomFilterCache(config)
	cache.Start()
	defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}

	barrier := make(chan struct{})

	go func(sig chan struct{}) {
		hook := func() {
			time.Sleep(1 * time.Second)
		}
		cache.rebuild(hook)
		close(sig)
	}(barrier)

	for i := 0; i < 100; i += 1 {
		block := RandomBlock()
		txs = RandomTxs()
		checkTxs = append(checkTxs, txs...)
		block.Number = uint64(i + 1)
		block.Transactions = txs
		block.Timestamp = time.Now().UnixNano()
		WriteTxBloomFilter(ns, txs)
		chain.PersistBlock(db.NewBatch(), block, true, true)
		chain.UpdateChain(ns, db.NewBatch(), block, false, true, true)
	}
	<-barrier
	if !checkExist(checkTxs, another) {
		c.Error("bloom filter check failed")
	}
}

func checkExist(existSet []*types.Transaction, notExistSet []*types.Transaction) bool {
	for _, tx := range existSet {
		res, _ := LookupTransaction(ns, tx.GetHash())
		if res != true {
			return false
		}
	}
	for _, tx := range notExistSet {
		res, _ := LookupTransaction(ns, tx.GetHash())
		if res != false {
			return false
		}
	}
	return true
}

func (suite *BloomSuite) BenchmarkWrite(c *checker.C) {
	var (
		err error
	)
	cache := NewBloomFilterCache(config)
	cache.Start()
	defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	for i := 0; i < c.N; i += 1 {
		WriteTxBloomFilter(ns, RandomTxs())
	}
}

func RandomTxs() []*types.Transaction {
	var txs []*types.Transaction
	hashBuf := make([]byte, 32)
	for i := 0; i < 50; i += 1 {
		rand.Read(hashBuf)
		txs = append(txs, &types.Transaction{TransactionHash: hashBuf})
	}
	return txs
}

func RandomBlock() *types.Block {
	blockHash := make([]byte, 32)
	parentHash := make([]byte, 32)
	rand.Read(blockHash)
	rand.Read(parentHash)
	return &types.Block{
		BlockHash:  blockHash,
		ParentHash: parentHash,
	}
}
