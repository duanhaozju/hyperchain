package peerComm

type PeerConfigNodes struct {
	ExternalAddress string `json:"external_address"`
	RPCPort         int    `json:"rpc_port"`
	Port            int    `json:"port"`
	ID              int    `json:"id"`
	Address         string `json:"address"`
}

func NewPeerConfigNodes(ip string, rpc_port int, port int, id int) *PeerConfigNodes {
	var peerConfigNode = PeerConfigNodes{
		ExternalAddress: ip,
		RPCPort:         rpc_port,
		Port:            port,
		ID:              id,
		Address:         ip,
	}
	return &peerConfigNode
}
