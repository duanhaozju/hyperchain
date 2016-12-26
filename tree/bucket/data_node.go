package bucket

import (
	"fmt"
	"hyperchain/common"
)

type dataNode struct {
	dataKey *dataKey
	value   []byte
}

func newDataNode(dataKey *dataKey, value []byte) *dataNode {
	return &dataNode{dataKey, value}
}

func unmarshalDataNodeFromBytes(keyBytes []byte, valueBytes []byte) *dataNode {
	return unmarshalDataNode(newDataKeyFromEncodedBytes(keyBytes), valueBytes)
}

func unmarshalDataNode(dataKey *dataKey, serializedBytes []byte) *dataNode {
	value := make([]byte, len(serializedBytes))
	copy(value, serializedBytes)
	return &dataNode{dataKey, value}
}

func (dataNode *dataNode) getCompositeKey() []byte {
	return dataNode.dataKey.compositeKey
}

func (dataNode *dataNode) isDelete() bool {
	return dataNode.value == nil
}

func (dataNode *dataNode) getKeyElements() (string, string) {
	return DecodeCompositeKey(dataNode.getCompositeKey())
}

func (dataNode *dataNode) getValue() []byte {
	return dataNode.value
}

func (dataNode *dataNode) String() string {
	return fmt.Sprintf("dataKey=[%s], value=[%s]", dataNode.dataKey, common.Bytes2Hex(dataNode.value))
}
