package executor

import "errors"

const (
	COMMITQUEUESIZE   = 1024
	VALIDATEQUEUESIZE = 1024

	VALIDATION_NORMAL = 0
	VALIDATION_IGNORE = 1

	BUSY = 1
	IDLE = 0
)

// Hyperchain protocol message definitions
const (
	// consensus
	NOTIFY_VALIDATION_RES = iota
	NOTIFY_VC_DONE
	NOTIFY_SYNC_DONE
	// p2p
	NOTIFY_UNICAST_INVALID
	NOTIFY_BROADCAST_DEMAND
	NOTIFY_UNICAST_BLOCK
	NOTIFY_BROADCAST_SINGLE
	NOTIFY_REQUEST_WORLD_STATE
	NOTIFY_SEND_WORLD_STATE_HANDSHAKE
	NOTIFY_SEND_WS_ACK
	NOTIFY_SEND_WORLD_STATE
	NOTIFY_SYNC_REPLICA
	NOTIFY_TRANSIT_BLOCK
	NOTIFY_NVP_SYNC
)

// Subscription type definitions
const (
	FILTER_NEW_BLOCK = iota
	FILTER_NEW_LOG
	FILTER_SNAPSHOT_RESULT
	FILTER_DELETE_SNAPSHOT
	FILTER_ARCHIVE
)

const (
	IDENTIFIER_VALIDATION = iota
	IDENTIFIER_COMMIT
	IDENTIFIER_REPLICA_SYNC
)

const (
	LatestBlockNumber     uint64 = 0
	MaxPendingSnapshotReq        = 1024
)

// Error message definitions
const (
	InvalidSnapshotReqMsg = "invalid snapshot request"
	InvalidDeletionReqMsg = "invalid snapshot deletion request"
	InvalidArchiveReqMsg  = "archive request not satisfied with requirement"
	InvalidContentMsg     = "snapshot content invalid"
	MakeSnapshotFailedMsg = "make snapshot failed"
	SnapshotNotExistMsg   = "snapshot doesn't exist"
	DeleteSnapshotMsg     = "delete snapshot failed"
	ArchiveFailedMsg      = "archive failed"
	ApplyWsErrMsg         = "apply world state failed"
	EmptyMessage          = ""
)

// Error definitions
var (
	SnapshotContentInvalidErr     = errors.New(InvalidContentMsg)
	ArchiveRequestNotSatisfiedErr = errors.New(InvalidArchiveReqMsg)
	SnapshotDoesntExistErr        = errors.New(SnapshotNotExistMsg)
	InvalidSnapshotDeletionErr    = errors.New(InvalidDeletionReqMsg)
	DeleteSnapshotFailedErr       = errors.New(DeleteSnapshotMsg)
	ArchiveFailedErr              = errors.New(ArchiveFailedMsg)
	ApplyWsErr                    = errors.New(ApplyWsErrMsg)
)
