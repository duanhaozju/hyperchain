package buckettree

import (
	"bytes"
	"sort"
	"hyperchain/core/hyperstate"
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
func newDataNodesDelta(stateDelta *hyperstate.StateDelta) *dataNodesDelta {
	dataNodesDelta := &dataNodesDelta{make(map[bucketKey]dataNodes)}
	accountIDs := stateDelta.GetUpdatedAccountIds(false)
	for _, accountID := range accountIDs {
		updates := stateDelta.GetUpdates(accountID)
		for key, updatedValue := range updates {
			if stateDelta.RollBackwards {
				dataNodesDelta.add(accountID, key, updatedValue.GetPreviousValue())
			} else {
				dataNodesDelta.add(accountID, key, updatedValue.GetValue())
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