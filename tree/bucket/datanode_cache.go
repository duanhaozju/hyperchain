package bucket

import (
	"github.com/hashicorp/golang-lru"
	"sync"
)

var (
	DefaultDataNodeCacheMaxSize = 10000
	GlobalDataNodeCacheSize     = 10000
	IsEnabledGlobal = true
	globalDataNodeCache         *GlobalDataNodeCache
)


func init() {
	globalDataNodeCache = &GlobalDataNodeCache{cacheMap: make(map[string]*lru.Cache), isEnable: IsEnabledGlobal}
}

type GlobalDataNodeCache struct {
	cacheMap map[string]*lru.Cache
	isEnable bool
	lock     sync.RWMutex
}
func (globalDataNodeCache *GlobalDataNodeCache) ClearAllCache() {
	globalDataNodeCache.cacheMap = make(map[string]*lru.Cache)
}

// Get - get a cache from global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Get(prefix string) *lru.Cache {
	globalDataNodeCache.lock.RLock()
	defer globalDataNodeCache.lock.RUnlock()
	return globalDataNodeCache.cacheMap[prefix]
}

// Add - add a new cache into global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Add(prefix string, cache *lru.Cache) {
	globalDataNodeCache.lock.Lock()
	defer globalDataNodeCache.lock.Unlock()
	globalDataNodeCache.cacheMap[prefix] = cache
}

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          *lru.Cache
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
}

func newDataNodeCache(treePrefix string, maxSizeMBs int) *DataNodeCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		log.Infof("Constructing datanode-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	if globalDataNodeCache.isEnable {
		if globalDataNodeCache.cacheMap[treePrefix] == nil {
			globalDataNodeCache.cacheMap[treePrefix], _ = lru.New(GlobalDataNodeCacheSize)
		} else {
			return &DataNodeCache{TreePrefix: treePrefix, c: globalDataNodeCache.cacheMap[treePrefix], maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
		}
	}
	cache, _ := lru.New(DefaultDataNodeCacheMaxSize)
	return &DataNodeCache{TreePrefix: treePrefix, c: cache, maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(bucketKey BucketKey) (dataNodes DataNodes, err error) {
	// step 0.
	if dataNodeCache.isEnabled == false {
		return fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix, &bucketKey)
	}

	// step 1.
	value, ok := dataNodeCache.c.Get(bucketKey)
	if ok {
		dataNodes = value.(DataNodes)
	}

	if dataNodes != nil && len(dataNodes) > 0 {
		return dataNodes, nil
	}

	// step 2.
	if globalDataNodeCache.isEnable {
		cache := globalDataNodeCache.Get(dataNodeCache.TreePrefix)
		if cache == nil {
			cache, _ = lru.New(GlobalDataNodeCacheSize)
			globalDataNodeCache.Add(dataNodeCache.TreePrefix, cache)
		}
		value, ok = cache.Get(bucketKey)
		if ok {
			dataNodes = value.(DataNodes)
		}
		if dataNodes != nil && len(dataNodes) > 0 {
			if dataNodeCache.isEnabled {
				dataNodeCache.c.Add(bucketKey, dataNodes)
			}
			return dataNodes, nil
		}
	}

	// step 3.
	dataNodes, err = fetchDataNodesFromDBByBucketKey(dataNodeCache.TreePrefix, &bucketKey)
	if err != nil {
		log.Error("fetchDataNodesFromDBByBucketKey Error")
		return dataNodes, err
	}
	if dataNodes == nil || len(dataNodes) == 0 {
		return dataNodes, nil
	} else {
		if dataNodeCache.isEnabled {
			dataNodeCache.c.Add(bucketKey, dataNodes)
		}
		if globalDataNodeCache.isEnable {
			if globalDataNodeCache.Get(dataNodeCache.TreePrefix) == nil {
				cache, _ := lru.New(GlobalDataNodeCacheSize)
				globalDataNodeCache.Add(dataNodeCache.TreePrefix, cache)
			}
			cache := globalDataNodeCache.Get(dataNodeCache.TreePrefix)
			cache.Add(bucketKey, dataNodes)
		}
		return dataNodes, nil
	}
}

func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c, _ = lru.New(DefaultDataNodeCacheMaxSize)
}
