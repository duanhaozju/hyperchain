package p2p

import (
	"hyperchain/p2p/info"
	"hyperchain/p2p/message"
	"hyperchain/p2p/msg"
	"hyperchain/p2p/network"
)

type Node struct {
	info      *info.Info
	namespace string
	net       *network.HyperNet
}

func NewNode(namespace string, id int, hostname string, net *network.HyperNet) *Node {
	node := &Node{
		info:      info.NewInfo(id, hostname, namespace),
		namespace: namespace,
		net:       net,
	}
	return node
}

//Bind msgType and handler for this namespace
func (node *Node) Bind(msgType message.MsgType, handler msg.MsgHandler) {
	node.net.RegisterHandler(node.namespace, msgType, handler)
}

// UnBindAll all handlers.
func (node *Node) UnBindAll() {
	node.net.DeRegisterHandlers(node.namespace)
}
