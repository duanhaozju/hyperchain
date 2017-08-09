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

const (
	FilterMakeSnapshot = "make_snapshot"
	FilterDeleteSnapshot = "delete_snapshot"
	FilterDoArchive = "do_archive"
)

type FilterArchive struct {
	Type     string      `json:"type"`
	FilterId string      `json:"filterId"`
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
}


