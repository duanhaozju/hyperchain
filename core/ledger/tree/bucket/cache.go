// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bucket

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"sync"
)

const (
	globalEnable          = true
	GlobalBucketCacheSize = 20
)

var (
	globalBucketCache *GlobalBucketCache
	globalOnce        sync.Once
)

func init() {
	InitGlobalBucketCache()
}

// BucketCache is used to store bucket data in memory.
// There are two reason to use BucketCache:
// 1. Since the validation and commit are seperate in hyperchain,
//    so the validation result should be saved in the cache to prevent write miss.
// 2. It will cost a lot if we get each hash node from database all the time,
//    store most-frequent hash node in cache is helpful to reduce IO pressure.
type bucketCache struct {
	prefix  string // tree identification
	enable  bool   // flag to represent the cache is enabled or not
	cache   *lru.Cache
	log     *logging.Logger
	db      db.Database
	maxSize int
}

// newBucketCache returns a bucket cache initialized by given config params.
func newBucketCache(db db.Database, log *logging.Logger, prefix string, maxSize int) *bucketCache {
	var c *lru.Cache
	if maxSize <= 0 {
		log.Notice("negative size for bucket cache, disable it")
		return &bucketCache{prefix: prefix, enable: false, log: log, db: db, maxSize: maxSize}
	}
	log.Infof("create bucket cache with max size = [%d]", maxSize)
	if globalEnable {
		// If a cache is saved in global with same tree id, use if directly.
		c = globalBucketCache.Get(ConstructPrefix(db.Namespace(), prefix))
		if c == nil {
			c, _ = lru.New(maxSize)
		}
	}
	return &bucketCache{prefix: prefix, cache: c, enable: true, log: log, db: db, maxSize: maxSize}
}

// get returns all entries stored in specified bucket.
// If cache miss hit, it will search in database instead.
func (bcache *bucketCache) get(db db.Database, pos Position) (bucket Bucket, err error) {
	// Retrieve data entries from database if cache is disable.
	if !bcache.enable {
		return RetrieveBucket(bcache.log, db, bcache.prefix, &pos)
	}

	// Retrieve data entries from cache. If hits, returns the result directly.
	value, ok := bcache.cache.Get(pos)
	if ok {
		bucket = value.(Bucket)
	}
	// Short circuit if cache hit.
	if bucket != nil && len(bucket) > 0 {
		return bucket, nil
	}

	// Retrieve data entries from global entries cache.
	if globalEnable {
		c := globalBucketCache.Get(ConstructPrefix(db.Namespace(), bcache.prefix))
		if c == nil {
			c, _ = lru.New(bcache.maxSize)
			globalBucketCache.Add(ConstructPrefix(db.Namespace(), bcache.prefix), c)
		} else {
			// Hit in global cache.
			value, ok = c.Get(pos)
			if ok {
				bucket = value.(Bucket)
			}
			if bucket != nil && len(bucket) > 0 {
				if bcache.enable {
					bcache.cache.Add(pos, bucket)
				}
				return bucket, nil
			}
		}
	}

	// Retrieve entries from database if non-hit in all caches.
	bucket, err = RetrieveBucket(bcache.log, db, bcache.prefix, &pos)
	if err != nil {
		bcache.log.Errorf("retrieve data entries from database failed, %s", err)
		return bucket, err
	}
	if bucket == nil || len(bucket) == 0 {
		return bucket, nil
	} else {
		// Update in cache.
		if bcache.enable {
			bcache.cache.Add(pos, bucket)
		}
		// Update in global cache
		if globalEnable {
			cache := globalBucketCache.Get(ConstructPrefix(db.Namespace(), bcache.prefix))
			if cache == nil {
				cache, _ = lru.New(bcache.maxSize)
				globalBucketCache.Add(ConstructPrefix(db.Namespace(), bcache.prefix), cache)
			}
			cache.Add(pos, bucket)
		}
		return bucket, nil
	}
}

// clear purge all stuff in cache.
func (cache *bucketCache) clear() {
	cache.cache.Purge()
}

// We can create a cache and keep all the merkle nodes pre-loaded.
// Since, the merkle nodes do not contain actual data and max possible
// merkles are pre-determined, the memory demand may not be very high or can easily
// be controlled - by keeping seletive merkles in the cache (most likely first few levels of the hash tree - because,
// higher the level of the bucket, more are the chances that the bucket would be required for recomputation of hash)
type merkleNodeCache struct {
	prefix string
	enable bool
	cache  *lru.Cache
	log    *logging.Logger
	db     db.Database
}

// TODO use a constant size to save all merkle nodes.
func newMerkleNodeCache(db db.Database, log *logging.Logger, prefix string, size int) *merkleNodeCache {
	if size <= 0 {
		return &merkleNodeCache{db: db, log: log, prefix: prefix, enable: false}
	} else {
		cache, _ := lru.New(size)
		return &merkleNodeCache{db: db, log: log, prefix: prefix, cache: cache, enable: true}
	}
}

// clear purge all stuff in cache.
func (cache *merkleNodeCache) clear() {
	cache.cache.Purge()
}

func (cache *merkleNodeCache) put(key Position, node *MerkleNode) {
	// Short circuit if cache is disable.
	if !cache.enable {
		return
	}
	node.deleted = false
	node.dirty = nil
	cache.cache.Add(key, node)
}

func (cache *merkleNodeCache) get(db db.Database, key Position) (*MerkleNode, error) {
	if !cache.enable {
		fmt.Println("retrieve")
		return RetrieveMerkleNode(cache.log, db, cache.prefix, &key)
	}
	v, ok := cache.cache.Get(key)
	if ok {
		return v.(*MerkleNode), nil
	}
	node, err := RetrieveMerkleNode(cache.log, db, cache.prefix, &key)
	if err != nil {
		return nil, err
	} else {
		return node, nil
	}
}

func (cache *merkleNodeCache) remove(key Position) {
	if !cache.enable {
		return
	}
	cache.cache.Remove(key)
}

// GlobalHashNodeCache
type GlobalBucketCache struct {
	cache    *lru.Cache
	enable   bool
	lock     sync.RWMutex
	cacheNum int
}

// InitGlobalHashNodeCache initializes a global cache to store all tree level cache.
func InitGlobalBucketCache() {
	globalOnce.Do(func() {
		cache, _ := lru.New(GlobalBucketCacheSize)
		globalBucketCache = &GlobalBucketCache{
			cache:    cache,
			enable:   globalEnable,
			cacheNum: GlobalBucketCacheSize,
		}
	})
}

func (cache *GlobalBucketCache) clear() {
	cache.cache.Purge()
}

// Get gets a cache from global cache thread safely.
func (cache *GlobalBucketCache) Get(prefix string) *lru.Cache {
	cache.lock.RLock()
	defer cache.lock.RUnlock()
	c, ok := cache.cache.Get(prefix)
	if ok {
		return c.(*lru.Cache)
	}
	return nil
}

// Add adds a new cache into global cache thread safely.
func (cache *GlobalBucketCache) Add(prefix string, c *lru.Cache) {
	cache.lock.Lock()
	defer cache.lock.Unlock()
	cache.cache.Add(prefix, c)
}

func ConstructPrefix(ns string, prefix string) string {
	return ns + prefix
}
