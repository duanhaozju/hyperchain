package bucket

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"sync"
)

type BucketNode struct {
	bucketKey          *BucketKey
	childrenCryptoHash [][]byte
	childrenUpdated    []bool
	markedForDeletion  bool
	lock               sync.RWMutex
	length             int32
}

type MemBucketNode struct {
	ChildrenCryptoHash [][]byte
	Length             int32
}

func newBucketNode(bucketKey *BucketKey) *BucketNode {
	maxChildren := conf.getMaxGroupingAtEachLevel()
	return &BucketNode{
		bucketKey:          bucketKey,
		childrenCryptoHash: make([][]byte, maxChildren),
		childrenUpdated:    make([]bool, maxChildren),
		markedForDeletion:  false,
	}
}

func unmarshalBucketNode(bucketKey *BucketKey, serializedBytes []byte) *BucketNode {
	bucketNode := newBucketNode(bucketKey)
	buffer := proto.NewBuffer(serializedBytes)
	for i := 0; i < conf.getMaxGroupingAtEachLevel(); i++ {
		childCryptoHash, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		//protobuf's buffer.EncodeRawBytes/buffer.DecodeRawBytes convert a nil into a zero length byte-array, so nil check would not work
		if len(childCryptoHash) != 0 {
			bucketNode.childrenCryptoHash[i] = childCryptoHash
		}
	}
	return bucketNode
}

func (bucketNode *BucketNode) marshal() []byte {
	buffer := proto.NewBuffer([]byte{})
	for i := 0; i < conf.getMaxGroupingAtEachLevel(); i++ {
		buffer.EncodeRawBytes(bucketNode.childrenCryptoHash[i])
	}
	data := buffer.Bytes()
	//log.Noticef("bucket node len %d", len(data))
	return data
}

func (bucketNode *BucketNode) setChildCryptoHash(childKey *BucketKey, cryptoHash []byte) {
	bucketNode.lock.Lock()
	defer bucketNode.lock.Unlock()
	i := bucketNode.bucketKey.getChildIndex(childKey)
	bucketNode.childrenCryptoHash[i] = cryptoHash
	bucketNode.childrenUpdated[i] = true
}

func (bucketNode *BucketNode) mergeBucketNode(anotherBucketNode *BucketNode) {
	bucketNode.lock.Lock()
	defer bucketNode.lock.Unlock()
	if !bucketNode.bucketKey.equals(anotherBucketNode.bucketKey) {
		panic(fmt.Errorf("Nodes with different keys can not be merged. BaseKey=[%#v], MergeKey=[%#v]", bucketNode.bucketKey, anotherBucketNode.bucketKey))
	}
	for i, childCryptoHash := range anotherBucketNode.childrenCryptoHash {
		if !bucketNode.childrenUpdated[i] {
			bucketNode.childrenCryptoHash[i] = childCryptoHash
		}
	}
}

func (bucketNode *BucketNode) computeCryptoHash(logger *logging.Logger) []byte {
	bucketNode.lock.RLock()
	defer bucketNode.lock.RUnlock()
	cryptoHashContent := []byte{}
	numChildren := 0
	for i, childCryptoHash := range bucketNode.childrenCryptoHash {
		if childCryptoHash != nil {
			numChildren++
			logger.Debugf("Appending crypto-hash for child bucket = [%s]", bucketNode.bucketKey.getChildKey(i))
			cryptoHashContent = append(cryptoHashContent, childCryptoHash...)
		}
	}
	if numChildren == 0 {
		logger.Debugf("Returning <nil> crypto-hash of bucket = [%s] - because, it has not children", bucketNode.bucketKey)
		bucketNode.markedForDeletion = true
		return nil
	}
	if numChildren == 1 {
		logger.Debugf("Propagating crypto-hash of single child node for bucket = [%s]", bucketNode.bucketKey)
		return cryptoHashContent
	}
	logger.Debugf("Computing crypto-hash for bucket [%s] by merging [%d] children", bucketNode.bucketKey, numChildren)
	if bucketNode.bucketKey.level <= 2 {
		return ComputeCryptoHash(cryptoHashContent)
	} else {
		return ComputeCryptoHashWithTrim(cryptoHashContent)
	}
}

func (bucketNode *BucketNode) String() string {
	bucketNode.lock.RLock()
	defer bucketNode.lock.RUnlock()
	numChildren := 0
	for i := range bucketNode.childrenCryptoHash {
		if bucketNode.childrenCryptoHash[i] != nil {
			numChildren++
		}
	}
	str := fmt.Sprintf("bucketKey={%s}\n NumChildren={%d}\n", bucketNode.bucketKey, numChildren)
	if numChildren == 0 {
		return str
	}

	str = str + "Childern crypto-hashes:\n"
	for i := range bucketNode.childrenCryptoHash {
		childCryptoHash := bucketNode.childrenCryptoHash[i]
		if childCryptoHash != nil {
			str = str + fmt.Sprintf("childNumber={%d}, cryptoHash={%x}\n", i, childCryptoHash)
		}
	}
	return str
}
