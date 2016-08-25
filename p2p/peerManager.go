// gRPC manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package p2p

import (
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/common"
	peer "hyperchain-alpha/p2p/peer"
	pb "hyperchain-alpha/p2p/peermessage"
	"hyperchain-alpha/p2p/peerEventManager"
	node "hyperchain-alpha/p2p/node"
	"time"
)


type PeerManager interface {
	// judge all peer are connected and return them
	JudgeAlivePeers(num int)(bool)
	GetAllPeers()([]*peer.Peer)
	Start(port int)
	GetClientId()common.Hash
	BroadcastPeers(payLoad []byte)

}

type Peer struct {


}

type  GrpcPeerManager struct{
	Message int

	EventManager *peerEventManager.PeerEventManager
	localNode *node.Node

}

func (this *GrpcPeerManager)GetClientId()common.Hash{
	return *new(common.Hash)

}
func (this *GrpcPeerManager)Start(port int)  {
	// start the grpc server
	this.localNode = node.NewNode(port)

	// init the event manager
	this.EventManager = peerEventManager.NewPeerEventManager()
	this.EventManager.RegisterEvent(pb.Message_HELLO,NewHelloHandler())
	this.EventManager.RegisterEvent(pb.Message_CONSUS,NewBroadCastHandler())
	this.EventManager.Start()

}
func (this *GrpcPeerManager)JudgeAlivePeers() bool  {
	//TODO 判断所有的节点是否存活
	return true
}

func (this *GrpcPeerManager)GetAllPeers()([]*peer.Peer)  {
	return nil
}

func (this *GrpcPeerManager)BroadcastPeers(msg *types.Msg)  {
	var broadCastMessage = pb.Message{
		MessageType:pb.Message_CONSUS,
		From:this.localNode,
		// TODO packaging the msg into payload
		Payload:[]byte("hhhh"),
		MsgTimeStamp: time.Now().Unix(),
	}
	this.EventManager.PostEvent(pb.Message_CONSUS,broadCastMessage)

}
