// author: chenquan
// date: 16-9-19
// last modified: 16-9-19 20:46
// last Modified Author: chenquan
// change log:
//
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

const MAX_PEER_NUM = 4

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	//本地节点
	LocalNode     *Node
	//节点池,相当于列表
	peersPool     *PeersPool
	//在peer /node中都需要使用,应当存储
	TEM           transport.TransportEncryptManager
	//连接节点的时候使用,不需要作为成员变量存在
	//peerStatus    map[uint64]bool
	//配置文件读取类
	configs       peerComm.Config
	//这个是动态变化的,不能作为初始化的时候的判断条件
	//MaxPeerNumber int
	//需要持有
	NodeID        uint64
	//这些信息都放在localAddr中就好了
	//Port          int64
	//IP            string
	//这个在peer和node中都要更新
	//Routers       pb.Routers
	//是否为创世节点,可能需要作为一个标识存在,但是
	Original      bool
	//是否上线
	IsOnline      bool
	//interducer information
	Introducer    pb.PeerAddress
	// N

	//N int

}

func NewGrpcManager(configPath string, nodeID int, isOriginal bool, introducerIP string, introducerPort uint64) *GrpcPeerManager {
	NodeID := uint64(nodeID)
	// configs
	var newgRPCManager GrpcPeerManager
	configUtil := peerComm.NewConfigUtil(configPath)
	newgRPCManager.configs = configUtil
	//newgRPCManager.MaxPeerNumber = newgRPCManager.configs.GetMaxPeerNumber()
	newgRPCManager.NodeID = NodeID
	newgRPCManager.LocalNode.N = MAX_PEER_NUM
	newgRPCManager.Original = isOriginal
	newgRPCManager.Introducer = peerComm.ExtractAddress(introducerIP, introducerPort, nodeID)
	//HSM only instanced once, so peersPool and Node Hsm are same instance
	newgRPCManager.TEM = transport.NewHandShakeManger()
	// start local node
	//init the flag map

	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(aliveChain chan int, eventMux *event.TypeMux) {
	if this.NodeID == 0 || this.configs == nil {
		log.Error("the gRPC Manager hasn't initlized")
		os.Exit(1)
	}
	//newgRPCManager.IP = newgRPCManager.configs.GetIP(newgRPCManager.NodeID)
	port := this.configs.GetPort(this.NodeID)
	this.LocalNode = NewNode(port, eventMux, this.NodeID, this.TEM, *this.peersPool)
	this.LocalNode.StartServer()
	// connect to peer
	// 如果进行单元测试,需要将参数设置为true
	// 重构peerpool 不采用单例模式进行管理
	this.peersPool = NewPeerPool(this.TEM)

	if this.Original {
		//连接其他节点
		this.connectToPeers()
		aliveChain <- 0
		this.IsOnline = true
	} else {
		//启动attend监听routine
		go this.LocalNode.attendNoticeProcess()
		//TODO 连接介绍人节点
		this.connectToIntroducer(this.Introducer)
		aliveChain <- 1
	}

	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

}

func (this *GrpcPeerManager)ConnectToOthers() {
	//TODO更新路由表之后进行连接
	allPeersWithTemp := this.peersPool.GetPeersWithTemp()
	newNodeMessage := pb.Message{
		MessageType:pb.Message_ATTEND,
		Payload:proto.Marshal(this.LocalNode.address),
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalNode.address,
	}
	for _, peer := range allPeersWithTemp {
		//review 返回值不做处理
		_, err := peer.Chat(newNodeMessage)
		if err != nil {
			log.Error("notice other node Attend Failed", err)
		}
	}
}

func (this *GrpcPeerManager) connectToIntroducer(introducerAddress pb.PeerAddress) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	peer, peerErr := NewPeerByIpAndPort(introducerAddress.IP, introducerAddress.Port, 0, this.TEM)
	//将介绍人的信息放入路由表中
	this.peersPool.PutPeer(peer)
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", introducerAddress.IP, ":", introducerAddress.Port, " can not connect!\n")
		return nil, peerErr
	} else {
		return peer, nil
	}

	//发送introduce 信息,取得路由表
	introduce_message := pb.Message{
		MessageType:pb.Message_INTRODUCE,
		Payload:proto.Marshal(this.LocalNode.address),
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalNode.address,
	}
	retMsg, sendErr := peer.Chat(introduce_message)
	if sendErr != nil {
		log.Error("get routing table error")
	}

	var routers pb.Routers
	unmarshalError := proto.Unmarshal(retMsg, routers)
	if unmarshalError != nil {
		log.Error("routing table unmarshal err ", unmarshalError)
	}
	this.peersPool.MergeFormRoutersToTemp(routers)
	this.LocalNode.N = len(this.GetAllPeersWithTemp())

}

func (this *GrpcPeerManager) connectToPeers() {
	var peerStatus  map[uint64]bool
	for i := 1; i <= MAX_PEER_NUM; i++ {
		_index := uint64(i)
		if _index == this.NodeID {
			peerStatus[_index] = true
		} else {
			peerStatus[_index] = false
		}
	}
	// connect other peers
	//TODO RETRY CONNECT 重试连接(未实现)
	for this.peersPool.GetAliveNodeNum() < MAX_PEER_NUM - 1 {
		log.Debug("node:", this.NodeID, "process connecting task...")
		log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
		nid := 1
		for range time.Tick(200 * time.Millisecond) {
			_index := uint64(nid)
			//log.Println("status map", nid, status)
			if nid > MAX_PEER_NUM {
				break
			}
			if peerStatus[_index] {
				nid++
				continue
			}
			//if this node is not online, connect it
			peerIp := this.configs.GetIP(_index)
			peerPort := this.configs.GetPort(_index)
			peerAddress := peerComm.ExtractAddress(peerIp, peerPort, _index)
			peer, connectErr := this.connectToPeer(peerAddress, _index)
			if connectErr != nil {
				// cannot connect to other peer
				log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n")

				continue
			} else {
				// add  peer to peer pool
				this.peersPool.PutPeer(*peerAddress, peer)
				//this.TEM.[peer.Addr.Hash]=peer.TEM
				peerStatus[_index] = true
				log.Debug("Peer Node hash:", peerAddress.Hash, "has connected!")
			}
		}
	}
	////todo 生成路由表
	//this.Routers = pb.Routers{
	//	Routers:this.peersPool.GetPeers(),
	//}

}

//connect to peer by ip address and port (why int32? because of protobuf limit)
func (this *GrpcPeerManager) connectToPeer(peerAddress *pb.PeerAddress, nid uint64) (*Peer, error) {
	//if this node is not online, connect it
	peer, peerErr := NewPeerByIpAndPort(peerAddress.IP, peerAddress.Port, nid, this.TEM)
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
func (this *GrpcPeerManager) GetAllPeersWithTemp() []*Peer {
	return this.peersPool.GetPeersWithTemp()
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
func broadcast(broadCastMessage pb.Message, pPool *PeersPool) {
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
	go func() {
		for _, p := range this.peersPool.GetPeers() {
			for _, NodeID := range peerList {
				// convert the uint64 to int
				// because the unicast node is not confirm so, here use double loop
				if p.ID == NodeID {
					log.Debug("send msg to ", NodeID)
					resMsg, err := p.Chat(syncMessage)
					if err != nil {
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
		perinfo.IP = per.RemoteAddr.IP
		perinfo.Port = per.RemoteAddr.Port
		retMsg, err := per.Client.Chat(context.Background(), &keepAliveMessage)
		if err != nil {
			perinfo.Status = STOP
		} else if retMsg.MessageType == pb.Message_RESPONSE {
			perinfo.Status = ALIVE
		} else if retMsg.MessageType == pb.Message_PENDING {
			perinfo.Status = PENDING
		}
		perinfo.IsPrimary = per.IsPrimary
		perinfo.Delay = this.LocalNode.DelayTable[per.ID]
		perinfo.ID = per.ID
		perinfos = append(perinfos, perinfo)
	}
	var self_info = PeerInfo{
		IP:        this.LocalNode.GetNodeAddr().IP,
		Port:      this.LocalNode.GetNodeAddr().Port,
		ID:        this.LocalNode.GetNodeAddr().ID,
		Status:    ALIVE,
		IsPrimary: this.LocalNode.IsPrimary,
		Delay:     this.LocalNode.DelayTable[this.NodeID],
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

func (this *GrpcPeerManager) UpdateRoutingTable(payload []byte) {
	//这里的payload 应该是前面传输过去的 address,里面应该只有一个

	var toUpdateAddress pb.PeerAddress
	err := proto.Unmarshal(payload, toUpdateAddress)
	if err != nil {
		log.Error(err)
	}
	//新节点peer
	newPeer := this.peersPool.peers[toUpdateAddress.Hash]
	//新消息
	attendResponseMsg := pb.Message{
		MessageType:pb.Message_ATTEND_RESPNSE,
		Payload:proto.Marshal(this.LocalNode.address),
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalNode.address,
	}

	if this.IsOnline {
		this.peersPool.MergeTempPeers(toUpdateAddress)
		//通知新节点进行接洽
		newPeer.Chat(attendResponseMsg)
		this.LocalNode.N += 1
	} else {
		//新节点
		//新节点在一开始的时候就已经将介绍人的节点列表加入了所以这里不需要处理
		//the new attend node
		this.IsOnline = true
		this.peersPool.MergeTempPeersForNewNode()
		this.LocalNode.N = this.peersPool.GetAliveNodeNum()
	}
}
//post
//this.LocalNode.higherEventManager.Post(event.RoutingTableUpdatedEvent{})
