package executor

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	checker "gopkg.in/check.v1"
	"testing"
	edb "hyperchain/core/db_utils"
	"hyperchain/common"
	tutil "hyperchain/core/test_util"
	"os"
	"path"
)

const (
	namespace             = "testing"
	configPath            = "./config/global.yaml"
	dbConfigPath          = "./config/db.yaml"
)

var (
	conf *common.Config
)

type SimulateSuite struct {
	executor *Executor
	owd      string
}

func TestSimulate(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&SimulateSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

// Run once when the suite starts running.
func (suite *SimulateSuite) SetUpSuite(c *checker.C) {
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
func (suite *SimulateSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *SimulateSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *SimulateSuite) TearDownSuite(c *checker.C) {
	os.RemoveAll("./build")
	os.RemoveAll("./db.log")
	os.Chdir(suite.owd)
}

/*
	Functional test
*/
func (suite *SimulateSuite) TestSimulate(c *checker.C) {
	if err := suite.simulateForTransafer(suite.executor); err != nil {
		c.Errorf("simulate transfer test failed. %s", err)
	}
}

func (suite *SimulateSuite) simulateForTransafer(executor *Executor) error {
	transaction := tutil.GenTransferTransactionRandomly()
	transaction.TransactionHash = transaction.Hash().Bytes()
	if err := executor.RunInSandBox(transaction); err != nil {
		return err
	}

	// check receipt existence
	if receipt := edb.GetReceipt(namespace, transaction.GetHash()); receipt == nil {
		return errors.New("no receipt found in database")
	}
	return nil
}

func (suite *SimulateSuite) simulateForDeploy() error {
	return nil
}

func (suite *SimulateSuite) simulateForInvocation() error {
	return nil
}

/*
	benchmarking
*/
func (suite *SimulateSuite) BenchmarkSimulate(c *checker.C) {
	for i := 0; i < c.N; i++ {
	}
}

