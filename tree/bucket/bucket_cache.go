package bucket

import (
	"sync"
	//"time"
	"unsafe"
)

var defaultBucketCacheMaxSize = 100 // MBs

// We can create a cache and keep all the bucket nodes pre-loaded.
// Since, the bucket nodes do not contain actual data and max possible
// buckets are pre-determined, the memory demand may not be very high or can easily
// be controlled - by keeping seletive buckets in the cache (most likely first few levels of the bucket tree - because,
// higher the level of the bucket, more are the chances that the bucket would be required for recomputation of hash)
type bucketCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[bucketKey]*bucketNode
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
}

func newBucketCache(treePrefix string,maxSizeMBs int) *bucketCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		logger.Infof("Constructing bucket-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	return &bucketCache{TreePrefix: treePrefix,c: make(map[bucketKey]*bucketNode), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled}
}

func (cache *bucketCache) clearAllCache(){
	isEnabled := true
	if cache.isEnabled{
	} else {
		logger.Infof("Constructing bucket-cache with max bucket cache size = [%d] MBs", cache.maxSize)
	}
	cache = &bucketCache{TreePrefix: cache.TreePrefix,c: make(map[bucketKey]*bucketNode), maxSize: uint64(cache.maxSize * 1024 * 1024), isEnabled: isEnabled}
}

// TODO cache will be done later
func (cache *bucketCache) loadAllBucketNodesFromDB() {
	if !cache.isEnabled {
		return
	}
}

func (cache *bucketCache) putWithoutLock(key bucketKey, node *bucketNode) {
	if !cache.isEnabled {
		return
	}
	node.markedForDeletion = false
	node.childrenUpdated = nil
	existingNode, ok := cache.c[key]
	size := uint64(0)
	if ok {
		size = node.size() - existingNode.size()
		cache.size += size
		if cache.size > cache.maxSize {
			delete(cache.c, key)
			cache.size -= (key.size() + existingNode.size())
		} else {
			cache.c[key] = node
		}
	} else {
		size = node.size()
		cache.size += size
		if cache.size > cache.maxSize {
			return
		}
		cache.c[key] = node
	}
}
// TODO performance status should be done
func (cache *bucketCache) get(key bucketKey) (*bucketNode, error) {
	//defer perfstat.UpdateTimeStat("timeSpent", time.Now())
	if !cache.isEnabled {
		return fetchBucketNodeFromDB(cache.TreePrefix,&key)
	}
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	bucketNode := cache.c[key]
	if bucketNode == nil {
		return fetchBucketNodeFromDB(cache.TreePrefix,&key)
	}
	return bucketNode, nil
}

func (cache *bucketCache) removeWithoutLock(key bucketKey) {
	if !cache.isEnabled {
		return
	}
	node, ok := cache.c[key]
	if ok {
		cache.size -= (key.size() + node.size())
		delete(cache.c, key)
	}
}

func (bk bucketKey) size() uint64 {
	return uint64(unsafe.Sizeof(bk))
}

func (bNode *bucketNode) size() uint64 {
	size := uint64(unsafe.Sizeof(*bNode))
	numChildHashes := len(bNode.childrenCryptoHash)
	if numChildHashes > 0 {
		size += uint64(numChildHashes * len(bNode.childrenCryptoHash[0]))
	}
	return size
}
