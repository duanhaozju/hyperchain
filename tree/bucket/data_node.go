package bucket

import (
	"fmt"
	"hyperchain/common"
)

type DataNode struct {
	dataKey *DataKey
	value   []byte
}

func newDataNode(dataKey *DataKey, value []byte) *DataNode {
	return &DataNode{dataKey, value}
}

func unmarshalDataNodeFromBytes(keyBytes []byte, valueBytes []byte) *DataNode {
	return unmarshalDataNode(newDataKeyFromEncodedBytes(keyBytes), valueBytes)
}

func unmarshalDataNode(dataKey *DataKey, serializedBytes []byte) *DataNode {
	value := make([]byte, len(serializedBytes))
	copy(value, serializedBytes)
	return &DataNode{dataKey, value}
}

func (dataNode *DataNode) getCompositeKey() []byte {
	return dataNode.dataKey.compositeKey
}

func (dataNode *DataNode) getActuallyKey() []byte {
	_, key := dataNode.getKeyElements()
	return []byte(key)
}

func (dataNode *DataNode) isDelete() bool {
	return dataNode.value == nil || len(dataNode.value) == 0 || common.EmptyHash(common.BytesToHash(dataNode.value))
}

func (dataNode *DataNode) getKeyElements() (string, string) {
	return DecodeCompositeKey(dataNode.getCompositeKey())
}

func (dataNode *DataNode) getValue() []byte {
	return dataNode.value
}

func (dataNode *DataNode) String() string {
	return fmt.Sprintf("dataKey=[%s], value=[%s]", dataNode.dataKey, common.Bytes2Hex(dataNode.value))
}