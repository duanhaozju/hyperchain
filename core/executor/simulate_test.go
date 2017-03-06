package executor

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	checker "gopkg.in/check.v1"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/tree/bucket"
	"testing"
)

const (
	testPort              = 8888
	testDir               = "~/tmp/hyperchain-test"
	defaultGas      int64 = 10000
	defaustGasPrice int64 = 10000
	genesisPath           = "../../config/genesis.json"
)

type SimulateSuite struct {
	rawPool   *Executor
	hyperPool *Executor
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
	core.InitDB(testDir, testPort)
	rawBlockPoolConf := BlockPoolConf{
		BlockVersion:       "1.0",
		TransactionVersion: "1.0",
		StateType:          "rawstate",
	}
	hyperBlockPoolConf := BlockPoolConf{
		BlockVersion:       "1.0",
		TransactionVersion: "1.0",
		StateType:          "hyperstate",
	}
	bucketConf := bucket.Conf{
		StateSize:         19,
		StateLevelGroup:   10,
		StorageSize:       19,
		StorageLevelGroup: 10,
	}
	suite.rawPool = NewBlockExecutor(nil, rawBlockPoolConf, bucketConf)
	suite.hyperPool = NewBlockExecutor(nil, hyperBlockPoolConf, bucketConf)
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
	var blockPool *Executor
	stateType := []string{"rawstate", "hyperstate"}
	for _, state := range stateType {
		switch state {
		case "rawstate":
			blockPool = suite.rawPool
			core.CreateInitBlock(genesisPath, state, "1.0", blockPool.bucketTreeConf)
			if err := suite.simulateForTransafer(blockPool); err != nil {
				c.Errorf("simulate transfer test failed. %s", err)
			}
		case "hyperstate":
			blockPool = suite.hyperPool
			core.CreateInitBlock(genesisPath, state, "1.0", blockPool.bucketTreeConf)
			if err := suite.simulateForTransafer(blockPool); err != nil {
				c.Errorf("simulate transfer test failed. %s", err)
			}
		}
	}
}

func (suite *SimulateSuite) simulateForTransafer(blockPool *Executor) error {
	value := types.NewTransactionValue(defaustGasPrice, defaultGas, 4, nil)
	data, err := proto.Marshal(value)
	if err != nil {
		return err
	}
	transaction := &types.Transaction{
		Version:   []byte("1.0"),
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        []byte("d1914e6e9845a66cae754752565786f6f52b339d"),
		Value:     data,
		Timestamp: 1483081034372214455,
		Nonce:     436124869695996140,
		Signature: []byte("39a3c574f95c825274ea9bd78054863ba3ca70f9ca04c602f1e55219ba177b24475c920f100dd1883ad317f20ba7294fb53d4ef976c169bc790f2630fe56ccf700"),
		Id:        1,
	}
	if err := blockPool.RunInSandBox(transaction); err != nil {
		return err
	}

	// check receipt existence
	if receipt := core.GetReceipt(transaction.GetTransactionHash()); receipt == nil {
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
