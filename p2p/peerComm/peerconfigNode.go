package peerComm

type PeerConfigNodes struct {
	ExternalAddress string `json:"external_address"`
	RPCPort int `json:"rpc_port"`
	Port int `json:"port"`
	ID int `json:"id"`
	Address string `json:"address"`
}

