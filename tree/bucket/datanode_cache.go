package bucket

import (
	"hyperchain/hyperdb"
	"sync"
)
var DataNodeCachePrefix = "-dataNodecache"
type DataNodeMap map[*DataKey]DataNode

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[BucketKey] *DataNodeMap
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
}
func newDataNodeCache(treePrefix string,maxSizeMBs int) *DataNodeCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		logger.Infof("Constructing datanode-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	return &DataNodeCache{TreePrefix: treePrefix,c: make(map[BucketKey]*DataNodeMap), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (datanodecache *DataNodeCache) Remove(bucketkey BucketKey,datakey *DataKey){
	if  datanodecache.c == nil || len(datanodecache.c) == 0 || datanodecache.c[bucketkey] == nil {
		logger.Error("There is no data in cache")
		return
	}
	delete(*datanodecache.c[bucketkey],datakey)
}

func (datanodecache *DataNodeCache) Put(bucketkey BucketKey,datakey *DataKey,datanode DataNode){
	if(datanodecache.c[bucketkey] == nil){
		datanodemap := make(DataNodeMap)
		datanodemap[datakey] = datanode
		datanodecache.c[bucketkey] = &datanodemap
	}else {
		datanodemap := datanodecache.c[bucketkey]
		(*datanodemap)[datakey] = datanode
	}
}

func (datanodecache *DataNodeCache) Get(bucket_key BucketKey,data_key *DataKey)  (*DataNode, error) {
	//defer perfstat.UpdateTimeStat("timeSpent", time.Now())
	/*if !datanodecache.isEnabled {
		return fetchDataNodeFromDB(data_key)
	}
	datanodecache.lock.RLock()
	defer datanodecache.lock.RUnlock()
	datanodeMap := datanodecache.c[bucket_key]
	if datanodeMap == nil {
		return fetchDataNodeFromDB(data_key)
	}
	return &((*datanodeMap)[data_key]), nil*/
	return nil,nil
}




func (datanodecache *DataNodeCache) fetchDataNodesFromCacheFor(treePrefix string,bucketKey BucketKey)  (dataNodes, error) {
	if(datanodecache.isEnabled == false){
		return fetchDataNodesFromDBFor(treePrefix,&bucketKey)
	}

	/*var datanodes dataNodes
	datanodemap := datanodecache.c[bucketKey]
	if(datanodemap == nil){
		return datanodes,nil
	}
	for _,datanode := range *datanodemap{
		datanodes = append(datanodes, &datanode)
	}
	logger.Criticalf("fetchDataNodesFromCacheFor the datanode to dataNodeCache %d",len(datanodes))
	return datanodes,nil*/

	db,_ := hyperdb.GetLDBDatabase()
	minimumDataKeyBytes := minimumPossibleDataKeyBytesFor(&bucketKey,treePrefix)
	minimumDataKeyBytes = append([]byte(DataNodeCachePrefix),minimumDataKeyBytes...)
	var res dataNodes
	// IMPORTANT return value obtained by iterator is sorted
	iter := db.NewIteratorWithPrefix(minimumDataKeyBytes)
	for iter.Next() {
		keyBytes := iter.Key()
		valueBytes := iter.Value()
		keyBytes = keyBytes[len(DataNodePrefix)+len(DataNodeCachePrefix):]

		dataKey := newDataKeyFromEncodedBytes(keyBytes)
		logger.Debugf("Retrieved data key [%s] from DB for bucket [%s]", dataKey, bucketKey)
		if !dataKey.getBucketKey().equals(&bucketKey) {
			logger.Errorf("Data key [%s] from DB does not belong to bucket = [%s]. Stopping further iteration and returning results [%v]", dataKey, bucketKey, res)
			return res, nil
		}
		dataNode := unmarshalDataNode(dataKey, valueBytes)
		logger.Debugf("Data node [%s] from DB belongs to bucket = [%s]. Including the key in results...", dataNode, bucketKey)
		res = append(res, dataNode)
	}
	return res, nil

}

func (datanodecache *DataNodeCache) clearDataNodeCache() {
	db, _ := hyperdb.GetLDBDatabase()
	iter := db.NewIteratorWithPrefix([]byte(DataNodeCachePrefix))
	for iter.Next() {
		logger.Debugf("delete clearDataNodeCache")
		keyBytes := iter.Key()
		db.Delete(keyBytes)
	}
	datanodecache = &DataNodeCache{TreePrefix: datanodecache.TreePrefix,c: make(map[BucketKey]*DataNodeMap), maxSize: uint64(datanodecache.maxSize * 1024 * 1024), isEnabled: datanodecache.isEnabled}
}

func (dataNodeCache *DataNodeCache) revertDataNodeCache(){

}