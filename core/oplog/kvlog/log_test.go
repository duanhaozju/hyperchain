package kvlog

import (
	"testing"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/oplog/proto"
	"github.com/hyperchain/hyperchain/manager/event"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func getConfig() *common.Config {
	config := common.NewRawConfig()
	config.Set("namespace.name", "global")
	config.Set("consensus.rbft.k", int64(40))
	return config
}

func TestLogAppendAndFetch(t *testing.T) {
	ast := assert.New(t)
	kvLogger := New(getConfig())

	event1 := &event.TransactionBlock{SeqNo: uint64(1)}
	payload1, err := proto.Marshal(event1)
	if err != nil {
		t.Errorf("TransactionBatch Marshal Error", err)
		return
	}
	entry1 := &oplog.LogEntry{
		Type:    oplog.LogEntry_TransactionList,
		Payload: payload1,
	}
	err = kvLogger.Append(entry1)
	ast.Nil(err, "The appropriate lid, should not return any error")

	entry2, err := kvLogger.Fetch(uint64(1))
	ast.Equal(entry1, entry2, "should be equal")

	event2 := &event.TransactionBlock{SeqNo: uint64(2)}
	payload2, err := proto.Marshal(event2)
	if err != nil {
		t.Errorf("TransactionBatch Marshal Error", err)
		return
	}
	entry2 = &oplog.LogEntry{
		Type:    oplog.LogEntry_TransactionList,
		Payload: payload2,
	}
	err = kvLogger.Append(entry2)
	ast.Nil(err, "The appropriate lid, should not return any error")
	ast.Equal(kvLogger.lastSet, uint64(2))
	ast.Equal(kvLogger.lastCommit, uint64(2))

	entry3 := &oplog.LogEntry{
		Type:    oplog.LogEntry_RollBack,
		Payload: []byte("3"),
	}
	err = kvLogger.Append(entry3)
	ast.Nil(err, "The appropriate lid, should not return any error")
	ast.Equal(kvLogger.lastSet, uint64(3))
	ast.Equal(kvLogger.lastCommit, uint64(2))

	lid, entry, err := kvLogger.getBySeqNo(uint64(2))
	ast.Nil(err)
	ast.Equal(uint64(2), lid)
	ast.Equal(entry2, entry)

	event4 := &event.TransactionBlock{SeqNo: uint64(3)}
	payload4, err := proto.Marshal(event4)
	if err != nil {
		t.Errorf("TransactionBatch Marshal Error", err)
		return
	}
	entry4 := &oplog.LogEntry{
		Type:    oplog.LogEntry_TransactionList,
		Payload: payload4,
	}
	err = kvLogger.Append(entry4)
	ast.Nil(err, "The appropriate lid, should not return any error")
	ast.Equal(kvLogger.lastSet, uint64(4))
	ast.Equal(kvLogger.lastCommit, uint64(3))

	lid, entry, err = kvLogger.getBySeqNo(uint64(3))
	ast.Nil(err)
	ast.Equal(uint64(4), lid)
	ast.Equal(entry4, entry)
}

//func TestLastSet(t *testing.T) {
//	ast := assert.New(t)
//	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
//	ast.Equal(nil, err, err)
//	logger := New(db)
//
//	logger.lastSet = 100
//	logger.storeLastSet()
//
//	logger.lastSet = 9
//	logger.restoreLastSet()
//
//	ast.Equal(uint64(100), logger.lastSet)
//}
