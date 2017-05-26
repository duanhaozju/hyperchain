package executor

import "time"

const (
	STATEDB               = "state"
	stateBucketSize       = "global.executor.buckettree.state.size"
	stateBucketLevelGroup = "global.executor.buckettree.state.levelGroup"
	stateBucketCacheSize  = "global.executor.buckettree.state.cacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.executor.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.executor.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.executor.buckettree.storage.cacheSize"
	syncReplicaInterval         = "global.executor.sync_replica.interval"
	syncReplicaEnable           = "global.executor.sync_replica.enable"
	syncChainBatchSize          = "global.executor.sync_chain.sync_batch_size"
	syncChainResendInterval     = "global.executor.sync_chain.sync_resend_interval"
	snapshotManifestPath        = "global.executor.archive.manifest"
	archiveMetaPath             = "global.executor.archive.archiveMeta"
	archiveForceConsistency     = "global.executor.archive.force_consistency"
	archiveThreshold            = "global.executor.archive.threshold"
)

// GetBucketSize - get bucket size.
func (executor *Executor) GetBucketSize(choice string) int {
	switch choice {
	case STATEDB:
		return executor.conf.GetInt(stateBucketSize)
	case STATEOBJECT:
		return executor.conf.GetInt(stateObjectBucketSize)
	default:
		executor.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketLevelGroup - get bucket level group
func (executor *Executor) GetBucketLevelGroup(choice string) int {
	switch choice {
	case STATEDB:
		return executor.conf.GetInt(stateBucketLevelGroup)
	case STATEOBJECT:
		return executor.conf.GetInt(stateObjectBucketLevelGroup)
	default:
		executor.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketCacheSize - get bucket cache size
func (executor *Executor) GetBucketCacheSize(choice string) int {
	switch choice {
	case STATEDB:
		return executor.conf.GetInt(stateBucketCacheSize)
	case STATEOBJECT:
		return executor.conf.GetInt(stateObjectBucketCacheSize)
	default:
		executor.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetSyncReplicaInterval - sync replica information interval.
func (executor *Executor) GetSyncReplicaInterval() time.Duration {
	return executor.conf.GetDuration(syncReplicaInterval)
}

// GetSyncReplicaEnable - sync replica switch value.
func (exector *Executor) GetSyncReplicaEnable() bool {
	return exector.conf.GetBool(syncReplicaEnable)
}

func (executor *Executor) GetSyncMaxBatchSize() uint64 {
	return uint64(executor.conf.GetInt64(syncChainBatchSize))
}

func (executor *Executor) GetSyncResendInterval() time.Duration {
	return executor.conf.GetDuration(syncChainResendInterval)
}

func (executor *Executor) GetManifestPath() string {
	return executor.conf.GetString(snapshotManifestPath)
}

func (executor *Executor) GetArchiveMetaPath() string {
	return executor.conf.GetString(archiveMetaPath)
}

func (executor *Executor) IsArchiveForceConsistency() bool {
	return executor.conf.GetBool(archiveForceConsistency)
}

func (executor *Executor) GetArchiveThreshold() int {
	return executor.conf.GetInt(archiveThreshold)
}