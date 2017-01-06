//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

const (
	ALIVE = iota
	PENDING
	STOP
)

type PeerInfo struct {

	Status int `json:"status"`
	IP     string `json:"ip"`
	Port   int `json:"port"`
	RPCPort int `json:"rpcport"`
	ID     int `json:"id"`
	IsPrimary bool `json:"isprimary"`
	Delay     int64 `json:"delay"`
}

type PeerInfos []PeerInfo

func NewPeerInfos(pis ...PeerInfo) PeerInfos {
	var peerInfos PeerInfos
	for _, pers := range pis {
		peerInfos = append(peerInfos, pers)
	}
	return peerInfos
}
func (this *PeerInfos) GetNumber() int {
	return len(*this)
}
