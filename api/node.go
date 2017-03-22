//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
	"hyperchain/common"
)

type NodeArgs struct {
	NodeHash string `json:"nodehash"`
}

type Node struct {
	eh *manager.EventHub
}

type NodeResult struct {
	Status      int         `json:"status"`
	CName       string      `json:"cName"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	delayTime   string      `json:"delayTime"`   //latency between nodes
	LatestBlock interface{} `json:"latestBlock"` //newest block of current block
}

func NewPublicNodeAPI(eh *manager.EventHub) *Node {
	return &Node{
		eh: eh,
	}
}

// GetNodes returns status of all the nodes
func (node *Node) GetNodes() (p2p.PeerInfos, error) {
	if node.eh == nil {
		return nil, &common.CallbackError{Message:"protocolManager is nil"}
	}

	return node.eh.GetNodeInfo(), nil
}

func (node *Node) GetNodeHash() (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message:"protocolManager is nil"}
	}
	return node.eh.GetPeerManager().GetLocalNodeHash(), nil
}

func (node *Node) DelNode(args NodeArgs) (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message:"protocolManager is nil"}
	}
	go node.eh.GetEventObject().Post(event.DelPeerEvent{
		Payload: []byte(args.NodeHash),
	})
	return "successful request", nil
}
