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

import (
	"time"

	"hyperchain/common"
)

const (
	// Replica status synchronization interval
	SyncReplicaInterval = "executor.sync_replica.interval"
	// Whether enable replica status synchronization or not
	SyncReplicaEnable = "executor.sync_replica.enable"
	// The maximum number of blocks a single block fetch request can get in synchronization
	SyncChainBatchSize = "executor.syncer.max_block_fetch"
	// The timeout of a single block fetch request
	SyncChainResendInterval = "executor.syncer.block_fetch_timeout"
	// Whether enable world state transition or not
	SyncWsEable = "executor.syncer.state_fetch_enable"
	// The maximum package size of a single world state fetch request
	SyncWsPacketSize = "executor.syncer.state_fetch_packet_size"
	// Snapshot meta file path
	SnapshotManifestPath = "executor.archive.snapshot_manifest"
	// Archive meta file path
	ArchiveMetaPath = "executor.archive.archive_manifest"
	// Whether the block data should keep consistent in historical database
	ArchiveForceConsistency = "executor.archive.force_consistency"
	// Block chain size threshold. The block chain size should not be less than
	// the threshold, either before and after archiving
	ArchiveThreshold = "executor.archive.threshold"
	// The non-verifying node exits when the block data is inconsistent
	exitFlag = "executor.nvp.exitflag"
)

// Config conf configuration reader.
type Config struct {
	conf      *common.Config
	namespace string
}

func (conf *Config) GetSyncReplicaInterval() time.Duration {
	return conf.conf.GetDuration(SyncReplicaInterval)
}

func (conf *Config) GetSyncReplicaEnable() bool {
	return conf.conf.GetBool(SyncReplicaEnable)
}

func (conf *Config) GetSyncMaxBatchSize() uint64 {
	return uint64(conf.conf.GetInt64(SyncChainBatchSize))
}

func (conf *Config) GetSyncResendInterval() time.Duration {
	return conf.conf.GetDuration(SyncChainResendInterval)
}

func (conf *Config) GetManifestPath() string {
	return common.GetPath(conf.namespace, conf.conf.GetString(SnapshotManifestPath))
}

func (conf *Config) GetArchiveMetaPath() string {
	return common.GetPath(conf.namespace, conf.conf.GetString(ArchiveMetaPath))
}

func (conf *Config) IsArchiveForceConsistency() bool {
	return conf.conf.GetBool(ArchiveForceConsistency)
}

func (conf *Config) GetArchiveThreshold() int {
	return conf.conf.GetInt(ArchiveThreshold)
}

func (conf *Config) IsSyncWsEable() bool {
	return conf.conf.GetBool(SyncWsEable)
}

func (conf *Config) isJvmEnable() bool {
	return conf.conf.GetBool(common.C_JVM_START)
}

func (conf *Config) GetStateFetchPacketSize() int {
	return conf.conf.GetInt(SyncWsPacketSize) * 1024
}

func (conf *Config) GetExitFlag() bool {
	return conf.conf.GetBool(exitFlag)
}
