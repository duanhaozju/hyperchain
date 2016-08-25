// gRPC manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25 Quan Chen
// change log:	modified the  PeerManager interface definition
//		implements GrpcPeerManager methods
package p2p

import (
	"hyperchain-alpha/common"
	node "hyperchain-alpha/p2p/node"
	peer "hyperchain-alpha/p2p/peer"
	peerComm "hyperchain-alpha/p2p/peerComm"
	"hyperchain-alpha/p2p/peerEventManager"
	pb "hyperchain-alpha/p2p/peermessage"
	"time"
)

type PeerManager interface {
	// judge all peer are connected and return them
	JudgeAlivePeers(num int) bool
	GetAllPeers() []*peer.Peer
	Start(path string, isFirst bool)
	GetClientId() common.Hash
	BroadcastPeers(payLoad []byte)
}

type GrpcPeerManager struct {
	Message int

	EventManager *peerEventManager.PeerEventManager
	localNode    *node.Node
}

func (this *GrpcPeerManager) GetClientId() common.Hash {
	return *new(common.Hash)

}

func (this *GrpcPeerManager) Start(path string, isFirst bool) {
	configs := peerComm.GetConfig(path)
	port := configs["port"].(int)
	this.localNode = node.NewNode(port)

	if !isFirst {
		//TODO Connect the peers

	}

	// init the event manager
	this.EventManager = peerEventManager.NewPeerEventManager()
	this.EventManager.RegisterEvent(pb.Message_HELLO, NewHelloHandler())
	this.EventManager.RegisterEvent(pb.Message_RESPONSE, NewResponseHandler())
	this.EventManager.RegisterEvent(pb.Message_CONSUS, NewBroadCastHandler())
	this.EventManager.Start()

}
func (this *GrpcPeerManager) JudgeAlivePeers() bool {
	//TODO 判断所有的节点是否存活
	return true
}

func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	return nil
}

func (this *GrpcPeerManager) BroadcastPeers(payLoad []byte) {
	var broadCastMessage = pb.Message{
		MessageType: pb.Message_CONSUS,
		From:        node.GetNodeAddr(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().Unix(),
	}
	this.EventManager.PostEvent(pb.Message_CONSUS, broadCastMessage)

}
