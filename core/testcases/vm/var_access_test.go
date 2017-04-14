package vm

import (
	"testing"
	"hyperchain/core/executor"
	"hyperchain/common"
	checker "gopkg.in/check.v1"
	"hyperchain/hyperdb"
	"hyperchain/core/db_utils"
	"os"
	"path"
	"fmt"
	util "hyperchain/core/test_util"
	"io/ioutil"
	"encoding/json"
)

var (
	configPath = "./namespaces/global/config/global.yaml"
	globalConfig *common.Config
)

const (
	GENESIS_ACCT = "6201cb0448964ac597faf6fdf1f472edf2a22b89"
)

type VarAccessSuite struct {
	executor *executor.Executor
	owd      string
}

type Operation struct {
	MethodName  string  `json:"method"`
	Payload     string  `json:"payload"`
	Expect      string  `json:"expect"`
}


func TestVarAccess(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&VarAccessSuite{})

// Run once when the suite starts running.
func (suite *VarAccessSuite) SetUpSuite(c *checker.C) {
	// switch execute location
	switchToExeLoc(suite)
	initLog()

	db_utils.InitDBForNamespace(globalConfig, "global")
	hyperdb.StartDatabase(globalConfig, "global")
	suite.executor = executor.NewExecutor("global", globalConfig, nil)
	suite.executor.CreateInitBlock(globalConfig)
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
	os.RemoveAll("./namespaces/global/data")
	switchBack(suite)
}

func (suite *VarAccessSuite) TestVisitVars(c *checker.C) {
	// deploy
	bin, _ := ioutil.ReadFile("../core/testcases/vm/test_resource/variable_access/bin")
	tx := util.GenContractTransactionRandomly(GENESIS_ACCT, "", string(bin))
	tx.To = nil
	suite.executor.FetchStateDb().MarkProcessStart(1)
	_, _, addr, err := suite.executor.ExecTransaction(suite.executor.FetchStateDb(), tx, 0, 1)
	if err != nil {
		fmt.Println(err.Error())
	}
	suite.executor.FetchStateDb().FetchBatch(1).Write()
	// query
	var operations []Operation
	buf, _ := ioutil.ReadFile("../core/testcases/vm/test_resource/variable_access/var_access.json")
	json.Unmarshal(buf, &operations)
	for idx, oper := range operations {
		tx := util.GenContractTransactionRandomly(GENESIS_ACCT, addr.Hex(), oper.Payload)
		receipt, _, _, _ := suite.executor.ExecTransaction(suite.executor.FetchStateDb(), tx, 0, 1)
		if common.Bytes2Hex(receipt.Ret) != oper.Expect {
			c.Errorf("operation %d get variable failed", idx)
		}
	}
}

func switchToExeLoc(suite *VarAccessSuite) {
	suite.owd, _ = os.Getwd()
	os.Chdir(path.Join(common.GetGoPath(), "src/hyperchain/configuration"))
}

func switchBack(suite *VarAccessSuite) {
	os.Chdir(suite.owd)
}

func initLog() {
	globalConfig = common.NewConfig(configPath)
	common.InitHyperLoggerManager(globalConfig)
	globalConfig.Set(common.NAMESPACE, "global")
	common.InitHyperLogger(globalConfig)
}
