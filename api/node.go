//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"hyperchain/manager"
	"fmt"
)

// This file implements the handler of Node service API which
// can be invoked by client in JSON-RPC request.

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

func NewPublicNodeAPI(namespace string, eh *manager.EventHub) *Node {
	return &Node{
		namespace: namespace,
		eh:        eh,
	}
}

// GetNodes returns all vp nodes information in the namespace.
func (node *Node) GetNodes() (p2p.PeerInfos, error) {
	if node.eh == nil {
		return nil, &common.CallbackError{Message: "EventHub is nil"}
	}

	return node.eh.GetPeerManager().GetPeerInfo(), nil
}

// GetNodeHash returns current node hash.
func (node *Node) GetNodeHash() (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	return node.eh.GetPeerManager().GetLocalNodeHash(), nil
}

// DeleteVP sends a request to delete vp node. Client can't judge from the returned results
// whether the deletion is successful. But client can call Node.GetNodes to determine whether
// the deletion is successful.
func (node *Node) DeleteVP(args NodeArgs) (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	go node.eh.GetEventObject().Post(event.DelVPEvent{
		Payload: []byte(args.NodeHash),
	})
	return fmt.Sprintf("successful request to delete vp node, hash %s", args.NodeHash), nil
}

// DeleteNVP sends a request to delete nvp node.
func (node *Node) DeleteNVP(args NodeArgs) (string, error) {
	if node.eh == nil {
		return "", &common.CallbackError{Message: "EventHub is nil"}
	}
	go node.eh.GetEventObject().Post(event.DelNVPEvent{
		Payload: []byte(args.NodeHash),
	})
	return fmt.Sprintf("successful request to delete nvp node, hash %s", args.NodeHash), nil
}

// DelNode is compatible with sdk for release1.2.
func (node *Node) DelNode(args NodeArgs) (string, error) {
	return node.DeleteVP(args)
}
