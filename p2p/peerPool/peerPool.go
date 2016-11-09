//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerPool

import (
	"errors"
	"github.com/op/go-logging"
	peer "hyperchain/p2p/peer"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"strconv"
	"strings"
)

type PeersPool struct {
	peers      map[string]*peer.Peer
	peerAddr   map[string]pb.PeerAddress
	peerKeys   map[pb.PeerAddress]string
	TEM        transport.TransportEncryptManager
	alivePeers int
}

// the peers pool instance
var prPoolIns PeersPool
var log *logging.Logger // package-level logger

//initialize the peers pool
func init() {
	log = logging.MustGetLogger("p2p/peerPool")
}

// NewPeerPool get a new peer pool instance
func NewPeerPool(TEM transport.TransportEncryptManager) *PeersPool {
	var newPrPoolIns PeersPool
	newPrPoolIns.peers = make(map[string]*peer.Peer)
	newPrPoolIns.peerAddr = make(map[string]pb.PeerAddress)
	newPrPoolIns.peerKeys = make(map[pb.PeerAddress]string)
	newPrPoolIns.TEM = TEM
	newPrPoolIns.alivePeers = 0
	return &newPrPoolIns

}

// PutPeer put a peer into the peer pool and get a peer point
func (this *PeersPool) PutPeer(addr pb.PeerAddress, client *peer.Peer) (*peer.Peer, error) {
	addrString := addr.Hash
	//log.Println("Add a peer:",addrString)
	if _, ok := this.peerKeys[addr]; ok {
		// the pool already has this client
		log.Error(addr.IP, addr.Port, "The client already in")
		return this.peers[addrString], errors.New("The client already in")

	} else {
		this.alivePeers += 1
		this.peerKeys[addr] = addrString
		this.peerAddr[addrString] = addr
		this.peers[addrString] = client
		return client, nil
	}

}

// GetPeer get a peer point by the peer address
func (this *PeersPool) GetPeer(addr pb.PeerAddress) *peer.Peer {
	if clientName, ok := this.peerKeys[addr]; ok {
		client := this.peers[clientName]
		return client
	} else {
		return nil
	}
}

// GetAliveNodeNum get all alive node num
func (this *PeersPool) GetAliveNodeNum() int {
	return this.alivePeers
}

// GetPeerByString get peer by address string
func GetPeerByString(addr string) *peer.Peer {
	address := strings.Split(addr, ":")
	p, err := strconv.Atoi(address[1])
	if err != nil {
		log.Error(`given string is not like "localhost:1234", pls check it`)
		return nil
	}
	pAddr := pb.PeerAddress{
		IP:   address[0],
		Port: int64(p),
	}
	if peerAddr, ok := prPoolIns.peerAddr[pAddr.String()]; ok {
		return prPoolIns.peers[prPoolIns.peerKeys[peerAddr]]
	} else {
		return nil
	}
}

// GetPeers  get peers from the peer pool
func (this *PeersPool) GetPeers() []*peer.Peer {
	var clients []*peer.Peer
	for _, cl := range this.peers {
		clients = append(clients, cl)
	}
	return clients
}

// DelPeer delete a peer by the given peer address (Not Used)
func DelPeer(addr pb.PeerAddress) {
	delete(prPoolIns.peers, prPoolIns.peerKeys[addr])
}
