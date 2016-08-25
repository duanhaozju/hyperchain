package node

import "hyperchain-alpha/common"

type Node interface {
	Start()
	GetNodeId() common.Hash


}

type GrpcNode struct {
	Id int
}

func (self *GrpcNode)Start()  {

}
func (self *GrpcNode)GetNodeId() common.Hash {

	return common.BytesToHash([]byte("hah"))
}