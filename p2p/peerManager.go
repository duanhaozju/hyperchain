package p2p

import (
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/common"
)

// TODO init the peer
// TODO init eventManager
// TODO init the handler
// TODO init the event register

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

func (self *GrpcPeerManager)GetClientId()common.Hash{
	return *new(common.Hash)

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
