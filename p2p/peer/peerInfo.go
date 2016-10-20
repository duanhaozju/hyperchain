// author: chenquan
// date: 16-9-13
// last modified: 16-9-13 14:18
// last Modified Author: chenquan
// change log:
//
package client

const (
	ALIVE = iota
	PENDING
	STOP
)

type PeerInfo struct {
	Status int
	IP     string
	Port   int64
	ID     uint64
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
