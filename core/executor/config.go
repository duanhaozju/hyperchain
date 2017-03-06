package executor

const (
	stateType          = "global.structure.state"
	blockVersion       = "global.version.blockversion"
	transactionVersion = "global.version.transactionversion"

	STATEDB               = "state"
	stateBucketSize       = "global.configs.buckettree.state.size"
	stateBucketLevelGroup = "global.configs.buckettree.state.levelGroup"
	stateBucketCacheSize  = "global.configs.buckettree.state.cacheSize"

	STATEOBJECT                 = "stateObject"
	stateObjectBucketSize       = "global.configs.buckettree.storage.size"
	stateObjectBucketLevelGroup = "global.configs.buckettree.storage.levelGroup"
	stateObjectBucketCacheSize  = "global.configs.buckettree.storage.cacheSize"
)

// GetStateType - get state type, "rawstate" or "hyperstate"
// "rawstate" is the old version which use patricia merkle tree to manage data structure
// "hyperstate" is the latest version which use direct k-v set and bucket tree to manage.
func (executor *Executor) GetStateType() string {
	return executor.conf.GetString(stateType)
}

// GetBlockVersion - get block data structure version tag.
func (executor *Executor) GetBlockVersion() string {
	return executor.conf.GetString(blockVersion)
}

// GetTransactionVersion - get transaction data structure version tag.
func (executor *Executor) GetTransactionVersion() string {
	return executor.conf.GetString(transactionVersion)
}

// GetReceiptVersion - get receipt data structure version tag, which is same with transaction.
func (executor *Executor) GetReceiptVersion() string {
	return executor.GetTransactionVersion()
}

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
