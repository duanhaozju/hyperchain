// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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

const (
	DemandSeqNo = iota
	DemandNumber
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
	NOTIFY_REQUEST_WORLD_STATE
	NOTIFY_SEND_WORLD_STATE_HANDSHAKE
	NOTIFY_SEND_WS_ACK
	NOTIFY_SEND_WORLD_STATE
	NOTIFY_SYNC_REPLICA
	NOTIFY_TRANSIT_BLOCK
	NOTIFY_NVP_SYNC
	NOTIFY_NVP_CONSULT
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
	VP = iota
	NVP
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
	ArchiveStoreFailedMsg = "archive store failed"
	ApplyWsErrMsg         = "apply world state failed"
	ParentGenesisMissMsg  = "miss parent genesis"
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
	NonAvailableParentGenesisErr  = errors.New(ParentGenesisMissMsg)
)
