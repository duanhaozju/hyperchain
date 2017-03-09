package executor

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	checker "gopkg.in/check.v1"
	"hyperchain/core"
	"hyperchain/core/types"
	"testing"
	edb "hyperchain/core/db_utils"
	"hyperchain/crypto"
	"hyperchain/common"
)

const (
	defaultGas      int64 = 10000
	defaustGasPrice int64 = 10000
	namespace             = "testing"
	configPath            = "../../config/global.yaml"
	dbConfigPath          = "../../config/db.yaml"
)

var (
	conf *common.Config = common.NewConfig(configPath)
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
	conf.MergeConfig(dbConfigPath)//todo:refactor it
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	edb.InitDBForNamespace(conf, namespace, dbConfigPath, 8001)
	suite.executor = NewExecutor(namespace, nil, conf, kec256Hash, encryption, nil)
}

// Run before each test or benchmark starts running.
func (suite *SimulateSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *SimulateSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *SimulateSuite) TearDownSuite(c *checker.C) {

}

/*
	Functional test
*/
func (suite *SimulateSuite) TestSimulate(c *checker.C) {
	core.CreateInitBlock(namespace, conf)
	if err := suite.simulateForTransafer(suite.executor); err != nil {
		c.Errorf("simulate transfer test failed. %s", err)
	}
}

func (suite *SimulateSuite) simulateForTransafer(executor *Executor) error {
	value := types.NewTransactionValue(defaustGasPrice, defaultGas, 3, nil, false)
	data, err := proto.Marshal(value)
	if err != nil {
		return err
	}
	transaction := &types.Transaction{
		Version:   []byte("1.0"),
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        []byte("d4084c9423785f3d8f69b0d236fb072d6b833142"),
		Value:     data,
		Timestamp: 1489051319894944577,
		Nonce:     5215324974019300734,
		Signature: []byte("42eaa8b4ea5c3e4d580757458e2b7f9382597cedb61a23d5019afa17f2bc9c0019c317b60a33e2a6f5997058c9e0e73c6d787d72cea32e61de4eed21a3e1fa4500"),
		Id:        1,
	}
	if err := executor.RunInSandBox(transaction); err != nil {
		return err
	}

	// check receipt existence
	if receipt := edb.GetReceipt(namespace, transaction.GetTransactionHash()); receipt == nil {
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

/*
	test utils
 */

func initConfig(configPath string) {
}
