//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/manager"
	"hyperchain/p2p"
	"github.com/pkg/errors"
)

type PublicNodeAPI struct{
	pm *manager.ProtocolManager
}

type NodeResult struct{
	Status 		int		`json:"status"`
	CName 		string		`json:"cName"`
	IP 		string		`json:"ip"`
	Port 		int		`json:"port"`
	delayTime	string		`json:"delayTime"`	// 节点与节点的延迟时间
	LatestBlock 	interface{}	`json:"latestBlock"`  //当前节点最新的区块
}

func NewPublicNodeAPI( pm *manager.ProtocolManager) *PublicNodeAPI{
	return &PublicNodeAPI{
		pm: pm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicNodeAPI) GetNodes() (p2p.PeerInfos, error) {
	//func (node *PublicNodeAPI) GetNodes() []*NodeResult{

	if node.pm == nil {
		return nil, errors.New("peerManager is nil")
	}

	return node.pm.GetNodeInfo(), nil
}

/*func outputNodeResult(node *client.PeerInfo) *NodeResult {

	return &NodeResult{
		Status: node.Status,
		//CName: node.CName,
		IP: node.IP,
		Port: node.Port,
	}

}*/

