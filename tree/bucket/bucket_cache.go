package bucket

import (
	"github.com/hashicorp/golang-lru"
	"hyperchain/hyperdb/db"
)


// We can create a cache and keep all the bucket nodes pre-loaded.
// Since, the bucket nodes do not contain actual data and max possible
// buckets are pre-determined, the memory demand may not be very high or can easily
// be controlled - by keeping seletive buckets in the cache (most likely first few levels of the bucket tree - because,
// higher the level of the bucket, more are the chances that the bucket would be required for recomputation of hash)
type BucketCache struct {
	TreePrefix string
	isEnabled  bool
	cache      *lru.Cache
}

func newBucketCache(treePrefix string, maxBucketCacheSize int) *BucketCache {
	isEnabled := true
	//log.Noticef("BucketCache Size is [%d]",maxBucketCacheSize)
	if maxBucketCacheSize <= 0 {
		isEnabled = false
		return &BucketCache{TreePrefix: treePrefix, isEnabled: isEnabled}
	} else {
		log.Infof("Constructing bucket-cache with max bucket cache size = [%d]", maxBucketCacheSize)
		cache,_ := lru.New(maxBucketCacheSize)
		return &BucketCache{TreePrefix: treePrefix, cache: cache, isEnabled: isEnabled}
	}
}

func (bucketCache *BucketCache) clearAllCache(){
	bucketCache.cache.Purge()
}

// TODO cache will be done later
func (bucketCache *BucketCache) loadAllBucketNodesFromDB() {
	if !bucketCache.isEnabled {
		return
	}
}

func (bucketCache *BucketCache) put(key BucketKey, node *BucketNode) {
	if !bucketCache.isEnabled {
		return
	}
	node.markedForDeletion = false
	node.childrenUpdated = nil
	bucketCache.cache.Add(key,node)
}

// TODO performance status should be done
func (bucketCache *BucketCache) fetchBucketNodeFromCache(db db.Database, key BucketKey) (*BucketNode, error) {
	if !bucketCache.isEnabled {
		return fetchBucketNodeFromDB(db, bucketCache.TreePrefix, &key)
	}
	bucketNodeValue,ok := bucketCache.cache.Get(key)
	if ok {
		return bucketNodeValue.(*BucketNode),nil
	}
	bucketNode,err := fetchBucketNodeFromDB(db, bucketCache.TreePrefix, &key)
	if err != nil{
		return nil,err
	}else {
		//cache.c.Add(key,bucketNode)
		return bucketNode,nil
	}

}

func (bucketCache *BucketCache) remove(key BucketKey) {
	if !bucketCache.isEnabled {
		return
	}
	bucketCache.cache.Remove(key)
}

