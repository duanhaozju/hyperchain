package executor

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
func (executor *Executor) GetBucketSize(choice string) int {
	switch choice {
	case STATEDB:
		return executor.conf.GetInt(stateBucketSize)
	case STATEOBJECT:
		return executor.conf.GetInt(stateObjectBucketSize)
	default:
		log.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
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
		log.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
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
		log.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJECT)
		return 0
	}
}
