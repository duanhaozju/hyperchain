package bucket

import (
	"github.com/hashicorp/golang-lru"
	"sync"
)

var (
	globalDataNodeCache *GlobalDataNodeCache
	globalOnce          sync.Once
)

// Just Initialize Once
func NewGlobalDataNodeCache(globalDataNodeCacheLength, globalDataNodeCacheSize int) {
	if globalDataNodeCache != nil {
		return
	}
	if globalDataNodeCacheSize <= 0 || globalDataNodeCacheLength <= 0 {
		IsEnabledGlobal = false
		globalDataNodeCache = &GlobalDataNodeCache{
			isEnable:                     IsEnabledGlobal,
			globalDataNodeCacheMaxLength: globalDataNodeCacheLength,
			globalDataNodeCacheMaxSize:   globalDataNodeCacheSize,
		}
	} else {
		globalLRUCache, _ := lru.New(globalDataNodeCacheLength)
		globalDataNodeCache = &GlobalDataNodeCache{
			globalLRUCache:               globalLRUCache,
			isEnable:                     IsEnabledGlobal,
			globalDataNodeCacheMaxLength: globalDataNodeCacheLength,
			globalDataNodeCacheMaxSize:   globalDataNodeCacheSize,
		}
	}
}

type GlobalDataNodeCache struct {
	globalLRUCache               *lru.Cache
	isEnable                     bool
	lock                         sync.RWMutex
	globalDataNodeCacheMaxSize   int
	globalDataNodeCacheMaxLength int
}

func (globalDataNodeCache *GlobalDataNodeCache) ClearAllCache() {
	globalDataNodeCache.globalLRUCache.Purge()
}

// Get - get a cache from global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Get(prefix string) *lru.Cache {
	globalDataNodeCache.lock.RLock()
	defer globalDataNodeCache.lock.RUnlock()
	cache, ok := globalDataNodeCache.globalLRUCache.Get(prefix)
	if ok {
		return cache.(*lru.Cache)
	}
	return nil
}

// Add - add a new cache into global cache thread safely.
func (globalDataNodeCache *GlobalDataNodeCache) Add(prefix string, cache *lru.Cache) {
	globalDataNodeCache.lock.Lock()
	defer globalDataNodeCache.lock.Unlock()
	globalDataNodeCache.globalLRUCache.Add(prefix, cache)
}

func ConstructPrefix(ns string, prefix string) string {
	return ns + prefix
}
