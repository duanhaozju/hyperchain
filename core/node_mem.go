package core

import (
	"hyperchain-alpha/core/node"
)

//-- 将Node存入内存
func PutNodeToMEM(key string, t node.Node) error{
	return PutNode(MemDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Node
func GetNodeFromMEM(key string) (node.Node, error){
	return GetNode(MemDB, []byte(key))
}

//-- 从MEM中删除Node
func DeleteNodeFromMEM(key string) error {
	return DeleteNode(MemDB, []byte(key))
}

//-- 从MEM中获取所有Node
func GetAllNodeFromMEM() ([]node.Node, error) {
	var ts []node.Node

	Keys := MemDB.Keys()

	var err error
	for _, key := range Keys {
		var t node.Node
		value , _ := MemDB.Get(key)
		err = decondeFromBytes(value, &t)
		ts = append(ts, t)
	}
	return ts, err
}
