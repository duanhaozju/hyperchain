package bucket

import (
	"fmt"
	"github.com/op/go-logging"
	"hash/fnv"
)

// ConfigNumBuckets - config name 'numBuckets' as it appears in yaml file
const ConfigNumBuckets = "numBuckets"

// ConfigMaxGroupingAtEachLevel - config name 'maxGroupingAtEachLevel' as it appears in yaml file
const ConfigMaxGroupingAtEachLevel = "maxGroupingAtEachLevel"

// ConfigHashFunction - config name 'hashFunction'. This is not exposed in yaml file. This configuration is used for testing with custom hash-function
const ConfigHashFunction = "hashFunction"

// ConfigBucketCacheMaxSize - config name 'bucketCacheMaxSize' as it appears in yaml file
const ConfigBucketCacheMaxSize = "bucketCacheMaxSize"

// ConfigDataNodeCacheMaxSize - config name 'dataNodeCacheMaxSize' as it appears in yaml file
const ConfigDataNodeCacheMaxSize = "dataNodeCacheMaxSize"

// DefaultNumBuckets - total buckets
const DefaultNumBuckets = 19

// DefaultMaxGroupingAtEachLevel - Number of max buckets to group at each level.
// Grouping is started from left. The last group may have less buckets
const DefaultMaxGroupingAtEachLevel = 10

// DefaultBucketCacheMaxSize - the cache of bucket
const DefaultBucketCacheMaxSize = 20

// DefaultDataNodeCacheMaxSize - the cache of dataNodes
const DefaultDataNodesCacheMaxSize = 40

var conf *config

type config struct {
	maxGroupingAtEachLevel int
	lowestLevel            int
	levelToNumBucketsMap   map[int]int
	hashFunc               hashFunc
	bucketCacheMaxSize     int
	dataNodeCacheMaxSize   int
}

func initConfig(log *logging.Logger, configs map[string]interface{}) {
	log.Infof("configs passed during initialization = %#v", configs)

	numBuckets, ok := configs[ConfigNumBuckets].(int)
	if !ok {
		numBuckets = DefaultNumBuckets
	}

	maxGroupingAtEachLevel, ok := configs[ConfigMaxGroupingAtEachLevel].(int)
	if !ok {
		maxGroupingAtEachLevel = DefaultMaxGroupingAtEachLevel
	}

	hashFunction, ok := configs[ConfigHashFunction].(hashFunc)
	if !ok {
		hashFunction = fnvHash
	}

	bucketCacheMaxSize, ok := configs[ConfigBucketCacheMaxSize].(int)
	if !ok {
		bucketCacheMaxSize = DefaultBucketCacheMaxSize
	}

	dataNodeCacheMaxSize, ok := configs[ConfigDataNodeCacheMaxSize].(int)
	if !ok {
		dataNodeCacheMaxSize = DefaultDataNodesCacheMaxSize
	}

	conf = newConfig(numBuckets, maxGroupingAtEachLevel, hashFunction, bucketCacheMaxSize, dataNodeCacheMaxSize)
	log.Infof("Initializing bucket tree state implemetation with configurations %+v", conf)
}

func newConfig(numBuckets int, maxGroupingAtEachLevel int, hashFunc hashFunc, bucketCacheMaxSize, dataNodeCacheMaxSize int) *config {
	conf := &config{maxGroupingAtEachLevel, -1, make(map[int]int), hashFunc,
		bucketCacheMaxSize, dataNodeCacheMaxSize}
	currentLevel := 0
	numBucketAtCurrentLevel := numBuckets
	levelInfoMap := make(map[int]int)
	levelInfoMap[currentLevel] = numBucketAtCurrentLevel
	for numBucketAtCurrentLevel > 1 {
		numBucketAtParentLevel := numBucketAtCurrentLevel / maxGroupingAtEachLevel
		if numBucketAtCurrentLevel%maxGroupingAtEachLevel != 0 {
			numBucketAtParentLevel++
		}

		numBucketAtCurrentLevel = numBucketAtParentLevel
		currentLevel++
		levelInfoMap[currentLevel] = numBucketAtCurrentLevel
	}

	conf.lowestLevel = currentLevel
	for k, v := range levelInfoMap {
		conf.levelToNumBucketsMap[conf.lowestLevel-k] = v
	}
	return conf
}

func (config *config) getNumBuckets(level int) int {
	if level < 0 || level > config.lowestLevel {
		panic(fmt.Errorf("level can only be between 0 and [%d]", config.lowestLevel))
	}
	return config.levelToNumBucketsMap[level]
}

func (config *config) computeBucketHash(data []byte) uint32 {
	return config.hashFunc(data)
}

func (config *config) getLowestLevel() int {
	return config.lowestLevel
}

func (config *config) getMaxGroupingAtEachLevel() int {
	return config.maxGroupingAtEachLevel
}

func (config *config) getNumBucketsAtLowestLevel() int {
	return config.getNumBuckets(config.getLowestLevel())
}

func (config *config) computeParentBucketNumber(bucketNumber int) int {
	parentBucketNumber := bucketNumber / config.getMaxGroupingAtEachLevel()
	if bucketNumber%config.getMaxGroupingAtEachLevel() != 0 {
		parentBucketNumber++
	}
	return parentBucketNumber
}

type hashFunc func(data []byte) uint32

func fnvHash(data []byte) uint32 {
	fnvHash := fnv.New32a()
	fnvHash.Write(data)
	return fnvHash.Sum32()
}