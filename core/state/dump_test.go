//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"hyperchain/common"
	"hyperchain/hyperdb"
	"math/big"
	"testing"
)

var toAddr = common.BytesToAddress

func setup() *StateDB {
	db, _ := hyperdb.NewMemDatabase()
	state, _ := New(common.Hash{}, db)
	// generate a few entries
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
	return state
}
func TestStateDB_Dump(t *testing.T) {

	state := setup()
	// check that dump contains the state objects that are in trie
	got := string(state.Dump())
	t.Log(got)
}
