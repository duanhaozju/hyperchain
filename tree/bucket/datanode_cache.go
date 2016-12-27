package bucket

import (
	"hyperchain/hyperdb"
	"sync"
)
var DataNodeCachePrefix = "-dataNodecache"
type dataNodeMap map[*dataKey] dataNode

type dataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[bucketKey] *dataNodeMap
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
}
func newDataNodeCache(treePrefix string,maxSizeMBs int) *dataNodeCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		logger.Infof("Constructing datanode-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	return &dataNodeCache{TreePrefix: treePrefix,c: make(map[bucketKey]*dataNodeMap), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (datanodecache *dataNodeCache) Remove(bucketkey bucketKey,datakey *dataKey){
	logger.Criticalf("Remove the datanode to dataNodeCache %d",len(datanodecache.c))
	if  datanodecache.c == nil || len(datanodecache.c) == 0 || datanodecache.c[bucketkey] == nil {
		logger.Error("There is no data in cache")
		return
	}
	delete(*datanodecache.c[bucketkey],datakey)
}

func (datanodecache *dataNodeCache) Put(bucketkey bucketKey,datakey *dataKey,datanode dataNode){
	logger.Criticalf("put the datanode to dataNodeCache %d",len(datanodecache.c))
	if(datanodecache.c[bucketkey] == nil){
		datanodemap := make(dataNodeMap)
		datanodemap[datakey] = datanode
		datanodecache.c[bucketkey] = &datanodemap
	}else {
		datanodemap := datanodecache.c[bucketkey]
		(*datanodemap)[datakey] = datanode
	}
}

func (datanodecache *dataNodeCache) Get(bucket_key bucketKey,data_key *dataKey)  (*dataNode, error) {
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




func (datanodecache *dataNodeCache) fetchDataNodesFromCacheFor(treePrefix string,bucketKey bucketKey)  (dataNodes, error) {
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
	var dataNodes dataNodes
	// IMPORTANT return value obtained by iterator is sorted
	iter := db.NewIteratorWithPrefix(minimumDataKeyBytes)
	for iter.Next() {
		keyBytes := iter.Key()
		valueBytes := iter.Value()
		keyBytes = keyBytes[len(DataNodePrefix)+len(DataNodeCachePrefix):]

		dataKey := newDataKeyFromEncodedBytes(keyBytes)
		logger.Debugf("Retrieved data key [%s] from DB for bucket [%s]", dataKey, bucketKey)
		if !dataKey.getBucketKey().equals(&bucketKey) {
			logger.Debugf("Data key [%s] from DB does not belong to bucket = [%s]. Stopping further iteration and returning results [%v]", dataKey, bucketKey, dataNodes)
			return dataNodes, nil
		}
		dataNode := unmarshalDataNode(dataKey, valueBytes)
		logger.Debugf("Data node [%s] from DB belongs to bucket = [%s]. Including the key in results...", dataNode, bucketKey)
		dataNodes = append(dataNodes, dataNode)
	}
	logger.Debugf("Returning results [%v]", dataNodes)
	return dataNodes, nil

}

func (datanodecache *dataNodeCache) clearDataNodeCache() {
	db, _ := hyperdb.GetLDBDatabase()
	iter := db.NewIteratorWithPrefix([]byte(DataNodeCachePrefix))
	for iter.Next() {
		logger.Criticalf("delete clearDataNodeCache")
		keyBytes := iter.Key()
		db.Delete(keyBytes)
	}
	datanodecache = &dataNodeCache{TreePrefix: datanodecache.TreePrefix,c: make(map[bucketKey]*dataNodeMap), maxSize: uint64(datanodecache.maxSize * 1024 * 1024), isEnabled: datanodecache.isEnabled}

}