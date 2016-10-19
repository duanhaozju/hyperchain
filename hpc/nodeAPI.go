package hpc

import (
	"hyperchain/manager"
	"hyperchain/p2p/peer"
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
func (node *PublicNodeAPI) GetNodes() client.PeerInfos{
//func (node *PublicNodeAPI) GetNodes() []*NodeResult{

	//block, err := lastestBlock() // 从一个节点的节点列表中 获取每个节点当前的最新区块
	//
	//fields := map[string]interface{}{
	//	"blockNumber": block.Number,
	//	"blockHash":  block.Hash,
	//	"avgTime":  block.AvgTime,
	//	"txcounts":  block.TxCounts,
	//}
	//
	//
	//n := node.pm.GetNodeInfo()
	//
	//nodes := make([]*NodeResult, len(n))
	//var err error
	//for i, nd := range n {
	//	//if nodes[i], err = formatTx(tx); err != nil {
	//	//	return nil, err
	//	//}
	//	nodes[i] = outputNodeResult(nd)
	//	nodes[i].LatestBlock = fields
	//
	//}




	return node.pm.GetNodeInfo()
}

/*func outputNodeResult(node *client.PeerInfo) *NodeResult {

	return &NodeResult{
		Status: node.Status,
		//CName: node.CName,
		IP: node.IP,
		Port: node.Port,
	}

}*/

