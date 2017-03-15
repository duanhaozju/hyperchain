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

type PublicNodeAPI struct {
	pm *manager.EventHub
}

type NodeResult struct {
	Status      int         `json:"status"`
	CName       string      `json:"cName"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	delayTime   string      `json:"delayTime"`   //latency between nodes
	LatestBlock interface{} `json:"latestBlock"` //newest block of current block
}

func NewPublicNodeAPI(pm *manager.EventHub) *PublicNodeAPI {
	return &PublicNodeAPI{
		pm: pm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicNodeAPI) GetNodes() (p2p.PeerInfos, error) {
	if node.pm == nil {
		return nil, &common.CallbackError{"protocolManager is nil"}
	}

	return node.pm.GetNodeInfo(), nil
}

func (node *PublicNodeAPI) GetNodeHash() (string, error) {
	if node.pm == nil {
		return "", &common.CallbackError{"protocolManager is nil"}
	}
	return node.pm.PeerManager.GetLocalNodeHash(), nil
}

func (node *PublicNodeAPI) DelNode(args NodeArgs) (string, error) {
	if node.pm == nil {
		return "", &common.CallbackError{"protocolManager is nil"}
	}
	go node.pm.GetEventObject().Post(event.DelPeerEvent{
		Payload: []byte(args.NodeHash),
	})
	return "successful request", nil
}
