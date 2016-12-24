package peermessage

import (
	"hyperchain/p2p/peerComm"
	"strconv"
)

// PeerAddress the peer address
type PeerAddr struct {
	IP      string
	Port    int
	RpcPort int
	Hash    string
	ID      int
}


//NewPeerAddr return a PeerAddr which can be used inner
func NewPeerAddr(IP string, Port int, JSONRPCPort int,ID int) (peerAddr *PeerAddr){
	peerAddr.IP = IP
	peerAddr.Port = Port
	peerAddr.RpcPort = JSONRPCPort
	peerAddr.ID = ID
	peerAddr.Hash = peerComm.GetHash(strconv.Itoa(IP)+":"+strconv.Itoa(Port)+":"+strconv.Itoa(ID))
	return
}
//ToPeerAddress convert the PeerAddr into PeerAddress which can transfer on the internet
func (pa *PeerAddr) ToPeerAddress() (peerAddress *PeerAddress){
	peerAddress.IP = pa.IP
	peerAddress.ID = int32(pa.ID)
	peerAddress.Port = int32(pa.Port)
	peerAddress.RpcPort = int32(pa.RpcPort)
	peerAddress.Hash = pa.Hash
	return
}

// recover the peerAddress into PeerAddr
func RecoverPeerAddr(peerAddress PeerAddress)(pa *PeerAddr){
	pa.Hash = peerAddress.Hash
	pa.IP = peerAddress.IP
	pa.ID = int(peerAddress.ID)
	pa.RpcPort = int(peerAddress.RpcPort)
	pa.Port = (peerAddress.Port)
	return
}