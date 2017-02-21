package peermessage

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAddr(t *testing.T) {
	peeraddr := NewPeerAddr("127.0.0.1",8001,8081,1)
	assert.Equal(t,peeraddr.Port,8001)
	assert.Equal(t,peeraddr.ID,1)
	assert.Equal(t,peeraddr.RPCPort,8081)
	assert.Equal(t,peeraddr.IP,"127.0.0.1")
	//127.0.0.1:8001:1
	assert.Equal(t,peeraddr.Hash,"c605d50c3ed56902ec31492ed43b238b36526df5d2fd6153c1858051b6635f6e")
}

func TestPeerAddr_ToPeerAddress(t *testing.T) {
	peeraddr := NewPeerAddr("127.0.0.1",8001,8081,1)
	peeraddress := peeraddr.ToPeerAddress()
	assert.Equal(t,peeraddress.GetHash(),"c605d50c3ed56902ec31492ed43b238b36526df5d2fd6153c1858051b6635f6e")
	assert.Equal(t,peeraddress.GetID(),int32(1))
	assert.Equal(t,peeraddress.GetIP(),"127.0.0.1")
	assert.Equal(t,peeraddress.GetPort(),int32(8001))
	assert.Equal(t,peeraddress.GetRPCPort(),int32(8081))
}

func TestRecoverPeerAddr(t *testing.T) {
	pa := PeerAddress{
		IP:"127.0.0.1",
		Port:int32(8001),
		RPCPort:int32(8081),
		Hash:GetHash("127.0.0.1:8001:1"),
		ID:int32(1),
	}
	peeraddr := RecoverPeerAddr(&pa)
	assert.Equal(t,peeraddr.Port,8001)
	assert.Equal(t,peeraddr.ID,1)
	assert.Equal(t,peeraddr.RPCPort,8081)
	assert.Equal(t,peeraddr.IP,"127.0.0.1")
	//127.0.0.1:8001:1
	assert.Equal(t,peeraddr.Hash,"c605d50c3ed56902ec31492ed43b238b36526df5d2fd6153c1858051b6635f6e")
}