package bucket

import (
	"bytes"
	"testing"
	"encoding/binary"
)

// TODO test
// 1.newDataNodesDelta

// TODO add


func TestBigInt(t *testing.T){
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, 8)
	log.Critical("size is",bytes)
	value := binary.LittleEndian.Uint64(bytes)
	log.Critical("size is",value)

}

func TestDataNodesSort(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	dataNodes := DataNodes{}
	newDataNodes := DataNodes{}
	dataNode1 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value1"))
	dataNode2 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value2"))
	dataNode3 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value3"))
	dataNode4 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value4"))
	dataNode5 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value5"))
	dataNode6 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value6"))
	dataNode7 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value7"))
	dataNode8 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value8"))
	dataNode9 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value9"))
	dataNode10 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value10"))
	dataNode11 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value11"))
	dataNodes = append(dataNodes, []*DataNode{dataNode2, dataNode4, dataNode3, dataNode1, dataNode5, dataNode6, dataNode8, dataNode7, dataNode9, dataNode10, dataNode11}...)

	dataNodesValue1 := dataNodes.Marshal()
	dataNodesValue2 := dataNodes.Marshal()
	dataNodesValue3 := dataNodes.Marshal()
	log.Critical("dataNodesValue1 is ", dataNodesValue1)
	log.Critical("dataNodesValue2 is ", dataNodesValue2)
	log.Critical("dataNodesValue3 is ", dataNodesValue3)

	err := UnmarshalDataNodes(dataNode1.dataKey.bucketKey, dataNodesValue1, &newDataNodes)
	if err != nil {
		log.Error("UnmarshalDataNodes error", err)
		return
	}

	for i := 0; i < len(dataNodes); i++ {
		if bytes.Compare(newDataNodes[i].getValue(), dataNodes[i].getValue()) != 0 ||
			bytes.Compare(newDataNodes[i].getCompositeKey(), dataNodes[i].getCompositeKey()) != 0 {
		}
	}
	log.Critical("newDataNodes is",len(newDataNodes))
	log.Critical("newDataNodes equals dataNodes")
}

func TestDataNodesSort2(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	dataNodes := DataNodes{}
	newDataNodes := DataNodes{}
	dataNode1 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value1"))
	dataNode2 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value2"))
	dataNode3 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value3"))
	dataNode4 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value4"))
	dataNode5 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value5"))
	dataNode6 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value6"))
	dataNode7 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value7"))
	dataNode8 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value8"))
	dataNode9 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value9"))
	dataNode10 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value10"))
	dataNode11 := newDataNode(newDataKey("chaincodeID1", "key1"), []byte("value11"))
	dataNodes = append(dataNodes, []*DataNode{dataNode2, dataNode4, dataNode3, dataNode1, dataNode5, dataNode6, dataNode8, dataNode7, dataNode9, dataNode10, dataNode11}...)

	dataNodesValue1 := dataNodes.Marshal()
	dataNodesValue2 := dataNodes.Marshal()
	dataNodesValue3 := dataNodes.Marshal()
	log.Critical("dataNodesValue1 is ", dataNodesValue1)
	log.Critical("dataNodesValue2 is ", dataNodesValue2)
	log.Critical("dataNodesValue3 is ", dataNodesValue3)

	err := UnmarshalDataNodes(dataNode1.dataKey.bucketKey, dataNodesValue1, &newDataNodes)
	if err != nil {
		log.Error("UnmarshalDataNodes error", err)
		return
	}

	for i := 0; i < len(dataNodes); i++ {
		if bytes.Compare(newDataNodes[i].getValue(), dataNodes[i].getValue()) != 0 ||
			bytes.Compare(newDataNodes[i].getCompositeKey(), dataNodes[i].getCompositeKey()) != 0 {
		}
	}
	log.Critical("newDataNodes equals dataNodes")
}
