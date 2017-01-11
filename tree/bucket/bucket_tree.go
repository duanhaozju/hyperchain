package bucket

import (
	"bytes"
	"github.com/op/go-logging"
	"hyperchain/hyperdb"
	"hyperchain/common"
	"math/big"
	"errors"
	"encoding/json"
	"time"
)

var (
	log = logging.MustGetLogger("buckettree")
	DataNodePrefix = "DataNode"
	BucketNodePrefix = "BucketNode"
	UpdatedValueSetPrefix = "UpdatedValueSet"
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
	bucketCache            *BucketCache
	dataNodeCache          *DataNodeCache
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
	//sort.Sort(key_valueMap)
	//log.Debug("Enter - PrepareWorkingSet()")
	if key_valueMap == nil || len(key_valueMap) == 0 {
		//log.Debug("Ignoring working-set as it is empty")
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
	log.Debug("Enter - ComputeCryptoHash()")
	// TODO there maybe have concurrent error
	if bucketTree.recomputeCryptoHash {
		log.Debug("Recomputing crypto-hash...")
		start_time := time.Now()
		err := bucketTree.processDataNodeDelta()
		if err != nil {
			return nil, err
		}
		if(bucketTree.treePrefix != "-bucket=state"){
			log.Criticalf("bucketTree.processDataNodeDelta cost time is ",time.Since(start_time))
		}
		start_time = time.Now()
		err = bucketTree.processBucketTreeDelta()
		if err != nil {
			return nil, err
		}
		if(bucketTree.treePrefix != "-bucket=state"){
			log.Criticalf("bucketTree.processBucketTreeDelta cost time is ",time.Since(start_time))
		}
		bucketTree.lastComputedCryptoHash = bucketTree.computeRootNodeCryptoHash()
		bucketTree.recomputeCryptoHash = false
	} else {
		log.Debug("Returing existing crypto-hash as recomputation not required")
	}
	return bucketTree.lastComputedCryptoHash, nil
}

func (bucketTree *BucketTree) processDataNodeDelta() error {
	afftectedBuckets := bucketTree.dataNodesDelta.getAffectedBuckets()
	var start_time_FetchDataNodesFromCache time.Duration
	var start_time_computeDataNodesCryptoHash time.Duration

	for _, bucketKey := range afftectedBuckets {
		updatedDataNodes := bucketTree.dataNodesDelta.getSortedDataNodesFor(bucketKey)
		start_time := time.Now()
		existingDataNodes, err := bucketTree.dataNodeCache.FetchDataNodesFromCache(*bucketKey)
		if err != nil {
			return err
		}
		start_time_FetchDataNodesFromCache += time.Since(start_time)

		start_time = time.Now()
		cryptoHashForBucket,newDataNodes := computeDataNodesCryptoHash(bucketKey, updatedDataNodes, existingDataNodes,bucketTree.updatedValueSet)
		start_time_computeDataNodesCryptoHash += time.Since(start_time)

		bucketTree.updateDataNodeCache(*bucketKey,newDataNodes)

		//log.Debugf("Crypto-hash for lowest-level bucket [%s] is [%x]", bucketKey, cryptoHashForBucket)
		parentBucket := bucketTree.bucketTreeDelta.getOrCreateBucketNode(bucketKey.getParentKey())
		parentBucket.setChildCryptoHash(bucketKey, cryptoHashForBucket)
		//log.Debugf("bucket tree prefix %s bucket key %s, bucket hash %s",
		//	bucketTree.treePrefix, bucketKey.String(), common.Bytes2Hex(cryptoHashForBucket))
	}
	if(bucketTree.treePrefix != "-bucket-state"){
		log.Criticalf("start_time_FetchDataNodesFromCache cost time is",start_time_FetchDataNodesFromCache)
		log.Criticalf("start_time_computeDataNodesCryptoHash cost time is",start_time_computeDataNodesCryptoHash)
	}

	return nil
}

func (bucketTree *BucketTree) processBucketTreeDelta() error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		bucketNodes := bucketTree.bucketTreeDelta.getBucketNodesAt(level)
		//log.Debugf("Bucket tree delta. Number of buckets at level [%d] are [%d]", level, len(bucketNodes))
		for _, bucketNode := range bucketNodes {
			//log.Debugf("bucketNode in tree-delta [%s]", bucketNode)
			dbBucketNode, err := bucketTree.bucketCache.get(*bucketNode.bucketKey)
			//log.Debugf("bucket node from db [%s]", dbBucketNode)
			if err != nil {
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
			cryptoHash := bucketNode.computeCryptoHash()
			//log.Debugf("cryptoHash for bucket [%s] is [%x]", bucketNode, cryptoHash)
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

func computeDataNodesCryptoHash(bucketKey *BucketKey, updatedNodes DataNodes, existingNodes DataNodes,updatedValueSet *UpdatedValueSet) ([]byte,DataNodes) {
	//log.Debugf("Computing crypto-hash for bucket [%s]. numUpdatedNodes=[%d], numExistingNodes=[%d]", bucketKey, len(updatedNodes), len(existingNodes))
	bucketHashCalculator := newBucketHashCalculator(bucketKey)
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
			compositeKey := string(updatedNode.getCompositeKey())
			updatedValueSet.Set(compositeKey,updatedNode.value, nil)
			i++
		case 0:
			nextNode = updatedNode
			if bytes.Compare(updatedNode.getValue(),existingNode.value) != 0{
				compositeKey := string(updatedNode.getCompositeKey())
				//log.Debugf("update updated value set, composite key %s, current value %s, origin value %s", compositeKey, common.Bytes2Hex(updatedNode.value), common.Bytes2Hex(existingNode.value))
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
			newDataNodes = append(newDataNodes,nextNode)
		}
	}

	var remainingNodes DataNodes
	if i < len(updatedNodes) {
		remainingNodes = updatedNodes[i:]
		for k :=0 ; k<len(remainingNodes);k++ {
			compositeKey := string(remainingNodes[k].getCompositeKey()[:])
			updatedValueSet.Set(compositeKey,remainingNodes[k].value,nil)
		}
	} else if j < len(existingNodes) {
		remainingNodes = existingNodes[j:]
	}

	for _, remainingNode := range remainingNodes {
		if !remainingNode.isDelete() {
			bucketHashCalculator.addNextNode(remainingNode)
			newDataNodes = append(newDataNodes,remainingNode)
		}
	}
	return bucketHashCalculator.computeCryptoHash(),newDataNodes
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
				//log.Debugf("Deleting data node key = %#v", datanode.dataKey)
				writeBatch.Delete(append([]byte(DataNodePrefix), datanode.dataKey.getEncodedBytes()...))
			} else {
				//log.Debugf("Adding data node with key = %s value = %s",datanode.dataKey,common.Bytes2Hex(datanode.value))
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
				writeBatch.Delete(append([]byte(BucketNodePrefix), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...))
			} else {
				writeBatch.Put(append([]byte(BucketNodePrefix), append([]byte(bucketTree.treePrefix), bucketNode.bucketKey.getEncodedBytes()...)...), bucketNode.marshal())
			}
		}
	}
}

// TODO it should be test later
func (bucketTree *BucketTree) addUpdatedValueSetForPersistence(writeBatch hyperdb.Batch) {
	updatedValueSet := bucketTree.updatedValueSet
	data, err := json.Marshal(updatedValueSet)
	if err != nil {
		log.Errorf("marshal updated value set failed")
		return
	}
	dbKey := append([]byte(UpdatedValueSetPrefix), updatedValueSet.BlockNum.Bytes()...)
	dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
	writeBatch.Put(dbKey, data)
}

func (updatedValueSet *UpdatedValueSet) Print(treePrefix string){
	//log.Debugf("UpdatedValueSet block number #%d", updatedValueSet.BlockNum)
	for k,v := range updatedValueSet.UpdatedKVs{
		realTreePrefix ,realKey := DecodeCompositeKey([]byte(k))
		if realTreePrefix != treePrefix {
			log.Errorf("Error the updatedValueSet Print Error")
		}
		log.Errorf("Print key is %v",realKey)
		log.Error("previous value is ",common.Bytes2Hex(v.PreviousValue))
		log.Error("value is ",common.Bytes2Hex(v.Value))
		//logger.Debug("previous value is ",common.Bytes2Hex(v.PreviousValue))
		//logger.Debug("value is ",common.Bytes2Hex(v.Value))
	}
}

// TODO test
func (bucketTree *BucketTree) updateCacheWithoutPersist(currentBlockNum *big.Int){
	value, ok := bucketTree.treeHashMap[currentBlockNum]
	if ok{
		log.Debug("the map has the block tree hash ",currentBlockNum)
		if(bytes.Compare(value,bucketTree.lastComputedCryptoHash)==0){
			log.Debug("the key hash is same as before ",value)
		}
	}
	bucketTree.treeHashMap[currentBlockNum] = bucketTree.lastComputedCryptoHash
	bucketTree.updateBucketCache()
	bucketTree.dataNodesDelta = nil
	bucketTree.bucketTreeDelta = nil
	bucketTree.updatedValueSet = nil
	bucketTree.recomputeCryptoHash = false
}



func (bucketTree *BucketTree) updateDataNodeCache(bucketKey BucketKey,newDataNodes DataNodes){
	if bucketTree.dataNodesDelta == nil {
		return
	}
	if(bucketTree.dataNodeCache.isEnabled){
		bucketTree.dataNodeCache.c[bucketKey] = newDataNodes
	}
	if(globalDataNodeCache.isEnable){
		if(globalDataNodeCache.cache[bucketTree.treePrefix] == nil){
			globalDataNodeCache.cache[bucketTree.treePrefix] = make(map[BucketKey] DataNodes)
		}
		globalDataNodeCache.cache[bucketTree.treePrefix][bucketKey] = newDataNodes
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
	log.Debug("Start RevertToTargetBlock, from ",currentBlockNum)
	db,_ := hyperdb.GetLDBDatabase()
	writeBatch := db.NewBatch()
	keyValueMap := NewKVMap()
	bucketTree.dataNodeCache.ClearDataNodeCache()
	bucketTree.bucketCache.clearAllCache()
	globalDataNodeCache.ClearAllCache()
	globalDataNodeCache.isEnable = false
	bucketTree.bucketCache.isEnabled = false
	bucketTree.dataNodeCache.isEnabled = false

	for i:= currentBlockNum.Int64() + 1;;i++{
		dbKey := append([]byte(UpdatedValueSetPrefix), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
		_,err := db.Get(dbKey)
		if err != nil{
			if  err.Error()== "leveldb: not found" {
				currentBlockNum = big.NewInt(i-1)
				break
			}else {
				return err
			}

		} else {
			continue
		}
	}

	for i := currentBlockNum.Int64(); i > toBlockNum.Int64(); i -- {
		dbKey := append([]byte(UpdatedValueSetPrefix), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey,[]byte(bucketTree.treePrefix)...)
		value, err := db.Get(dbKey)
		if err != nil{
			if err.Error() == "leveldb: not found" {
				log.Debug("current block has no change",i)
				continue
			} else {
				log.Debug("Current BlockNum is",i,"Test RevertToTargetBlock Error",err.Error())
				return err
			}
			continue
		}
		if value == nil || len(value) == 0{
			log.Debugf("There is no value update")
			continue
		}
		updatedValueSet := newUpdatedValueSet(big.NewInt(i))
		err = json.Unmarshal(value, updatedValueSet)
		if err != nil {
			log.Errorf("unmarshal bucket updated values failed. for #%d", i)
		}

		revertToTargetBlock(bucketTree.treePrefix, big.NewInt(i), updatedValueSet, &keyValueMap)
		bucketTree.PrepareWorkingSet(keyValueMap,big.NewInt(i))
		bucketTree.AddChangesForPersistence(writeBatch,big.NewInt(i))
		keyValueMap = NewKVMap()
		writeBatch.Delete(dbKey)
		writeBatch.Write()
	}
	writeBatch.Write()
	/*for i := currentBlockNum.Int64();i > toBlockNum.Int64(); i -- {
		dbKey := append([]byte(UpdatedValueSetPrefix), big.NewInt(i).Bytes()...)
		dbKey = append(dbKey, []byte(bucketTree.treePrefix)...)
		_, err := db.Get(dbKey)
		if err != nil {
			logger.Error("test blockNum is ",i,"error is ",err.Error())
		}
	}*/

	bucketTree.dataNodeCache.ClearDataNodeCache()
	bucketTree.bucketCache.clearAllCache()
	bucketTree.bucketCache.isEnabled = true
	bucketTree.dataNodeCache.isEnabled = true
	globalDataNodeCache.isEnable = GLOBAL
	globalDataNodeCache.ClearAllCache()
	return nil
}

// TODO add verify about value and previousvalue
// TODO there should be some errors in the func
func revertToTargetBlock(treePrefix string, blockNum *big.Int,updatedValueSet *UpdatedValueSet,keyValueMap *K_VMap)  {
	for key, updatedValue := range updatedValueSet.UpdatedKVs {
		realTreePrefix,realKey := DecodeCompositeKey([]byte(key))
		(*keyValueMap)[realKey] = updatedValue.PreviousValue
		if(treePrefix != realTreePrefix){
			log.Debugf("RevertToTargetBlock error realTreePrefix",realTreePrefix)
			log.Debugf("RevertToTargetBlock error treePrefix",treePrefix)
		}
	}
}
