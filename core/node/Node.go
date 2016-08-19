// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了节点(node)的结构信息
package node

import (
	"hyperchain-alpha/encrypt"
)

// Node的结构定义
type Node struct {
	//提供RPC/HTTP服务的IP
	P2PIP string `json:"p2pip"`
	//提供RPC服务的端口
	P2PPort int `json:"p2pport"`
	//提供http服务的端口
	HttpPort int `json:"httpport"`
	// 节点对外公布的地址
	CoinBase string
	//私钥对外不可见
	privateKey string
	//公钥对外可见
	publicKey string


}

type Nodes []Node

func NewNode(P2PIP string,P2PPORT int,HttpPort int) Node{
	privatekey := encrypt.GetPrivateKey()
	publickey := encrypt.GetPublicKey(privatekey)
	coinbase :=  encrypt.EncodePublicKey(publickey)

	var newNode = Node{
		P2PIP:P2PIP,
		P2PPort:P2PPORT,
		HttpPort:HttpPort,
		CoinBase:coinbase,
		privateKey:privatekey,
		publicKey:publickey,
	}
	return newNode
}