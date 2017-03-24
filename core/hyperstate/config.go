package hyperstate

const (
	STATEDB               = "state"
	stateBucketSize       = "global.configs.buckettree.state.size"
	stateBucketLevelGroup = "global.configs.buckettree.state.levelGroup"
	stateBucketCacheSize  = "global.configs.buckettree.state.cacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.configs.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.configs.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.configs.buckettree.storage.cacheSize"
)

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
