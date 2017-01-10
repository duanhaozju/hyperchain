package bucket

import (
	"github.com/golang/protobuf/proto"
	"crypto/sha1"
)

type bucketHashCalculator struct {
	bucketKey          *BucketKey
	currentChaincodeID string
	dataNodes          []*DataNode
	hashingData        []byte
}

func newBucketHashCalculator(bucketKey *BucketKey) *bucketHashCalculator {
	return &bucketHashCalculator{bucketKey, "", nil, nil}
}

// addNextNode - this method assumes that the datanodes are added in the increasing order of the keys
func (c *bucketHashCalculator) addNextNode(dataNode *DataNode) {
	c.hashingData = append(c.hashingData,append(dataNode.getCompositeKey(),dataNode.getValue()...))
}

func (c *bucketHashCalculator) computeCryptoHash() []byte {
	if c.hashingData == nil {
		return nil
	}
	h := sha1.New()
	h.Write([]byte(c.hashingData))
	bs := h.Sum(nil)
	return bs
}

func (c *bucketHashCalculator) appendCurrentChaincodeData() {
	if c.currentChaincodeID == "" {
		return
	}
	c.appendSizeAndData([]byte(c.currentChaincodeID))
	c.appendSize(len(c.dataNodes))
	for _, dataNode := range c.dataNodes {
		_, key := dataNode.getKeyElements()
		value := dataNode.getValue()
		c.appendSizeAndData([]byte(key))
		c.appendSizeAndData(value)
	}
}

func (c *bucketHashCalculator) appendSizeAndData(b []byte) {
	c.appendSize(len(b))
	c.hashingData = append(c.hashingData, b...)
}

func (c *bucketHashCalculator) appendSize(size int) {
	c.hashingData = append(c.hashingData, proto.EncodeVarint(uint64(size))...)
}
