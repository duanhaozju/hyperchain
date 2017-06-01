package executor

import "errors"

const (
	COMMITQUEUESIZE = 1024
	VALIDATEQUEUESIZE = 1024

	VALIDATION_NORMAL = 0
	VALIDATION_IGNORE = 1

	BUSY = 1
	IDLE = 0
)

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
	NOTIFY_SEND_WORLD_STATE
	NOTIFY_SYNC_REPLICA
)


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

const LatestBlockNumber uint64 = 0

const InvalidSnapshotReqErr   = "invalid snapshot request"
const InvalidDeletionReqErr   = "invalid snapshot deletion request"
const MakeSnapshotFailedErr   = "make snapshot failed"
const SnapshotNotExistErr     = "snapshot doesn't exist"
const DeleteSnapshotErr       = "delete snapshot failed"

const ArchiveFailedErr        = "archive failed"

const EmptyMessage = ""
var   SnapshotContentInvalidErr = errors.New("snapshot content invalid")
var   ArchiveRequestNotSatisfiedErr = errors.New("archive request not satisfied with requirement")
var   SnapshotDoesntExistErr = errors.New("snapshot does not exist")
