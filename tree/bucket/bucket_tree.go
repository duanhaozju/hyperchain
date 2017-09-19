package bucket

import (
	"bytes"
	"errors"
	"github.com/hashicorp/golang-lru"
	"github.com/op/go-logging"
	"hyperchain/common"
	tp "hyperchain/common/threadpool"
	hdb "hyperchain/hyperdb/db"
	"math/big"
	"runtime"
	"sync"
)

var (
	DataNodesPrefix       = "DataNodes"
	BucketNodePrefix      = "BucketNode"
	UpdatedValueSetPrefix = "UpdatedValueSet"
)

type K_VMap map[string][]byte

func NewKVMap() K_VMap {
	ret := make(map[string][]byte)
	return ret
}

// StateImpl - implements the interface - 'statemgmt.HashableState'
type BucketTree struct {
	db                     hdb.Database
	treePrefix             string
	dataNodesDelta         *dataNodesDelta
	bucketTreeDelta        *bucketTreeDelta
	persistedStateHash     []byte
	lastComputedCryptoHash []byte
	recomputeCryptoHash    bool
	bucketCache            *BucketCache
	dataNodeCache          *DataNodeCache
	treeHashMap            map[*big.Int][]byte
	log                    *logging.Logger
}

type Conf struct {
	StateSize         int
	StateLevelGroup   int
	StorageSize       int
	StorageLevelGroup int
}

// NewStateImpl constructs a new StateImpl
func NewBucketTree(db hdb.Database, tree_prefix string) *BucketTree {
	return &BucketTree{
		treePrefix: tree_prefix,
		db:         db,
		log:        common.GetLogger(db.Namespace(), "bucket"),
	}
}

// Initialize - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) Initialize(configs map[string]interface{}) error {
	initConfig(bucketTree.log, configs)
	if err := bucketTree.initLastComputed(); err != nil {
		return err
	}
	bucketTree.bucketCache = newBucketCache(bucketTree.treePrefix, configs[ConfigBucketCacheMaxSize].(int))
	bucketTree.bucketCache.loadAllBucketNodesFromDB()
	bucketTree.dataNodeCache = newDataNodeCache(bucketTree.log, bucketTree.db.Namespace(), bucketTree.treePrefix, configs[ConfigDataNodeCacheMaxSize].(int))
	bucketTree.treeHashMap = make(map[*big.Int][]byte)
	return nil
}

// PrepareWorkingSet - method implementation for interface 'statemgmt.HashableState'
// TODO test the stateImpl just accept the stateDelta which accountID equals
func (bucketTree *BucketTree) PrepareWorkingSet(key_valueMap K_VMap, blockNum *big.Int) error {
	//sort.Sort(key_valueMap)
	//log.Debug("Enter - PrepareWorkingSet()")
	if key_valueMap == nil || len(key_valueMap) == 0 {
		//log.Debug("Ignoring working-set as it is empty")
		return nil
	}
	bucketTree.dataNodesDelta = newDataNodesDelta(bucketTree.treePrefix, key_valueMap)
	bucketTree.bucketTreeDelta = newBucketTreeDelta()
	bucketTree.recomputeCryptoHash = true
	return nil
}

// ClearWorkingSet - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) ClearWorkingSet(changesPersisted bool) {
	//log.Debug("Enter - ClearWorkingSet()")
	if changesPersisted {
		bucketTree.persistedStateHash = bucketTree.lastComputedCryptoHash
		bucketTree.updateBucketCache()
	} else {
		//bucketTree.lastComputedCryptoHash = bucketTree.persistedStateHash
	}
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.recomputeCryptoHash = false
}

// ComputeCryptoHash - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) ComputeCryptoHash() ([]byte, error) {
	bucketTree.log.Debug("Enter - ComputeCryptoHash()")
	// TODO there maybe have concurrent error
	if bucketTree.recomputeCryptoHash {
		bucketTree.log.Debug("Recomputing crypto-hash...")
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
		bucketTree.log.Debug("Returing existing crypto-hash as recomputation not required")
	}
	return bucketTree.lastComputedCryptoHash, nil
}

func (bucketTree *BucketTree) processDataNodeDelta() error {
	var wg sync.WaitGroup
	afftectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()

	wg.Add(len(afftectedBuckets))

	handler := func(obj interface{}) interface{} {
		defer wg.Done()
		bucketKey, ok := obj.(*BucketKey)
		if !ok {
			return errors.New("expect bucket key")
		}

		updatedDataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(bucketKey)
		// thread safe
		existingDataNodes, err := bucketTree.dataNodeCache.FetchDataNodesFromCache(bucketTree.db, *bucketKey)
		if err != nil {
			bucketTree.log.Errorf("fetch datanodes failed. %s", err.Error())
			return err
		}
		// thread safe
		cryptoHashForBucket, newDataNodes := computeDataNodesCryptoHash(bucketKey, updatedDataNodes, existingDataNodes)

		// thread safe
		bucketTree.updateDataNodeCache(*bucketKey, newDataNodes)

		//log.Debugf("Crypto-hash for lowest-level bucket [%s] is [%x]", bucketKey, cryptoHashForBucket)
		// thread safe
		parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketKey.getParentKey())
		// thread safe
		parentBucket.setChildCryptoHash(bucketKey, cryptoHashForBucket)
		//log.Debugf("bucket tree prefix %s bucket key %s, bucket hash %s",
		//	bucketTree.treePrefix, bucketKey.String(), common.Bytes2Hex(cryptoHashForBucket))
		return nil
	}

	pool, _ := tp.CreatePool(runtime.NumCPU()+1, handler).Open()
	defer pool.Close()

	for _, bucketKey := range afftectedBuckets {
		pool.SendWorkAsync(bucketKey, nil)
	}
	wg.Wait()
	return nil
}

func (bucketTree *BucketTree) processBucketTreeDelta() error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		// thread safe
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		//log.Debugf("Bucket tree delta. Number of buckets at level [%d] are [%d]", level, len(bucketNodes))
		var wg sync.WaitGroup
		wg.Add(len(bucketNodes))

		handler := func(obj interface{}) interface{} {
			defer wg.Done()
			bucketNode, ok := obj.(*BucketNode)
			if !ok {
				return errors.New("expect to be bucketnode")
			}
			//log.Debugf("bucketNode in tree-delta [%s]", bucketNode)
			dbBucketNode, err := bucketTree.bucketCache.fetchBucketNodeFromCache(bucketTree.db, *bucketNode.bucketKey)
			//log.Debugf("bucket node from db [%s]", dbBucketNode)
			if err != nil {
				bucketTree.log.Errorf("get bucketnode from cache failed. %s", err.Error())
				return err
			}
			if dbBucketNode != nil {
				bucketNode.mergeBucketNode(dbBucketNode)
				//log.Debugf("After merge... bucketNode in tree-delta [%s]", bucketNode)
			}
			if level == 0 {
				return nil
			}
			//log.Debugf("Computing cryptoHash for bucket [%s]", bucketNode)
			cryptoHash := bucketNode.computeCryptoHash(bucketTree.log)
			//log.Debugf("cryptoHash for bucket [%s] is [%x]", bucketNode, cryptoHash)
			parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketNode.bucketKey.getParentKey())
			parentBucket.setChildCryptoHash(bucketNode.bucketKey, cryptoHash)
			return nil
		}

		pool, _ := tp.CreatePool(runtime.NumCPU()+1, handler).Open()

		for _, bucketNode := range bucketNodes {
			pool.SendWorkAsync(bucketNode, nil)
		}
		wg.Wait()
		pool.Close()
	}
	return nil
}

func (bucketTree *BucketTree) computeRootNodeCryptoHash() []byte {
	return bucketTree.bucketTreeDelta.getRootNode().computeCryptoHash(bucketTree.log)
}

// TODO test
func (bucketTree *BucketTree) GetTreeHash(blockNum *big.Int) ([]byte, error) {
	value, ok := bucketTree.treeHashMap[blockNum]
	if ok {
		return value, nil
	} else {
		return nil, errors.New("has no hash of this blockNum")
	}
}

func computeDataNodesCryptoHash(bucketKey *BucketKey, updatedNodes DataNodes, existingNodes DataNodes) ([]byte, DataNodes) {
	//log.Debugf("Computing crypto-hash for bucket [%s]. numUpdatedNodes=[%d], numExistingNodes=[%d]", bucketKey, len(updatedNodes), len(existingNodes))
	bucketHashCalculator := newBucketHashCalculator()
	i := 0
	j := 0
	var newDataNodes DataNodes
	for i < len(updatedNodes) && j < len(existingNodes) {
		updatedNode := updatedNodes[i]
		existingNode := existingNodes[j]
		c := bytes.Compare(updatedNode.dataKey.compositeKey, existingNode.dataKey.compositeKey)
		var nextNode *DataNode
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
			newDataNodes = append(newDataNodes, nextNode)
		}
	}

	var remainingNodes DataNodes
	if i < len(updatedNodes) {
		remainingNodes = updatedNodes[i:]
		for _, remainingNode := range remainingNodes {
			if !remainingNode.isDelete() {
				newDataNodes = append(newDataNodes, remainingNode)
			}
		}
	} else if j < len(existingNodes) {
		remainingNodes = existingNodes[j:]
		newDataNodes = append(newDataNodes, remainingNodes...)
	}
	hashingDataArray := make([][]byte, len(newDataNodes))
	for i, dataNode := range newDataNodes {
		hashingDataArray[i] = dataNode.getValue()
	}
	bucketHashCalculator.setHashingData(JoinBytes(hashingDataArray, ""))
	return bucketHashCalculator.computeCryptoHash(), newDataNodes
}

// AddChangesForPersistence - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) AddChangesForPersistence(writeBatch hdb.Batch, currentBlockNum *big.Int) error {
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
	bucketTree.updateCacheWithoutPersist(currentBlockNum)

	return nil
}

// TODO it should be test later
func (bucketTree *BucketTree) addDataNodeChangesForPersistence(writeBatch hdb.Batch) {
	affectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	for _, affectedBucket := range affectedBuckets {
		value, ok := bucketTree.dataNodeCache.cache.Get(*affectedBucket)
		if ok {
			dataNodes := value.(DataNodes)
			if dataNodes == nil || len(dataNodes) == 0 {
				writeBatch.Delete(append([]byte(bucketTree.treePrefix), append([]byte(DataNodesPrefix), affectedBucket.getEncodedBytes()...)...))
			} else {
				writeBatch.Put(append([]byte(bucketTree.treePrefix), append([]byte(DataNodesPrefix), affectedBucket.getEncodedBytes()...)...), dataNodes.Marshal())
			}
		}
	}
}

// TODO it should be test later
func (bucketTree *BucketTree) addBucketNodeChangesForPersistence(writeBatch hdb.Batch) {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		for _, bucketNode := range bucketNodes {
			if bucketNode.markedForDeletion {
				writeBatch.Delete(append([]byte(BucketNodePrefix), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...))
			} else {
				writeBatch.Put(append([]byte(BucketNodePrefix), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...), bucketNode.marshal())
			}
		}
	}
}

// TODO test
func (bucketTree *BucketTree) updateCacheWithoutPersist(currentBlockNum *big.Int) {
	value, ok := bucketTree.treeHashMap[currentBlockNum]
	if ok {
		bucketTree.log.Debug("the map has the block tree hash ", currentBlockNum)
		if bytes.Compare(value, bucketTree.lastComputedCryptoHash) == 0 {
			bucketTree.log.Debug("the key hash is same as before ", value)
		}
	}
	bucketTree.treeHashMap[currentBlockNum] = bucketTree.lastComputedCryptoHash
	bucketTree.updateBucketCache()
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.recomputeCryptoHash = false
}

func (bucketTree *BucketTree) updateDataNodeCache(bucketKey BucketKey, newDataNodes DataNodes) {
	if bucketTree.dataNodesDelta == nil {
		return
	}
	if bucketTree.dataNodeCache.isEnabled {
		bucketTree.dataNodeCache.cache.Add(bucketKey, newDataNodes)
	}
	if globalDataNodeCache.isEnable {
		cache := globalDataNodeCache.Get(ConstructPrefix(bucketTree.db.Namespace(), bucketTree.treePrefix))
		if cache == nil {
			cache, _ = lru.New(globalDataNodeCache.globalDataNodeCacheMaxSize)
		}
		cache.Add(bucketKey, newDataNodes)
		globalDataNodeCache.Add(ConstructPrefix(bucketTree.db.Namespace(), bucketTree.treePrefix), cache)
	}

}

// TODO to do test with cache
func (bucketTree *BucketTree) updateBucketCache() {
	if bucketTree.bucketTreeDelta == nil || bucketTree.bucketTreeDelta.isEmpty() {
		return
	}
	secondLastLevel := conf.getLowestLevel() - 1
	for level := 0; level <= secondLastLevel; level++ {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		for _, bucketNode := range bucketNodes {
			key := *bucketNode.bucketKey
			if bucketNode.markedForDeletion {
				bucketTree.bucketCache.remove(key)
			} else {
				bucketTree.bucketCache.put(key, bucketNode)
			}
		}
	}
}

// PerfHintKeyChanged - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) PerfHintKeyChanged(accountID string, key string) {
	// We can create a cache. Pull all the keys for the bucket (to which given key belongs) in a separate thread
	// This prefetching can help making method 'ComputeCryptoHash' faster.
}

// TODO test
// it should be used when the statedb reset
func (bucket *BucketTree) Reset() {
	bucket.ClearWorkingSet(false)
}

// TODO test important
// the func can make the buckettree revert to target block
func (bucketTree *BucketTree) RevertToTargetBlock(writeBatch hdb.Batch, currentBlockNum, toBlockNum *big.Int, flush, sync bool) error {
	bucketTree.log.Debug("Start RevertToTargetBlock, from ", currentBlockNum)
	keyValueMap := NewKVMap()
	bucketTree.dataNodeCache.ClearDataNodeCache()
	bucketTree.bucketCache.clearAllCache()
	globalDataNodeCache.ClearAllCache()
	globalDataNodeCache.isEnable = false
	bucketTree.bucketCache.isEnabled = true
	bucketTree.dataNodeCache.isEnabled = true

	for i := currentBlockNum.Int64() + 1; ; i++ {
		dbKey := append([]byte(UpdatedValueSetPrefix), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey, []byte(bucketTree.treePrefix)...)
		_, err := bucketTree.db.Get(dbKey)
		if err != nil {
			if err.Error() == hdb.DB_NOT_FOUND.Error() {
				currentBlockNum = big.NewInt(i - 1)
				break
			} else {
				return err
			}

		} else {
			continue
		}
	}

	for i := currentBlockNum.Int64(); i > toBlockNum.Int64(); i-- {
		dbKey := append([]byte(UpdatedValueSetPrefix), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey, []byte(bucketTree.treePrefix)...)
		value, err := bucketTree.db.Get(dbKey)

		if err != nil {
			if err.Error() == hdb.DB_NOT_FOUND.Error() {
				bucketTree.log.Debug("current block has no change", i)
				continue
			} else {
				bucketTree.log.Debug("Current BlockNum is", i, "Test RevertToTargetBlock Error", err.Error())
				return err
			}
			continue
		}
		if value == nil || len(value) == 0 {
			bucketTree.log.Debugf("There is no value update")
			continue
		}

		bucketTree.PrepareWorkingSet(keyValueMap, big.NewInt(i))
		bucketTree.AddChangesForPersistence(writeBatch, big.NewInt(i))

		keyValueMap = NewKVMap()
		writeBatch.Delete(dbKey)
		writeBatch.Write()
	}
	bucketTree.dataNodeCache.ClearDataNodeCache()
	bucketTree.bucketCache.clearAllCache()
	globalDataNodeCache.ClearAllCache()
	bucketTree.bucketCache.isEnabled = true
	bucketTree.dataNodeCache.isEnabled = true
	globalDataNodeCache.isEnable = IsEnabledGlobal
	if flush {
		if sync {
			writeBatch.Write()
		} else {
			go writeBatch.Write()
		}
	}
	return nil
}

func (bucketTree *BucketTree) ClearAllCache() {
	bucketTree.dataNodeCache.ClearDataNodeCache()
	bucketTree.bucketCache.clearAllCache()
	globalDataNodeCache.ClearAllCache()
	bucketTree.recomputeCryptoHash = false
	bucketTree.initLastComputed()
}

func (bucketTree *BucketTree) initLastComputed() error {
	rootBucketNode, err := fetchBucketNodeFromDB(bucketTree.db, bucketTree.treePrefix, constructRootBucketKey())
	if err != nil {
		return err
	}
	if rootBucketNode != nil {
		bucketTree.persistedStateHash = rootBucketNode.computeCryptoHash(bucketTree.log)
		bucketTree.lastComputedCryptoHash = bucketTree.persistedStateHash
	}
	return nil
}
