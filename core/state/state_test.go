//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"bytes"
	"fmt"
	checker "gopkg.in/check.v1"
	"math/big"
	"testing"

	"hyperchain/common"
	"hyperchain/hyperdb"
)

type StateSuite struct {
	state *StateDB
}

func Test(t *testing.T) { checker.TestingT(t) }

var _ = checker.Suite(&StateSuite{})

var toAddr = common.BytesToAddress

func (s *StateSuite) TestGetAccounts(c *checker.C) {
	// generate a few entries
	obj1 := s.state.GetOrNewStateObject(toAddr([]byte{0x01}))
	obj1.AddBalance(big.NewInt(22))
	obj2 := s.state.GetOrNewStateObject(toAddr([]byte{0x01, 0x02}))
	obj2.SetCode([]byte{3, 3, 3, 3, 3, 3, 3})
	obj3 := s.state.GetOrNewStateObject(toAddr([]byte{0x02}))
	obj3.SetBalance(big.NewInt(44))

	// write some of them to the trie
	s.state.UpdateStateObject(obj1)
	s.state.UpdateStateObject(obj2)
	s.state.Commit()
}
func (s *StateSuite) TestDump(c *checker.C) {
	// generate a few entries
	obj1 := s.state.GetOrNewStateObject(toAddr([]byte{0x01}))
	obj1.AddBalance(big.NewInt(22))
	obj2 := s.state.GetOrNewStateObject(toAddr([]byte{0x01, 0x02}))
	obj2.SetCode([]byte{3, 3, 3, 3, 3, 3, 3})
	obj3 := s.state.GetOrNewStateObject(toAddr([]byte{0x02}))
	obj3.SetBalance(big.NewInt(44))

	// write some of them to the trie
	s.state.UpdateStateObject(obj1)
	s.state.UpdateStateObject(obj2)
	s.state.Commit()

	// check that dump contains the state objects that are in trie
	got := string(s.state.Dump())
	/*
		want := `{
			    "root": "15d37a330fa35eddd0cb4928d78abab3f5376e7e30f2e977ddbbb9846be0be40",
			    "accounts": {
			        "0000000000000000000000000000000000000001": {
			            "balance": "22",
			            "nonce": 0,
			            "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			            "codeHash": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
			            "code": "",
						"abi": "",
			            "storage": {}
			        },
			        "0000000000000000000000000000000000000002": {
			            "balance": "44",
			            "nonce": 0,
			            "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			            "codeHash": "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
			            "code": "",
						"abi": "",
			            "storage": {}
			        },
			        "0000000000000000000000000000000000000102": {
			            "balance": "0",
			            "nonce": 0,
			            "root": "56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
			            "codeHash": "87874902497a5bb968da31a2998d8f22e949d1ef6214bcdedd8bae24cca4b9e3",
			            "code": "03030303030303",
						"abi": "",
			            "storage": {}
			        }
			    }
			}`
	*/
	c.Log(got)
}
func (s *StateSuite) SetUpTest(c *checker.C) {
	db, _ := hyperdb.NewMemDatabase()
	s.state, _ = New(common.Hash{}, db)
}

func TestNull(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	state, _ := New(common.Hash{}, db)

	address := common.HexToAddress("0x823140710bf13990e4500136726d8b55")
	state.CreateAccount(address)
	//value := common.FromHex("0x823140710bf13990e4500136726d8b55")
	var value common.Hash
	state.SetState(address, common.Hash{}, value)
	state.Commit()
	value = state.GetState(address, common.Hash{})
	if !common.EmptyHash(value) {
		t.Errorf("expected empty hash. got %x", value)
	}
}

func (s *StateSuite) TestSnapshot(c *checker.C) {
	stateobjaddr := toAddr([]byte("aa"))
	var storageaddr common.Hash
	data1 := common.BytesToHash([]byte{42})
	data2 := common.BytesToHash([]byte{43})

	// set initial state object value
	s.state.SetState(stateobjaddr, storageaddr, data1)
	// get snapshot of current state
	snapshot := s.state.Copy()

	// set new state object value
	s.state.SetState(stateobjaddr, storageaddr, data2)
	// restore snapshot
	s.state.Set(snapshot)

	// get state storage value
	res := s.state.GetState(stateobjaddr, storageaddr)

	c.Assert(data1, checker.DeepEquals, res)
}
func (s *StateSuite) TestReplication(c *checker.C) {
	db, _ := hyperdb.NewMemDatabase()
	state1, _ := New(common.Hash{}, db)

	obj1 := state1.GetOrNewStateObject(toAddr([]byte{0x01}))
	obj1.AddBalance(big.NewInt(22))
	obj2 := state1.GetOrNewStateObject(toAddr([]byte{0x01, 0x02}))
	obj2.SetCode([]byte{3, 3, 3, 3, 3, 3, 3})
	obj3 := state1.GetOrNewStateObject(toAddr([]byte{0x02}))
	obj3.SetBalance(big.NewInt(44))
	obj4 := state1.GetOrNewStateObject(toAddr([]byte{0x03}))
	obj4.SetABI([]byte{1, 2, 3, 4, 5, 6, 7})
	// write some of them to the trie
	state1.UpdateStateObject(obj1)
	state1.UpdateStateObject(obj2)
	state1.UpdateStateObject(obj3)
	state1.UpdateStateObject(obj4)
	root, _ := state1.Commit()
	fmt.Println(root.Hex())

	state2, _ := New(root, db)

	it := state1.trie.Iterator()
	for it.Next() {
		key := state1.trie.GetKey(it.Key)
		val2 := state2.trie.Get(key)
		c.Assert(it.Value, checker.DeepEquals, val2)
	}
	c.Assert(state1.Dump(), checker.DeepEquals, state2.Dump())

	state1.Delete(obj1.address)
}

// use testing instead of checker because checker does not support
// printing/logging in tests (-check.vv does not work)
func TestSnapshot2(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	state, _ := New(common.Hash{}, db)

	stateobjaddr0 := toAddr([]byte("so0"))
	stateobjaddr1 := toAddr([]byte("so1"))
	var storageaddr common.Hash

	data0 := common.BytesToHash([]byte{17})
	data1 := common.BytesToHash([]byte{18})

	state.SetState(stateobjaddr0, storageaddr, data0)
	state.SetState(stateobjaddr1, storageaddr, data1)

	// db, trie are already non-empty values
	so0 := state.GetStateObject(stateobjaddr0)
	so0.BalanceData = big.NewInt(42)
	so0.nonce = 43
	so0.SetCode([]byte{'c', 'a', 'f', 'e'})
	so0.remove = true
	so0.deleted = false
	so0.dirty = false
	state.SetStateObject(so0)

	// and one with deleted == true
	so1 := state.GetStateObject(stateobjaddr1)
	so1.BalanceData = big.NewInt(52)
	so1.nonce = 53
	so1.SetCode([]byte{'c', 'a', 'f', 'e', '2'})
	so1.remove = true
	so1.deleted = true
	so1.dirty = true
	state.SetStateObject(so1)

	so1 = state.GetStateObject(stateobjaddr1)
	if so1 != nil {
		t.Fatalf("deleted object not nil when getting")
	}

	snapshot := state.Copy()
	state.Set(snapshot)

	so0Restored := state.GetStateObject(stateobjaddr0)
	so1Restored := state.GetStateObject(stateobjaddr1)
	// non-deleted is equal (restored)
	compareStateObjects(so0Restored, so0, t)
	// deleted should be nil, both before and after restore of state copy
	if so1Restored != nil {
		t.Fatalf("deleted object not nil after restoring snapshot")
	}
}

func compareStateObjects(so0, so1 *StateObject, t *testing.T) {
	if so0.address != so1.address {
		t.Fatalf("Address mismatch: have %v, want %v", so0.address, so1.address)
	}
	if so0.BalanceData.Cmp(so1.BalanceData) != 0 {
		t.Fatalf("Balance mismatch: have %v, want %v", so0.BalanceData, so1.BalanceData)
	}
	if so0.nonce != so1.nonce {
		t.Fatalf("Nonce mismatch: have %v, want %v", so0.nonce, so1.nonce)
	}
	if !bytes.Equal(so0.codeHash, so1.codeHash) {
		t.Fatalf("CodeHash mismatch: have %v, want %v", so0.codeHash, so1.codeHash)
	}
	if !bytes.Equal(so0.code, so1.code) {
		t.Fatalf("Code mismatch: have %v, want %v", so0.code, so1.code)
	}
	if !bytes.Equal(so0.abi, so1.abi) {
		t.Fatalf("InitCode mismatch: have %v, want %v", so0.abi, so1.abi)
	}

	for k, v := range so1.storage {
		if so0.storage[k] != v {
			t.Fatalf("Storage key %s mismatch: have %v, want %v", k, so0.storage[k], v)
		}
	}
	for k, v := range so0.storage {
		if so1.storage[k] != v {
			t.Fatalf("Storage key %s mismatch: have %v, want none.", k, v)
		}
	}

	if so0.remove != so1.remove {
		t.Fatalf("Remove mismatch: have %v, want %v", so0.remove, so1.remove)
	}
	if so0.deleted != so1.deleted {
		t.Fatalf("Deleted mismatch: have %v, want %v", so0.deleted, so1.deleted)
	}
	if so0.dirty != so1.dirty {
		t.Fatalf("Dirty mismatch: have %v, want %v", so0.dirty, so1.dirty)
	}
}
