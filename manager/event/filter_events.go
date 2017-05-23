package event

import (
	"hyperchain/core/types"
	"hyperchain/core/vm"
)

type FilterNewBlockEvent struct {
	Block *types.Block
}

type FilterNewLogEvent struct {
	Logs    []*vm.Log
}


/*
	Archive
 */

type FilterSnapshotEvent struct {
	FilterId string
	Success  bool
	Message  string
}

type FilterDeleteSnapshotEvent struct {
	FilterId string
	Success  bool
	Message  string
}
