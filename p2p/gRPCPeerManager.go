// author: chenquan
// date: 16-9-19
// last modified: 16-9-19 20:46
// last Modified Author: chenquan
// change log:
//
package p2p

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"hyperchain/event"
	node "hyperchain/p2p/node"
	peer "hyperchain/p2p/peer"
	"hyperchain/p2p/peerComm"
	"hyperchain/p2p/peerPool"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"os"
	"strconv"
	"time"
)

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	//localNodeHash
	LocalNode     *node.Node
	peersPool     *peerPool.PeersPool
	TEM           transport.TransportEncryptManager
	peerStatus    map[int]bool
	configs       peerComm.Config
	MaxPeerNumber int
	NodeId        int
	CName         string
	Port          int
	IP            string
}

func NewGrpcManager(configPath string, NodeId int) *GrpcPeerManager {
	// configs
	var newgRPCManager GrpcPeerManager
	configUtil := peerComm.NewConfigUtil(configPath)
	newgRPCManager.configs = configUtil
	newgRPCManager.MaxPeerNumber = newgRPCManager.configs.GetMaxPeerNumber()
	newgRPCManager.NodeId = NodeId
	newgRPCManager.IP = newgRPCManager.configs.GetIP(newgRPCManager.NodeId)
	newgRPCManager.Port = newgRPCManager.configs.GetPort(newgRPCManager.NodeId)
	newgRPCManager.CName = newgRPCManager.configs.GetCname(newgRPCManager.NodeId)
	newgRPCManager.TEM = transport.NewHandShakeManger()
	// start local node
	newgRPCManager.peerStatus = make(map[int]bool)
	//初始化flag map
	for i := 1; i <= newgRPCManager.MaxPeerNumber; i++ {
		if i == newgRPCManager.NodeId {
			newgRPCManager.peerStatus[i] = true
		} else {
			newgRPCManager.peerStatus[i] = false
		}
	}
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(aliveChain chan bool, eventMux *event.TypeMux) {
	if this.NodeId == 0 || this.configs == nil {
		log.Error("the gRPC Manager hasn't initlized")
		os.Exit(1)
	}
	this.LocalNode = node.NewNode(this.Port, eventMux, this.NodeId, this.CName, this.TEM)
	this.LocalNode.StartServer()
	// connect to peer
	// 如果进行单元测试,需要将参数设置为true
	// 重构peerpool 不采用单例模式进行管理
	this.peersPool = peerPool.NewPeerPool(this.TEM)
	// 读取待连接的节点信息
	this.connectToPeers()
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

	aliveChain <- true
}

func (this *GrpcPeerManager) connectToPeers() {
	// connect other peers
	//TODO RETRY CONNECT 重试连接(未实现)
	for this.peersPool.GetAliveNodeNum() < this.MaxPeerNumber-1 {
		log.Debug("node:", this.NodeId, "process connecting task...")
		log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
		nid := 1
		for range time.Tick(200 * time.Millisecond) {
			//log.Println("status map", nid, status)
			if nid > this.MaxPeerNumber {
				break
			}
			if this.peerStatus[nid] {
				nid++
				continue
			}
			//if this node is not online, connect it
			peerIp := this.configs.GetIP(nid)
			peerPort := this.configs.GetPort(nid)
			peerAddress := peerComm.ExtractAddress(peerIp, peerPort, int32(nid))
			peer, connectErr := this.connectToPeer(peerAddress, int32(nid))
			if connectErr != nil {
				// cannot connect to other peer
				log.Error("Node: ", peerAddress.Ip, ":", peerAddress.Port, " can not connect!\n")

				continue
			} else {
				// add  peer to peer pool
				this.peersPool.PutPeer(*peerAddress, peer)
				//this.TEM.[peer.Addr.Hash]=peer.TEM
				this.peerStatus[nid] = true
				log.Debug("Peer Node hash:", peerAddress.Hash, "has connected!")
			}
		}
	}
}

//connect to peer by ip address and port (why int32? because of protobuf limit)
func (this *GrpcPeerManager) connectToPeer(peerAddress *pb.PeerAddress, nid int32) (*peer.Peer, error) {
	//if this node is not online, connect it
	peer, peerErr := peer.NewPeerByIpAndPort(peerAddress.Ip, peerAddress.Port, nid, this.TEM, this.LocalNode.GetNodeAddr())
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", peerAddress.Address+" can not connect!\n")
		return nil, peerErr
	} else {
		return peer, nil
	}

}

// GetAllPeers get all connected peer in the peer pool
func (this *GrpcPeerManager) GetAllPeers() []*peer.Peer {
	return this.peersPool.GetPeers()
}

// BroadcastPeers Broadcast Massage to connected peers
func (this *GrpcPeerManager) BroadcastPeers(payLoad []byte) {
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         this.LocalNode.GetNodeAddr(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	go broadcast(broadCastMessage, this.peersPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(broadCastMessage pb.Message, pPool *peerPool.PeersPool) {
	for _, peer := range pPool.GetPeers() {
		//REVIEW 这里没有返回值,不知道本次通信是否成功
		//log.Notice(string(broadCastMessage.Payload))
		//TODO 其实这里不需要处理返回值，需要将其go起来
		//REVIEW Chat 方法必须要传实例，否则将会重复加密，请一定要注意！！
		//REVIEW Chat Function must give a message instance, not a point, if not the encrypt will break the payload!
		go peer.Chat(broadCastMessage)

	}

}

// SendMsgToPeers Send msg to specific peer peerlist
func (this *GrpcPeerManager) SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType) {
	var mpPaylod = &recovery.Message{
		MessageType:  MessageType,
		MsgTimeStamp: time.Now().UnixNano(),
		Payload:      payLoad,
	}
	realPayload, err := proto.Marshal(mpPaylod)
	if err != nil {
		log.Error("marshal failed")
	}
	localNodeAddr := this.LocalNode.GetNodeAddr()
	var syncMessage = pb.Message{
		MessageType:  pb.Message_SYNCMSG,
		From:         localNodeAddr,
		Payload:      realPayload,
		MsgTimeStamp: time.Now().UnixNano(),
	}

	// broadcast to special peers
	//TODO for stateUpdate
	go func() {
		for _, peer := range this.peersPool.GetPeers() {

			for _, nodeID := range peerList {
				nodeid := strconv.FormatUint(nodeID, 10)

				nid, _ := strconv.Atoi(nodeid)
				//if peerId==nodeID{
				if peer.ID == nid {
					log.Error(nid)
					resMsg, err := peer.Chat(syncMessage)
					if err != nil {
						log.Error("enter error")
						log.Error("Broadcast failed,Node", peer.Addr)
					} else {
						log.Info("resMsg:", string(resMsg.Payload))
						//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
					}
				}
			}

		}
	}()

}

func (this *GrpcPeerManager) GetPeerInfos() peer.PeerInfos {
	peers := this.peersPool.GetPeers()
	var perinfo peer.PeerInfo
	localNodeAddr := this.LocalNode.GetNodeAddr()

	var keepAliveMessage = pb.Message{
		MessageType:  pb.Message_KEEPALIVE,
		From:         localNodeAddr,
		Payload:      []byte("Query Status"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	var perinfos peer.PeerInfos
	for _, per := range peers {
		log.Debug("rage the peer")
		perinfo.IP = per.Addr.Ip
		perinfo.Port = int(per.Addr.Port)
		perinfo.CName = per.CName
		retMsg, err := per.Client.Chat(context.Background(), &keepAliveMessage)
		if err != nil {
			perinfo.Status = peer.STOP
		} else if retMsg.MessageType == pb.Message_RESPONSE {
			perinfo.Status = peer.ALIVE
		} else if retMsg.MessageType == pb.Message_PENDING {
			perinfo.Status = peer.PENDING
		}
		perinfos = append(perinfos, &perinfo)
	}
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (this *GrpcPeerManager) GetNodeId() int {
	return this.NodeId
}
