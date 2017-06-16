package hyperstate

import (
	"testing"
	"hyperchain/hyperdb/mdb"
	"hyperchain/common"
	"fmt"
)

func TestStorageIterator(t *testing.T) {
	db, _ := mdb.NewMemDatabase()
	stateDb, _ := New(common.Hash{}, db, db, nil, 0, "test")
	stateDb.MarkProcessStart(1)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key1")), []byte("value1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key2")), []byte("value2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key3")), []byte("value3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key4")), []byte("value4"), 0)
	stateDb.Commit()
	batch := stateDb.FetchBatch(1)
	batch.Write()
	stateDb.MarkProcessFinish(1)

	stateDb.MarkProcessStart(2)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key1")), []byte("newvalue1"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key2")), []byte("newvalue2"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key3")), []byte("newvalue3"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key5")), []byte("value5"), 0)
	stateDb.SetState(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key6")), []byte("value6"), 0)

	iter, _ := stateDb.NewIterator(common.BytesToAddress([]byte("address001")), common.BytesToHash([]byte("key0")), common.BytesToHash([]byte("key10")))
	for iter.Next() {
		fmt.Println(common.Bytes2Hex(iter.Key()))
		fmt.Println(common.Bytes2Hex(iter.Value()))
	}

}
