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
)

const MAXPEERNODE = 4

type PeerManager interface {
	// judge all peer are connected and return them
	//JudgeAlivePeers(*chan bool)
	GetAllPeers() []*peer.Peer
	Start(path string, NodeId int, aliveChan chan bool,isTest bool)
	GetClientId() common.Hash
	BroadcastPeers(payLoad []byte)
}

type GrpcPeerManager struct {
	EventManager *peerEventManager.PeerEventManager
	localNode    *node.Node
	aliveChain   *chan bool
}

func (this *GrpcPeerManager) GetClientId() common.Hash {
	addr := node.GetNodeAddr()
	return common.BytesToHash([]byte(addr.String()))
	//return *new(common.Hash)

}

func (this *GrpcPeerManager) Start(path string, NodeId int, aliveChan chan bool,isTest bool) {
	configs := peerComm.GetConfig(path)
	port, _ := strconv.Atoi(configs["port"+strconv.Itoa(NodeId)])
	// start local node
	this.localNode = node.NewNode(port,isTest)
	log.Println("本地节点HASH:",hex.EncodeToString(this.GetClientId().Bytes()))
	//
	this.aliveChain = &aliveChan

	// init the event manager
	this.EventManager = peerEventManager.NewPeerEventManager()

	this.EventManager.RegisterEvent(pb.Message_HELLO, peerEventHandler.NewHelloHandler(this.EventManager))
	this.EventManager.RegisterEvent(pb.Message_RESPONSE, peerEventHandler.NewResponseHandler(this.EventManager))
	this.EventManager.RegisterEvent(pb.Message_CONSUS, peerEventHandler.NewBroadCastHandler(this.EventManager))
	this.EventManager.RegisterEvent(pb.Message_KEEPALIVE, peerEventHandler.NewKeepAliveHandler(this.EventManager))

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
	log.Println("Status map:",alivePeerMap)
	// connect other peers
	for peerPool.GetAliveNodeNum() < MAXPEERNODE - 1{
		log.Println("node:",NodeId,"connecting...")
		nid := 1
		for range time.Tick(3 * time.Second) {
			status := alivePeerMap[nid]
			//log.Println("status map", nid, status)
			if !status {
				//if this node is not online connect it
				peerAddr := configs["node"+strconv.Itoa(nid)] + ":" + configs["port"+strconv.Itoa(nid)]
				log.Println("连接到:", peerAddr)
				peer, peerErr := peer.NewPeer(peerAddr)
				if peerErr != nil {
					// cannot connect to other peer
					log.Println("节点:"+peerAddr+"无法连接\n", peerErr)
					//continue
				} else {
					// add  peer to peer pool
					peerPort, _ := strconv.Atoi(configs["port"+strconv.Itoa(nid)])
					peerPool.PutPeer(pb.PeerAddress{
						Ip:   configs["node"+strconv.Itoa(nid)],
						Port: int32(peerPort),
					}, peer)
					alivePeerMap[nid] = true
					log.Println("Alive Peer Node num:", peerPool.GetAliveNodeNum())
				}
			}
			nid += 1
			if nid > MAXPEERNODE{
				break
			}
		}
	}
	log.Println("完成同步啦...")
	*this.aliveChain <- true
}

//tell the main thread the peers are already

//func (this *GrpcPeerManager) JudgeAlivePeers(c *chan bool){
//	//TODO 判断所有的节点是否存活
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

func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	peerPool := peerPool.NewPeerPool(false,false)
	return peerPool.GetPeers()
}


func (this *GrpcPeerManager) BroadcastPeers(payLoad []byte) {
	localNodeAddr := node.GetNodeAddr()
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         &localNodeAddr,
		Payload:      payLoad,
		MsgTimeStamp: time.Now().Unix(),
	}
	this.EventManager.PostEvent(pb.Message_CONSUS, broadCastMessage)


}
