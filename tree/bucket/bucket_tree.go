package bucket

import (
	"bytes"
	"github.com/op/go-logging"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"math/big"
	"github.com/golang/protobuf/proto"
	"errors"
)

var logger = logging.MustGetLogger("buckettree")

type K_VMap map[string][]byte

func NewKVMap() K_VMap {
	ret := make(map[string][]byte)
	return ret
}



// StateImpl - implements the interface - 'statemgmt.HashableState'
type BucketTree struct {
	treePrefix             string
	dataNodesDelta         *dataNodesDelta
	bucketTreeDelta        *bucketTreeDelta
	persistedStateHash     []byte
	lastComputedCryptoHash []byte
	recomputeCryptoHash    bool
	bucketCache            *bucketCache
	updatedValueSet        *UpdatedValueSet
	treeHashMap	       map[*big.Int][]byte
}

type Conf struct {
	StateSize         int
	StateLevelGroup   int
	StorageSize       int
	StorageLevelGroup int
}

// NewStateImpl constructs a new StateImpl
func NewBucketTree(tree_prefix string) *BucketTree {
	return &BucketTree{treePrefix:tree_prefix}
}

// Initialize - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) Initialize(configs map[string]interface{}) error {
	initConfig(configs)
	rootBucketNode, err := fetchBucketNodeFromDB(bucketTree.treePrefix, constructRootBucketKey())
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
	bucketTree.bucketCache = newBucketCache(bucketTree.treePrefix, bucketCacheMaxSize)
	bucketTree.bucketCache.loadAllBucketNodesFromDB()
	bucketTree.treeHashMap = make(map[*big.Int][]byte)
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
	bucketTree.dataNodesDelta = newDataNodesDelta(bucketTree.treePrefix, key_valueMap)
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
		existingDataNodes, err := fetchDataNodesFromDBFor(bucketTree.treePrefix, bucketKey)
		if err != nil {
			return err
		}
		// TODO test, add the logic of record the UpdatedValueSet
		cryptoHashForBucket := computeDataNodesCryptoHash(bucketKey, updatedDataNodes, existingDataNodes,bucketTree.updatedValueSet)
		logger.Debugf("Crypto-hash for lowest-level bucket [%s] is [%x]", bucketKey, cryptoHashForBucket)
		parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketKey.getParentKey())
		parentBucket.setChildCryptoHash(bucketKey, cryptoHashForBucket)
		logger.Noticef("bucket tree prefix %s bucket key %s, bucket hash %s",
			bucketTree.treePrefix, bucketKey.String(), common.Bytes2Hex(cryptoHashForBucket))
	}
	return nil
}

func (bucketTree *BucketTree) processBucketTreeDelta() error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		logger.Noticef("Bucket tree delta. Number of buckets at level [%d] are [%d]", level, len(bucketNodes))
		for _, bucketNode := range bucketNodes {
			logger.Noticef("bucketNode in tree-delta [%s]", bucketNode)
			dbBucketNode, err := bucketTree.bucketCache.get(*bucketNode.bucketKey)
			logger.Noticef("bucket node from db [%s]", dbBucketNode)
			if err != nil {
				return err
			}
			if dbBucketNode != nil {
				bucketNode.mergeBucketNode(dbBucketNode)
				logger.Noticef("After merge... bucketNode in tree-delta [%s]", bucketNode)
			}
			if level == 0 {
				return nil
			}
			logger.Noticef("Computing cryptoHash for bucket [%s]", bucketNode)
			cryptoHash := bucketNode.computeCryptoHash()
			logger.Noticef("cryptoHash for bucket [%s] is [%x]", bucketNode, cryptoHash)
			parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketNode.bucketKey.getParentKey())
			parentBucket.setChildCryptoHash(bucketNode.bucketKey, cryptoHash)
		}
	}
	return nil
}

func (bucketTree *BucketTree) computeRootNodeCryptoHash() []byte {
	return bucketTree.bucketTreeDelta.getRootNode().computeCryptoHash()
}

// TODO test
func (bucketTree *BucketTree) GetTreeHash(blockNum *big.Int) ([]byte,error){
	value,ok := bucketTree.treeHashMap[blockNum]
	if ok{
		return value,nil
	}else {
		return nil,errors.New("has no hash of this blockNum")
	}
}

func computeDataNodesCryptoHash(bucketKey *bucketKey, updatedNodes dataNodes, existingNodes dataNodes,updatedValueSet *UpdatedValueSet) []byte {
	logger.Noticef("Computing crypto-hash for bucket [%s]. numUpdatedNodes=[%d], numExistingNodes=[%d]", bucketKey, len(updatedNodes), len(existingNodes))
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
			compositeKey := string(updatedNode.getCompositeKey())
			updatedValueSet.Set(compositeKey,updatedNode.value,nil)
			i++
		case 0:
			nextNode = updatedNode
			if bytes.Compare(updatedNode.getValue(),existingNode.getValue()) != 0{
				compositeKey := string(updatedNode.getCompositeKey())
				updatedValueSet.Set(compositeKey,updatedNode.value,existingNode.getValue())
			}

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
		for i := 0;i<len(updatedNodes);i++ {
			compositeKey := string(updatedNodes[i].getCompositeKey()[:])
			updatedValueSet.Set(compositeKey,updatedNodes[i].value,nil)
		}
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
func (bucketTree *BucketTree) AddChangesForPersistence(writeBatch hyperdb.Batch,currentBlockNum *big.Int) error {

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
	bucketTree.addUpdatedValueSetForPersistence(writeBatch)
	// TODO test
	// 1.add the computehash to a temp array
	// 2.add the bucketcache record to bucketCache
	bucketTree.updateBucketCacheWithoutPersist(currentBlockNum)
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
				writeBatch.Delete(append([]byte("DataNode"), dataNode.dataKey.getEncodedBytes()...))
			} else {
				logger.Debugf("Adding data node with value = %#v", dataNode.value)
				writeBatch.Put(append([]byte("DataNode"), dataNode.dataKey.getEncodedBytes()...), dataNode.value)
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
				writeBatch.Delete(append([]byte("BucketNode"), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...))
			} else {
				writeBatch.Put(append([]byte("BucketNode"), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...), bucketNode.marshal())
			}
		}
	}
}

// TODO it should be test later
func (bucketTree *BucketTree) addUpdatedValueSetForPersistence(writeBatch hyperdb.Batch) {
	buffer := proto.NewBuffer([]byte{})
	updatedValueSet := bucketTree.updatedValueSet
	updatedValueSet.Marshal(buffer)
	writeBatch.Put(append([]byte("UpdatedValueSet"),updatedValueSet.BlockNum.Bytes()...),buffer.Bytes())
}

// TODO test
func (bucketTree *BucketTree) updateBucketCacheWithoutPersist(currentBlockNum *big.Int){
	value, ok := bucketTree.treeHashMap[currentBlockNum]
	if ok{
		logger.Debug("the map has the block tree hash ",currentBlockNum)
		if(bytes.Compare(value,bucketTree.lastComputedCryptoHash)==0){
			logger.Debug("the key hash is same as before ",value)
		}
	}
	bucketTree.treeHashMap[currentBlockNum] = bucketTree.lastComputedCryptoHash
	bucketTree.updateBucketCache()
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.recomputeCryptoHash = false
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

// TODO test
// it should be used when the statedb reset
func (bucket *BucketTree) Reset() {
	bucket.ClearWorkingSet(false)
}

// TODO test important
// the func can make the buckettree revert to target block
func (bucketTree *BucketTree) RevertToTargetBlock(currentBlockNum, toBlockNum *big.Int) (error){
	logger.Debug("Start RevertToTargetBlock, from ",currentBlockNum)
	db,_ := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()
	keyValueMap := NewKVMap()
	bucketTree.bucketCache.clearAllCache()

	for i:= currentBlockNum.Int64()+1;;i++{
		value,err := db.Get(append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...))
		if err != nil{
			if  err.Error()=="leveldb not found" {
				logger.Debug("all UpdatedValueSet before ",i,"has been clear")
				break
			}else {
				return err
			}

		}
		logger.Debug("delete the UpdatedValueSet of block ",i,"before ",value)
	}

	for i := currentBlockNum.Int64();i > toBlockNum.Int64(); i -- {
		value,err := db.Get(append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...))
		if err != nil{
			logger.Debug("Test RevertToTargetBlock Error",err.Error())
			return err
		}
		if value == nil || len(value)== 0{
			logger.Debug("There is no value update")
			continue
		}
		logger.Debug(i,"ValueValue",value)
		updatedValueSet := newUpdatedValueSet(big.NewInt(i))
		buffer := proto.NewBuffer(value)
		updatedValueSet.UnMarshal(buffer)
		revertToTargetBlock(bucketTree.treePrefix,big.NewInt(i),updatedValueSet,&keyValueMap)
		bucketTree.PrepareWorkingSet(keyValueMap,big.NewInt(i))
		bucketTree.AddChangesForPersistence(writeBatch,big.NewInt(i))
		keyValueMap = NewKVMap()
		writeBatch.Delete(append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...))
	}

	writeBatch.Write()
	logger.Debug("End RevertToTargetBlock, to ", toBlockNum)
	return nil
}

// TODO add verify about value and previousvalue
func revertToTargetBlock(treePrefix string,blockNum *big.Int,updatedValueSet *UpdatedValueSet,keyValueMap *K_VMap)  {
	length := len(treePrefix) + len(blockNum.Bytes())
	for key,updatedValue := range updatedValueSet.UpdatedKVs{
		logger.Debug(key[length:],"------------------previous value is ",updatedValue.PreviousValue,"------------------current value is ",updatedValue.Value)
		(*keyValueMap)[key[length:]] = updatedValue.PreviousValue
	}
}