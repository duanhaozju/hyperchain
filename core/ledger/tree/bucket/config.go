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
	"github.com/op/go-logging"
	"hash/fnv"
	"sync"
)

const (
	// The capacity of the underlying hash table.
	// The greater the hash table capacity, the greater the storage capacity of the tree,
	// but will cause the tree height to increase, the number of merkle nodes increased.
	hashTableCapcity = "capacity"

	// The number of hash table bucket gathers into a Merkle node.
	aggreation = "aggreation"

	// The capacity of a cache which used to store hash node in memory.
	bucketCacheSize = "bucketCache"

	// The capacity of a cache which used to store merkle node in memory.
	merkleNodeCacheSize = "merkleNodeCache"

	// The type of hash function used to calculate index.
	entryHashFunc = "hashFunc"
)

const (
	defaultHashTableCap        = 10009
	defaultAggreation          = 10
	defaultBucketCacheSize     = 10000
	defaultMerkleNodeCacheSize = 10000
)

var (
	conf *config // TODO remove the global config instance
	once sync.Once
)

type config struct {
	aggreation      int
	lowest          int
	levelInfo       map[int]int
	hashFunc        hashFunc
	nodeCacheSize   int
	bucketCacheSize int
}

func initConfig(log *logging.Logger, configs map[string]interface{}) *config {
	cap, ok := configs[hashTableCapcity].(int)
	if !ok {
		cap = defaultHashTableCap
	}
	aggr, ok := configs[aggreation].(int)
	if !ok {
		aggr = defaultAggreation
	}
	hashFn, ok := configs[entryHashFunc].(hashFunc)
	if !ok {
		hashFn = fnvHash
	}
	merkleNodeCache, ok := configs[merkleNodeCacheSize].(int)
	if !ok {
		merkleNodeCache = defaultMerkleNodeCacheSize
	}
	bucketCache, ok := configs[bucketCacheSize].(int)
	if !ok {
		bucketCache = defaultBucketCacheSize
	}
	once.Do(func() {
		conf = newConfig(cap, aggr, hashFn, merkleNodeCache, bucketCache)
	})
	log.Infof("Initializing tree with configurations %+v", conf)
	return conf
}

func newConfig(cap, aggr int, hashFunc hashFunc, merkleNodeCacheSize, bucketCacheSize int) *config {
	conf := &config{aggr, -1, make(map[int]int), hashFunc,
		merkleNodeCacheSize, bucketCacheSize}
	var (
		curlevel  int
		curSize   int = cap
		levelInfo     = make(map[int]int)
	)
	levelInfo[curlevel] = curSize
	for curSize > 1 {
		parSize := curSize / aggr
		if curSize%aggr != 0 {
			parSize++
		}
		curSize = parSize
		curlevel++
		levelInfo[curlevel] = curSize
	}
	conf.lowest = curlevel
	for k, v := range levelInfo {
		conf.levelInfo[conf.lowest-k] = v
	}
	return conf
}

// getMaxIndex returns max index at the given level.
func (config *config) getMaxIndex(level int) int {
	if level < 0 || level > config.lowest {
		panic(fmt.Errorf("level can only be between 0 and [%d]", config.lowest))
	}
	return config.levelInfo[level]
}

// computeIndex returns index calculated with given data.
func (config *config) computeIndex(data []byte) uint32 {
	return config.hashFunc(data)
}

// getLowestLevel returns the lowest level number.
func (config *config) getLowestLevel() int {
	return config.lowest
}

// getAggreation returns aggreation.
func (config *config) getAggreation() int {
	return config.aggreation
}

// getCapacity returns tree's capacity.
func (config *config) getCapacity() int {
	return config.getMaxIndex(config.getLowestLevel())
}

func (config *config) computeParentIndex(bucketNumber int) int {
	parentBucketNumber := bucketNumber / config.getAggreation()
	if bucketNumber%config.getAggreation() != 0 {
		parentBucketNumber++
	}
	return parentBucketNumber
}

type hashFunc func(data []byte) uint32

// fnvHash returns a index by hash the given content.
func fnvHash(data []byte) uint32 {
	fnvHash := fnv.New32a()
	fnvHash.Write(data)
	return fnvHash.Sum32()
}
