package bucket

import (
	"testing"
	"hyperchain/tree/bucket/testutil"
)
func TestBucketTreeDeltaBasic(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketTreeDelta := newBucketTreeDelta()
	b1 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(2, 1))
	testutil.AssertSame(t, bucketTreeDelta.getOrCreateBucketNode(newBucketKey(2, 1)), b1)
	b2 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(2, 2))
	b3 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(2, 3))
	testutil.AssertContainsAll(t, bucketTreeDelta.getBucketNodesAt(2), []*bucketNode{b1, b2, b3})

	b4 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(1, 1))
	b5 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(1, 2))
	testutil.AssertContainsAll(t, bucketTreeDelta.getBucketNodesAt(1), []*bucketNode{b4, b5})

	b6 := bucketTreeDelta.getOrCreateBucketNode(newBucketKey(0, 1))
	testutil.AssertContainsAll(t, bucketTreeDelta.getBucketNodesAt(0), []*bucketNode{b6})
	testutil.AssertContainsAll(t, bucketTreeDelta.getBucketNodesAt(1), []*bucketNode{b4, b5})
	testutil.AssertContainsAll(t, bucketTreeDelta.getBucketNodesAt(2), []*bucketNode{b1, b2, b3})

	testutil.AssertEquals(t, len(bucketTreeDelta.getBucketNodesAt(0)), 1)
	testutil.AssertEquals(t, len(bucketTreeDelta.getBucketNodesAt(1)), 2)
	testutil.AssertEquals(t, len(bucketTreeDelta.getBucketNodesAt(2)), 3)

	testutil.AssertSame(t, bucketTreeDelta.getRootNode(), b6)
}

func TestBucketTreeDeltaGetRootWithoutProcessing(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketTreeDelta := newBucketTreeDelta()
	bucketKey1 := newBucketKey(2, 1)
	bucketTreeDelta.getOrCreateBucketNode(bucketKey1)
	defer testutil.AssertPanic(t, "A panic should have occured. Because, asking for root node without fully prosessing the bucket tree delta")
	bucketTreeDelta.getRootNode()
}
