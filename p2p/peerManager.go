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
	"encoding/hex"
	"hyperchain/event"
	"hyperchain/crypto"

	"github.com/op/go-logging"
	"hyperchain/p2p/transport"
)


const MAXPEERNODE = 4

var DESKEY = []byte("sfe023f_sefiel#fi32lf3e!")



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

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}

// GetClientId GetLocalClientId
func (this *GrpcPeerManager) GetClientId() common.Hash{
	addr := node.GetNodeAddr()
	addrString := addr.String()
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	return hasher.Hash(addrString)
	//return *new(common.Hash)

}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(path string, NodeId int, aliveChan chan bool,isTest bool,eventMux *event.TypeMux) {

	configs := peerComm.GetConfig(path)
	port, _ := strconv.Atoi(configs["port"+strconv.Itoa(NodeId)])
	// start local node
	this.localNode = node.NewNode(port,isTest,eventMux)
	log.Info("Local Node Hash:",hex.EncodeToString(this.GetClientId().Bytes()))
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
		log.Debug("node:",NodeId,"process connecting task...")
		nid := 1
		for range time.Tick(3 * time.Second) {
			status := alivePeerMap[nid]
			//log.Println("status map", nid, status)
			if !status {
				//if this node is not online, connect it
				peerIp   := configs["node"+strconv.Itoa(nid)]
				peerPort_s := configs["port"+strconv.Itoa(nid)]
				peerPort_i,err := strconv.Atoi(peerPort_s)
				if err != nil{
					log.Error("port cannot convert into int",err)
				}
				peerPort_i32 := int32(peerPort_i)
				peerAddress := pb.PeerAddress{
					Ip:peerIp,
					Port:peerPort_i32,
				}
				peerAddrString := peerIp + ":" + peerPort_s
				log.Info("Connecting to: ", peerAddrString)
				hasher := crypto.NewKeccak256Hash("keccak256Hanser")
				peerNodeHash := hex.EncodeToString(hasher.Hash(peerAddress.String()).Bytes())

				peer, peerErr := peer.NewPeerByString(peerAddrString)
				if peerErr != nil {
					// cannot connect to other peer
					log.Error("Node: "+peerAddrString+" can not connect!\n", peerErr)

					//continue
				} else {
					// add  peer to peer pool
					peerPort, _ := strconv.Atoi(configs["port"+strconv.Itoa(nid)])
					peerPool.PutPeer(pb.PeerAddress{
						Ip:   configs["node"+strconv.Itoa(nid)],
						Port: int32(peerPort),
					}, peer)
					alivePeerMap[nid] = true
					log.Debug("Peer Node hash:", peerNodeHash)
				}
			}
			nid += 1
			if nid > MAXPEERNODE{
				break
			}
		}
	}
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

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
	result, err := transport.TripleDesEncrypt(payLoad, DESKEY)
	if err!=nil{
		log.Fatal("TripleDesEncrypt Failed!")
	}
	localNodeAddr := node.GetNodeAddr()
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         &localNodeAddr,
		Payload:      result,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	pPool := peerPool.NewPeerPool(false, false)
	//go this.EventManager.PostEvent(pb.Message_CONSUS, broadCastMessage)
	go broadcast(broadCastMessage,&pPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(broadCastMessage pb.Message,pPool *peerPool.PeersPool){
	for _, peer := range pPool.GetPeers() {
		resMsg, err := peer.Chat(&broadCastMessage)
		if err != nil {
			log.Error("Broadcast failed,Node", peer.Addr)
		} else {

			log.Info("resMsg:", string(resMsg.Payload))
			//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
		}
	}
}
