package hpc

import (
	"hyperchain/manager"
	"hyperchain/p2p/peer"
)

type PublicNodeAPI struct{
	pm *manager.ProtocolManager
}

func NewPublicNodeAPI( pm *manager.ProtocolManager) *PublicNodeAPI{
	return &PublicNodeAPI{
		pm: pm,
	}
}

// GetNodes returns status of all the nodes
func (node *PublicNodeAPI) GetNodes() client.PeerInfos{
	return node.pm.GetNodeInfo()
}