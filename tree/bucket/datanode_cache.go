package bucket

import (
	"sync"
)
var (
	globalDataNodeCache = &GlobalDataNodeCache{cache:make(map[string] (map[BucketKey] DataNodes)),isEnable:false}
)
type GlobalDataNodeCache struct{
	cache map[string] (map[BucketKey] DataNodes)
	isEnable bool
}
func (g *GlobalDataNodeCache) ClearAllCache (){
	g.cache = make(map[string] (map[BucketKey] DataNodes))
}

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
	if(globalDataNodeCache.isEnable){
		if(globalDataNodeCache.cache[treePrefix] == nil){
			globalDataNodeCache.cache[treePrefix] = make(map[BucketKey] DataNodes)
		}
		return &DataNodeCache{TreePrefix: treePrefix,c: globalDataNodeCache.cache[treePrefix], maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
	}
	return &DataNodeCache{TreePrefix: treePrefix,c: make(map[BucketKey] DataNodes), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (DataNodes,error) {
	if(dataNodeCache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	}
	dataNodes := dataNodeCache.c[bucketKey]
	if(dataNodeCache.TreePrefix != "-bucket-state"){
		log.Debugf("FetchDataNodesFromCache bucketKey is",bucketKey," length is",len(dataNodes))
	}

	if dataNodes == nil || len(dataNodes) == 0 {
		log.Debugf("The bucket is nil, bucketLevel is [%d] bucketNumber [%d]",bucketKey.level,bucketKey.bucketNumber)
		dbDataNodes,err := fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
		if err != nil{
			log.Error("fetchDataNodesFromDBByBucketKey Error")
			return dbDataNodes,err
		}
		dataNodeCache.c[bucketKey] = dataNodes
		return dbDataNodes,nil
	}
	return dataNodes,nil
}


func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c = make(map[BucketKey] DataNodes)
}