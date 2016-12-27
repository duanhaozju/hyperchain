package bucket

import (
	"github.com/golang/protobuf/proto"
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
	chaincodeID, _ := dataNode.getKeyElements()
	if chaincodeID != c.currentChaincodeID {
		c.appendCurrentChaincodeData()
		c.currentChaincodeID = chaincodeID
		c.dataNodes = nil
	}
	c.dataNodes = append(c.dataNodes, dataNode)
}

func (c *bucketHashCalculator) computeCryptoHash() []byte {
	if c.currentChaincodeID != "" {
		c.appendCurrentChaincodeData()
		c.currentChaincodeID = ""
		// TODO
		c.dataNodes = nil
	}
	logger.Errorf("Hashable content for bucket [%s]: length=%d, contentInStringForm=[%s]", c.bucketKey, len(c.hashingData), string(c.hashingData))
	if c.hashingData == nil {
		return nil
	}
	return ComputeCryptoHash(c.hashingData)
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
