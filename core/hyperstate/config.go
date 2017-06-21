package hyperstate

const (
	GlobalDataNodeCacheSize  = "global.executor.buckettree.global.globalDataNodeCacheSize"
	GlobalDataNodeCacheLength  = "global.executor.buckettree.global.globalDataNodeCacheLength"

	STATEDB               = "state"
	stateBucketSize       = "global.executor.buckettree.state.size"
	stateBucketLevelGroup = "global.executor.buckettree.state.levelGroup"
	stateBucketCacheSize  = "global.executor.buckettree.state.bucketCacheSize"
	stateDataNodeCacheSize  = "global.executor.buckettree.state.dataNodeCacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.executor.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.executor.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.executor.buckettree.storage.bucketCacheSize"
	stateObjectDataNodeCacheSize  = "global.executor.buckettree.storage.dataNodeCacheSize"
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
		return stateDB.bktConf.GetInt(stateBucketSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(stateObjectBucketSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketLevelGroup - get bucket level group
func (stateDB *StateDB) GetBucketLevelGroup(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(stateBucketLevelGroup)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(stateObjectBucketLevelGroup)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetBucketCacheSize - get bucket cache size
func (stateDB *StateDB) GetBucketCacheSize(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(stateBucketCacheSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(stateObjectBucketCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}

// GetDataNodeCacheSize - get dataNode cache size
func (stateDB *StateDB) GetDataNodeCacheSize(choice string) int {
	switch choice {
	case STATEDB:
		return stateDB.bktConf.GetInt(stateDataNodeCacheSize)
	case STATEOBJECT:
		return stateDB.bktConf.GetInt(stateObjectDataNodeCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}