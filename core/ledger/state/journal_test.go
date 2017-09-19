package state

import (
	"bytes"
	"encoding/binary"
	"fmt"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	tutil "hyperchain/core/test_util"
	//"hyperchain/core/vm"
	"hyperchain/hyperdb/mdb"
	"math"
	//"math/big"
	"hyperchain/core/types"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

var (
	configPath = "../../configuration/namespaces/global/config/namespace.toml"
)

func init() {
	common.InitHyperLoggerManager(tutil.InitConfig(configPath))
	common.InitRawHyperLogger(common.DEFAULT_NAMESPACE)
}

type JournalSuite struct {
}

func TestJournal(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&JournalSuite{})

// Run once when the suite starts running.
func (suite *JournalSuite) SetUpSuite(c *checker.C) {
}

// Run before each test or benchmark starts running.
func (suite *JournalSuite) SetUpTest(c *checker.C) {
}

// Run after each test or benchmark runs.
func (suite *JournalSuite) TearDownTest(c *checker.C) {
}

// Run once after all tests or benchmarks have finished running.
func (suite *JournalSuite) TearDownSuite(c *checker.C) {
}

func (suite *JournalSuite) TestSnapshotRandom(c *checker.C) {
	config := &quick.Config{MaxCount: 3}
	err := quick.Check((*snapshotTest).run, config)
	if cerr, ok := err.(*quick.CheckError); ok {
		test := cerr.In[0].(*snapshotTest)
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
type snapshotTest struct {
	addrs     []common.Address // all account addresses
	actions   []testAction     // modifications to the state
	snapshots []int            // actions indexes at which snapshot is taken
	err       error            // failure details are reported through this field
}

type testAction struct {
	name   string
	fn     func(testAction, *StateDB)
	args   []int64
	noAddr bool
}

// newTestAction creates a random action that changes state.
func newTestAction(addr common.Address, r *rand.Rand) testAction {
	actions := []testAction{
		{
			name: "SetBalance",
			fn: func(a testAction, s *StateDB) {
				s.SetBalance(addr, big.NewInt(a.args[0]))
			},
			args: make([]int64, 1),
		},
		{
			name: "AddBalance",
			fn: func(a testAction, s *StateDB) {
				s.AddBalance(addr, big.NewInt(a.args[0]))
			},
			args: make([]int64, 1),
		},
		{
			name: "SetNonce",
			fn: func(a testAction, s *StateDB) {
				s.SetNonce(addr, uint64(a.args[0]))
			},
			args: make([]int64, 1),
		},
		{
			name: "SetState",
			fn: func(a testAction, s *StateDB) {
				var key, val common.Hash
				binary.BigEndian.PutUint16(key[:], uint16(a.args[0]))
				binary.BigEndian.PutUint16(val[:], uint16(a.args[1]))
				s.SetState(addr, key, val.Bytes(), 0)
			},
			args: make([]int64, 2),
		},
		{
			name: "SetCode",
			fn: func(a testAction, s *StateDB) {
				code := make([]byte, 16)
				binary.BigEndian.PutUint64(code, uint64(a.args[0]))
				binary.BigEndian.PutUint64(code[8:], uint64(a.args[1]))
				s.SetCode(addr, code)
			},
			args: make([]int64, 2),
		},
		{
			name: "SetCreator",
			fn: func(a testAction, s *StateDB) {
				creator := make([]byte, 16)
				binary.BigEndian.PutUint64(creator, uint64(a.args[0]))
				binary.BigEndian.PutUint64(creator[8:], uint64(a.args[1]))
				s.SetCreator(addr, common.BytesToAddress(creator))
			},
			args: make([]int64, 2),
		},
		{
			name: "SetStatus",
			fn: func(a testAction, s *StateDB) {
				if a.args[0] < 10000/2 {
					s.SetStatus(addr, STATEOBJECT_STATUS_NORMAL)
				} else {
					s.SetStatus(addr, STATEOBJECT_STATUS_FROZON)
				}
			},
			args: make([]int64, 1),
		},
		{
			name: "DeployedContractChange",
			fn: func(a testAction, s *StateDB) {
				contract := make([]byte, 16)
				binary.BigEndian.PutUint64(contract, uint64(a.args[0]))
				binary.BigEndian.PutUint64(contract[8:], uint64(a.args[1]))
				s.AddDeployedContract(addr, common.BytesToAddress(contract))
			},
			args: make([]int64, 2),
		},
		{
			name: "SetCreateTime",
			fn: func(a testAction, s *StateDB) {
				s.SetCreateTime(addr, uint64(a.args[0]))
			},
			args: make([]int64, 1),
		},
		{
			name: "CreateAccount",
			fn: func(a testAction, s *StateDB) {
				s.CreateAccount(addr)
			},
		},
		{
			name: "Suicide",
			fn: func(a testAction, s *StateDB) {
				s.Delete(addr)
			},
		},
		{
			name: "AddLog",
			fn: func(a testAction, s *StateDB) {
				data := make([]byte, 16)
				binary.BigEndian.PutUint64(data, uint64(a.args[0]))
				binary.BigEndian.PutUint64(data[8:], uint64(a.args[1]))
				s.AddLog(&types.Log{Address: addr, Data: data})
			},
			args: make([]int64, 2),
		},
	}
	action := actions[r.Intn(len(actions))]
	var nameargs []string
	if !action.noAddr {
		nameargs = append(nameargs, addr.Hex())
	}
	for _, i := range action.args {
		action.args[i] = rand.Int63n(10000)
		nameargs = append(nameargs, fmt.Sprint(action.args[i]))
	}
	action.name += "\t"
	action.name += strings.Join(nameargs, ", ")
	return action
}

// Generate returns a new snapshot test of the given size. All randomness is
// derived from r.
func (*snapshotTest) Generate(r *rand.Rand, size int) reflect.Value {
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
	// Generate snapshot indexes.
	nsnapshots := int(math.Sqrt(float64(size)))
	if size > 0 && nsnapshots == 0 {
		nsnapshots = 1
	}
	snapshots := make([]int, nsnapshots)
	snaplen := len(actions) / nsnapshots
	for i := range snapshots {
		// Try to place the snapshots some number of actions apart from each other.
		snapshots[i] = (i * snaplen) + r.Intn(snaplen)
	}
	return reflect.ValueOf(&snapshotTest{addrs, actions, snapshots, nil})
}

func (test *snapshotTest) String() string {
	out := new(bytes.Buffer)
	sindex := 0
	for i, action := range test.actions {
		if len(test.snapshots) > sindex && i == test.snapshots[sindex] {
			fmt.Fprintf(out, "---- snapshot %d ----\n", sindex)
			sindex++
		}
		fmt.Fprintf(out, "%4d: %s\n", i, action.name)
	}
	return out.String()
}

func (test *snapshotTest) run() bool {
	// Run all actions and create snapshots.
	var (
		db, _        = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
		state, _     = New(common.Hash{}, db, db, tutil.InitConfig(configPath), 10, common.DEFAULT_NAMESPACE)
		snapshotRevs = make([]int, len(test.snapshots))
		sindex       = 0
	)
	// Test snapshot
	for i, action := range test.actions {
		if len(test.snapshots) > sindex && i == test.snapshots[sindex] {
			snapshotRevs[sindex] = state.Snapshot().(int)
			sindex++
		}
		action.fn(action, state)
	}

	// Revert all snapshots in reverse order. Each revert must yield a state
	// that is equivalent to fresh state with all actions up the snapshot applied.
	for sindex--; sindex >= 0; sindex-- {
		checkstate, _ := New(common.Hash{}, db, db, tutil.InitConfig(configPath), 10, common.DEFAULT_NAMESPACE)
		for _, action := range test.actions[:test.snapshots[sindex]] {
			action.fn(action, checkstate)
		}
		state.RevertToSnapshot(snapshotRevs[sindex])
		if err := test.checkEqual(state, checkstate); err != nil {
			test.err = fmt.Errorf("state mismatch after revert to snapshot %d\n%v", sindex, err)
			return false
		}
	}
	return true
}

// checkEqual checks that methods of state and checkstate return the same values.
func (test *snapshotTest) checkEqual(state, checkstate *StateDB) error {
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

func RandomString(length int) string {
	var letters = []byte("abcdef0123456789")
	b := make([]byte, length)
	for i := 0; i < length; i += 1 {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
