package bucket

import (
	"sync"
)
var (
	GLOBAL = false
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
	// step 0.
	if(dataNodeCache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	}
	// step 1.
	dataNodes := dataNodeCache.c[bucketKey]
	if(dataNodeCache.TreePrefix != "-bucket-state"){
		log.Debugf("FetchDataNodesFromCache bucketKey is",bucketKey," length is",len(dataNodes))
	}
	if(dataNodes != nil && len(dataNodes) > 0){
		return dataNodes,nil
	}
	// step 2.
	if(globalDataNodeCache.isEnable){
		dataNodes = globalDataNodeCache.cache[dataNodeCache.TreePrefix][bucketKey]
		if(dataNodes != nil && len(dataNodes) > 0){
			if(dataNodeCache.isEnabled){
				dataNodeCache.c[bucketKey] = dataNodes
			}
			return dataNodes,nil
		}
	}

	// step 3.
	dataNodes,err := fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	if err != nil{
		log.Error("fetchDataNodesFromDBByBucketKey Error")
		return dataNodes,err
	}
	if(dataNodeCache.isEnabled){
		dataNodeCache.c[bucketKey] = dataNodes
	}
	if(globalDataNodeCache.isEnable){
		if(globalDataNodeCache.cache[dataNodeCache.TreePrefix] == nil){
			globalDataNodeCache.cache[dataNodeCache.TreePrefix] = make(map[BucketKey] DataNodes)
		}
		globalDataNodeCache.cache[dataNodeCache.TreePrefix][bucketKey] = dataNodes
	}
	return dataNodes,nil
}


func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c = make(map[BucketKey] DataNodes)
}