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
	"errors"
	"github.com/hashicorp/golang-lru"
	"github.com/op/go-logging"
	"hyperchain/common"
	tp "hyperchain/common/threadpool"
	hdb "hyperchain/hyperdb/db"
	"runtime"
	"sync"
)

type Entries map[string][]byte

func NewEntries() Entries {
	ret := make(map[string][]byte)
	return ret
}

// BucketTree a tree structure that combines a hash table with a merkle tree,
// used to quickly calculate the status hash.
type BucketTree struct {
	db        hdb.Database     // database handler
	prefix    string           // tree prefix
	buckets   *Buckets         // dirty hash buckets collection
	nodeDelta *MerkleNodeDelta // dirty merkle node collection
	hash      []byte           // latest computed hash
	dirty     bool             // dirty flag, used to indicates whether hash is needed to recompute
	mcache    *merkleNodeCache // merkle node cache
	bcache    *bucketCache     // bucket cache
	log       *logging.Logger  // logger
}

func NewBucketTree(db hdb.Database, prefix string) *BucketTree {
	return &BucketTree{
		prefix: prefix,
		db:     db,
		log:    common.GetLogger(db.Namespace(), "buckettree"),
	}
}

// Initialize sets up bucket tree with given configuration.
// The configuration parameters include:
//   (1) the number of hash bucket
//   (2) the intensity of aggreation
//   (3) the size of cache
func (tree *BucketTree) Initialize(configs map[string]interface{}) error {
	initConfig(tree.log, configs)
	if err := tree.loadLatestHash(); err != nil {
		return err
	}
	tree.mcache = newMerkleNodeCache(tree.db, tree.log, tree.prefix, conf.nodeCacheSize)
	tree.bcache = newBucketCache(tree.db, tree.log, tree.prefix, conf.bucketCacheSize)
	return nil
}

// Prepare init the dirty bucket collection with an incoming set of entires.
func (bucketTree *BucketTree) Prepare(entries Entries) error {
	// Short circuit if entries is empty.
	if entries == nil || len(entries) == 0 {
		return nil
	}
	bucketTree.buckets = newBuckets(bucketTree.prefix, entries)
	bucketTree.nodeDelta = newMerkleNodeDelta(bucketTree.log)
	bucketTree.dirty = true
	return nil
}

// Process calculates the new hash for tree.
// During the calculation:
// (1) calculates all dirty buckets and set the new hash to its parent children list.
// (2) calculates all involved merkle nodes and set the new hash to its parent children list.
// (3) iter to the root, and calculates root node, regards it as the tree's hash.
func (bucketTree *BucketTree) Process() ([]byte, error) {
	bucketTree.log.Debug("begin to calculate tree hash")
	if bucketTree.dirty {
		bucketTree.log.Debug("dirty tree, begin to re-calculate")
		err := bucketTree.processBuckets()
		if err != nil {
			return nil, err
		}
		err = bucketTree.processNodes()
		if err != nil {
			return nil, err
		}
		bucketTree.hash = bucketTree.computeRootNodeCryptoHash()
		bucketTree.dirty = false
	} else {
		bucketTree.log.Debugf("clear tree, return cached hash %s directly", common.Bytes2Hex(bucketTree.hash))
	}
	return bucketTree.hash, nil
}

// processBuckets recompute all dirty buckets.
// The bucket is a collection of data entry, which consists of a key-valu pair.
// Dirty bucket is merged with the original bucket and make all data entries sorted by lexicographical order.
func (bucketTree *BucketTree) processBuckets() error {
	var (
		wg             sync.WaitGroup
		dirtyBucketNum = bucketTree.buckets.getAllPosition()
	)
	wg.Add(len(dirtyBucketNum))

	handler := func(obj interface{}) interface{} {
		defer wg.Done()
		pos, ok := obj.(*Position)
		if !ok {
			return errors.New("expect bucket key")
		}

		modifiedBucket := bucketTree.buckets.get(pos)
		oldBucket, err := bucketTree.bcache.get(bucketTree.db, *pos)
		if err != nil {
			bucketTree.log.Errorf("fetch original bucket failed. %s", err.Error())
			return err
		}
		newBucket := oldBucket.merge(modifiedBucket)
		cryptoHash := newBucket.computeCryptoHash()
		bucketTree.updateBucketCache(*pos, newBucket)
		parentBucket := bucketTree.nodeDelta.getOrCreate(pos.getParent())
		parentBucket.setChild(pos, cryptoHash)
		return nil
	}
	// Run all hash calculation processes via threadpool.
	// The reason is that each execution goroutine may involve IO operations.
	// If so, the goroutine(G) will block the whole thread(M).
	// If there is no more thread(M) available, system will create new thread and assign it
	// with new goroutine. This situation can make memory usage increase rapidly.
	pool, _ := tp.CreatePool(runtime.NumCPU()+1, handler).Open()
	defer pool.Close()

	for _, bucketKey := range dirtyBucketNum {
		pool.SendWorkAsync(bucketKey, nil)
	}
	wg.Wait()
	return nil
}

// processNodes recompute all dirty merkle nodes.
// Dirty merkle nodes will merged with original merkle node,
// Hash value is derived from merkle node's children list.
func (bucketTree *BucketTree) processNodes() error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		nodes := bucketTree.nodeDelta.getLevel(level)
		var wg sync.WaitGroup
		wg.Add(len(nodes))

		handler := func(obj interface{}) interface{} {
			defer wg.Done()
			node, ok := obj.(*MerkleNode)
			if !ok {
				return errors.New("expect to be merkle node")
			}
			oldNode, err := bucketTree.mcache.get(bucketTree.db, *node.pos)
			if err != nil {
				bucketTree.log.Errorf("get bucketnode from cache failed. %s", err.Error())
				return err
			}
			if oldNode != nil {
				node.merge(oldNode)
			}
			if level == 0 {
				return nil
			}
			cryptoHash := node.computeCryptoHash()
			parent := bucketTree.nodeDelta.getOrCreate(node.pos.getParent())
			parent.setChild(node.pos, cryptoHash)
			return nil
		}
		// The same with processBuckets function.
		pool, _ := tp.CreatePool(runtime.NumCPU()+1, handler).Open()

		for _, node := range nodes {
			pool.SendWorkAsync(node, nil)
		}
		wg.Wait()
		pool.Close()
	}
	return nil
}

// computeRootNodeCryptoHash calculate root node's hash.
func (bucketTree *BucketTree) computeRootNodeCryptoHash() []byte {
	return bucketTree.nodeDelta.getRoot().computeCryptoHash()
}

// Commit commits all modified content to db write batch.
func (bucketTree *BucketTree) Commit(dbw hdb.Batch) (err error) {
	// Short circult if the dirty buckets is empty.
	if bucketTree.buckets == nil {
		return nil
	}
	if bucketTree.dirty {
		_, err := bucketTree.Process()
		if err != nil {
			return nil
		}
	}
	// commits all dirty buckets
	err = bucketTree.commitBuckets(dbw)
	if err != nil {
		return err
	}
	// commits all dirty merkle nodes
	err = bucketTree.commitNodes(dbw)
	if err != nil {
		return err
	}
	bucketTree.updateNodeCache()
	bucketTree.clearStatus()
	return nil
}

func (bucketTree *BucketTree) commitBuckets(dbw hdb.Batch) error {
	dirtyPos := bucketTree.buckets.getAllPosition()
	for _, pos := range dirtyPos {
		value, ok := bucketTree.bcache.cache.Get(*pos)
		if ok {
			bucket := value.(Bucket)
			if err := PersistBucket(bucketTree.prefix, bucket, pos, dbw); err != nil {
				bucketTree.log.Errorf("commit buckets for %s failed, err msg %s", pos.String(), err.Error())
				return err
			}
		}
	}
	return nil
}

func (bucketTree *BucketTree) commitNodes(dbw hdb.Batch) error {
	secondLastLevel := conf.getLowestLevel() - 1
	for level := secondLastLevel; level >= 0; level-- {
		nodes := bucketTree.nodeDelta.getLevel(level)
		for _, node := range nodes {
			if err := PersistMerkleNode(bucketTree.prefix, node, dbw); err != nil {
				bucketTree.log.Errorf("commit merkle node for %s failed, err msg %s", node.pos.String(), err.Error())
				return err
			}
		}
	}
	return nil
}

func (bucketTree *BucketTree) clearStatus() {
	bucketTree.buckets = nil
	bucketTree.nodeDelta = nil
	bucketTree.dirty = false
}

func (bucketTree *BucketTree) updateBucketCache(pos Position, bucket Bucket) {
	// Short circuit if buckets is empty.
	if bucketTree.buckets == nil {
		return
	}
	// Update to bcache.
	if bucketTree.bcache.enable {
		bucketTree.bcache.cache.Add(pos, bucket)
	}
	// Update to global cache.
	if globalBucketCache.enable {
		cache := globalBucketCache.Get(ConstructPrefix(bucketTree.db.Namespace(), bucketTree.prefix))
		if cache == nil {
			cache, _ = lru.New(bucketTree.bcache.maxSize)
		}
		cache.Add(pos, bucket)
		globalBucketCache.Add(ConstructPrefix(bucketTree.db.Namespace(), bucketTree.prefix), cache)
	}
}

func (bucketTree *BucketTree) updateNodeCache() {
	// Short circuit if node delta is empty.
	if bucketTree.nodeDelta == nil || bucketTree.nodeDelta.empty() {
		return
	}
	secondLastLevel := conf.getLowestLevel() - 1
	for level := 0; level <= secondLastLevel; level++ {
		nodes := bucketTree.nodeDelta.getLevel(level)
		for _, node := range nodes {
			key := *node.pos
			if node.deleted {
				bucketTree.mcache.remove(key)
			} else {
				bucketTree.mcache.put(key, node)
			}
		}
	}
}

// Clear clears all cached stuffs.
func (bucketTree *BucketTree) Clear() {
	bucketTree.bcache.clear()
	bucketTree.mcache.clear()
	globalBucketCache.clear()
	bucketTree.dirty = false
	bucketTree.loadLatestHash()
}

// loadLatestHash load root node from database and record its hash value.
func (bucketTree *BucketTree) loadLatestHash() error {
	rootBucketNode, err := RetrieveMerkleNode(bucketTree.log, bucketTree.db, bucketTree.prefix, rootPosition())
	if err != nil {
		return err
	}
	if rootBucketNode != nil {
		bucketTree.hash = rootBucketNode.computeCryptoHash()
	}
	return nil
}
