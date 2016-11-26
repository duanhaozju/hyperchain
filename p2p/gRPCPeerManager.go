//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"hyperchain/event"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"os"
	"strconv"
	"time"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	//localNodeHash
	LocalNode     *Node
	peersPool     *PeersPool
	TEM           transport.TransportEncryptManager
	peerStatus    map[uint64]bool
	configs       peerComm.Config
	MaxPeerNumber int
	NodeID        uint64
	Port          int64
	IP            string
}

func NewGrpcManager(configPath string, nodeID int) *GrpcPeerManager {
	NodeID := uint64(nodeID)

	// configs
	var newgRPCManager GrpcPeerManager
	configUtil := peerComm.NewConfigUtil(configPath)
	newgRPCManager.configs = configUtil
	//get the maxpeer from config
	newgRPCManager.MaxPeerNumber = newgRPCManager.configs.GetMaxPeerNumber()
	newgRPCManager.NodeID = NodeID

	newgRPCManager.IP = newgRPCManager.configs.GetIP(newgRPCManager.NodeID)
	newgRPCManager.Port = newgRPCManager.configs.GetPort(newgRPCManager.NodeID)

	//HSM only instanced once, so peersPool and Node Hsm are same instance
	newgRPCManager.TEM = transport.NewHandShakeManger()
	// start local node
	newgRPCManager.peerStatus = make(map[uint64]bool)
	//init the flag map
	for i := 1; i <= newgRPCManager.MaxPeerNumber; i++ {
		_index := uint64(i)
		if _index == newgRPCManager.NodeID {
			newgRPCManager.peerStatus[_index] = true
		} else {
			newgRPCManager.peerStatus[_index] = false
		}
	}
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(aliveChain chan bool, eventMux *event.TypeMux, isReconnect bool) {
	if this.NodeID == 0 || this.configs == nil {
		log.Error("the gRPC Manager hasn't initlized")
		os.Exit(1)
	}
	// 重构peerpool 不采用单例模式进行管理
	this.peersPool = NewPeerPool(this.TEM)
	this.LocalNode = NewNode(this.Port, eventMux, this.NodeID, this.peersPool)
	this.LocalNode.StartServer()
	// connect to peer
	// 读取待连接的节点信息
	this.connectToPeers(isReconnect)

	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

	aliveChain <- true
}

func (this *GrpcPeerManager) connectToPeers(isReconnect bool) {
	// connect other peers
	//TODO RETRY CONNECT 重试连接(未实现)
	for this.peersPool.GetAliveNodeNum() < this.MaxPeerNumber-1 {
		log.Debug("node:", this.NodeID, "process connecting task...")
		log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
		nid := 1
		for range time.Tick(200 * time.Millisecond) {
			_index := uint64(nid)
			//log.Println("status map", nid, status)
			if nid > this.MaxPeerNumber {
				break
			}
			if this.peerStatus[_index] {
				nid++
				continue
			}
			//if this node is not online, connect it
			peerIp := this.configs.GetIP(_index)
			peerPort := this.configs.GetPort(_index)
			peerAddress := peerComm.ExtractAddress(peerIp, peerPort, _index)
			peer, connectErr := this.connectToPeer(peerAddress, _index, isReconnect)
			if connectErr != nil {
				// cannot connect to other peer
				log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)

				continue
			} else {
				// add  peer to peer pool
				this.peersPool.PutPeer(*peerAddress, peer)
				//this.TEM.[peer.Addr.Hash]=peer.TEM
				this.peerStatus[_index] = true
				log.Debug("Peer Node hash:", peerAddress.Hash, "has connected!")
			}
		}
	}
}




//connect to peer by ip address and port (why int32? because of protobuf limit)
func (this *GrpcPeerManager) connectToPeer(peerAddress *pb.PeerAddress, nid uint64, isReconnect bool) (*Peer, error) {
	//if this node is not online, connect it
	var peer *Peer
	var peerErr error
	if isReconnect {
		peer, peerErr = NewPeerByIpAndPortReconnect(peerAddress.IP, peerAddress.Port, nid, this.TEM, this.LocalNode.GetNodeAddr(), this.peersPool)

	} else {
		peer, peerErr = NewPeerByIpAndPort(peerAddress.IP, peerAddress.Port, nid, this.TEM, this.LocalNode.GetNodeAddr(), this.peersPool)
	}
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n")
		return nil, peerErr
	} else {
		return peer, nil
	}

}

// GetAllPeers get all connected peer in the peer pool
func (this *GrpcPeerManager) GetAllPeers() []*Peer {
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
	go broadcast(this,broadCastMessage, this.peersPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(grpcPeerManager *GrpcPeerManager,broadCastMessage pb.Message, pPool *PeersPool) {
	for _, peer := range pPool.GetPeers() {
		//REVIEW 这里没有返回值,不知道本次通信是否成功
		//log.Notice(string(broadCastMessage.Payload))
		//TODO 其实这里不需要处理返回值，需要将其go起来
		//REVIEW Chat 方法必须要传实例，否则将会重复加密，请一定要注意！！
		//REVIEW Chat Function must give a message instance, not a point, if not the encrypt will break the payload!
		go func(){
			start := time.Now().UnixNano()
			_,err := peer.Chat(broadCastMessage)
			if err != nil {
				grpcPeerManager.LocalNode.DelayChan <- UpdateTable{updateID:peer.Addr.ID,updateTime:time.Now().UnixNano() - start}
			}
		}()


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
	go func() {
		for _, p := range this.peersPool.GetPeers() {
			for _, NodeID := range peerList {
				// convert the uint64 to int
				// because the unicast node is not confirm so, here use double loop
				if p.ID == NodeID {
					log.Debug("send msg to ", NodeID)
					start := time.Now().UnixNano()
					resMsg, err := p.Chat(syncMessage)

					if err != nil {
						this.LocalNode.DelayChan <- UpdateTable{updateID:p.Addr.ID,updateTime:time.Now().UnixNano() - start}
						log.Error("Broadcast failed,Node", p.Addr)
					} else {
						log.Debug("resMsg:", string(resMsg.Payload))
						//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
					}
				}
			}

		}
	}()

}

func (this *GrpcPeerManager) GetPeerInfo() PeerInfos {
	peers := this.peersPool.GetPeers()
	localNodeAddr := this.LocalNode.GetNodeAddr()

	var keepAliveMessage = pb.Message{
		MessageType:  pb.Message_KEEPALIVE,
		From:         localNodeAddr,
		Payload:      []byte("Query Status"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	var perinfos PeerInfos
	for _, per := range peers {
		var perinfo PeerInfo
		log.Debug("rage the peer")
		perinfo.IP = per.Addr.IP
		perinfo.Port = per.Addr.Port
		retMsg, err := per.Client.Chat(context.Background(), &keepAliveMessage)
		if err != nil {
			perinfo.Status = STOP
		} else if retMsg.MessageType == pb.Message_RESPONSE {
			perinfo.Status = ALIVE
		} else if retMsg.MessageType == pb.Message_PENDING {
			perinfo.Status = PENDING
		}
		perinfo.IsPrimary = per.IsPrimary
		this.LocalNode.delayTableMutex.RLock()
		perinfo.Delay = this.LocalNode.delayTable[per.ID]
		this.LocalNode.delayTableMutex.RUnlock()
		perinfo.ID = per.ID
		perinfos = append(perinfos, perinfo)
	}
	var self_info = PeerInfo{
		IP:        this.LocalNode.GetNodeAddr().IP,
		Port:      this.LocalNode.GetNodeAddr().Port,
		ID:        this.LocalNode.GetNodeAddr().ID,
		Status:    ALIVE,
		IsPrimary: this.LocalNode.IsPrimary,
		Delay:     this.LocalNode.delayTable[this.NodeID],
	}
	perinfos = append(perinfos, self_info)
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (this *GrpcPeerManager) GetNodeId() int {
	_id := strconv.FormatUint(this.NodeID, 10)
	_node_id, _err := strconv.Atoi(_id)
	if _err != nil {
		log.Error("convert err", _err)
	}
	return _node_id
}

func (this *GrpcPeerManager) SetPrimary(id uint64) error {
	peers := this.peersPool.GetPeers()
	this.LocalNode.IsPrimary = false
	for _, per := range peers {
		per.IsPrimary = false
	}

	for _, per := range peers {
		if per.ID == id {
			per.IsPrimary = true
		}
		if this.NodeID == id {
			this.LocalNode.IsPrimary = true
		}
	}
	return nil
}

func (this *GrpcPeerManager) GetLocalNode() *Node {
	return this.LocalNode
}
