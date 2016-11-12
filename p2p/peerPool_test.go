//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"testing"
	"hyperchain/p2p/transport"
	"hyperchain/p2p/peerComm"
	"github.com/stretchr/testify/assert"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/membersrvc"
)

var fakePeerPool *PeersPool

var fakePeer *Peer

var fakeAddr *pb.PeerAddress

func init(){
	membersrvc.Start("../../config/test/local_membersrvc.yaml",1)
	fakePeerPool = NewPeerPool(transport.NewHandShakeManger())
	fakeAddr = peerComm.ExtractAddress("127.0.0.1",int64(8001),uint64(1))
	fakePeer, _ = NewPeerByIpAndPort("127.0.0.1", int64(8001), uint64(1), transport.NewHandShakeManger(), fakeAddr)


}

func TestPeersPool_PutPeer(t *testing.T) {
	fakePeerPool.PutPeer(*fakeAddr,fakePeer)
	assert.Exactly(t,1,fakePeerPool.GetAliveNodeNum())
}

func TestDelPeer(t *testing.T) {
	tempPeer := fakePeerPool.GetPeer(*fakeAddr)
	assert.Exactly(t,fakePeer,tempPeer)
}