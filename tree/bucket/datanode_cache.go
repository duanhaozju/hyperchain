package bucket

import (
	"sync"
)
var (
	globalDataNodeCache = make(map[string] (map[BucketKey] DataNodes))
)
type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[BucketKey] DataNodes
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
}
func newDataNodeCache(treePrefix string,maxSizeMBs int) *DataNodeCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		log.Infof("Constructing datanode-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	if(globalDataNodeCache[treePrefix] == nil){
		globalDataNodeCache[treePrefix] = make(map[BucketKey] DataNodes)
	}
	return &DataNodeCache{TreePrefix: treePrefix,c: globalDataNodeCache[treePrefix], maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (DataNodes,error) {
	if(dataNodeCache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	}
	dataNodes := dataNodeCache.c[bucketKey]
	if(dataNodeCache.TreePrefix != "-bucket-state"){
		log.Criticalf("FetchDataNodesFromCache bucketKey is",bucketKey," length is",len(dataNodes))
	}

	if dataNodes == nil || len(dataNodes) == 0 {
		log.Debugf("The bucket is nil, bucketLevel is [%d] bucketNumber [%d]",bucketKey.level,bucketKey.bucketNumber)
		dbDataNodes,err := fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
		if err != nil{
			log.Error("fetchDataNodesFromDBByBucketKey Error")
			return dbDataNodes,err
		}
		return dbDataNodes,nil
	}
	return dataNodeCache.c[bucketKey],nil
}


func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c = make(map[BucketKey] DataNodes)
}