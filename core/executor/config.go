package executor

import (
	"hyperchain/common"
	"time"
)

const (
	SyncReplicaInterval     = "executor.sync_replica.interval"
	SyncReplicaEnable       = "executor.sync_replica.enable"
	SyncChainBatchSize      = "executor.syncer.max_block_fetch"
	SyncChainResendInterval = "executor.syncer.block_fetch_timeout"
	SyncWsEable             = "executor.syncer.state_fetch_enable"
	SnapshotManifestPath    = "executor.archive.snapshot_manifest"
	ArchiveMetaPath         = "executor.archive.archive_manifest"
	ArchiveForceConsistency = "executor.archive.force_consistency"
	ArchiveThreshold        = "executor.archive.threshold"
	exitFlag                = "executor.nvp.exitflag"
)

// GetSyncReplicaInterval - sync replica information interval.
func (executor *Executor) GetSyncReplicaInterval() time.Duration {
	return executor.conf.GetDuration(SyncReplicaInterval)
}

// GetSyncReplicaEnable - sync replica switch value.
func (exector *Executor) GetSyncReplicaEnable() bool {
	return exector.conf.GetBool(SyncReplicaEnable)
}

func (executor *Executor) GetSyncMaxBatchSize() uint64 {
	return uint64(executor.conf.GetInt64(SyncChainBatchSize))
}

func (executor *Executor) GetSyncResendInterval() time.Duration {
	return executor.conf.GetDuration(SyncChainResendInterval)
}

func (executor *Executor) GetManifestPath() string {
	return executor.conf.GetString(SnapshotManifestPath)
}

func (executor *Executor) GetArchiveMetaPath() string {
	return executor.conf.GetString(ArchiveMetaPath)
}

func (executor *Executor) IsArchiveForceConsistency() bool {
	return executor.conf.GetBool(ArchiveForceConsistency)
}

func (executor *Executor) GetArchiveThreshold() int {
	return executor.conf.GetInt(ArchiveThreshold)
}

func (executor *Executor) IsSyncWsEable() bool {
	return executor.conf.GetBool(SyncWsEable)
}

func (executor *Executor) isJvmEnable() bool {
	return executor.conf.GetBool(common.C_JVM_START)
}


func (executor *Executor) GetExitFlag() bool {
	return executor.conf.GetBool(exitFlag)
}