package common

type PeerConfig struct {
	SelfConfig  Self              `json:"self"`
	PeerNodes   []PeerConfigNodes `json:"nodes"`
	Maxpeernode int               `json:"maxpeernode"`
}

type Self struct {
	IsOrigin              bool   `json:"is_origin"`
	IsVP                  bool   `json:"is_vp"`
	NodeID                int    `json:"node_id"`
	GrpcPort              int    `json:"grpc_port"`
	LocalIP               string `json:"local_ip"`
	JsonrpcPort           int    `json:"jsonrpc_port"`
	IntroducerIP          string `json:"introducer_ip"`
	IntroducerPort        int    `json:"introducer_port"`
	IntroducerJSONRPCPort int    `json:"introducer_jsonrpcport"`
	IntroducerID          int    `json:"introducer_id"`
}
