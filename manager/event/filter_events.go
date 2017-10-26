package event

import (
	"github.com/hyperchain/hyperchain/core/types"
	"time"
)

type FilterNewBlockEvent struct {
	Block *types.Block
}

type FilterNewLogEvent struct {
	Logs []*types.Log
}

type FilterSystemStatusEvent struct {
	Module    string    `json:"module"`
	Status    bool      `json:"status"`
	Subtype   string    `json:"subType"`
	ErrorCode int       `json:"errorCode"`
	Message   string    `json:"message"`
	Date      time.Time `json:"date"`
}

/*
	Archive
*/

const (
	FilterMakeSnapshot   = "make_snapshot"
	FilterDeleteSnapshot = "delete_snapshot"
	FilterDoArchive      = "do_archive"
)

type FilterArchive struct {
	Type     string `json:"type"`
	FilterId string `json:"filterId"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}
