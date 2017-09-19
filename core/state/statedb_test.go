package state

import (
	"testing"
	//"testing/quick"
	"bytes"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"hyperchain/hyperdb/mdb"
)

type StateDbSuite struct {
}

func TestStateSuite(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&StateDbSuite{})

// Run once when the suite starts running.
func (suite *StateDbSuite) SetUpSuite(c *checker.C) {
}

// Run before each test or benchmark starts running.
func (suite *StateDbSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *StateDbSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *StateDbSuite) TearDownSuite(c *checker.C) {
}

func (suite *StateDbSuite) TestGetState(c *checker.C) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	state, _ := New(common.Hash{}, db, db, nil, 0, "global")
	addr := common.BytesToAddress([]byte("address"))
	key := common.BytesToHash([]byte("key"))
	value := []byte("value")
	state.MarkProcessStart(1)
	state.CreateAccount(addr)
	state.SetState(addr, key, value, 0)
	state.Commit()
	batch := state.FetchBatch(1)
	batch.Write()
	state.MarkProcessFinish(1)

	exist, newValue := state.GetState(addr, key)
	if bytes.Compare(value, newValue) != 0 || exist == false {
		c.Error("get state failed")
	}
}
