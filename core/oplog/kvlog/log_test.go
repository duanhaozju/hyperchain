package kvlog

import (

	"testing"

	"github.com/hyperchain/hyperchain/common"
	mdb "github.com/hyperchain/hyperchain/hyperdb/mdb"
	"github.com/hyperchain/hyperchain/core/oplog/proto"

	"github.com/stretchr/testify/assert"
)

func TestLogAppendAndFetch(t *testing.T) {
	ast := assert.New(t)
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	ast.Equal(nil, err, err)
	logger := New(db)

	entry100 := &oplog.LogEntry{
		Lid:	uint64(100),
	}
	err = logger.Append(entry100)
	ast.NotNil(err, "lid is too large, should return an error")

	entry1 := &oplog.LogEntry{
		Lid:	uint64(1),
		Type:	oplog.LogEntry_TransactionList,
		Payload: []byte("1"),
	}
	err = logger.Append(entry1)
	ast.Nil(err, "The appropriate lid, should not return any error")

	entry2 , err := logger.Fetch(uint64(1))
	ast.Equal(entry1, entry2, "should be equal")

	entry2 = &oplog.LogEntry{
		Lid:	uint64(2),
		Type:	oplog.LogEntry_TransactionList,
		Payload: []byte("2"),
	}
	err = logger.Append(entry2)
	ast.Nil(err, "The appropriate lid, should not return any error")

	err = logger.Reset(100)
	ast.NotNil(err, "lid is too large, should return an error")
	err = logger.Reset(1)
	ast.Nil(err, "The appropriate lid, should not return any error")
}

func TestLastSet(t *testing.T) {
	ast := assert.New(t)
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	ast.Equal(nil, err, err)
	logger := New(db)

	logger.lastSet = 100
	logger.storeLastSet()

	logger.lastSet = 9
	logger.restoreLastSet()

	ast.Equal(uint64(100), logger.lastSet)
}