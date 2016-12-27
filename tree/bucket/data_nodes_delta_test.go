package bucket

import (
	"testing"
	"sort"
	"hyperchain/tree/bucket/testutil"
)
// TODO test
// 1.newDataNodesDelta

// TODO add

func TestDataNodesSort(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	dataNodes := dataNodes{}
	dataNode1 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value1_1"))
	dataNode2 := newDataNode(newDataKey("chaincodeID1", "key2"), []byte("value1_2"))
	dataNode3 := newDataNode(newDataKey("chaincodeID2", "key1"), []byte("value2_1"))
	dataNode4 := newDataNode(newDataKey("chaincodeID2", "key2"), []byte("value2_2"))
	dataNodes = append(dataNodes, []*DataNode{dataNode2, dataNode4, dataNode3, dataNode1}...)
	sort.Sort(dataNodes)
	testutil.AssertSame(t, dataNodes[0], dataNode1)
	testutil.AssertSame(t, dataNodes[1], dataNode2)
	testutil.AssertSame(t, dataNodes[2], dataNode3)
	testutil.AssertSame(t, dataNodes[3], dataNode4)
}

func TestDataNodesDelta(t *testing.T) {
	/*conf = newConfig(26, 3, fnvHash)
	stateDelta := NewStateDelta()
	stateDelta.Set("chaincodeID1", "key1", []byte("value1_1"), nil)
	stateDelta.Set("chaincodeID1", "key2", []byte("value1_2"), nil)
	stateDelta.Set("chaincodeID2", "key1", []byte("value2_1"), nil)
	stateDelta.Set("chaincodeID2", "key2", []byte("value2_2"), nil)

	dataNodesDelta := newDataNodesDelta(stateDelta)
	affectedBuckets := dataNodesDelta.getAffectedBuckets()
	t.Log("the num of affectedBuckets is ",len(affectedBuckets))
	testutil.AssertContains(t, affectedBuckets, newDataKey("chaincodeID1", "key1").getBucketKey())
	testutil.AssertContains(t, affectedBuckets, newDataKey("chaincodeID1", "key2").getBucketKey())
	testutil.AssertContains(t, affectedBuckets, newDataKey("chaincodeID2", "key1").getBucketKey())
	testutil.AssertContains(t, affectedBuckets, newDataKey("chaincodeID2", "key2").getBucketKey())*/
}
