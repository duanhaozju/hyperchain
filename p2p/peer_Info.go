//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

const (
	ALIVE = iota
	PENDING
	STOP
)

type PeerInfo struct {
	ID        int    `json:"id"`
	IP        string   `json:"ip"`
	Port      string    `json:"port"`
	Namespace   string `json:"namespace"`
	Hash        string `json:"hash"`
	Hostname    string `json:"hostname"`
	IsPrimary   bool	`json:"isPrimary"`
	IsVP        bool   `json:"isvp"`
	Status    int    `json:"status"`
	Delay     int64  `json:"delay"`
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
