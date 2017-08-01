package executor

import (
	"time"
	"hyperchain/common"
)

const (
	STATEDB               = "state"
	stateBucketSize       = "executor.buckettree.state.size"
	stateBucketLevelGroup = "executor.buckettree.state.levelGroup"
	stateBucketCacheSize  = "executor.buckettree.state.cacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.executor.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.executor.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.executor.buckettree.storage.cacheSize"

	syncReplicaInterval         = "executor.sync_replica.interval"
	syncReplicaEnable           = "executor.sync_replica.enable"

	syncChainBatchSize          = "executor.sync_chain.sync_batch_size"
	syncChainResendInterval     = "executor.sync_chain.sync_resend_interval"
	exitFlag                    = "executor.nvp.exitflag"

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

func (executor *Executor) isJvmEnable() bool {
	return executor.conf.GetBool(common.C_JVM_START)
}

func (executor *Executor) GetExitFlag() bool {
	return executor.conf.GetBool(exitFlag)
}