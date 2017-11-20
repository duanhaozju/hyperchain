package oplog

import "github.com/hyperchain/hyperchain/core/oplog/proto"

//Provide operation log manipulations

type Log interface {
	//Append append a log entry.
	Append(entry *oplog.LogEntry) error
	//Fetch fetch log entry by log id.
	Fetch(lid uint64) (*oplog.LogEntry, error)
}

type OpLog interface {
	Log
	Reset(lid uint64) error //reset committed log id to target lid.
}
