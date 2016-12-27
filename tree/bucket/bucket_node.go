package bucket

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

type BucketNode struct {
	bucketKey          *BucketKey
	childrenCryptoHash [][]byte
	childrenUpdated    []bool
	markedForDeletion  bool
}

func newBucketNode(bucketKey *BucketKey) *BucketNode {
	maxChildren := conf.getMaxGroupingAtEachLevel()
	return &BucketNode{bucketKey, make([][]byte, maxChildren), make([]bool, maxChildren), false}
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
	return buffer.Bytes()
}

func (bucketNode *BucketNode) setChildCryptoHash(childKey *BucketKey, cryptoHash []byte) {
	i := bucketNode.bucketKey.getChildIndex(childKey)
	bucketNode.childrenCryptoHash[i] = cryptoHash
	bucketNode.childrenUpdated[i] = true
	logger.Errorf("DEBUG set idx = %d is true", i)
}

func (bucketNode *BucketNode) mergeBucketNode(anotherBucketNode *BucketNode) {
	logger.Error("mergeBucketNode")
	if !bucketNode.bucketKey.equals(anotherBucketNode.bucketKey) {
		panic(fmt.Errorf("Nodes with different keys can not be merged. BaseKey=[%#v], MergeKey=[%#v]", bucketNode.bucketKey, anotherBucketNode.bucketKey))
	}
	logger.Error("before mergeBucketNode", bucketNode.childrenCryptoHash)
	for i, childCryptoHash := range anotherBucketNode.childrenCryptoHash {
		if !bucketNode.childrenUpdated[i] {
			logger.Errorf("overwrite idx = %d value = %v", i, childCryptoHash)
			bucketNode.childrenCryptoHash[i] = childCryptoHash
		}
	}
	logger.Error("after mergeBucketNode", bucketNode.childrenCryptoHash)
	logger.Error("another", anotherBucketNode.childrenCryptoHash)
}

func (bucketNode *BucketNode) computeCryptoHash() []byte {
	cryptoHashContent := []byte{}
	numChildren := 0
	for i, childCryptoHash := range bucketNode.childrenCryptoHash {
		if childCryptoHash != nil {
			numChildren++
			logger.Errorf("Appending crypto-hash for child bucket = [%s]", bucketNode.bucketKey.getChildKey(i))
			cryptoHashContent = append(cryptoHashContent, childCryptoHash...)
		}
	}
	if numChildren == 0 {
		logger.Errorf("Returning <nil> crypto-hash of bucket = [%s] - because, it has not children", bucketNode.bucketKey)
		bucketNode.markedForDeletion = true
		return nil
	}
	if numChildren == 1 {
		logger.Errorf("Propagating crypto-hash of single child node for bucket = [%s]", bucketNode.bucketKey)
		return cryptoHashContent
	}
	logger.Errorf("Computing crypto-hash for bucket [%s] by merging [%d] children", bucketNode.bucketKey, numChildren)
	return ComputeCryptoHash(cryptoHashContent)
}

func (bucketNode *BucketNode) String() string {
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
