package executor

import "time"

const (
	STATEDB               = "state"
	stateBucketSize       = "global.configs.buckettree.state.size"
	stateBucketLevelGroup = "global.configs.buckettree.state.levelGroup"
	stateBucketCacheSize  = "global.configs.buckettree.state.cacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.configs.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.configs.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.configs.buckettree.storage.cacheSize"

	syncReplicaInterval         = "global.configs.replicainfo.interval"
	syncReplicaEnable           = "global.configs.replicainfo.enable"
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
