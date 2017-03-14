package executor

import (
	"github.com/davecgh/go-spew/spew"
	checker "gopkg.in/check.v1"
	tutil "hyperchain/core/test_util"
	"testing"
	"hyperchain/core/types"
	"os"
	"path"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
)

type ValidationSuite struct {
	executor *Executor
	owd      string
}

var _ = checker.Suite(&ValidationSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

func TestValidation(t *testing.T) {
	checker.TestingT(t)
}

// Run once when the suite starts running.
func (suite *ValidationSuite) SetUpSuite(c *checker.C) {
	// initialize block pool
	suite.owd, _ = os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain"))
	os.RemoveAll("./build")
	conf = tutil.InitConfig(configPath, dbConfigPath)
	edb.InitDBForNamespace(conf, namespace, dbConfigPath, 8001)
	suite.executor = NewExecutor(namespace, conf, nil)
	suite.executor.CreateInitBlock(conf)
	suite.executor.initialize()
}

// Run before each test or benchmark starts running.
func (suite *ValidationSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *ValidationSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *ValidationSuite) TearDownSuite(c *checker.C) {
	os.RemoveAll("./build")
	os.RemoveAll("./db.log")
	os.Chdir(suite.owd)
}

func (suite *ValidationSuite) TestCheckSign(c *checker.C) {
	var txs []*types.Transaction
	for i := 0; i < 5; i += 1 {
		txs = append(txs, tutil.GenTransferTransactionRandomly())
	}
	txs[0].Signature = []byte("fake sign")
	txs[1].Signature = []byte("fake sign")
	txs[4].Signature = []byte("fake sign")
	ctxs := make([]*types.Transaction, len(txs))
	copy(ctxs, txs)
	invalidtxs, ntxs := suite.executor.checkSign(txs)
	c.Assert(len(invalidtxs), checker.Equals, 3)
	c.Assert(len(ntxs), checker.Equals, 2)

	c.Assert(txs[0].GetHash().Hex(), checker.Equals, ctxs[2].GetHash().Hex())
	c.Assert(txs[1].GetHash().Hex(), checker.Equals, ctxs[3].GetHash().Hex())
}


