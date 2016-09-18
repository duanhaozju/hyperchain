package hpc

import "hyperchain/manager"

type PublicNodeAPI struct{
	pm *manager.ProtocolManager
}

func NewPublicNodeAPI( pm *manager.ProtocolManager) *PublicNodeAPI{
	return &PublicNodeAPI{
		pm: pm,
	}
}

// TODO 得到节点状态
func (node *PublicNodeAPI) GetNodes() map[string]bool{
	return node.pm.GetNodeInfo()
}