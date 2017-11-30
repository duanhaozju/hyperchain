package oplog

import "github.com/hyperchain/hyperchain/core/oplog/proto"

//Provide operation log manipulations

type Log interface {
	//Append append a log entry.
	Append(entry *oplog.LogEntry) (uint64, error)
	//Fetch fetch log entry by log id.
	Fetch(lid uint64) (*oplog.LogEntry, error)
}

type OpLog interface {
	Log
	Reset(lid uint64) error //reset committed log id to target lid.
	Iterator() Iterator
	GetLastSet() uint64      //GetLastSet last committed log entry index.
	GetLastBlockNum() uint64 //GetLastBlockNum last stored block number.
	SetStableCheckpoint(checkpoint uint64)
	GetHeightAndDigest() (uint64, string, error)
	GetLidBySeqNo(seqNo uint64) (uint64, error)
}

type Iterator interface {
	Key() []byte
	Value() []byte
	Seek(lid uint64) bool
	Next() bool
	Prev() bool
	Error() error
	Release()
}
