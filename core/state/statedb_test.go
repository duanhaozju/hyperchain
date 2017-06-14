//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"bytes"
	"fmt"
	"hyperchain/common"
	"hyperchain/core/evm"
	"hyperchain/hyperdb"
	"math/big"
	"reflect"
	"testing"
)

func TestStateDB_GetAccounts(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	state, _ := New(common.Hash{}, db)
	obj1 := state.GetOrNewStateObject(toAddr([]byte{0x01}))
	obj1.AddBalance(big.NewInt(22))
	obj2 := state.GetOrNewStateObject(toAddr([]byte{0x01, 0x02}))
	obj2.SetCode([]byte{3, 3, 3, 3, 3, 3, 3})
	obj3 := state.GetOrNewStateObject(toAddr([]byte{0x02}))
	obj3.SetBalance(big.NewInt(44))

	// write some of them to the trie
	state.UpdateStateObject(obj1)
	state.UpdateStateObject(obj2)
	state.Commit()

	res := state.GetAccounts()
	if len(res) != 3 {
		t.Fatal("getaccounts error,num not match")
	}
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

func TestStateDB_Copy(t *testing.T) {
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

func TestReplication(t *testing.T) {
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
		if !reflect.DeepEqual(it.Value, val2) {
			t.Errorf("mismatch: have %v, want %v", it.Value, val2)
		}
	}

	if !reflect.DeepEqual(state1.Dump(), state2.Dump()) {
		t.Errorf("dump mismatch: have %v, want %v", state1.Dump(), state2.Dump())
	}
	//c.Assert(state1.Dump(), checker.DeepEquals, state2.Dump())

	state1.Delete(obj1.address)
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

func TestLog(t *testing.T) {
	thash := common.BytesToHash([]byte{1, 2, 3})
	bhash := common.BytesToHash([]byte{4, 5, 6})
	ti := 1
	state := setup()
	state.StartRecord(thash, bhash, ti)
	var log *evm.Log
	var topic []common.Hash
	topic = append(topic, common.Hash{})
	topic = append(topic, common.Hash{})
	log = evm.NewLog(toAddr([]byte{0x01}), topic, []byte{1}, uint64(1))
	state.AddLog(log)
	if len(state.Logs()) != 1 {
		t.Error("state addlog error")
	}

	getlog := state.GetLogs(thash)[0]
	if getlog.BlockHash != bhash {
		t.Error("getlogs error")
	}
	if getlog.TxHash != thash {
		t.Error("getlogs error")
	}
	if getlog.TxIndex != uint(ti) {
		t.Error("getlogs error")
	}
}
func TestStateDB_AddRefund(t *testing.T) {
	state := setup()
	state.AddRefund(common.Big("5"))
	if state.GetRefund().Cmp(common.Big("5")) != 0 {
		t.Error("state addrefund error")
	}
}
func TestStateDB_SetGet(t *testing.T) {
	state := setup()
	addr := toAddr([]byte{0x01})
	code := []byte{2, 2, 2, 2, 2}
	abi := []byte{3, 3, 3, 3, 3, 3, 3}
	state.SetABI(addr, abi)
	state.SetCode(addr, code)
	state.SetNonce(addr, uint64(8))

	if x := state.GetABI(addr); bytes.Compare(x, abi) != 0 {
		t.Error("state set get abi error")
	}
	if x := state.GetCode(addr); bytes.Compare(x, code) != 0 {
		t.Error("state set get code error")
	}
	if state.GetNonce(addr) != uint64(8) {
		t.Error("state set get nonce error")
	}
	balance0 := state.GetBalance(addr)
	state.AddBalance(addr, common.Big("5"))
	if x := state.GetBalance(addr); x.Cmp(balance0.Add(balance0, common.Big("5"))) != 0 {
		t.Error("state add&get banlance error")
	}
}

func TestStateDB_IntermediateRoot(t *testing.T) {
	state := setup()
	root := state.IntermediateRoot()
	if len(root) == 0 {
		t.Error("root should not be nil")
	}
}
