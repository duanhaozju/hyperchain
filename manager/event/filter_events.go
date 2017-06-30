package event

import (
	"hyperchain/core/types"
	"time"
	"hyperchain/manager/exception"
)

type FilterNewBlockEvent struct {
	Block *types.Block
}

type FilterNewLogEvent struct {
	Logs []*types.Log
}

type FilterExceptionEvent struct {
	Module    string
	Exception exception.ExceptionError
}

type FilterExceptionData struct {
	Module    string	`json:"module"`
	SubType   string	`json:"subType"`
	ErrorCode int		`json:"errorCode"`
	Message   string	`json:"message"`
	Date      time.Time	`json:"date"`
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

