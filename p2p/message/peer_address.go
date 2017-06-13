package message

import (
	"strconv"
)

// PeerAddress the peer address
type PeerAddr struct {
	IP      string
	Port    int
	RPCPort int
	Hash    string
	ID      int
}

//NewPeerAddr return a PeerAddr which can be used inner
func NewPeerAddr(IP string, Port int, JSONRPCPort int, ID int) *PeerAddr {
	var peerAddr PeerAddr
	peerAddr.IP = IP
	peerAddr.Port = Port
	peerAddr.RPCPort = JSONRPCPort
	peerAddr.ID = ID
	peerAddr.Hash = GetHash(strconv.Itoa(ID))
	return &peerAddr
}

//ToPeerAddress convert the PeerAddr into PeerAddress which can transfer on the internet
func (pa *PeerAddr) ToPeerAddress() *PeerAddress {
	var peerAddress PeerAddress
	peerAddress.IP = pa.IP
	peerAddress.ID = int32(pa.ID)
	peerAddress.Port = int32(pa.Port)
	peerAddress.RPCPort = int32(pa.RPCPort)
	peerAddress.Hash = pa.Hash
	return &peerAddress
}

// recover the peerAddress into PeerAddr
func RecoverPeerAddr(peerAddress *PeerAddress) *PeerAddr {
	var pa PeerAddr
	pa.Hash = peerAddress.Hash
	pa.IP = peerAddress.IP
	pa.ID = int(peerAddress.ID)
	pa.RPCPort = int(peerAddress.RPCPort)
	pa.Port = int(peerAddress.Port)
	return &pa
}
