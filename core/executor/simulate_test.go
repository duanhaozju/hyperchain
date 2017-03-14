package executor

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	checker "gopkg.in/check.v1"
	"hyperchain/core/types"
	"testing"
	edb "hyperchain/core/db_utils"
	"hyperchain/common"
	"hyperchain/core/test_util"
	"os"
)

const (
	defaultGas      int64 = 10000
	defaustGasPrice int64 = 10000
	namespace             = "testing"
	configPath            = "./config/global.yaml"
	dbConfigPath          = "./config/db.yaml"
)

var (
	conf *common.Config
)

type SimulateSuite struct {
	executor *Executor
}

func Test(t *testing.T) {
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
	os.Chdir("../..")
	os.RemoveAll("./build")
	conf = test_util.InitConfig(configPath, dbConfigPath)
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
	value := types.NewTransactionValue(defaustGasPrice, defaultGas, 9, nil, false)
	data, err := proto.Marshal(value)
	if err != nil {
		return err
	}
	transaction := &types.Transaction{
		Version:   []byte("1.0"),
		From:      common.Hex2Bytes("0xe93b92f1da08f925bdee44e91e7768380ae83307"),
		To:        common.Hex2Bytes("0xc8f233096e6d0b4241939593340255353460cd63"),
		Value:     data,
		Timestamp: 1489390658651237328,
		Nonce:     5306948822540864594,
		Signature: []byte("2e5ecff88359e5bb8590d2efbc609c4b3fe1ead066bc81b4153b1557d4f93fcc2d8db2abe0d3ff6e586c4f7ce7797ef85aadc85accfd5bf7c28db354738370bb00"),
		Id:        1,
	}
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

