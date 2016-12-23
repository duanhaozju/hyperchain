package peerComm
type PeerConfig struct {
	PeerNodes []PeerConfigNodes `json:"nodes"`
	Maxpeernode int `json:"maxpeernode"`
}
