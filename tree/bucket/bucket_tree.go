package bucket

import (
	"bytes"
	"github.com/op/go-logging"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"math/big"
	"errors"
	"encoding/json"
)

var (
	logger = logging.MustGetLogger("buckettree")
	DataNodePrefix = "DataNode"
)
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
	dataNodeCache          *dataNodeCache
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
	bucketTree.dataNodeCache = newDataNodeCache(bucketTree.treePrefix, bucketCacheMaxSize)
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
func (bucketTree *BucketTree) PrepareWorkingSet(key_valueMap K_VMap, blockNum *big.Int) error {
	logger.Debug("Enter - PrepareWorkingSet()")
	if key_valueMap == nil || len(key_valueMap) == 0 {
		logger.Debug("Ignoring working-set as it is empty")
		return nil
	}
	bucketTree.dataNodesDelta = newDataNodesDelta(bucketTree.treePrefix, key_valueMap)
	bucketTree.bucketTreeDelta = newBucketTreeDelta()
	bucketTree.recomputeCryptoHash = true
	bucketTree.updatedValueSet = newUpdatedValueSet(blockNum)
	return nil
}

// ClearWorkingSet - method implementation for interface 'statemgmt.HashableState'
func (bucketTree *BucketTree) ClearWorkingSet(changesPersisted bool) {
	logger.Debug("Enter - ClearWorkingSet()")
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
	//hash,_ := bucketTree.ComputeCryptoHash()
	for _, bucketKey := range afftectedBuckets {

		//logger.Critical("blockhash is",common.Bytes2Hex(hash),"---------------------updatedDataNodes length is ",len(updatedDataNodes))
		//logger.Critical("blockhash is",common.Bytes2Hex(hash),"---------------------existingDataNodes length is ",len(existingDataNodes))
		updatedDataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(bucketKey)
		existingDataNodes, err := bucketTree.dataNodeCache.fetchDataNodesFromCacheFor(bucketTree.treePrefix, *bucketKey)
		if (bucketTree.treePrefix == "-bucket-state") {
			logger.Criticalf("updatedDataNodes length = %d",len(updatedDataNodes))
			logger.Criticalf("existingDataNodes length = %d",len(existingDataNodes))
		}
		if err != nil {
			return err
		}
		// TODO test, add the logic of record the UpdatedValueSet
		cryptoHashForBucket := computeDataNodesCryptoHash(bucketKey, updatedDataNodes, existingDataNodes,bucketTree.updatedValueSet)
		logger.Debugf("Crypto-hash for lowest-level bucket [%s] is [%x]", bucketKey, cryptoHashForBucket)
		parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketKey.getParentKey())
		parentBucket.setChildCryptoHash(bucketKey, cryptoHashForBucket)
		logger.Debugf("bucket tree prefix %s bucket key %s, bucket hash %s",
			bucketTree.treePrefix, bucketKey.String(), common.Bytes2Hex(cryptoHashForBucket))
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
			compositeKey := string(updatedNode.getCompositeKey())
			updatedValueSet.Set(compositeKey,updatedNode.value,nil)
			i++
		case 0:
			nextNode = updatedNode
			if bytes.Compare(updatedNode.getValue(),existingNode.value) != 0{
				compositeKey := string(updatedNode.getCompositeKey())
				updatedValueSet.Set(compositeKey,updatedNode.value,existingNode.value)
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
	bucketTree.updateCacheWithoutPersist(currentBlockNum)

	return nil
}

// TODO it should be test later
func (bucketTree *BucketTree) addDataNodeChangesForPersistence(writeBatch hyperdb.Batch) {
	affectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	for _, affectedBucket := range affectedBuckets {
		dataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(affectedBucket)
		for _, datanode := range dataNodes {
			if datanode.isDelete() {
				logger.Debugf("Deleting data node key = %#v", datanode.dataKey)
				writeBatch.Delete(append([]byte(DataNodePrefix), datanode.dataKey.getEncodedBytes()...))
			} else {
				logger.Debugf("Adding data node with value = %s", common.Bytes2Hex(datanode.value))
				writeBatch.Put(append([]byte(DataNodePrefix), datanode.dataKey.getEncodedBytes()...), datanode.value)
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
	updatedValueSet := bucketTree.updatedValueSet
	data, err := json.Marshal(updatedValueSet)
	if err != nil {
		logger.Error("marshal updated value set failed")
		return
	}
	dbKey := append([]byte("UpdatedValueSet"),updatedValueSet.BlockNum.Bytes()...)
	dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
	writeBatch.Put(dbKey, data)
	if bucketTree.treePrefix == "-bucket-state" {
		logger.Critical("addUpdatedValueSetForPersistence-----------------------------start",updatedValueSet.BlockNum.Int64())
		updatedValueSet.Print()
		logger.Critical("addUpdatedValueSetForPersistence-----------------------------end",updatedValueSet.BlockNum.Int64())
	}
}

func (updatedValueSet *UpdatedValueSet) Print(){
	logger.Criticalf("block number #%d", updatedValueSet.BlockNum)
	for k,v := range updatedValueSet.UpdatedKVs{
		logger.Critical("Print key is ",k)
		logger.Critical("previous value is ",common.Bytes2Hex(v.PreviousValue))
		logger.Critical("value is ",common.Bytes2Hex(v.Value))
	}
}

// TODO test
func (bucketTree *BucketTree) updateCacheWithoutPersist(currentBlockNum *big.Int){
	value, ok := bucketTree.treeHashMap[currentBlockNum]
	if ok{
		logger.Critical("the map has the block tree hash ",currentBlockNum)
		if(bytes.Compare(value,bucketTree.lastComputedCryptoHash)==0){
			logger.Critical("the key hash is same as before ",value)
		}
	}
	bucketTree.treeHashMap[currentBlockNum] = bucketTree.lastComputedCryptoHash
	bucketTree.updateDataNodeCache()
	bucketTree.updateBucketCache()
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.updatedValueSet = nil
	bucketTree.recomputeCryptoHash = false
}



func (bucketTree *BucketTree) updateDataNodeCache(){
	if bucketTree.dataNodesDelta == nil {
		return
	}
	db,_ := hyperdb.GetLDBDatabase()
	affectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	for _, affectedBucket := range affectedBuckets {
		dataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(affectedBucket)
		for _, dataNode := range dataNodes {
			if dataNode.isDelete() {
				bucketTree.dataNodeCache.Remove(*affectedBucket,dataNode.dataKey)
				dbkey := append([]byte(DataNodePrefix), dataNode.dataKey.getEncodedBytes()...)
				dbkey = append([]byte(DataNodeCachePrefix),dbkey...)
				db.Delete(dbkey)

			} else {
				bucketTree.dataNodeCache.Put(*affectedBucket,dataNode.dataKey,*dataNode)
				dbkey := append([]byte(DataNodePrefix), dataNode.dataKey.getEncodedBytes()...)
				dbkey = append([]byte(DataNodeCachePrefix),dbkey...)
				db.Put(dbkey, dataNode.value)
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
	bucketTree.dataNodeCache.clearDataNodeCache()
	bucketTree.dataNodeCache.isEnabled = false

	for i:= currentBlockNum.Int64() + 1;;i++{
		dbKey := append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
		_,err := db.Get(dbKey)
		if err != nil{
			if  err.Error()== "leveldb: not found" {
				logger.Error("all UpdatedValueSet before", i ,"has been clear")
				currentBlockNum = big.NewInt(i-1)
				break
			}else {
				logger.Error("clear UpdatedValueSet error",err)
				return err
			}

		} else {
			continue
		}
	}

	for i := currentBlockNum.Int64(); i > toBlockNum.Int64(); i -- {
		dbKey := append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
		value, err := db.Get(dbKey)
		if err != nil{
			logger.Error("Current BlockNum is",i,"bucketTree.treePrefix is ",bucketTree.treePrefix,"Test RevertToTargetBlock Error",err.Error())
			if err.Error() == "leveldb: not found" {
				logger.Debug("current block has no change",i)
				continue
			} else {
				logger.Debug("Current BlockNum is",i,"Test RevertToTargetBlock Error",err.Error())
				return err
			}
			continue
		}
		if value == nil || len(value) == 0{
			logger.Error("There is no value update")
			continue
		}
		hash,_ := bucketTree.ComputeCryptoHash()
		logger.Critical(bucketTree.treePrefix, "before revert to ", i, "current hash is ",common.Bytes2Hex(hash))
		updatedValueSet := newUpdatedValueSet(big.NewInt(i))
		err = json.Unmarshal(value, updatedValueSet)
		if err != nil {
			logger.Errorf("unmarshal bucket updated values failed. for #%d", i)
		}
		if bucketTree.treePrefix == "-bucket-state" {
			logger.Critical("before RevertToTargetBlock-------------------------start",i)
			updatedValueSet.Print()
			logger.Critical("before RevertToTargetBlock-------------------------end",i)
		}
		revertToTargetBlock(bucketTree.treePrefix, big.NewInt(i), updatedValueSet, &keyValueMap)
		if bucketTree.treePrefix == "-bucket-state" {
			logger.Critical("after RevertToTargetBlock-------------------------start",i)
			for k,v := range keyValueMap{
				logger.Critical("the key is ",k)
				logger.Critical("the value is ",common.Bytes2Hex(ComputeCryptoHash(v)))
			}
			logger.Critical("after RevertToTargetBlock-------------------------end",i)
		}
		bucketTree.PrepareWorkingSet(keyValueMap,big.NewInt(i))
		bucketTree.AddChangesForPersistence(writeBatch,big.NewInt(i))
		hash,_ = bucketTree.ComputeCryptoHash()
		logger.Critical(bucketTree.treePrefix,"after current blockTree is",i,"hash is",common.Bytes2Hex(hash))
		keyValueMap = NewKVMap()
		writeBatch.Delete(dbKey)
		writeBatch.Write()
	}
	writeBatch.Write()
	/*for i := currentBlockNum.Int64();i > toBlockNum.Int64(); i -- {
		dbKey := append([]byte("UpdatedValueSet"), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey, []byte(bucketTree.treePrefix)...)
		_, err := db.Get(dbKey)
		if err != nil {
			logger.Error("test blockNum is ",i,"error is ",err.Error())
		}
	}*/
	logger.Debug("End RevertToTargetBlock, to ", toBlockNum)
	hash, _ := bucketTree.ComputeCryptoHash()
	bucketTree.dataNodeCache.isEnabled = true
	logger.Criticalf("revert bucket tree current block number %d, current block hash %s", toBlockNum, common.Bytes2Hex(hash))
	return nil
}

// TODO add verify about value and previousvalue
func revertToTargetBlock(treePrefix string,blockNum *big.Int,updatedValueSet *UpdatedValueSet,keyValueMap *K_VMap)  {
	length := len(treePrefix) + len(blockNum.Bytes())
	for key, updatedValue := range updatedValueSet.UpdatedKVs {
		(*keyValueMap)[key[length:]] = updatedValue.PreviousValue
	}
}
