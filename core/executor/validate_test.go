package executor

import (
	"github.com/davecgh/go-spew/spew"
	checker "gopkg.in/check.v1"
	tutil "hyperchain/core/test_util"
	"testing"
)

type ValidationSuite struct {
	//executor *Executor
}

var _ = checker.Suite(&ValidationSuite{})

func init() {
	spew.Config.Indent = "    "
	spew.Config.DisableMethods = true
}

func TestValidation(t *testing.T) {
	checker.TestingT(t)
}

func (suite *ValidationSuite) TestCheckSign(c *checker.C) {
	tutil.GenTransferTransactionRandomly()
}
