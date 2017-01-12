package bucket

import (
	"bytes"
	"sort"
	"encoding/json"
	"github.com/pkg/errors"
)

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

func (dataNodes DataNodes) Marshal() ([]byte){
	dataNodesKVDeltas := make([][]byte,len(dataNodes)*2)
	for i,dataNode := range dataNodes{
		dataNodesKVDeltas[2*i] = dataNode.getCompositeKey()
		dataNodesKVDeltas[2*i+1] = dataNode.getValue()
	}
	data,err := json.Marshal(dataNodesKVDeltas)
	if err != nil{
		log.Error("DataNodes Marshal error ",err)
		return nil
	}
	dataPrefix := append([]byte(DataNodesPrefix),byte(len(dataNodes)))
	data = append(dataPrefix,data...)
	return data
}

func UnmarshalDataNodes(bucketKey *BucketKey,data []byte, v interface{}) error {
	dataNodes, ok := v.(DataNodes)
	if ok == false {
		return errors.New("invalid type")
	}
	length := (int)(data[len(DataNodesPrefix)])
	dataNodesKVDeltas := make([][]byte,length*2)
	err := json.Unmarshal(data[len(DataNodesPrefix):], dataNodesKVDeltas)
	for i := 0;i < length;i++ {
		dataKey := &DataKey{bucketKey,dataNodesKVDeltas[2*i]}
		dataNodes = append(dataNodes,&DataNode{dataKey, dataNodesKVDeltas[2*i+1]})
	}
	return err
}

type dataNodesDelta struct {
	byBucket map[BucketKey]DataNodes
}

// TODO should be test by rjl and zhz
func newDataNodesDelta(treePrefix string,key_valueMap K_VMap) *dataNodesDelta {
	dataNodesDelta := &dataNodesDelta{make(map[BucketKey]DataNodes)}
	// TODO optimized
	for key, value := range key_valueMap {
		dataNodesDelta.add(treePrefix, key,value)
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