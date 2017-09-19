package state

const (
	STATEDB                = "state"
	StateBucketSize        = "executor.buckettree.state.capacity"
	StateBucketLevelGroup  = "executor.buckettree.state.aggreation"
	StateBucketCacheSize   = "executor.buckettree.state.merklenode_cache"
	StateDataNodeCacheSize = "executor.buckettree.state.bucket_cache"

	STATEOBJECT                  = "stateObject"
	StateObjectBucketSize        = "executor.buckettree.storage.capacity"
	StateObjectBucketLevelGroup  = "executor.buckettree.storage.aggreation"
	StateObjectBucketCacheSize   = "executor.buckettree.storage.merklenode_cache"
	StateObjectDataNodeCacheSize = "executor.buckettree.storage.bucket_cache"
)

// GetBucketSize - get bucket size.
func (stateDB *StateDB) GetBucketSize(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(StateBucketSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(StateObjectBucketSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketLevelGroup - get bucket level group
func (stateDB *StateDB) GetBucketLevelGroup(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(StateBucketLevelGroup)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(StateObjectBucketLevelGroup)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketCacheSize - get bucket cache size
func (stateDB *StateDB) GetBucketCacheSize(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(StateBucketCacheSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(StateObjectBucketCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetDataNodeCacheSize - get dataNode cache size
func (stateDB *StateDB) GetDataNodeCacheSize(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(StateDataNodeCacheSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(StateObjectDataNodeCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}
