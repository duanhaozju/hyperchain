package bucket

import (
	"sync"
	//"time"
	"unsafe"
	"hyperchain/hyperdb/db"
	"github.com/op/go-logging"
)
const (
	HASHLEN = 20
)
var defaultBucketCacheMaxSize = 100 // MBs

// We can create a cache and keep all the bucket nodes pre-loaded.
// Since, the bucket nodes do not contain actual data and max possible
// buckets are pre-determined, the memory demand may not be very high or can easily
// be controlled - by keeping seletive buckets in the cache (most likely first few levels of the bucket tree - because,
// higher the level of the bucket, more are the chances that the bucket would be required for recomputation of hash)
type BucketCache struct {
	TreePrefix string
	isEnabled  bool
	c          map[BucketKey]*BucketNode
	lock       sync.RWMutex
	size       uint64
	maxSize    uint64
	logger     *logging.Logger
}

func newBucketCache(treePrefix string, maxSizeMBs int, logger *logging.Logger) *BucketCache {
	isEnabled := true
	if maxSizeMBs <= 0 {
		isEnabled = false
	} else {
		logger.Infof("Constructing bucket-cache with max bucket cache size = [%d] MBs", maxSizeMBs)
	}
	return &BucketCache{TreePrefix: treePrefix, c: make(map[BucketKey]*BucketNode), maxSize: uint64(maxSizeMBs * 1024 * 1024), isEnabled: isEnabled, logger: logger}
}

func (cache *BucketCache) clearAllCache() *BucketCache {
	isEnabled := true
	if cache.isEnabled {
	} else {
		cache.logger.Infof("Constructing bucket-cache with max bucket cache size = [%d] MBs", cache.maxSize)
	}
	tmp := &BucketCache{TreePrefix: cache.TreePrefix, c: make(map[BucketKey]*BucketNode), maxSize: uint64(cache.maxSize), isEnabled: isEnabled}
	return tmp
}

// TODO cache will be done later
func (cache *BucketCache) loadAllBucketNodesFromDB() {
	if !cache.isEnabled {
		return
	}
}

func (cache *BucketCache) putWithoutLock(key BucketKey, node *BucketNode) {
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
func (cache *BucketCache) get(db db.Database, key BucketKey) (*BucketNode, error) {
	//defer perfstat.UpdateTimeStat("timeSpent", time.Now())
	if !cache.isEnabled {
		return fetchBucketNodeFromDB(db, cache.TreePrefix, &key)
	}
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	bucketNode := cache.c[key]
	if bucketNode == nil {
		return fetchBucketNodeFromDB(db, cache.TreePrefix, &key)
	}
	return bucketNode, nil
}

func (cache *BucketCache) removeWithoutLock(key BucketKey) {
	if !cache.isEnabled {
		return
	}
	node, ok := cache.c[key]
	if ok {
		cache.size -= (key.size() + node.size())
		delete(cache.c, key)
	}
}

func (bk BucketKey) size() uint64 {
	return uint64(unsafe.Sizeof(bk))
}

func (bNode *BucketNode) size() uint64 {
	size := uint64(unsafe.Sizeof(*bNode))
	length := len(bNode.childrenCryptoHash)
	if length > 0 {
		size += uint64(length * HASHLEN)
	}
	return size
}
