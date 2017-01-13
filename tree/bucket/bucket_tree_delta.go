package bucket

import "math/big"

type byBucketNumber map[int]*BucketNode

type bucketTreeDelta struct {
	byLevel map[int]byBucketNumber
}

func newBucketTreeDelta() *bucketTreeDelta {
	return &bucketTreeDelta{make(map[int]byBucketNumber)}
}

func newUpdatedValueSet(blockNumber *big.Int) *UpdatedValueSet {
	return &UpdatedValueSet{BlockNum: blockNumber, UpdatedKVs: make(map[string]*UpdatedValue)}
}

func (bucketTreeDelta *bucketTreeDelta) getOrCreateBucketNode(bucketKey *BucketKey) *BucketNode {
	byBucketNumber := bucketTreeDelta.byLevel[bucketKey.level]
	if byBucketNumber == nil {
		byBucketNumber = make(map[int]*BucketNode)
		bucketTreeDelta.byLevel[bucketKey.level] = byBucketNumber
	}
	bucketNode := byBucketNumber[bucketKey.bucketNumber]
	if bucketNode == nil {
		bucketNode = newBucketNode(bucketKey)
		byBucketNumber[bucketKey.bucketNumber] = bucketNode
	}
	return bucketNode
}

func (bucketTreeDelta *bucketTreeDelta) isEmpty() bool {
	return bucketTreeDelta.byLevel == nil || len(bucketTreeDelta.byLevel) == 0
}

func (bucketTreeDelta *bucketTreeDelta) getBucketNodesAt(level int) []*BucketNode {
	bucketNodes := []*BucketNode{}
	byBucketNumber := bucketTreeDelta.byLevel[level]
	if byBucketNumber == nil {
		return nil
	}
	for _, bucketNode := range byBucketNumber {
		bucketNodes = append(bucketNodes, bucketNode)
	}
	return bucketNodes
}

func (bucketTreeDelta *bucketTreeDelta) getRootNode() *BucketNode {
	bucketNodes := bucketTreeDelta.getBucketNodesAt(0)
	if bucketNodes == nil || len(bucketNodes) == 0 {
		panic("This method should be called after processing is completed (i.e., the root node has been created)")
	}
	return bucketNodes[0]
}
