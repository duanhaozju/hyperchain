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
	"hyperchain/core/types"
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
	edb.InitDBForNamespace(conf, namespace)
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
	if err := suite.simulateForInvalidTransafer(suite.executor); err.Error() != "OUTOFBALANCE" {
		c.Errorf("simulate invalid transfer test failed. %s", err)
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

	//for _,v := range transactions {
	//	intValue, _ := strconv.ParseInt(string(v.Value), 10, 0)
	//	value := types.NewTransactionValue(10000, 10000, intValue, nil, false)
	//	data, err := proto.Marshal(value)
	//	if err != nil {
	//		return err
	//	}
	//	v.Value = data
	//	v.TransactionHash = v.Hash().Bytes()
	//	if err := executor.RunInSandBox(v); err != nil {
	//		return err
	//	}
	//	if receipt := edb.GetReceipt(namespace, v.GetHash()); receipt == nil {
	//		return errors.New("no receipt found in database")
	//	}
	//}
	return nil
}

func (suite *SimulateSuite) simulateForInvalidTransafer(executor *Executor) error {
	transaction := tutil.GenInvalidTransferTransactionRandomly()
	transaction.TransactionHash = transaction.Hash().Bytes()
	if err := executor.RunInSandBox(transaction); err != nil {
		return err
	}

	receipt := edb.GetReceipt(namespace, transaction.GetHash())
	if receipt == nil {
		errType, _ := edb.GetInvaildTxErrType(namespace, transaction.GetHash().Bytes())
		return errors.New(errType.String())
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

var transactions = []*types.Transaction{
	&types.Transaction{
		Version:   []byte("1.0"),
		From:      common.Hex2Bytes("0x6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        common.Hex2Bytes("0x7a0f59b160ddcb5e58f3f11a7c97e19a3b62933e"),
		Value:     []byte("1"),
		Timestamp: 1489469156170800342,
		Nonce:     3748177677647696353,
		Signature: []byte("31dccc3cdada52a0f3c9461213db568211c49e6ec4aee2432ea5cfb6b8c5cb6d0ee98a89170d33e175149da2e4d0e83c01b7ee4fc5acb345e58f4db7df6aa45501"),
		Id:        1,
	},
	&types.Transaction{
		Version:   []byte("1.0"),
		From:      common.Hex2Bytes("0xb18c8575e3284e79b92100025a31378feb8100d6"),
		To:        common.Hex2Bytes("0x77becdee7c1fc54ed63f910ef085fc40668eed05"),
		Value:     []byte("5"),
		Timestamp: 1489469219189091668,
		Nonce:     281010826848111663,
		Signature: []byte("d0fe62123b3a9035d638beaca14fe19ed24039f95c6c7edecd54f6873fa975734c187315119a9a49fd134ac1ab07412b1cf281ec27618f3056b328f56478e0a600"),
		Id:        1,
	},
	&types.Transaction{
		Version:   []byte("1.0"),
		From:      common.Hex2Bytes("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"),
		To:        common.Hex2Bytes("0xac461094785b42981e8312d7dca72545ed9f61db"),
		Value:     []byte("10"),
		Timestamp: 1489468913867871255,
		Nonce:     7651229937773117423,
		Signature: []byte("3c4abe87077bacf3284fdd5beac8323bcfe343088b1a2806a27630fe07a04800245b29d83039510cefdd4d8081c963fe03d33a2df590c2e85a91f0a9671da45a01"),
		Id:        1,
	},
	&types.Transaction{
		Version:   []byte("1.0"),
		From:      common.Hex2Bytes("0xb18c8575e3284e79b92100025a31378feb8100d6"),
		To:        common.Hex2Bytes("0x7a594e3049bfd91f668472f236c87da8df77de29"),
		Value:     []byte("100"),
		Timestamp: 1489469268546881606,
		Nonce:     6999655251710596396,
		Signature: []byte("5c57375135296d7d1eaf4cccc9f173cacec083ac2461dc10f69a1a16b8ae25827bd3a98276b8e5be1199aa6312d7ef0504ae6824c94802bb2f807e1a72152bf501"),
		Id:        1,
	},
}
