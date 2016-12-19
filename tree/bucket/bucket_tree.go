package bucket

import (
	"bytes"
	"github.com/op/go-logging"
	"hyperchain/hyperdb"
)

var logger = logging.MustGetLogger("buckettree")
type K_VMap map[string][]byte

// StateImpl - implements the interface - 'statemgmt.HashableState'
type BucketTree struct {
	treePrefix              string
	dataNodesDelta         *dataNodesDelta
	bucketTreeDelta        *bucketTreeDelta
	persistedStateHash     []byte
	lastComputedCryptoHash []byte
	recomputeCryptoHash    bool
	bucketCache            *bucketCache
}

// NewStateImpl constructs a new StateImpl
func NewBucketTree(tree_prefix string) *BucketTree {
	return &BucketTree{treePrefix:tree_prefix}
}

// Initialize - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) Initialize(configs map[string]interface{}) error {
	initConfig(configs)
	rootBucketNode, err := fetchBucketNodeFromDB(bucketTree.treePrefix,constructRootBucketKey())
	if err != nil {
		return err
	}
	if rootBucketNode != nil {
		bucketTree.persistedStateHash = rootBucketNode.computeCryptoHash()
		bucketTree.lastComputedCryptoHash = bucketTree.persistedStateHash
	}

	bucketCacheMaxSize, ok := configs["bucketCacheSize"].(int)
	if !ok {
		bucketCacheMaxSize = defaultBucketCacheMaxSize
	}
	bucketTree.bucketCache = newBucketCache(bucketTree.treePrefix,bucketCacheMaxSize)
	bucketTree.bucketCache.loadAllBucketNodesFromDB()
	return nil
}

// Get - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) Get(key string) ([]byte, error) {
	dataKey := newDataKey(bucketTree.treePrefix, key)
	dataNode, err := fetchDataNodeFromDB(dataKey)
	if err != nil {
		return nil, err
	}
	if dataNode == nil {
		return nil, nil
	}
	return dataNode.value, nil
}

// PrepareWorkingSet - method implementation for interface 'statemgmt.HashableState'
// TODO test the stateImpl just accept the stateDelta which accountID equals
func (bucketTree *BucketTree) PrepareWorkingSet(key_valueMap K_VMap) error {
	logger.Debug("Enter - PrepareWorkingSet()")

	if key_valueMap == nil || len(key_valueMap) == 0 {
		logger.Debug("Ignoring working-set as it is empty")
		return nil
	}
	bucketTree.dataNodesDelta = newDataNodesDelta(bucketTree.treePrefix,key_valueMap)
	bucketTree.bucketTreeDelta = newBucketTreeDelta()
	bucketTree.recomputeCryptoHash = true
	return nil
}

// ClearWorkingSet - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) ClearWorkingSet(changesPersisted bool) {
	logger.Debug("Enter - ClearWorkingSet()")
	if changesPersisted {
		bucketTree.persistedStateHash = bucketTree.lastComputedCryptoHash
		bucketTree.updateBucketCache()
	} else {
		bucketTree.lastComputedCryptoHash = bucketTree.persistedStateHash
	}
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.recomputeCryptoHash = false
}

// ComputeCryptoHash - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) ComputeCryptoHash() ([]byte, error) {
	logger.Debug("Enter - ComputeCryptoHash()")
	if bucketTree.recomputeCryptoHash {
		logger.Debug("Recomputing crypto-hash...")
		err := bucketTree.processDataNodeDelta()
		if err != nil {
			return nil, err
		}
		err = bucketTree.processBucketTreeDelta()
		if err != nil {
			return nil, err
		}
		bucketTree.lastComputedCryptoHash = bucketTree.computeRootNodeCryptoHash()
		bucketTree.recomputeCryptoHash = false
	} else {
		logger.Debug("Returing existing crypto-hash as recomputation not required")
	}
	return bucketTree.lastComputedCryptoHash, nil
}

func (bucketTree *BucketTree) processDataNodeDelta() error {
	afftectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	for _, bucketKey := range afftectedBuckets {
		updatedDataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(bucketKey)
		existingDataNodes, err := fetchDataNodesFromDBFor(bucketKey)
		if err != nil {
			return err
		}
		cryptoHashForBucket := computeDataNodesCryptoHash(bucketKey, updatedDataNodes, existingDataNodes)
		logger.Debugf("Crypto-hash for lowest-level bucket [%s] is [%x]", bucketKey, cryptoHashForBucket)
		parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketKey.getParentKey())
		parentBucket.setChildCryptoHash(bucketKey, cryptoHashForBucket)
		logger.Notice("*******************the cryptoHashForBucket is ",cryptoHashForBucket)
	}
	return nil
}

func (bucketTree *BucketTree) processBucketTreeDelta() error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		logger.Debugf("Bucket tree delta. Number of buckets at level [%d] are [%d]", level, len(bucketNodes))
		for _, bucketNode := range bucketNodes {
			logger.Debugf("bucketNode in tree-delta [%s]", bucketNode)
			dbBucketNode, err := bucketTree.bucketCache.get(*bucketNode.bucketKey)
			logger.Debugf("bucket node from db [%s]", dbBucketNode)
			if err != nil {
				return err
			}
			if dbBucketNode != nil {
				bucketNode.mergeBucketNode(dbBucketNode)
				logger.Debugf("After merge... bucketNode in tree-delta [%s]", bucketNode)
			}
			if level == 0 {
				return nil
			}
			logger.Debugf("Computing cryptoHash for bucket [%s]", bucketNode)
			cryptoHash := bucketNode.computeCryptoHash()
			logger.Debugf("cryptoHash for bucket [%s] is [%x]", bucketNode, cryptoHash)
			parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketNode.bucketKey.getParentKey())
			parentBucket.setChildCryptoHash(bucketNode.bucketKey, cryptoHash)
		}
	}
	return nil
}

func (bucketTree *BucketTree) computeRootNodeCryptoHash() []byte {
	return bucketTree.bucketTreeDelta.getRootNode().computeCryptoHash()
}

func computeDataNodesCryptoHash(bucketKey *bucketKey, updatedNodes dataNodes, existingNodes dataNodes) []byte {
	logger.Debugf("Computing crypto-hash for bucket [%s]. numUpdatedNodes=[%d], numExistingNodes=[%d]", bucketKey, len(updatedNodes), len(existingNodes))
	bucketHashCalculator := newBucketHashCalculator(bucketKey)
	i := 0
	j := 0
	for i < len(updatedNodes) && j < len(existingNodes) {
		updatedNode := updatedNodes[i]
		existingNode := existingNodes[j]
		c := bytes.Compare(updatedNode.dataKey.compositeKey, existingNode.dataKey.compositeKey)
		var nextNode *dataNode
		switch c {
		case -1:
			nextNode = updatedNode
			i++
		case 0:
			nextNode = updatedNode
			i++
			j++
		case 1:
			nextNode = existingNode
			j++
		}
		if !nextNode.isDelete() {
			bucketHashCalculator.addNextNode(nextNode)
		}
	}

	var remainingNodes dataNodes
	if i < len(updatedNodes) {
		remainingNodes = updatedNodes[i:]
	} else if j < len(existingNodes) {
		remainingNodes = existingNodes[j:]
	}

	for _, remainingNode := range remainingNodes {
		if !remainingNode.isDelete() {
			bucketHashCalculator.addNextNode(remainingNode)
		}
	}
	return bucketHashCalculator.computeCryptoHash()
}

// AddChangesForPersistence - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) AddChangesForPersistence(writeBatch hyperdb.Batch) error {

	if bucketTree.dataNodesDelta == nil {
		return nil
	}

	if bucketTree.recomputeCryptoHash {
		_, err := bucketTree.ComputeCryptoHash()
		if err != nil {
			return nil
		}
	}
	bucketTree.addDataNodeChangesForPersistence(writeBatch)
	bucketTree.addBucketNodeChangesForPersistence(writeBatch)
	return nil
}

// TODO it should be test later
func (bucketTree *BucketTree) addDataNodeChangesForPersistence(writeBatch hyperdb.Batch) {
	affectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	for _, affectedBucket := range affectedBuckets {
		dataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(affectedBucket)
		for _, dataNode := range dataNodes {
			if dataNode.isDelete() {
				logger.Debugf("Deleting data node key = %#v", dataNode.dataKey)
				writeBatch.Delete(dataNode.dataKey.getEncodedBytes())
			} else {
				logger.Debugf("Adding data node with value = %#v", dataNode.value)
				writeBatch.Put(dataNode.dataKey.getEncodedBytes(), dataNode.value)
			}
		}
	}
}

// TODO it should be test later
func (bucketTree *BucketTree) addBucketNodeChangesForPersistence(writeBatch hyperdb.Batch) {

	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		for _, bucketNode := range bucketNodes {
			if bucketNode.markedForDeletion {
				writeBatch.Delete(bucketNode.bucketKey.getEncodedBytes(bucketTree.treePrefix))
			} else {
				writeBatch.Put(bucketNode.bucketKey.getEncodedBytes(bucketTree.treePrefix), bucketNode.marshal())
			}
		}
	}
}

// TODO to do test with cache
func (bucketTree *BucketTree) updateBucketCache() {
	if bucketTree.bucketTreeDelta == nil || bucketTree.bucketTreeDelta.isEmpty() {
		return
	}
	bucketTree.bucketCache.lock.Lock()
	defer bucketTree.bucketCache.lock.Unlock()
	secondLastLevel := conf.getLowestLevel() - 1
	for level := 0; level <= secondLastLevel; level++ {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		for _, bucketNode := range bucketNodes {
			key := *bucketNode.bucketKey
			if bucketNode.markedForDeletion {
				bucketTree.bucketCache.removeWithoutLock(key)
			} else {
				bucketTree.bucketCache.putWithoutLock(key, bucketNode)
			}
		}
	}
}

// PerfHintKeyChanged - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) PerfHintKeyChanged(accountID string, key string) {
	// We can create a cache. Pull all the keys for the bucket (to which given key belongs) in a separate thread
	// This prefetching can help making method 'ComputeCryptoHash' faster.
}
