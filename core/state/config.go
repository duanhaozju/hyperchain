package state

const (
	GlobalDataNodeCacheSize   = "executor.buckettree.global.globalDataNodeCacheSize"
	GlobalDataNodeCacheLength = "executor.buckettree.global.globalDataNodeCacheLength"

	STATEDB                = "state"
	StateBucketSize        = "executor.buckettree.state.size"
	StateBucketLevelGroup  = "executor.buckettree.state.levelGroup"
	StateBucketCacheSize   = "executor.buckettree.state.bucketCacheSize"
	StateDataNodeCacheSize = "executor.buckettree.state.dataNodeCacheSize"

	STATEOBJECT                  = "stateObject"
	StateObjectBucketSize        = "executor.buckettree.storage.size"
	StateObjectBucketLevelGroup  = "executor.buckettree.storage.levelGroup"
	StateObjectBucketCacheSize   = "executor.buckettree.storage.bucketCacheSize"
	StateObjectDataNodeCacheSize = "executor.buckettree.storage.dataNodeCacheSize"
)

// GetGlobalDataNodeCacheSize - get size of every global data node cache
func (stateDB *StateDB) GetGlobalDataNodeCacheSize() int {
	return stateDB.bktConf.GetInt(GlobalDataNodeCacheSize)
}

// GetGlobalDataNodeCacheLength - get the length of global data node cache
func (stateDB *StateDB) GetGlobalDataNodeCacheLength() int {
	return stateDB.bktConf.GetInt(GlobalDataNodeCacheLength)
}

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
