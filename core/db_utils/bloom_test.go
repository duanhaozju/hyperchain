package db_utils

import (
	"crypto/rand"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
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
	config     *common.Config
	ns         string   = common.DEFAULT_NAMESPACE
	dbconf     string   = "../../configuration/namespaces/global/config/db.yaml"
	globalconf string   = "../../configuration/namespaces/global/config/global.yaml"
	dbList     []string = []string{"Archive", "blockchain", "Consensus", "namespaces"}
)

// Run once when the suite starts running.
func (suite *BloomSuite) SetUpSuite(c *checker.C) {
	// init conf
	config = common.NewConfig(globalconf)
	config.MergeConfig(dbconf)
	config.Set(RebuildTime, 3)
	config.Set(RebuildInterval, 24)
	config.Set(BloomBit, 10000)
	// init logger
	config.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(ns, config)
	// init db
	hyperdb.StartDatabase(config, ns)
	// init chan
	InitializeChain(ns)

}

// Run before each test or benchmark starts running.
func (suite *BloomSuite) SetUpTest(c *checker.C) {
	db, _ := hyperdb.GetDBDatabaseByNamespace(ns)
	putChain(db.NewBatch(), &types.Chain{}, true, true)
}

// Run after each test or benchmark runs.
func (suite *BloomSuite) TearDownTest(c *checker.C) {
	for _, dbname := range dbList {
		os.RemoveAll(dbname)
	}
}

// Run once after all tests or benchmarks have finished running.
func (suite *BloomSuite) TearDownSuite(c *checker.C) {}

func (suite *BloomSuite) TestLook(c *checker.C) {
	var (
		err     error
		txs     []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
	)
	cache := NewBloomCache(config)
	// defer cache.Close()
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
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns)
	cache := NewBloomCache(config)
	// defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	block := test_util.BlockCases
	block.Number = 1
	block.Transactions = txs
	block.Timestamp = time.Now().UnixNano()

	WriteTxBloomFilter(ns, txs)
	PersistBlock(db.NewBatch(), &block, true, true)
	UpdateChain(ns, db.NewBatch(), &block, false, true, true)

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
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns)
	cache := NewBloomCache(config)
	// defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	block := test_util.BlockCases
	block.Number = 1
	block.Transactions = txs
	block.Timestamp = time.Now().UnixNano() - 2*24*time.Hour.Nanoseconds()

	WriteTxBloomFilter(ns, txs)
	PersistBlock(db.NewBatch(), &block, true, true)
	UpdateChain(ns, db.NewBatch(), &block, false, true, true)

	cache.rebuild(nil)

	if !checkExist(nil, append(txs, another...)) {
		c.Error("bloom filter check failed")
	}

}

func (suite *BloomSuite) TestBuildConcurrently(c *checker.C) {
	var (
		err     error
		txs     []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
		db      db.Database
	)
	db, _ = hyperdb.GetDBDatabaseByNamespace(ns)
	cache := NewBloomCache(config)
	// defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	block := test_util.BlockCases
	block.Number = 1
	block.Transactions = txs
	block.Timestamp = time.Now().UnixNano()

	barrier := make(chan struct{})

	go func(sig chan struct{}) {
		hook := func() {
			time.Sleep(2 * time.Second)
		}
		cache.rebuild(hook)
		close(sig)
	}(barrier)

	WriteTxBloomFilter(ns, txs)
	PersistBlock(db.NewBatch(), &block, true, true)
	UpdateChain(ns, db.NewBatch(), &block, false, true, true)

	<-barrier

	if !checkExist(txs, another) {
		c.Error("bloom filter check failed")
	}
}

func checkExist(existSet []*types.Transaction, notExistSet []*types.Transaction) bool {
	for _, tx := range existSet {
		_, res := LookupTransaction(ns, tx.GetHash())
		if res != true {
			return false
		}
	}
	for _, tx := range notExistSet {
		_, res := LookupTransaction(ns, tx.GetHash())
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
	cache := NewBloomCache(config)
	// defer cache.Close()
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
