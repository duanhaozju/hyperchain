package kvlog

import (
	"fmt"

	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/hyperchain/hyperchain/core/oplog/proto"

	"github.com/golang/protobuf/proto"
)


type kvLoggerImpl struct {
	lastSet	uint64
	db		db.Database
}

func New(db db.Database) *kvLoggerImpl {
	logger := &kvLoggerImpl{
		db: db,
	}
	logger.lastSet = uint64(0)
	return logger
}

func (logger *kvLoggerImpl) Append(entry *oplog.LogEntry) error {

	if entry.Lid != logger.lastSet + 1 {
		return fmt.Errorf("not matched with lastSet")
	}

	raw, err := proto.Marshal(entry)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("entry.%d", logger.lastSet + 1)
	if err = logger.db.Put([]byte(key), raw); err == nil {
		logger.lastSet += 1
	}
	return err
}

func (logger *kvLoggerImpl) Fetch(lid uint64) (*oplog.LogEntry, error) {

	if lid > logger.lastSet {
		return nil, fmt.Errorf("lid is too large")
	}

	key := fmt.Sprintf("entry.%d", lid)
	raw, err := logger.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	entry := &oplog.LogEntry{}
	if err = proto.Unmarshal(raw, entry); err != nil{
		return nil, err
	} else {
		return entry, nil
	}
}

func (logger *kvLoggerImpl) Reset(lid uint64) error {

	if lid > logger.lastSet {
		return fmt.Errorf("lid is too large")
	}

	logger.lastSet = lid
	return nil
}