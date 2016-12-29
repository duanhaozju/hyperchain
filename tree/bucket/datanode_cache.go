package bucket

import (
	"sync"
	"encoding/json"
	"github.com/pkg/errors"
)
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
}

func (datanodecache *DataNodeCache) Get(bucket_key BucketKey,data_key *DataKey)  (*DataNode, error) {
	if !datanodecache.isEnabled {
		return fetchDataNodeFromDB(data_key)
	}
	datanodecache.lock.RLock()
	defer datanodecache.lock.RUnlock()
	datanodeMap := datanodecache.c[bucket_key]
	if datanodeMap == nil {
		return fetchDataNodeFromDB(data_key)
	}
	dataKeyBytes,err := json.Marshal(*data_key)
	if err != nil {
		logger.Errorf("json.Marshal Error",err)
		return nil,err
	}
	return datanodeMap[string(dataKeyBytes)], nil
}

func (datanodecache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (dataNodes DataNodes,err error) {

	if(datanodecache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(datanodecache.TreePrefix,&bucketKey)
	}

	dataNodeMap := datanodecache.c[bucketKey]

	if dataNodeMap == nil || len(dataNodeMap) == 0 {
		logger.Errorf("The bucket is nil, bucketLevel is [%d] bucketNumber [%d]",bucketKey.level,bucketKey.bucketNumber)
		dataNodes,err = fetchDataNodesFromDBByBucketKey(datanodecache.TreePrefix,&bucketKey)
		if err != nil{
			logger.Errorf("fetchDataNodesFromDBByBucketKey Error")
			return dataNodes,err
		}
		for _,dataNode := range dataNodes {
			datanodecache.Put(dataNode)
		}
	}else {
		for _,dataNode := range dataNodeMap {
			dataNodes = append(dataNodes, dataNode)
		}
	}
	logger.Debugf("FetchDataNodesFromCacheFor the datanode to dataNodeCache [%v]",dataNodes)
	return dataNodes, nil
}


func (datanodecache *DataNodeCache) ClearDataNodeCache() {
	datanodecache.c = make(map[BucketKey] DataNodeMap)
}