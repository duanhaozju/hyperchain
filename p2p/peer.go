package p2p

import (
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/common"
)
// manager grpc connect,client stands for local node,peers stand for remote servers
type PeerManager interface {

	// judge all peer are connected and return them
	JudgeAlivePeers()(bool)
	GetAllPeers()([]*Peer)
	Start()
	GetClientId()common.Hash
	BroadcastPeers(msg *types.Msg)

}


type Peer struct {

	
}
type  GrpcPeerManager struct{
	Message int
	

}
func (self *GrpcPeerManager)Start()  {

}
func (self *GrpcPeerManager)JudgeAlivePeers() bool  {

	return true
}

func (self *GrpcPeerManager)GetAllPeers()([]*Peer)  {
	return nil
}

func (self *GrpcPeerManager)BroadcastPeers(msg *types.Msg)  {

}

