package bucket

import (
	"sync"
	"encoding/json"
	"github.com/pkg/errors"
	"sort"
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

func (dataNodeCache *DataNodeCache) Remove(dataNode *DataNode) error{
	bucketKey := *(dataNode.dataKey.bucketKey)
	dataKey := dataNode.dataKey
	if  dataNodeCache.c == nil || len(dataNodeCache.c) == 0 || dataNodeCache.c[bucketKey] == nil {
		return errors.New("There is no data in cache")
	}
	delete(dataNodeCache.c[bucketKey],string(dataKey.compositeKey))
	return nil
}

func (dataNodeCache *DataNodeCache) Put (dataNode *DataNode){
	bucketKey := *(dataNode.dataKey.bucketKey)
	dataKey := dataNode.dataKey
	if(dataNodeCache.c[bucketKey] == nil){
		dataNodeCache.c[bucketKey] = make(DataNodeMap)
		dataNodeCache.c[bucketKey][string(dataKey.compositeKey)] = dataNode
	}else {
		dataNodeCache.c[bucketKey][string(dataKey.compositeKey)] = dataNode
	}
}

func (dataNodeCache *DataNodeCache) Get(bucket_key BucketKey,data_key *DataKey)  (*DataNode, error) {
	if !dataNodeCache.isEnabled {
		return fetchDataNodeFromDB(data_key)
	}
	dataNodeCache.lock.RLock()
	defer dataNodeCache.lock.RUnlock()
	datanodeMap := dataNodeCache.c[bucket_key]
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

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (dataNodes DataNodes,err error) {
	if(dataNodeCache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	}

	dataNodeMap := dataNodeCache.c[bucketKey]

	if dataNodeMap == nil || len(dataNodeMap) == 0 {
		logger.Debugf("The bucket is nil, bucketLevel is [%d] bucketNumber [%d]",bucketKey.level,bucketKey.bucketNumber)
		dataNodes,err = fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
		if err != nil{
			logger.Error("fetchDataNodesFromDBByBucketKey Error")
			return dataNodes,err
		}
		for _,dataNode := range dataNodes {
			logger.Debugf("FetchDataNodesFromCache put dataNode to dataNodeCache [%v]",dataNode)
			dataNodeCache.Put(dataNode)
		}
	}else {
		for _, dataNode := range dataNodeMap {
			logger.Debugf("Get datanode from cache [%v]", dataNode)
			dataNodes = append(dataNodes, dataNode)
		}
		//
		sort.Sort(dataNodes)
	}
	logger.Debugf("FetchDataNodesFromCacheFor the datanode to dataNodeCache [%v]",dataNodes)
	return dataNodes, nil
}


func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c = make(map[BucketKey] DataNodeMap)
}