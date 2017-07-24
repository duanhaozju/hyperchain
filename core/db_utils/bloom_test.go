package db_utils

import (
	"hyperchain/common"
	"testing"
	"hyperchain/hyperdb"
	checker "gopkg.in/check.v1"
	"os"
	"hyperchain/core/types"
	"crypto/rand"
)

type BloomSuite struct{}

func TestBloom(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&BloomSuite{})

var (
	config *common.Config
	ns     string = common.DEFAULT_NAMESPACE
	dbconf string = "../../configuration/namespaces/global/config/db.yaml"
	globalconf string = "../../configuration/namespaces/global/config/global.yaml"
	dbList []string = []string{"Archive", "blockchain", "Consensus", "namespaces"}
)

// Run once when the suite starts running.
func (suite *BloomSuite) SetUpSuite(c *checker.C) {
	// init conf
	config = common.NewConfig(globalconf)
	config.MergeConfig(dbconf)
	config.Set(RebuildTime, 3)
	config.Set(RebuildInterval, 24)
	config.Set(BloomBit, 10000)
	// init db
	hyperdb.StartDatabase(config, ns)
	// init logger
	config.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, config)
}

// Run before each test or benchmark starts running.
func (suite *BloomSuite) SetUpTest(c *checker.C) {
	db, _ := hyperdb.GetDBDatabaseByNamespace(common.DEFAULT_NAMESPACE)
	putChain(db.NewBatch(), &types.Chain{}, true, true)
}

// Run after each test or benchmark runs.
func (suite *BloomSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *BloomSuite) TearDownSuite(c *checker.C) {
	for _, dbname := range dbList {
		os.RemoveAll(dbname)
	}
}
func (suite *BloomSuite) TestLook(c *checker.C) {
	var (
		err error
		txs []*types.Transaction = RandomTxs()
		another []*types.Transaction = RandomTxs()
	)
	cache := NewBloomCache(config)
	// defer cache.Close()
	err = cache.Register(ns)
	if err != nil {
		c.Error(err.Error())
	}
	WriteTxBloomFilter(common.DEFAULT_NAMESPACE, txs)
	for _, tx := range txs {
		_, res := LookupTransaction(common.DEFAULT_NAMESPACE, tx.GetHash())
		if res != true {
			c.Error("expect to be true")
		}
	}
	for _, tx := range another {
		_, res := LookupTransaction(common.DEFAULT_NAMESPACE, tx.GetHash())
		if res != false {
			c.Error("expect to be true")
		}
	}
}

func (suite *blockSuite) BenchmarkWrite(c *checker.C) {
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
		WriteTxBloomFilter(common.DEFAULT_NAMESPACE, RandomTxs())
	}
}

func RandomTxs() []*types.Transaction {
	var txs []*types.Transaction
	hashBuf := make([]byte, 32)
	for i := 0; i < 50; i += 1 {
		rand.Read(hashBuf)
		txs = append(txs, &types.Transaction{TransactionHash:hashBuf})
	}
	return txs
}


