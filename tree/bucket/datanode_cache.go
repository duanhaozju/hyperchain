package bucket

import (
	"github.com/hashicorp/golang-lru"
	"hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"hyperchain/common"
)

var (
	IsEnabledGlobal = true
)

type DataNodeCache struct {
	TreePrefix string
	isEnabled  bool
	cache      *lru.Cache
}

func newDataNodeCache(log *logging.Logger, ns string, treePrefix string, dataNodeCacheMaxSize int) *DataNodeCache {
	isEnabled := true
	if dataNodeCacheMaxSize <= 0 {
		isEnabled = false
		log.Error("dataNodeCacheMaxSize is ",dataNodeCacheMaxSize)
		return &DataNodeCache{TreePrefix: treePrefix, isEnabled: isEnabled}
	} else {
		log.Infof("Constructing datanode-cache with max datanode cache size = [%d]", dataNodeCacheMaxSize)
	}

	if IsEnabledGlobal {
		c := globalDataNodeCache.Get(ConstructPrefix(ns, treePrefix))
		if c == nil {
			c, _ = lru.New(globalDataNodeCache.globalDataNodeCacheMaxSize)
			globalDataNodeCache.Add(ConstructPrefix(ns, treePrefix), c)
		}
	}
	cache, _ := lru.New(dataNodeCacheMaxSize)
	return &DataNodeCache{TreePrefix: treePrefix, cache: cache, isEnabled: isEnabled}
}

func (dataNodeCache *DataNodeCache) FetchDataNodesFromCache(db db.Database, bucketKey BucketKey) (dataNodes DataNodes, err error) {
	// step 0.
	log := common.GetLogger(db.Namespace(), "bucket")
	if !dataNodeCache.isEnabled{
		return fetchDataNodesFromDBByBucketKey(db, dataNodeCache.TreePrefix, &bucketKey)
	}

	// step 1.
	value, ok := dataNodeCache.cache.Get(bucketKey)
	if ok {
		dataNodes = value.(DataNodes)
	}

	if dataNodes != nil && len(dataNodes) > 0 {
		return dataNodes, nil
	}

	// step 2.
	if IsEnabledGlobal {
		cache := globalDataNodeCache.Get(ConstructPrefix(db.Namespace(), dataNodeCache.TreePrefix))
		if cache == nil {
			cache, _ = lru.New(globalDataNodeCache.globalDataNodeCacheMaxSize)
			globalDataNodeCache.Add(ConstructPrefix(db.Namespace(), dataNodeCache.TreePrefix), cache)
		}
		value, ok = cache.Get(bucketKey)
		if ok {
			dataNodes = value.(DataNodes)
		}
		if dataNodes != nil && len(dataNodes) > 0 {
			if dataNodeCache.isEnabled {
				dataNodeCache.cache.Add(bucketKey, dataNodes)
			}
			return dataNodes, nil
		}
	}

	// step 3.
	dataNodes, err = fetchDataNodesFromDBByBucketKey(db, dataNodeCache.TreePrefix, &bucketKey)
	if err != nil {
		log.Error("fetchDataNodesFromDBByBucketKey Error")
		return dataNodes, err
	}
	if dataNodes == nil || len(dataNodes) == 0 {
		return dataNodes, nil
	} else {
		if dataNodeCache.isEnabled {
			dataNodeCache.cache.Add(bucketKey, dataNodes)
		}
		if IsEnabledGlobal {
			cache := globalDataNodeCache.Get(ConstructPrefix(db.Namespace(), dataNodeCache.TreePrefix))
			if cache == nil {
				cache, _ = lru.New(globalDataNodeCache.globalDataNodeCacheMaxSize)
				globalDataNodeCache.Add(ConstructPrefix(db.Namespace(), dataNodeCache.TreePrefix), cache)
			}
			cache.Add(bucketKey, dataNodes)
		}
		return dataNodes, nil
	}
}

func (dataNodeCache *DataNodeCache) ClearDataNodeCache() {
	dataNodeCache.cache.Purge()
}
