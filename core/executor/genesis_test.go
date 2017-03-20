package executor

import (
	"github.com/davecgh/go-spew/spew"
	checker "gopkg.in/check.v1"
	"os"
	tutil "hyperchain/core/test_util"
	edb "hyperchain/core/db_utils"
	"bytes"
	"testing"
	"path"
	"hyperchain/common"
)

type GenesisSuite struct {
	executor *Executor
	owd      string
}

var _ = checker.Suite(&GenesisSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

func TestGenesis(t *testing.T) {
	checker.TestingT(t)
}
// Run once when the suite starts running.
func (suite *GenesisSuite) SetUpSuite(c *checker.C) {
	// initialize block pool
	suite.owd, _ = os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/configuration"))
	os.RemoveAll("./namespaces/global/data")
	conf = tutil.InitConfig(configPath)
	edb.InitDBForNamespace(conf, namespace)
	suite.executor = NewExecutor(namespace, conf, nil)
}

// Run before each test or benchmark starts running.
func (suite *GenesisSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *GenesisSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *GenesisSuite) TearDownSuite(c *checker.C) {
	os.RemoveAll("./namespaces/global/data")
	os.Chdir(suite.owd)
}

func (suite *GenesisSuite) TestInitGenesis(c *checker.C) {
	suite.executor.CreateInitBlock(conf)
	genesisBlk, err := edb.GetBlockByNumber(namespace, 0)
	if err != nil {
		c.Error("load genesis block failed")
	}
	if bytes.Compare(genesisBlk.BlockHash, edb.GetLatestBlockHash(namespace)) != 0 {
		c.Error("block hash mismatch")
	}
	if bytes.Compare(genesisBlk.ParentHash, edb.GetParentBlockHash(namespace)) != 0 {
		c.Error("parent hash mismatch")
	}
	if edb.GetTxSumOfChain(namespace) != 0 {
		c.Error("tx sum not equal with 0")
	}
	if edb.GetHeightOfChain(namespace) != 0 {
		c.Error("chain height not equal with 0")
	}
}

