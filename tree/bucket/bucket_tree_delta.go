package bucket

import (
	"sync"
)

type byBucketNumber map[int]*BucketNode

type bucketTreeDelta struct {
	byLevel map[int]byBucketNumber
	lock    sync.RWMutex
}

func newBucketTreeDelta() *bucketTreeDelta {
	return &bucketTreeDelta{
		byLevel:make(map[int]byBucketNumber),
	}
}


func (bucketTreeDelta *bucketTreeDelta) getOrCreateBucketNode(bucketKey *BucketKey) *BucketNode {
	bucketTreeDelta.lock.Lock()
	defer bucketTreeDelta.lock.Unlock()
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
	bucketTreeDelta.lock.RLock()
	defer bucketTreeDelta.lock.RUnlock()
	return bucketTreeDelta.byLevel == nil || len(bucketTreeDelta.byLevel) == 0
}

func (bucketTreeDelta *bucketTreeDelta) getBucketNodesAt(level int) []*BucketNode {
	bucketTreeDelta.lock.RLock()
	defer bucketTreeDelta.lock.RUnlock()
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
