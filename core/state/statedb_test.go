package state

import (
	"testing"
	"hyperchain/hyperdb"
)

func TestCommit(t *testing.T){
	db,_ := hyperdb.GetLDBDatabase()
	statedb,_  := New(db)
	t.Log("test commit")
	batch := statedb.db.NewBatch()

	batch.Write()
}
