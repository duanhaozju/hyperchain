package bucket

import (
	"hyperchain/hyperdb"
	"sync"
	"encoding/json"
	"github.com/pkg/errors"
)
var DataNodeCachePrefix = "-dataNodecache"
type DataNodeMap map[string] *DataNode

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[BucketKey] DataNodeMap
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
	return &DataNodeCache{TreePrefix: treePrefix,c: make(map[BucketKey] DataNodeMap), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (datanodecache *DataNodeCache) Remove(dataNode *DataNode) error{
	bucketKey := *(dataNode.dataKey.bucketKey)
	dataKey := dataNode.dataKey
	dataKeyBytes,err := json.Marshal(*dataKey)
	if err != nil {
		logger.Error("Remove Error ",err)
		return err
	}
	if  datanodecache.c == nil || len(datanodecache.c) == 0 || datanodecache.c[bucketKey] == nil {
		return errors.New("There is no data in cache")
	}

	delete(datanodecache.c[bucketKey],string(dataKeyBytes))

	db,_ := hyperdb.GetLDBDatabase()
	dbkey := append([]byte(DataNodePrefix), dataNode.dataKey.getEncodedBytes()...)
	dbkey = append([]byte(DataNodeCachePrefix),dbkey...)
	db.Delete(dbkey)

	return nil
}

func (datanodecache *DataNodeCache) Put(dataNode *DataNode){
	bucketKey := *(dataNode.dataKey.bucketKey)
	dataKey := dataNode.dataKey
	dataKeyBytes,err := json.Marshal(*dataKey)
	if err != nil {
		logger.Error("Remove Error ",err)
	}
	if(datanodecache.c[bucketKey] == nil){
		datanodecache.c[bucketKey] = make(DataNodeMap)
		datanodecache.c[bucketKey][string(dataKeyBytes)] = dataNode
	}else {
		datanodecache.c[bucketKey][string(dataKeyBytes)] = dataNode
	}

	db,_ := hyperdb.GetLDBDatabase()
	dbkey := append([]byte(DataNodePrefix), dataNode.dataKey.getEncodedBytes()...)
	dbkey = append([]byte(DataNodeCachePrefix),dbkey...)
	db.Put(dbkey, dataNode.getValue())
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




func (datanodecache *DataNodeCache) fetchDataNodesFromCacheFor(bucketKey BucketKey) (dataNodes DataNodes,err error) {

	if(datanodecache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(datanodecache.TreePrefix,&bucketKey)
	}

	dataNodeMap := datanodecache.c[bucketKey]
	if(dataNodeMap == nil){
		logger.Errorf("The bucket is nil, bucketLevel is [%d] bucketNumber [%d]",bucketKey.level,bucketKey.bucketNumber)
		return dataNodes,nil
	}

	if len(dataNodeMap) == 0{
		dataNodes,err = fetchDataNodesFromDBByBucketKey(datanodecache.TreePrefix,&bucketKey)
		if err != nil{
			logger.Errorf("fetchDataNodesFromDBByBucketKey Error")
			return dataNodes,err
		}
		for _,dataNode := range dataNodes {
			datanodecache.Put(dataNode)
		}
	}else {
		for _,dataNode := range dataNodeMap{
			dataNodes = append(dataNodes, dataNode)
		}
	}
	logger.Debugf("FetchDataNodesFromCacheFor the datanode to dataNodeCache [%v]",dataNodes)
	return dataNodes, nil

	/*db,_ := hyperdb.GetLDBDatabase()
	minimumDataKeyBytes := minimumPossibleDataKeyBytesFor(&bucketKey,datanodecache.TreePrefix)
	minimumDataKeyBytes = append([]byte(DataNodeCachePrefix),minimumDataKeyBytes...)
	// IMPORTANT return value obtained by iterator is sorted
	iter := db.NewIteratorWithPrefix(minimumDataKeyBytes)
	num := 0
	for iter.Next() {
		num ++
		keyBytes := iter.Key()
		valueBytes := iter.Value()
		keyBytes = keyBytes[len(DataNodePrefix)+len(DataNodeCachePrefix):]

		dataKey := newDataKeyFromEncodedBytes(keyBytes)
		logger.Debugf("Retrieved data key [%s] from DB for bucket [%s]", dataKey, bucketKey)
		if !dataKey.getBucketKey().equals(&bucketKey) {
			logger.Errorf("Data key [%s] from DB does not belong to bucket = [%s]. Stopping further iteration and returning results [%v]", dataKey, bucketKey, dataNodes)
			return dataNodes, nil
		}
		dataNode := unmarshalDataNode(dataKey, valueBytes)
		logger.Debugf("Data node [%s] from DB belongs to bucket = [%s]. Including the key in results...", dataNode, bucketKey)
		dataNodes = append(dataNodes, dataNode)
	}
	if num <= 0{
		dataNodes,err = fetchDataNodesFromDBByBucketKey(datanodecache.TreePrefix,&bucketKey)
		if err != nil{
			logger.Errorf("fetchDataNodesFromDBByBucketKey Error")
			return dataNodes,err
		}
		for _,dataNode := range dataNodes {
			datanodecache.Put(dataNode)
		}
	}
	return dataNodes, nil*/

}

func (datanodecache *DataNodeCache) clearDataNodeCache() {
	db, _ := hyperdb.GetLDBDatabase()
	iter := db.NewIteratorWithPrefix([]byte(DataNodeCachePrefix))
	for iter.Next() {
		keyBytes := iter.Key()
		db.Delete(keyBytes)
	}
	datanodecache.c = make(map[BucketKey] DataNodeMap)
}

func (dataNodeCache *DataNodeCache) revertDataNodeCacheFromDB() error{
	/*dataNodeCache.c = make(map[BucketKey]DataNodeMap)
	dataNodes,err := fetchDataNodesFromDB(dataNodeCache.TreePrefix)
	if (err != nil){
		return err
	}
	for _,dataNode := range dataNodes {
		dataNodeCache.Put(dataNode)
	}*/

	return nil
}