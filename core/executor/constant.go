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
	NOTIFY_REMOVE_CACHE = iota
	NOTIFY_VALIDATION_RES
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
	WsShardLen            = 1024 * 1024 // 1MB
	MaxPendingSnapshotReq = 1024
)

// Error message definitions
const (
	InvalidSnapshotReqErr = "invalid snapshot request"
	InvalidDeletionReqErr = "invalid snapshot deletion request"
	MakeSnapshotFailedErr = "make snapshot failed"
	SnapshotNotExistErr   = "snapshot doesn't exist"
	DeleteSnapshotErr     = "delete snapshot failed"
	ArchiveFailedErr      = "archive failed"
	EmptyMessage          = ""
)

// Error definitions
var (
	SnapshotContentInvalidErr     = errors.New("snapshot content invalid")
	ArchiveRequestNotSatisfiedErr = errors.New("archive request not satisfied with requirement")
	SnapshotDoesntExistErr        = errors.New("snapshot does not exist")
	ApplyWsErr                    = errors.New("apply world state failed")
)
