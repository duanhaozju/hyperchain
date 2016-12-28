package peermessage

import (
	"strconv"
	"encoding/hex"
	"hyperchain/crypto"
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
func NewPeerAddr(IP string, Port int, JSONRPCPort int,ID int) *PeerAddr{
	var peerAddr PeerAddr
	peerAddr.IP = IP
	peerAddr.Port = Port
	peerAddr.RpcPort = JSONRPCPort
	peerAddr.ID = ID
	peerAddr.Hash = getHash(IP+":"+strconv.Itoa(Port)+":"+strconv.Itoa(ID))
	return &peerAddr
}
//ToPeerAddress convert the PeerAddr into PeerAddress which can transfer on the internet
func (pa *PeerAddr) ToPeerAddress() *PeerAddress{
	var peerAddress PeerAddress
	peerAddress.IP = pa.IP
	peerAddress.ID = int32(pa.ID)
	peerAddress.Port = int32(pa.Port)
	peerAddress.RpcPort = int32(pa.RpcPort)
	peerAddress.Hash = pa.Hash
	return &peerAddress
}

// recover the peerAddress into PeerAddr
func RecoverPeerAddr(peerAddress *PeerAddress) *PeerAddr{
	var pa PeerAddr
	pa.Hash = peerAddress.Hash
	pa.IP = peerAddress.IP
	pa.ID = int(peerAddress.ID)
	pa.RpcPort = int(peerAddress.RpcPort)
	pa.Port = int(peerAddress.Port)
	return &pa
}

func getHash(needHashString string) string {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	return hex.EncodeToString(hasher.Hash(needHashString).Bytes())
}