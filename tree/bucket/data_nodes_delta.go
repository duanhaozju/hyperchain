package buckettree

import (
	"bytes"
	"sort"
	"github.com/hyperledger/fabric/core/ledger/statemgmt"
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

// 将stateDelta中的所有数据都加入到dataNodesDelta里
// stateDelta中记录的是所有更改了的chaincode的原生数据
// dataNodesDelta是记录了bucket tree的叶子数据
func newDataNodesDelta(stateDelta *statemgmt.StateDelta) *dataNodesDelta {
	dataNodesDelta := &dataNodesDelta{make(map[bucketKey]dataNodes)}
	chaincodeIDs := stateDelta.GetUpdatedChaincodeIds(false)
	for _, chaincodeID := range chaincodeIDs {
		updates := stateDelta.GetUpdates(chaincodeID)
		for key, updatedValue := range updates {
			if stateDelta.RollBackwards {
				dataNodesDelta.add(chaincodeID, key, updatedValue.GetPreviousValue())
			} else {
				dataNodesDelta.add(chaincodeID, key, updatedValue.GetValue())
			}
		}
	}
	for _, dataNodes := range dataNodesDelta.byBucket {
		sort.Sort(dataNodes)
	}
	return dataNodesDelta
}

func (dataNodesDelta *dataNodesDelta) add(chaincodeID string, key string, value []byte) {
	dataKey := newDataKey(chaincodeID, key)
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
