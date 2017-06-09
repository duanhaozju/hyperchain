package p2p

type PeerMap interface {
	Get(identifier string) (*Peer, error)
}

type MsgHandler struct {
	peer *Peer

}

type peerMapImpl struct {
	peermap map[string]*Peer

}

