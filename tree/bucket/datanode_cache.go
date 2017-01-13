package bucket

import (
	"sync"
	"github.com/hashicorp/golang-lru"
)
var (
	defaultDataNodeCacheMaxSize = 1000000
	GLOBAL = true
	GlobalDataNodeCacheSize = 10
	globalDataNodeCache *GlobalDataNodeCache
)
type GlobalDataNodeCache struct{
	cacheMap map[string] *lru.Cache
	marshacacheMap map[string] *lru.Cache
	isEnable bool
}

func init(){
	globalDataNodeCache = &GlobalDataNodeCache{cacheMap:make(map[string] *lru.Cache),isEnable:true}
}

func (globalDataNodeCache *GlobalDataNodeCache) ClearAllCache (){
	globalDataNodeCache.cacheMap = make(map[string] *lru.Cache)
}

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          *lru.Cache
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
		if(globalDataNodeCache.cacheMap[treePrefix] == nil){
			globalDataNodeCache.cacheMap[treePrefix],_ = lru.New(GlobalDataNodeCacheSize)
		}else{
			return &DataNodeCache{TreePrefix: treePrefix,c:globalDataNodeCache.cacheMap[treePrefix] , maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
		}
	}
	cache,_ := lru.New(defaultDataNodeCacheMaxSize)
	return &DataNodeCache{TreePrefix: treePrefix, c:cache, maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (dataNodes DataNodes,err error) {
	// step 0.
	if(dataNodeCache.isEnabled == false){
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	}

	// step 1.
	value,ok := dataNodeCache.c.Get(bucketKey)
	if(ok){
		dataNodes = value.(DataNodes)
	}

	if(dataNodes != nil && len(dataNodes) > 0){
		return dataNodes,nil
	}
	log.Critical("dataNode cache is empty")

	// step 2.
	if(globalDataNodeCache.isEnable){
		cache := globalDataNodeCache.cacheMap[dataNodeCache.TreePrefix]
		if cache == nil{
			cache,_ = lru.New(GlobalDataNodeCacheSize)
			globalDataNodeCache.cacheMap[dataNodeCache.TreePrefix] = cache
		}
		value,ok = cache.Get(bucketKey)
		if(ok){
			dataNodes = value.(DataNodes)
		}
		if(dataNodes != nil && len(dataNodes) > 0){
			if(dataNodeCache.isEnabled){
				dataNodeCache.c.Add(bucketKey,dataNodes)
			}
			return dataNodes,nil
		}
	}
	log.Critical("globalDataNodeCache is empty")

	// step 3.
	dataNodes,err = fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix,&bucketKey)
	if err != nil{
		log.Error("fetchDataNodesFromDBByBucketKey Error")
		return dataNodes,err
	}
	if dataNodes == nil || len(dataNodes) == 0 {
		return dataNodes,nil
	}else {
		if(dataNodeCache.isEnabled){
			dataNodeCache.c.Add(bucketKey,dataNodes)
		}
		if(globalDataNodeCache.isEnable){
			if(globalDataNodeCache.cacheMap[dataNodeCache.TreePrefix] == nil){
				cache,_ := lru.New(GlobalDataNodeCacheSize)
				globalDataNodeCache.cacheMap[dataNodeCache.TreePrefix] = cache
			}
			globalDataNodeCache.cacheMap[dataNodeCache.TreePrefix].Add(bucketKey,dataNodes)
		}
		return dataNodes,nil
	}
}


func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c,_ = lru.New(defaultDataNodeCacheMaxSize)
}