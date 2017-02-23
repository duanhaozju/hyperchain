package bucket

import (
	"bytes"
	"github.com/pkg/errors"
	"sort"
	"github.com/golang/protobuf/proto"
	"fmt"
)

const MAXDATANODESSIZE  = 8

// Code for managing changes in data nodes
type DataNodes []*DataNode

func (dataNodes DataNodes) Len() int {
	return len(dataNodes)
}

func (dataNodes DataNodes) Swap(i, j int) {
	dataNodes[i], dataNodes[j] = dataNodes[j], dataNodes[i]
}

func (dataNodes DataNodes) Less(i, j int) bool {
	return bytes.Compare(dataNodes[i].dataKey.compositeKey, dataNodes[j].dataKey.compositeKey) < 0
}

func (dataNodes DataNodes) Marshal() []byte {
	buffer := proto.NewBuffer([]byte{})
	buffer.EncodeFixed64(uint64(len(dataNodes)))
	for _, dataNode := range dataNodes {
		buffer.EncodeRawBytes(dataNode.getActuallyKey())
		buffer.EncodeRawBytes(dataNode.getValue())
	}
	return buffer.Bytes()
}

func UnmarshalDataNodes(prefix string, bucketKey *BucketKey, data []byte, v interface{}) error {
	dataNodes, ok := v.(*DataNodes)
	if ok == false {
		return errors.New("invalid type")
	}

	buffer := proto.NewBuffer(data)
	length, err := buffer.DecodeFixed64()
	if err != nil {
		log.Errorf("decode datanodes failed")
		return err
	}
	for i := 0; i < int(length); i++ {
		key, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		value, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		dataKey := &DataKey{bucketKey, ConstructCompositeKey(prefix, string(key))}
		*dataNodes = append(*dataNodes, newDataNode(dataKey, value))
	}
	return err
}

type dataNodesDelta struct {
	byBucket map[BucketKey]DataNodes
}

// TODO should be test by rjl and zhz
func newDataNodesDelta(treePrefix string, key_valueMap K_VMap) *dataNodesDelta {
	dataNodesDelta := &dataNodesDelta{make(map[BucketKey]DataNodes)}
	// TODO optimized
	for key, value := range key_valueMap {
		dataNodesDelta.add(treePrefix, key, value)
	}
	for _, dataNodes := range dataNodesDelta.byBucket {
		sort.Sort(dataNodes)
	}
	return dataNodesDelta
}

func (dataNodesDelta *dataNodesDelta) add(treePrefix string, key string, value []byte) {
	dataKey := newDataKey(treePrefix, key)
	bucketKey := dataKey.getBucketKey()
	dataNode := newDataNode(dataKey, value)
	//log.Debugf("Adding dataNode=[%s] against bucketKey=[%s]", dataNode, bucketKey)
	dataNodesDelta.byBucket[*bucketKey] = append(dataNodesDelta.byBucket[*bucketKey], dataNode)
}

func (dataNodesDelta *dataNodesDelta) getAffectedBuckets() []*BucketKey {
	changedBuckets := []*BucketKey{}
	for bucketKey := range dataNodesDelta.byBucket {
		copyOfBucketKey := bucketKey.clone()
		//log.Debugf("Adding changed bucket [%s]", copyOfBucketKey)
		changedBuckets = append(changedBuckets, copyOfBucketKey)
	}
	//log.Debugf("Changed buckets are = [%s]", changedBuckets)
	return changedBuckets
}

func (dataNodesDelta *dataNodesDelta) getSortedDataNodesFor(bucketKey *BucketKey) DataNodes {
	return dataNodesDelta.byBucket[*bucketKey]
}
