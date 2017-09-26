//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

/*
    This file implements the handler of node service API
	which can be invoked by client in JSON-RPC request.
 */

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"hyperchain/p2p"
)

type NodeArgs struct {
	NodeHash string `json:"nodehash"`
}

type Node struct {
	namespace string
	eh        *manager.EventHub
}
// TODO add annotation, fix bug about GetNodes returns null
type NodeResult struct {
	Status      int         `json:"status"`		//
	CName       string      `json:"cName"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	delayTime   string      `json:"delayTime"`   //latency between nodes
	LatestBlock interface{} `json:"latestBlock"` //newest block of current block
}

func NewPublicNodeAPI(namespace string, eh *manager.EventHub) *Node {
	return &Node{
		namespace: namespace,
		eh:        eh,
	}
}

// GetNodes returns all nodes information.
func (node *Node) GetNodes() (p2p.PeerInfos, error) {
	if node.eh == nil {
		return nil, &common.CallbackError{Message: "EventHub is nil"}
	}

	return node.eh.GetPeerManager().GetPeerInfo(), nil
}

func (node *Node) GetNodeHash() (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	return node.eh.GetPeerManager().GetLocalNodeHash(), nil
}

func (node *Node) DeleteVP(args NodeArgs) (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	go node.eh.GetEventObject().Post(event.DelVPEvent{
		Payload: []byte(args.NodeHash),
	})
	return fmt.Sprintf("successful request to delete vp node, hash %s", args.NodeHash), nil
}

func (node *Node) DeleteNVP(args NodeArgs) (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	go node.eh.GetEventObject().Post(event.DelNVPEvent{
		Payload: []byte(args.NodeHash),
	})
	return "successful request to delete nvp node", nil
}

// DelNode is in order to be compatible with sdk for release1.2
func (node *Node) DelNode(args NodeArgs) (string, error) {
	return node.DeleteVP(args)
}
