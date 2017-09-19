package state

import (
	"bytes"
	"fmt"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	tutil "hyperchain/core/test_util"
	"hyperchain/hyperdb/mdb"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

type RevertSuite struct {
}

func TestRevert(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&RevertSuite{})

// Run once when the suite starts running.
func (suite *RevertSuite) SetUpSuite(c *checker.C) {
}

// Run before each test or benchmark starts running.
func (suite *RevertSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *RevertSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *RevertSuite) TearDownSuite(c *checker.C) {
}

func (suite *RevertSuite) TestRevertRandom(c *checker.C) {
	config := &quick.Config{MaxCount: 3}
	err := quick.Check((*revertTest).run, config)
	if cerr, ok := err.(*quick.CheckError); ok {
		test := cerr.In[0].(*revertTest)
		c.Errorf("%v:\n%s", test.err, test)
	} else if err != nil {
		c.Error(err)
	}
}

// The test works as follows:
//
// A new state is created and all actions are applied to it. Several snapshots are taken
// in between actions. The test then reverts each snapshot. For each snapshot the actions
// leading up to it are replayed on a fresh, empty state. The behaviour of all public
// accessor methods on the reverted state must match the return value of the equivalent
// methods on the replayed state.
type revertTest struct {
	addrs        []common.Address // all account addresses
	actions      []testAction     // modifications to the state
	actionsAfter []testAction     // modifications to the state
	err          error            // failure details are reported through this field
}

// Generate returns a new snapshot test of the given size. All randomness is
// derived from r.
func (*revertTest) Generate(r *rand.Rand, size int) reflect.Value {
	// Generate random actions.
	addrs := make([]common.Address, 10)
	for i := range addrs {
		addrs[i] = common.HexToAddress(RandomString(40))
	}
	actions := make([]testAction, size)
	for i := range actions {
		addr := addrs[r.Intn(len(addrs))]
		actions[i] = newTestAction(addr, r)
	}

	actionsAfter := make([]testAction, size)
	for i := range actionsAfter {
		addr := addrs[r.Intn(len(addrs))]
		actionsAfter[i] = newTestAction(addr, r)
	}
	return reflect.ValueOf(&revertTest{addrs, actions, actionsAfter, nil})
}

func (test *revertTest) String() string {
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "Action Before, totally %d", len(test.actions))
	for i, action := range test.actions {
		fmt.Fprintf(out, "%4d: %s\n", i, action.name)
	}
	fmt.Fprintf(out, "Action After, totally %d", len(test.actionsAfter))
	for i, action := range test.actionsAfter {
		fmt.Fprintf(out, "%4d: %s\n", i, action.name)
	}
	return out.String()
}

func (test *revertTest) run() bool {
	// Run all actions and create snapshots.
	var (
		db, _         = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
		state, _      = New(common.Hash{}, db, db, tutil.InitConfig(configPath), 10, common.DEFAULT_NAMESPACE)
		checkDb, _    = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
		checkState, _ = New(common.Hash{}, checkDb, checkDb, tutil.InitConfig(configPath), 10, common.DEFAULT_NAMESPACE)
		immediateRoot []byte
	)

	immediateRoot = applyActions(state, test.actions, 11)
	applyActions(state, test.actionsAfter, 12)

	batch := db.NewBatch()
	state.RevertToJournal(11, 12, immediateRoot, batch)
	batch.Write()

	applyActions(checkState, test.actions, 11)
	if err := test.checkEqual(state, checkState); err != nil {
		test.err = fmt.Errorf("state mismatch after revert to snapshot \n%v", err)
		return false
	}
	return true
}

// checkEqual checks that methods of state and checkstate return the same values.
func (test *revertTest) checkEqual(state, checkstate *StateDB) error {
	for _, addr := range test.addrs {
		var err error
		checkeq := func(op string, a, b interface{}) bool {
			if err == nil && !reflect.DeepEqual(a, b) {
				err = fmt.Errorf("got %s(%s) == %v, want %v", op, addr.Hex(), a, b)
				return false
			}
			return true
		}
		// Check basic accessor methods.
		checkeq("Exist", state.Exist(addr), checkstate.Exist(addr))
		checkeq("IsDeleted", state.IsDeleted(addr), checkstate.IsDeleted(addr))
		checkeq("GetBalance", state.GetBalance(addr), checkstate.GetBalance(addr))
		checkeq("GetNonce", state.GetNonce(addr), checkstate.GetNonce(addr))
		checkeq("GetCode", state.GetCode(addr), checkstate.GetCode(addr))
		checkeq("GetCodeHash", state.GetCodeHash(addr), checkstate.GetCodeHash(addr))
		checkeq("GetCodeSize", state.GetCodeSize(addr), checkstate.GetCodeSize(addr))
		// Check storage.
		if obj := state.GetStateObject(addr); obj != nil {
			for key, val := range obj.cachedStorage {
				_, checkvalue := checkstate.GetState(addr, key)
				checkeq("GetState("+key.Hex()+")", val, checkvalue)
			}
		}
		if obj := checkstate.GetStateObject(addr); obj != nil {
			for key, val := range obj.cachedStorage {
				_, value := state.GetState(addr, key)
				checkeq("GetState("+key.Hex()+")", val, value)
			}
		}
		if err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(state.GetLogs(common.Hash{}), checkstate.GetLogs(common.Hash{})) {
		return fmt.Errorf("got GetLogs(common.Hash{}) == %v, want GetLogs(common.Hash{}) == %v",
			state.GetLogs(common.Hash{}), checkstate.GetLogs(common.Hash{}))
	}
	return nil
}

func applyActions(s *StateDB, actions []testAction, seqNo uint64) []byte {
	s.MarkProcessStart(seqNo)
	for _, action := range actions {
		action.fn(action, s)
	}
	hash, _ := s.Commit()
	batch := s.FetchBatch(seqNo)
	batch.Write()
	s.MarkProcessFinish(seqNo)
	s.Purge()
	return hash.Bytes()
}
