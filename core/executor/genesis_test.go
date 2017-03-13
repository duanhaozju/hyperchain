package executor

import (
	"github.com/davecgh/go-spew/spew"
	checker "gopkg.in/check.v1"
	"os"
	"hyperchain/core/test_util"
	edb "hyperchain/core/db_utils"
)

type GenesisSuite struct {
	executor *Executor
}

var _ = checker.Suite(&GenesisSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

// Run once when the suite starts running.
func (suite *GenesisSuite) SetUpSuite(c *checker.C) {
	// initialize block pool
	os.Chdir("../..")
	os.RemoveAll("./build")
	conf = test_util.InitConfig(configPath, dbConfigPath)
	edb.InitDBForNamespace(conf, namespace, dbConfigPath, 8001)
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
	os.RemoveAll("./build")
	os.RemoveAll("./db.log")
}

func (suite *GenesisSuite) TestInitGenesis(c *checker.C) {
	suite.executor.CreateInitBlock(conf)
	_, err := edb.GetBlockByNumber(namespace, 0)
	if err != nil {
		c.Error("load genesis block failed")
	}
}

