// gRPC manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25 Quan Chen
// change log:	modified the  PeerManager interface definition
//		implements GrpcPeerManager methods
package p2p

import (
	"hyperchain/common"
	node "hyperchain/p2p/node"
	peer "hyperchain/p2p/peer"
	"hyperchain/p2p/peerComm"
	"hyperchain/p2p/peerEventHandler"
	"hyperchain/p2p/peerEventManager"
	pb "hyperchain/p2p/peermessage"
	"strconv"
	"time"
	"hyperchain/p2p/peerPool"
	"log"
	"encoding/hex"
	"hyperchain/event"
	"fmt"
)

const MAXPEERNODE = 4

type PeerManager interface {
	// judge all peer are connected and return them
	//JudgeAlivePeers(*chan bool)
	GetAllPeers() []*peer.Peer
	Start(path string, NodeId int, aliveChan chan bool,isTest bool,eventMux *event.TypeMux)
	GetClientId() common.Hash
	BroadcastPeers(payLoad []byte)
}

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	EventManager *peerEventManager.PeerEventManager
	localNode    *node.Node
	aliveChain   *chan bool
}

// GetClientId GetLocalClientId
func (this *GrpcPeerManager) GetClientId() common.Hash {
	addr := node.GetNodeAddr()
	return common.BytesToHash([]byte(addr.String()))
	//return *new(common.Hash)

}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(path string, NodeId int, aliveChan chan bool,isTest bool,eventMux *event.TypeMux) {

	configs := peerComm.GetConfig(path)
	port, _ := strconv.Atoi(configs["port"+strconv.Itoa(NodeId)])
	// start local node
	this.localNode = node.NewNode(port,isTest,eventMux)
	log.Println("Local Node hash:",hex.EncodeToString(this.GetClientId().Bytes()))
	//
	this.aliveChain = &aliveChan

	// init the event manager
	this.EventManager = peerEventManager.NewPeerEventManager()

	this.EventManager.RegisterEvent(pb.Message_HELLO, peerEventHandler.NewHelloHandler())
	this.EventManager.RegisterEvent(pb.Message_RESPONSE, peerEventHandler.NewResponseHandler())
	this.EventManager.RegisterEvent(pb.Message_CONSUS, peerEventHandler.NewBroadCastHandler())
	this.EventManager.RegisterEvent(pb.Message_KEEPALIVE, peerEventHandler.NewKeepAliveHandler())

	this.EventManager.Start()

	// connect to peer
	// 如果进行单元测试,需要将参数设置为true
	peerPool := peerPool.NewPeerPool(isTest,!isTest)
	alivePeerMap := make(map[int]bool)
	for i := 1; i <= MAXPEERNODE; i++ {
		if i == NodeId {
			alivePeerMap[i] = true
		}else{
			alivePeerMap[i] = false
		}
	}
	// connect other peers
	for peerPool.GetAliveNodeNum() < MAXPEERNODE - 1{
		log.Println("node:",NodeId,"process connecting task...")
		nid := 1
		for range time.Tick(3 * time.Second) {
			status := alivePeerMap[nid]
			//log.Println("status map", nid, status)
			if !status {
				//if this node is not online connect it
				peerAddr := configs["node"+strconv.Itoa(nid)] + ":" + configs["port"+strconv.Itoa(nid)]
				log.Println("Connecting to: ", peerAddr)
				peer, peerErr := peer.NewPeer(peerAddr)
				if peerErr != nil {
					// cannot connect to other peer
					log.Println("Node: "+peerAddr+"can not connect!\n", peerErr)
					//continue
				} else {
					// add  peer to peer pool
					peerPort, _ := strconv.Atoi(configs["port"+strconv.Itoa(nid)])
					peerPool.PutPeer(pb.PeerAddress{
						Ip:   configs["node"+strconv.Itoa(nid)],
						Port: int32(peerPort),
					}, peer)
					alivePeerMap[nid] = true
					log.Println("Alive Peer Node ID:", peerPool.GetAliveNodeNum())
				}
			}
			nid += 1
			if nid > MAXPEERNODE{
				break
			}
		}
	}
	log.Println("##########################################")
	log.Println("#                                        #")
	log.Println("# All the nodes have been connected...   #")
	log.Println("#                                        #")
	log.Println("##########################################")

	*this.aliveChain <- true
}

//tell the main thread the peers are already

//func (this *GrpcPeerManager) JudgeAlivePeers(c *chan bool){
//	//TODO 判断所有节点是否存活
//	var keepAliveMessage = pb.Message{
//		MessageType: pb.Message_KEEPALIVE,
//		From:        node.GetNodeAddr(),
//		Payload:      []byte("KeepAliveMessage"),
//		MsgTimeStamp: time.Now().Unix(),
//	}
//	this.EventManager.PostEvent(pb.Message_KEEPALIVE, keepAliveMessage)
//	peerPool := peerPool.NewPeerPool(false)
//	for {
//		if peerPool.GetAliveNodeNum() == MAXPEERNODE{
//			c <- true
//			break
//		}
//	}
//
//
//}

// GetAllPeers get all connected peer in the peer pool
func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	peerPool := peerPool.NewPeerPool(false,false)
	return peerPool.GetPeers()
}

// BroadcastPeers Broadcast Massage to connected peers
func (this *GrpcPeerManager) BroadcastPeers(payLoad []byte) {
	log.Println("enter peer manager broadcast")
	localNodeAddr := node.GetNodeAddr()
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         &localNodeAddr,
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	pPool := peerPool.NewPeerPool(false, false)
	//go this.EventManager.PostEvent(pb.Message_CONSUS, broadCastMessage)
	go broadcast(broadCastMessage,pPool)
}

func broadcast(broadCastMessage pb.Message,pPool *peerPool.PeersPool){
	for _, peer := range pPool.GetPeers() {
		fmt.Println("广播....")
		resMsg, err := peer.Chat(&broadCastMessage)
		if err != nil {
			log.Println("Broadcast failed,Node", peer.Addr)
		} else {
			log.Println("resMsg:", string(resMsg.Payload))
			//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
		}
	}
}
