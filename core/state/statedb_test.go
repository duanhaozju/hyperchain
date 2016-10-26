package state

import (
	"math/big"
	"testing"

	"hyperchain/common"
	"hyperchain/hyperdb"
)

// Tests that updating a state trie does not leak any database writes prior to
// actually committing the state.
func TestUpdateLeaks(t *testing.T) {
	// Create an empty state database
	db, _ := hyperdb.NewMemDatabase()
	state, _ := New(common.Hash{}, db)
	// Update it with some accounts
	for i := byte(0); i < 255; i++ {
		obj := state.GetOrNewStateObject(common.BytesToAddress([]byte{i}))
		obj.AddBalance(big.NewInt(int64(11 * i)))
		obj.SetNonce(uint64(42 * i))
		if i%2 == 0 {
			obj.SetState(common.BytesToHash([]byte{i, i, i}), common.BytesToHash([]byte{i, i, i, i}))
		}
		if i%3 == 0 {
			obj.SetCode([]byte{i, i, i, i, i})
		}
		state.UpdateStateObject(obj)
	}
	// Ensure that no data was leaked into the database
	for _, key := range db.Keys() {
		value, _ := db.Get(key)
		t.Errorf("State leaked into database: %x -> %x", key, value)
	}
}

// Tests that no intermediate state of an object is stored into the database,
// only the one right before the commit.
func TestIntermediateLeaks(t *testing.T) {
	// Create two state databases, one transitioning to the final state, the other final from the beginning
	transDb, _ := hyperdb.NewMemDatabase()
	finalDb, _ := hyperdb.NewMemDatabase()
	transState, _ := New(common.Hash{}, transDb)
	finalState, _ := New(common.Hash{}, finalDb)

	// Update the states with some objects
	for i := byte(0); i < 255; i++ {
		// Create a new state object with some data into the transition database
		obj := transState.GetOrNewStateObject(common.BytesToAddress([]byte{i}))
		obj.SetBalance(big.NewInt(int64(11 * i)))
		obj.SetNonce(uint64(42 * i))
		if i%2 == 0 {
			obj.SetState(common.BytesToHash([]byte{i, i, i, 0}), common.BytesToHash([]byte{i, i, i, i, 0}))
		}
		if i%3 == 0 {
			obj.SetCode([]byte{i, i, i, i, i, 0})
		}
		transState.UpdateStateObject(obj)

		// Overwrite all the data with new values in the transition database
		obj.SetBalance(big.NewInt(int64(11*i + 1)))
		obj.SetNonce(uint64(42*i + 1))
		if i%2 == 0 {
			obj.SetState(common.BytesToHash([]byte{i, i, i, 0}), common.Hash{})
			obj.SetState(common.BytesToHash([]byte{i, i, i, 1}), common.BytesToHash([]byte{i, i, i, i, 1}))
		}
		if i%3 == 0 {
			obj.SetCode([]byte{i, i, i, i, i, 1})
		}
		transState.UpdateStateObject(obj)

		// Create the final state object directly in the final database
		obj = finalState.GetOrNewStateObject(common.BytesToAddress([]byte{i}))
		obj.SetBalance(big.NewInt(int64(11*i + 1)))
		obj.SetNonce(uint64(42*i + 1))
		if i%2 == 0 {
			obj.SetState(common.BytesToHash([]byte{i, i, i, 1}), common.BytesToHash([]byte{i, i, i, i, 1}))
		}
		if i%3 == 0 {
			obj.SetCode([]byte{i, i, i, i, i, 1})
		}
		finalState.UpdateStateObject(obj)
	}
	transState.Commit()
	finalState.Commit()
	// Cross check the databases to ensure they are the same
	for _, key := range finalDb.Keys() {
		if _, err := transDb.Get(key); err != nil {
			val, _ := finalDb.Get(key)
			t.Errorf("entry missing from the transition database: %x -> %x", key, val)
		}
	}

	for _, key := range transDb.Keys() {
		if _, err := finalDb.Get(key); err != nil {
			val, _ := transDb.Get(key)
			t.Errorf("extra entry in the transition database: %x -> %x", key, val)
		}
	}
}
