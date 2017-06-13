package bucket

import (
	"github.com/hashicorp/golang-lru"
	"sync"
	"hyperchain/hyperdb/db"
	"github.com/op/go-logging"
)

var (
	DefaultDataNodeCacheMaxSize = 400000
	GlobalDataNodeCacheSize     = 400000
	IsEnabledGlobal = true
	globalDataNodeCache         *GlobalDataNodeCache
)


func init() {
	globalDataNodeCache = &GlobalDataNodeCache{cacheMap: make(map[string]map[string]*lru.Cache), isEnable: IsEnabledGlobal}
}

type GlobalDataNodeCache struct {
	cacheMap map[string]map[string]*lru.Cache
	isEnable bool
	lock     sync.RWMutex
}
func (globalDataNodeCache *GlobalDataNodeCache) ClearAllCache() {
	globalDataNodeCache.cacheMap = make(map[string]map[string]*lru.Cache)
}

// Get - get a cache from global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Get(namespace, prefix string) *lru.Cache {
	globalDataNodeCache.lock.RLock()
	defer globalDataNodeCache.lock.RUnlock()
	cs, existed := globalDataNodeCache.cacheMap[namespace]
	if existed == false {
		return nil
	}
	return cs[prefix]
}

// Add - add a new cache into global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Add(namespace, prefix string, cache *lru.Cache) {
	globalDataNodeCache.lock.Lock()
	defer globalDataNodeCache.lock.Unlock()

	if cs, existed := globalDataNodeCache.cacheMap[namespace]; existed == false {
		tmp := make(map[string]*lru.Cache)
		tmp[prefix] = cache
		globalDataNodeCache.cacheMap[namespace] = tmp
	} else {
		cs[prefix] = cache
	}
}

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	c          *lru.Cache
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
	logger     *logging.Logger
}

func newDataNodeCache(treePrefix string, maxSizeMBs int, logger *logging.Logger, ns string) *DataNodeCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		logger.Infof("Constructing datanode-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	if globalDataNodeCache.isEnable {
		if c := globalDataNodeCache.Get(ns, treePrefix); c == nil {
			c, _ := lru.New(GlobalDataNodeCacheSize)
			globalDataNodeCache.Add(ns, treePrefix, c)
		} else {
			dataNodeCache := &DataNodeCache{TreePrefix: treePrefix, c: c, maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled, logger: logger}
			globalDataNodeCache.Add(ns, treePrefix, nil)
			return dataNodeCache
		}
	}
	cache, _ := lru.New(DefaultDataNodeCacheMaxSize)
	return &DataNodeCache{TreePrefix: treePrefix, c: cache, maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled, logger: logger}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(db db.Database, bucketKey BucketKey) (dataNodes DataNodes, err error) {
	// step 0.
	if dataNodeCache.isEnabled == false {
		return fetchDataNodesFromDBByBucketKey(db, dataNodeCache.TreePrefix, &bucketKey)
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
		cache := globalDataNodeCache.Get(db.Namespace(), dataNodeCache.TreePrefix)
		if cache == nil {
			cache, _ = lru.New(GlobalDataNodeCacheSize)
			globalDataNodeCache.Add(db.Namespace(), dataNodeCache.TreePrefix, cache)
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
	dataNodes, err = fetchDataNodesFromDBByBucketKey(db, dataNodeCache.TreePrefix, &bucketKey)
	if err != nil {
		dataNodeCache.logger.Error("fetchDataNodesFromDBByBucketKey Error")
		return dataNodes, err
	}
	if dataNodes == nil || len(dataNodes) == 0 {
		return dataNodes, nil
	} else {
		if dataNodeCache.isEnabled {
			dataNodeCache.c.Add(bucketKey, dataNodes)
		}
		if globalDataNodeCache.isEnable {
			if globalDataNodeCache.Get(db.Namespace(), dataNodeCache.TreePrefix) == nil {
				cache, _ := lru.New(GlobalDataNodeCacheSize)
				globalDataNodeCache.Add(db.Namespace(), dataNodeCache.TreePrefix, cache)
			}
			cache := globalDataNodeCache.Get(db.Namespace(), dataNodeCache.TreePrefix)
			cache.Add(bucketKey, dataNodes)
		}
		return dataNodes, nil
	}
}

func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.c, _ = lru.New(DefaultDataNodeCacheMaxSize)
}
