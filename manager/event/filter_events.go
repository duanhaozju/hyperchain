package event

import (
	"hyperchain/core/types"
	"time"
)

type FilterNewBlockEvent struct {
	Block *types.Block
}

type FilterNewLogEvent struct {
	Logs []*types.Log
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

type FilterArchive struct {
	FilterId string
	Success  bool
	Message  string
}

const (
	// definition format: <ExceptionModule> + <Module>
	ExceptionModule_P2P      = "p2p"
	ExceptionModule_Consenus = "consensus"
	ExceptionModule_Executor = "executor"
	// etc ...
)

const (
	// definition format: <ExceptionCode> + <Module> + <SubType>
	ExceptionCode_Executor_Viewchange int =  -1 * iota
	// etc ...
)

type FilterException struct {
	Module    string	`json:"module"`
	SubType   string	`json:"subType"`
	ErrorCode int		`json:"errorCode"`
	Message   string	`json:"message"`
	Date      time.Time	`json:"date"`
}
