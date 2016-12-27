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
	if value == nil {
		logger.Error("DEBUG newDataNode empty value")
	}
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

func (dataNode *DataNode) isDelete() bool {
	return dataNode.value == nil
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
