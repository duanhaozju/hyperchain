package core

import (
	"hyperchain-alpha/core/node"
	"sync"
)

type memNode struct {
	data map[string]node.Node
	lock sync.RWMutex
}

func newMemNode() *memNode {
	return &memNode{
		data: make(map[string]node.Node),
	}
}

var memNodeMap *memNode;

//-- 将Node存入内存
func PutNodeToMEM(key string, t node.Node){
	memNodeMap.lock.Lock()
	defer memNodeMap.lock.Unlock()
	memNodeMap.data[key] = t
}

//-- 在MEM中 根据Key获取的Node
func GetNodeFromMEM(key string) node.Node{
	memNodeMap.lock.RLock()
	defer memNodeMap.lock.RUnlock()
	return memNodeMap.data[key]
}

//-- 从MEM中删除Node
func DeleteNodeFromMEM(key string) {
	memNodeMap.lock.Lock()
	defer memNodeMap.lock.Unlock()
	delete(memNodeMap.data, key)
}

//-- 从MEM中获取所有Node
func GetAllNodeFromMEM() ([]node.Node) {
	memNodeMap.lock.RLock()
	defer memNodeMap.lock.RUnlock()
	var ts []node.Node
	for _, m := range memNodeMap.data {
		ts = append(ts, m)
	}
	return ts
}

/*//-- 将Node存入内存
func PutNodeToMEM(key string, t node.Node) error{
	return putNode(memDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Node
func GetNodeFromMEM(key string) (node.Node, error){
	return getNode(memDB, []byte(key))
}

//-- 从MEM中删除Node
func DeleteNodeFromMEM(key string) error {
	return deleteNode(memDB, []byte(key))
}

//-- 从MEM中获取所有Node
func GetAllNodeFromMEM() ([]node.Node, error) {
	var ts []node.Node

	Keys := memDB.Keys()

	var err error
	for _, key := range Keys {
		if string(key[:len(nodeHeaderKey)]) == string(nodeHeaderKey) {
			var t node.Node
			value, _ := memDB.Get(key)
			err = decondeFromBytes(value, &t)
			ts = append(ts, t)
		}
	}
	return ts, err
}*/
