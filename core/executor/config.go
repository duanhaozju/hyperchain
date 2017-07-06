package executor

import (
	"time"
	"hyperchain/common"
)

const (
	SyncReplicaInterval         = "global.executor.sync_replica.interval"
	SyncReplicaEnable           = "global.executor.sync_replica.enable"
	SyncChainBatchSize          = "global.executor.sync_chain.sync_batch_size"
	SyncChainResendInterval     = "global.executor.sync_chain.sync_resend_interval"
	SyncWsEable                 = "global.executor.sync_chain.sync_ws_enable"
	SnapshotManifestPath        = "global.executor.archive.manifest"
	ArchiveMetaPath             = "global.executor.archive.archiveMeta"
	ArchiveForceConsistency     = "global.executor.archive.force_consistency"
	ArchiveThreshold            = "global.executor.archive.threshold"
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