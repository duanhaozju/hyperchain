package p2p

import "hyperchain-alpha/core/types"

type PeerManager interface {

	// judge all peer are connected and return them
	JudgeAlivePeers()(bool)
	GetAllPeers()([]*Peer)
	Start()
	BroadcastPeers(msg *types.Msg)

}


type Peer struct {

	
}
type  GrpcPeer struct{
	

}

func (self *GrpcPeer)GetAlivePeers()  {
	
}

func (self *GrpcPeer)BroadCastMsg()  {

}

