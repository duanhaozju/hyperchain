//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
)

type NodeArgs struct {
	NodeHash string `json:"nodehash"`
}

type PublicNodeAPI struct {
	pm *manager.ProtocolManager
}

type NodeResult struct {
	Status      int         `json:"status"`
	CName       string      `json:"cName"`
	IP          string      `json:"ip"`
	Port        int         `json:"port"`
	delayTime   string      `json:"delayTime"`   // 节点与节点的延迟时间
	LatestBlock interface{} `json:"latestBlock"` //当前节点最新的区块
}

func NewPublicNodeAPI(pm *manager.ProtocolManager) *PublicNodeAPI {
	return &PublicNodeAPI{
		pm: pm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicNodeAPI) GetNodes() (p2p.PeerInfos, error) {
	if node.pm == nil {
		return nil, &CallbackError{"protocolManager is nil"}
	}

	return node.pm.GetNodeInfo(), nil
}

func (node *PublicNodeAPI) GetNodeHash() (string, error) {
	if node.pm == nil {
		return "", &CallbackError{"protocolManager is nil"}
	}
	return node.pm.Peermanager.GetLocalNodeHash(), nil
}

func (node *PublicNodeAPI) DelNode(args NodeArgs) error {
	if node.pm == nil {
		return &CallbackError{"protocolManager is nil"}
	}
	node.pm.GetEventObject().Post(event.DelPeerEvent{
		Payload: []byte(args.NodeHash),
	})
	return nil
}

/*func outputNodeResult(node *client.PeerInfo) *NodeResult {

	return &NodeResult{
		Status: node.Status,
		//CName: node.CName,
		IP: node.IP,
		Port: node.Port,
	}

}*/
