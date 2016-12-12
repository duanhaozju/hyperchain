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
	"encoding/hex"
	"hyperchain/crypto"
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

func NewGrpcManager(configPath string, nodeID int, isOriginal bool, introducerIP string, introducerPort int64,introducer_ID uint64) *GrpcPeerManager {
	NodeID := uint64(nodeID)
	// configs
	var newgRPCManager GrpcPeerManager
	configUtil := peerComm.NewConfigUtil(configPath)
	newgRPCManager.configs = configUtil
	//get the maxpeer from config
	newgRPCManager.NodeID = NodeID

	newgRPCManager.Original = isOriginal
	newgRPCManager.Introducer = *peerComm.ExtractAddress(introducerIP, introducerPort, introducer_ID)
	//HSM only instanced once, so peersPool and Node Hsm are same instance
	newgRPCManager.TEM = transport.NewHandShakeManger()
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(aliveChain chan int, eventMux *event.TypeMux, isReconnect bool, GRPCProt int64) {
	if this.NodeID == 0 || this.configs == nil {
		log.Error("the gRPC Manager hasn't initlized")
		os.Exit(1)
	}
	//newgRPCManager.IP = newgRPCManager.configs.GetIP(newgRPCManager.NodeID)
	// 重构peerpool 不采用单例模式进行管理
	port := this.configs.GetPort(this.NodeID)
	if port == int64(0) {
		port = GRPCProt
	}

	this.peersPool = NewPeerPool(this.TEM,GRPCProt,this.NodeID)
	this.LocalNode = NewNode(port, eventMux, this.NodeID, this.TEM, this.peersPool)
	this.LocalNode.StartServer()
	this.LocalNode.N = MAX_PEER_NUM
	// connect to peer
	if this.Original {
		// 读取待连接的节点信息
		this.connectToPeers(isReconnect)
		//log.Critical("路由表:", this.peersPool.peerAddr)

		aliveChain <- 0
		this.IsOnline = true
	} else {
		//启动attend监听routine
		go this.LocalNode.attendNoticeProcess(this.LocalNode.N)
		//TODO 连接介绍人节点
		this.connectToIntroducer(this.Introducer)
		//this.ConnectToOthers()
		aliveChain <- 1
	}



	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  │")
	log.Notice("└────────────────────────────┘")

}

func (this *GrpcPeerManager)ConnectToOthers() {
	//TODO更新路由表之后进行连接
	allPeersWithTemp := this.peersPool.GetPeersWithTemp()
	payload, _ := proto.Marshal(this.LocalNode.address)
	newNodeMessage := pb.Message{
		MessageType:pb.Message_ATTEND,
		Payload:payload,
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

func (this *GrpcPeerManager) connectToIntroducer( introducerAddress pb.PeerAddress) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	peer, peerErr := NewPeerByIpAndPort(introducerAddress.IP, introducerAddress.Port, introducerAddress.ID, this.TEM, this.LocalNode.address,this.peersPool)
	//将介绍人的信息放入路由表中
	this.peersPool.PutPeer(*peer.RemoteAddr,peer)
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", introducerAddress.IP, ":", introducerAddress.Port, " casn not connect!\n")
		return
	}

	//发送introduce 信息,取得路由表
	payload, _ := proto.Marshal(this.LocalNode.address)
	log.Warning("address: ", this.LocalNode.address)
	introduce_message := pb.Message{
		MessageType:pb.Message_INTRODUCE,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalNode.address,
	}
	retMsg, sendErr := peer.Chat(introduce_message)
	if sendErr != nil {
		log.Error("get routing table error")
	}

	var routers pb.Routers
	unmarshalError := proto.Unmarshal(retMsg.Payload, &routers)
	if unmarshalError != nil {
		log.Error("routing table unmarshal err ", unmarshalError)
	}
	log.Warning("合并路由表并链接", routers)
	this.peersPool.MergeFormRoutersToTemp(routers)
	for _, p := range this.peersPool.GetPeersWithTemp() {
		log.Warning("路由表中的节点", p)
		//review 		this.LocalNode.attendChan <- 1
		//attend_message := pb.Message{
		//	MessageType:pb.Message_ATTEND,
		//	Payload:payload,
		//	MsgTimeStamp:time.Now().UnixNano(),
		//	From:this.LocalNode.address,
		//}
		//retMsg,err := p.Chat(attend_message)
		//if err != nil{
		//	log.Error(err)
		//}else{
		//	retMsg
		//}
	}
	this.LocalNode.N = len(this.GetAllPeersWithTemp())
}

func (this *GrpcPeerManager) connectToPeers(isReconnect bool) {
	var peerStatus  map[uint64]bool
	peerStatus = make(map[uint64]bool)
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
		log.Debug("node:", this.NodeID, "连接节点...")
		log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
		nid := 1
		for range time.Tick(200 * time.Millisecond) {
			_index := uint64(nid)

			if nid > MAX_PEER_NUM {
				break
			}
			if peerStatus[_index] {
				nid++
				continue
			}
			log.Debug("status map", nid, peerStatus[_index])
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
				//log.Critical("将地址加入到地址池", *peerAddress)
				this.peersPool.PutPeer(*peerAddress, peer)
				//this.TEM.[peer.Addr.Hash]=peer.TEM
				peerStatus[_index] = true
				log.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
				log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
			}

		}
	}
	////todo 生成路由表
	//this.Routers = pb.Routers{
	//	Routers:this.peersPool.GetPeers(),
	//}

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
		//log.Critical("连接到节点", nid,isReconnect)
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
	//log.Warning("P2P broadcast")
	if !this.IsOnline {
		log.Warning("IsOnline")
		return
	}
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         this.LocalNode.GetNodeAddr(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(this, broadCastMessage, this.peersPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(grpcPeerManager *GrpcPeerManager,broadCastMessage pb.Message, pPool *PeersPool) {
	for _, peer := range pPool.GetPeers() {
		//REVIEW 这里没有返回值,不知道本次通信是否成功
		//log.Notice(string(broadCastMessage.Payload))
		//TODO 其实这里不需要处理返回值，需要将其go起来
		//REVIEW Chat 方法必须要传实例，否则将会重复加密，请一定要注意！！
		//REVIEW Chat Function must give a message instance, not a point, if not the encrypt will break the payload!
		go func(p2 *Peer) {
			start := time.Now().UnixNano()
			_, err := p2.Chat(broadCastMessage)
			if err == nil {
				grpcPeerManager.LocalNode.DelayChan <- UpdateTable{updateID:p2.Addr.ID, updateTime:time.Now().UnixNano() - start}
			} else {
				log.Error("chat failed", err);
			}
		}(peer)

	}
}
func (this *GrpcPeerManager) SetOnline() {
	this.IsOnline = true
	this.peersPool.MergeTempPeersForNewNode()
}

func (this *GrpcPeerManager) GetLocalAddressPayload() (payload []byte) {
	payload, _ = proto.Marshal(this.LocalNode.address)
	return
}

// SendMsgToPeers Send msg to specific peer peerlist
func (this *GrpcPeerManager) SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType) {
	//log.Critical("need send message to ", peerList)
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
	for _, NodeID := range peerList {
		peers := this.peersPool.GetPeers()
		for _, p := range peers {
			//log.Critical("range nodeid", p.RemoteAddr)
			//log.Critical("range nodeid", p.ID)
			// convert the uint64 to int
			// because the unicast node is not confirm so, here use double loop
			if p.ID == NodeID {
				log.Debug("send msg to ", NodeID)
				start := time.Now().UnixNano()
				resMsg, err := p.Chat(syncMessage)
				if err != nil {
					log.Error("Broadcast failed,Node", p.Addr)
				} else {
					this.LocalNode.DelayChan <- UpdateTable{updateID:p.Addr.ID,updateTime:time.Now().UnixNano() - start}
					log.Debug("resMsg:", string(resMsg.Payload))
					//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
				}
			}
		}

	}

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

func (this *GrpcPeerManager) UpdateRoutingTable(payload []byte) {
	//这里的payload 应该是前面传输过去的 address,里面应该只有一个

	var toUpdateAddress pb.PeerAddress
	err := proto.Unmarshal(payload, &toUpdateAddress)
	if err != nil {
		log.Error(err)
	}
	//新节点peer
	//newPeer := this.peersPool.tempPeers[this.peersPool.tempPeerKeys[toUpdateAddress]]
	log.Debugf("hash: %v",toUpdateAddress )
	newPeer, err := NewPeerByAddress(&toUpdateAddress, toUpdateAddress.ID, this.TEM, this.LocalNode.address)
	if err != nil {
		log.Error(err)
	}

	log.Debug("newPeer: %v", newPeer)
	//新消息
	payload, _ = proto.Marshal(this.LocalNode.address)

	attendResponseMsg := pb.Message{
		MessageType:pb.Message_ATTEND_RESPNSE,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalNode.address,
	}

	if this.IsOnline {
		this.peersPool.MergeTempPeers(newPeer)
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

/*********************************
 * delete LocalNode part
 ********************************/
func (this *GrpcPeerManager) GetLocalNodeHash() string{
	return this.LocalNode.address.Hash
}

func (this *GrpcPeerManager) GetRouterHashifDelete(hash string) (string,uint64){
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	routers := this.peersPool.ToRoutingTableWithout(hash)
	hash = hex.EncodeToString(hasher.Hash(routers).Bytes())

	var ID uint64
	localHash := this.LocalNode.address.Hash
	for _,rs := range routers.Routers{
		log.Debug("RS hash: ", rs.Hash)
		if rs.Hash == localHash{
			log.Notice("rs hash: ", rs.Hash)
			log.Notice("id: ", rs.ID)
			ID=rs.ID;
		}
	}
	return hex.EncodeToString(hasher.Hash(routers).Bytes()),ID
}


func (this *GrpcPeerManager)  DeleteNode(hash string) error{

	if this.LocalNode.address.Hash == hash {
		// delete local node and stop all server
		this.LocalNode.StopServer()

	} else{
		// delete the specific node
		for _,pers := range this.peersPool.GetPeers(){
			if pers.Addr.Hash == hash{
				this.peersPool.DeletePeer(pers)
			}
		}
		//TODO update node id
		hasher := crypto.NewKeccak256Hash("keccak256Hanser")
		routers := this.peersPool.ToRoutingTableWithout(hash)
		hash = hex.EncodeToString(hasher.Hash(routers).Bytes())

		if hash == this.LocalNode.address.Hash {
			log.Critical("THIS NODE WAS BEEN CLOSED...")
			return nil
		}

		for _,per :=range this.peersPool.GetPeers(){
			if per.Addr.Hash == hash{
				this.peersPool.DeletePeer(per)
			}else{
				for _,router := range routers.Routers{
					if router.Hash == per.Addr.Hash{
						per.Addr = *peerComm.ExtractAddress(router.IP,router.Port,router.ID)
					}
				}
			}

		}
		return nil


	}
	return nil
}