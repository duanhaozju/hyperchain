package db

import (
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/storage"
)

type NodeHanle struct {

}
func (nh NodeHanle) NodeSave(node node.Node) bool{

}
func (nh NodeHanle) NodeGet(key string){

}

func main() {
	nh := storage.NodeHandler()
	nh.NodeGet("hahah")

}