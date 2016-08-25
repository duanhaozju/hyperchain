// gRPC manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package p2p

import (

	"hyperchain-alpha/common"
)


type PeerManager interface {

	// judge all peer are connected and return them
	JudgeAlivePeers(num int)(bool)
	GetAllPeers()([]*Peer)
	Start(path string,isFirst bool)
	GetClientId()common.Hash
	BroadcastPeers(payLoad []byte)

}

type Peer struct {


}

type  GrpcPeerManager struct{
	Message int


}

func (self *GrpcPeerManager)GetClientId()common.Hash{
	return *new(common.Hash)

}
func (self *GrpcPeerManager)Start(path string,isFirst bool)  {

}
func (self *GrpcPeerManager)JudgeAlivePeers(num int) bool  {

	return true
}

func (self *GrpcPeerManager)GetAllPeers()([]*Peer)  {
	return nil
}

func (self *GrpcPeerManager)BroadcastPeers(payLoad []byte)  {

}
