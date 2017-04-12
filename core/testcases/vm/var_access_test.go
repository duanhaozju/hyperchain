package vm

import (
	"testing"
	"hyperchain/core/executor"
	"hyperchain/common"
	checker "gopkg.in/check.v1"
	"hyperchain/hyperdb"
	"hyperchain/core/db_utils"
)

var (
	configPath = "../../../configuration/namespaces/global/config/global.yaml"
	globalConfig *common.Config
)
func init() {
	globalConfig = common.NewConfig(configPath)
	common.InitHyperLoggerManager(globalConfig)
	globalConfig.Set(common.NAMESPACE, "global")
	common.InitHyperLogger(globalConfig)
}
type VarAccessSuite struct {
	executor *executor.Executor
}

func TestVarAccess(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&VarAccessSuite{})

// Run once when the suite starts running.
func (suite *VarAccessSuite) SetUpSuite(c *checker.C) {
	db_utils.InitDBForNamespace(globalConfig, "global")
	hyperdb.StartDatabase(globalConfig, "global")
	suite.executor = executor.NewExecutor("global", globalConfig, nil)
	suite.executor.Start()
}

// Run before each test or benchmark starts running.
func (suite *VarAccessSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *VarAccessSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *VarAccessSuite) TearDownSuite(c *checker.C) {
}

func (suite *VarAccessSuite) TestVisitVars(c *checker.C) {

}

