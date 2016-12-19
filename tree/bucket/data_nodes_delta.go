package bucket

import (
	"bytes"
	"sort"
)

// Code for managing changes in data nodes
type dataNodes []*dataNode

func (dataNodes dataNodes) Len() int {
	return len(dataNodes)
}

func (dataNodes dataNodes) Swap(i, j int) {
	dataNodes[i], dataNodes[j] = dataNodes[j], dataNodes[i]
}

func (dataNodes dataNodes) Less(i, j int) bool {
	return bytes.Compare(dataNodes[i].dataKey.compositeKey, dataNodes[j].dataKey.compositeKey) < 0
}

type dataNodesDelta struct {
	byBucket map[bucketKey]dataNodes
}

// TODO should be test by rjl and zhz
func newDataNodesDelta(accountID string,key_valueMap K_VMap) *dataNodesDelta {
	dataNodesDelta := &dataNodesDelta{make(map[bucketKey]dataNodes)}
	// TODO optimized
	for key, value := range key_valueMap {
		dataNodesDelta.add(accountID, key,value)
	}
	for _, dataNodes := range dataNodesDelta.byBucket {
		sort.Sort(dataNodes)
	}
	return dataNodesDelta
}

func (dataNodesDelta *dataNodesDelta) add(accountID string, key string, value []byte) {
	dataKey := newDataKey(accountID, key)
	bucketKey := dataKey.getBucketKey()
	dataNode := newDataNode(dataKey, value)
	logger.Debugf("Adding dataNode=[%s] against bucketKey=[%s]", dataNode, bucketKey)
	dataNodesDelta.byBucket[*bucketKey] = append(dataNodesDelta.byBucket[*bucketKey], dataNode)
}

func (dataNodesDelta *dataNodesDelta) getAffectedBuckets() []*bucketKey {
	changedBuckets := []*bucketKey{}
	for bucketKey := range dataNodesDelta.byBucket {
		copyOfBucketKey := bucketKey.clone()
		logger.Debugf("Adding changed bucket [%s]", copyOfBucketKey)
		changedBuckets = append(changedBuckets, copyOfBucketKey)
	}
	logger.Debugf("Changed buckets are = [%s]", changedBuckets)
	return changedBuckets
}

func (dataNodesDelta *dataNodesDelta) getSortedDataNodesFor(bucketKey *bucketKey) dataNodes {
	return dataNodesDelta.byBucket[*bucketKey]
}