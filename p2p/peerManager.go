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
	"hyperchain-alpha/p2p/peerEventHandler"
	pb "hyperchain-alpha/p2p/peermessage"
	"time"
	"strconv"
	//"github.com/labstack/gommon/log"
	"hyperchain-alpha/p2p/peerPool"
	"log"

)

const MAXPEERNODE = 4

type PeerManager interface {
	// judge all peer are connected and return them
	JudgeAlivePeers(*chan int)
	GetAllPeers() []*peer.Peer
	Start(path string, NodeId int)
	GetClientId() common.Hash
	BroadcastPeers(payLoad []byte)
}

type GrpcPeerManager struct {
	EventManager *peerEventManager.PeerEventManager
	localNode    *node.Node
}

func (this *GrpcPeerManager) GetClientId() common.Hash {
	return *new(common.Hash)

}

func (this *GrpcPeerManager) Start(path string, NodeId int) {
	configs := peerComm.GetConfig(path)
	port := configs["port"+strconv.Itoa(NodeId)].(int)
	this.localNode = node.NewNode(port)

	peerPool := peerPool.NewPeerPool(false)

	for i :=1;i<=MAXPEERNODE;i++{
		if i == NodeId{
			continue
		}
		peerAddr := configs["node"+strconv.Itoa(NodeId)].(string) + ":" +configs["port"+strconv.Itoa(NodeId)].(string)
		peer,peerErr := peer.NewPeer(configs["node"+strconv.Itoa(NodeId)].(string) + ":" +configs["port"+strconv.Itoa(NodeId)].(string))
		if peerErr != nil {
			log.Println("节点:" + peerAddr +"无法连接\n",peerErr)
			continue
		}else{
			peerPool.PutPeer(pb.PeerAddress{
				Ip:configs["node"+strconv.Itoa(NodeId)].(string),
				Port:configs["port"+strconv.Itoa(NodeId)].(int32),
			},peer)
		}
	}

	// init the event manager
	this.EventManager = peerEventManager.NewPeerEventManager()
	this.EventManager.RegisterEvent(pb.Message_HELLO, peerEventHandler.NewHelloHandler())
	this.EventManager.RegisterEvent(pb.Message_RESPONSE, peerEventHandler.NewResponseHandler())
	this.EventManager.RegisterEvent(pb.Message_CONSUS, peerEventHandler.NewBroadCastHandler())
	this.EventManager.RegisterEvent(pb.Message_KEEPALIVE, peerEventHandler.NewKeepAliveHandler())
	this.EventManager.Start()

}
func (this *GrpcPeerManager) JudgeAlivePeers(c *chan bool){
	//TODO 判断所有的节点是否存活
	var keepAliveMessage = pb.Message{
		MessageType: pb.Message_KEEPALIVE,
		From:        node.GetNodeAddr(),
		Payload:      []byte("KeepAliveMessage"),
		MsgTimeStamp: time.Now().Unix(),
	}
	this.EventManager.PostEvent(pb.Message_KEEPALIVE, keepAliveMessage)
	peerPool := peerPool.NewPeerPool(false)
	for {
		if peerPool.GetAliveNodeNum() == MAXPEERNODE{
			c <- true
			break
		}
	}


}

func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	peerPool := peerPool.NewPeerPool(false)
	return peerPool.GetPeers()
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
