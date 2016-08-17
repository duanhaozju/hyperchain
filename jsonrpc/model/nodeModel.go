package model
//定义节点
type Node struct {
	P2PAddr string `json:"p2pip"`
	P2PPort int `json:"p2pport"`
	HTTPPORT int `json:"httpport"`
}

type Nodes []Node


// TODO 添加节点

//func appendNode(n )