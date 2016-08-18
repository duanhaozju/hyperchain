// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了节点(node)的结构信息
package node

// Node的结构定义
type Node struct {
	Saver
	//提供RPC/HTTP服务的IP
	P2PIp string `json:"p2pip"`
	//提供RPC服务的端口
	P2PPort int `json:"p2pport"`
	//提供http服务的端口
	HttpPort int `json:"httpport"`
	// 节点公钥，可信传输的时候需要携带
	Pubkey string `json:"pubkey"`
}

type Saver interface {
	Save()
}