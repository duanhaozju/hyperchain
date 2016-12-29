//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
//changelog:修改新的handshankemanager的实现。
//last Modified Author: zhangkejie
//date:2016-12-23
package p2p

import (
	"hyperchain/event"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"strconv"
	"time"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"encoding/hex"
	"hyperchain/crypto"
	"hyperchain/common"
	"math"
)

const MAX_PEER_NUM = 4

// gRPC peer manager struct, which to manage the gRPC peers
type GrpcPeerManager struct {
	LocalAddr  *pb.PeerAddr
	LocalNode  *Node
	peersPool  PeersPool
	TEM        transport.TransportEncryptManager
	configs    peerComm.Config
	IsOriginal bool
	IsOnline   bool
	//interducer information
	Introducer *pb.PeerAddr

}

func NewGrpcManager(conf *common.Config) *GrpcPeerManager {
	var newgRPCManager GrpcPeerManager
	config := peerComm.NewConfigReader(conf.GetString("global.configs.peers"))
	//log.Critical(config.GetLocalJsonRPCPort())
	// configs
	newgRPCManager.configs = config

	newgRPCManager.LocalAddr = pb.NewPeerAddr(config.GetLocalIP(),config.GetLocalGRPCPort(),config.GetLocalJsonRPCPort(),config.GetLocalID())
	//log.Critical("local ID",newgRPCManager.LocalAddr.ID)
	//get the maxpeer from config
	newgRPCManager.IsOriginal = config.IsOrigin()
	newgRPCManager.Introducer = pb.NewPeerAddr(config.GetIntroducerIP(),config.GetIntroducerPort(),config.GetIntroducerJSONRPCPort(),config.GetIntroducerID())
	//HSM only instanced once, so peersPool and Node Hsm are same instance

	//handshakemanager
	newgRPCManager.TEM = transport.NewHandShakeMangerNew();
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GrpcPeerManager) Start(aliveChain chan int, eventMux *event.TypeMux) {
	if this.LocalAddr.ID == 0 || this.configs == nil {
		panic("the PeerManager hasn't initlized")
	}
	this.peersPool = NewPeerPoolIml(this.TEM,this.LocalAddr)
	this.LocalNode = NewNode(this.LocalAddr,eventMux,this.TEM, this.peersPool)
	this.LocalNode.StartServer()
	this.LocalNode.N = MAX_PEER_NUM
	// connect to peer
	if this.IsOriginal {
		// load the waiting connecting node information
		this.connectToPeers()
		//log.Critical("Routing table:", this.peersPool.peerAddr)

		aliveChain <- 0
		this.IsOnline = true
	} else {
		// start attend routine
		go this.LocalNode.attendNoticeProcess(this.LocalNode.N)
		//review connect to introducer
		this.connectToIntroducer(*this.Introducer)
		aliveChain <- 1
	}
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  |")
	log.Notice("└────────────────────────────┘")

}

func (this *GrpcPeerManager)ConnectToOthers() {
	//TODO更新路由表之后进行连接
	allPeersWithTemp := this.peersPool.GetPeersWithTemp()
	payload, _ := proto.Marshal(this.LocalAddr.ToPeerAddress())
	newNodeMessage := pb.Message{
		MessageType:pb.Message_ATTEND,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalAddr.ToPeerAddress(),
	}
	for _, peer := range allPeersWithTemp {
		//review 返回值不做处理
		_, err := peer.Chat(newNodeMessage)
		if err != nil {
			log.Error("notice other node Attend Failed", err)
		}
	}
}

func (this *GrpcPeerManager) connectToIntroducer( introducerAddress pb.PeerAddr) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	peer, peerErr := NewPeer(&introducerAddress,this.LocalAddr,this.TEM)
	//将介绍人的信息放入路由表中
	this.peersPool.PutPeer(*peer.PeerAddr,peer)
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", introducerAddress.IP, ":", introducerAddress.Port, " casn not connect!\n")
		return
	}

	//发送introduce 信息,取得路由表
	payload, _ := proto.Marshal(this.LocalAddr.ToPeerAddress())
	log.Warning("address: ", this.LocalAddr)
	introduce_message := pb.Message{
		MessageType:pb.Message_INTRODUCE,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalAddr.ToPeerAddress(),
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
	this.peersPool.MergeFromRoutersToTemp(routers)
	for _, p := range this.peersPool.GetPeersWithTemp() {
		log.Warning("路由表中的节点", p)
		//review 		this.LocalNode.attendChan <- 1
		attend_message := pb.Message{
			MessageType:pb.Message_ATTEND,
			Payload:payload,
			MsgTimeStamp:time.Now().UnixNano(),
			From:this.LocalNode.localAddr.ToPeerAddress(),
		}
		retMsg,err := p.Chat(attend_message)
		if err != nil{
			log.Error(err)
		}else{
			retMsg
		}
	}
	this.LocalNode.N = len(this.GetAllPeersWithTemp())
}

func (this *GrpcPeerManager) connectToPeers() {
	var peerStatus  map[int]bool
	peerStatus = make(map[int]bool)
	for i := 1; i <= MAX_PEER_NUM; i++ {
		if i == this.LocalAddr.ID {
			peerStatus[i] = true
		} else {
			peerStatus[i] = false
		}
	}
	N := MAX_PEER_NUM
	F := int(math.Floor((MAX_PEER_NUM - 1)/3))

	MaxNum := 0
	if this.configs.IsOrigin() {
		MaxNum = N-1
	}else{
		MaxNum = N - F - 1
	}

	// connect other peers
	//TODO RETRY CONNECT 重试连接(未实现)
	for this.peersPool.GetAliveNodeNum() < MaxNum{
		log.Debug("node:", this.LocalAddr.ID, "连接节点...")
		log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
		nid := 1
		for range time.Tick(200 * time.Millisecond) {

			_index := nid

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
			//peerAddress := peerComm.ExtractAddress(peerIp, peerPort, _index)
			//TODO fix the getJSONRPC PORT
			peerAddress := pb.NewPeerAddr(peerIp, peerPort, this.configs.GetPort(_index)+80,_index)
			peer, connectErr := this.connectToPeer(peerAddress, _index)
			if connectErr != nil {
				// cannot connect to other peer
				log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
				continue
			} else {
				// add  peer to peer pool
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
func (this *GrpcPeerManager) connectToPeer(peerAddress *pb.PeerAddr, nid int) (*Peer, error) {
	//if this node is not online, connect it
	var peer *Peer
	var peerErr error
	//if isReconnect {
	//	peer, peerErr = NewPeerReconnect(peerAddress,this.LocalAddr,this.TEM)
	//} else {
	peer, peerErr = NewPeer(peerAddress,this.LocalAddr,this.TEM)
	//}

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
		From:         this.LocalNode.GetNodeAddr().ToPeerAddress(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(this, broadCastMessage, this.peersPool)
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(grpcPeerManager *GrpcPeerManager,broadCastMessage pb.Message, pPool PeersPool) {
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
				grpcPeerManager.LocalNode.DelayChan <- UpdateTable{updateID:p2.LocalAddr.ID, updateTime:time.Now().UnixNano() - start}
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
	payload, _ = proto.Marshal(this.LocalAddr.ToPeerAddress())
	return
}

// SendMsgToPeers Send msg to specific peer peerlist
func (this *GrpcPeerManager) SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType) {

	log.Debug("need send message to ", peerList)
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
		From:         localNodeAddr.ToPeerAddress(),
		Payload:      realPayload,
		MsgTimeStamp: time.Now().UnixNano(),
	}

	// broadcast to special peers
	for _, NodeID := range peerList {
		peers := this.peersPool.GetPeers()
		for _, p := range peers {
			log.Debug("range nodeid", p.PeerAddr)
			 //convert the uint64 to int
			 //because the unicast node is not confirm so, here use double loop
			if p.PeerAddr.ID == int(NodeID) {
				log.Debug("send msg to ", NodeID)
				start := time.Now().UnixNano()
				resMsg, err := p.Chat(syncMessage)
				if err != nil {
					log.Error("Broadcast failed,Node", p.LocalAddr.ID)
				} else {
					this.LocalNode.DelayChan <- UpdateTable{updateID:p.LocalAddr.ID,updateTime:time.Now().UnixNano() - start}
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
		From:         localNodeAddr.ToPeerAddress(),
		Payload:      []byte("Query Status"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	var perinfos PeerInfos
	for _, per := range peers {
		var perinfo PeerInfo
		log.Debug("rage the peer")
		perinfo.IP = per.PeerAddr.IP
		perinfo.Port = per.PeerAddr.Port
		perinfo.RPCPort = per.PeerAddr.RpcPort
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
		perinfo.Delay = this.LocalNode.delayTable[per.PeerAddr.ID]
		this.LocalNode.delayTableMutex.RUnlock()
		perinfo.ID = per.PeerAddr.ID
		perinfos = append(perinfos, perinfo)
	}
	var self_info = PeerInfo{
		IP:        this.LocalNode.GetNodeAddr().IP,
		Port:      this.LocalNode.GetNodeAddr().Port,
		ID:        this.LocalNode.GetNodeAddr().ID,
		RPCPort:   this.LocalAddr.RpcPort,
		Status:    ALIVE,
		IsPrimary: this.LocalNode.IsPrimary,
		Delay:     this.LocalNode.delayTable[this.LocalAddr.ID],
	}
	perinfos = append(perinfos, self_info)
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (this *GrpcPeerManager) GetNodeId() int {
	return this.LocalAddr.ID
}

func (this *GrpcPeerManager) SetPrimary(_id uint64) error {
	__id := strconv.FormatUint(_id, 10)
	id, _err := strconv.Atoi(__id)
	if _err != nil {
		log.Error("convert err", _err)
	}
	peers := this.peersPool.GetPeers()
	this.LocalNode.IsPrimary = false
	for _, per := range peers {
		per.IsPrimary = false
	}

	for _, per := range peers {
		if per.PeerAddr.ID == id {
			per.IsPrimary = true
		}
		if this.LocalAddr.ID == id {
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
	newPeer, err := NewPeer(pb.RecoverPeerAddr(&toUpdateAddress),this.LocalAddr, this.TEM)
	if err != nil {
		log.Error(err)
	}

	log.Debug("newPeer: %v", newPeer)
	//新消息
	payload, _ = proto.Marshal(this.LocalAddr.ToPeerAddress())

	attendResponseMsg := pb.Message{
		MessageType:pb.Message_ATTEND_RESPNSE,
		Payload:payload,
		MsgTimeStamp:time.Now().UnixNano(),
		From:this.LocalAddr.ToPeerAddress(),
	}

	if this.IsOnline {
		this.peersPool.MergeTempPeers(newPeer)
		//通知新节点进行接洽
		newPeer.Chat(attendResponseMsg)
		this.LocalNode.N += 1
		this.configs.AddNodesAndPersist(*this.peersPool.GetPeers())
	} else {
		//新节点
		//新节点在一开始的时候就已经将介绍人的节点列表加入了所以这里不需要处理
		//the new attend node
		this.IsOnline = true
		this.peersPool.MergeTempPeersForNewNode()
		this.LocalNode.N = this.peersPool.GetAliveNodeNum()
		this.configs.AddNodesAndPersist(*this.peersPool.GetPeers())
	}
}

/*********************************
 * delete LocalNode part
 ********************************/
func (this *GrpcPeerManager) GetLocalNodeHash() string{
	return this.LocalAddr.Hash
}

func (this *GrpcPeerManager) GetRouterHashifDelete(hash string) (string,uint64){
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	routers := this.peersPool.ToRoutingTableWithout(hash)
	hash = hex.EncodeToString(hasher.Hash(routers).Bytes())

	var ID uint64
	localHash := this.LocalAddr.Hash
	for _,rs := range routers.Routers{
		log.Debug("RS hash: ", rs.Hash)
		if rs.Hash == localHash{
			log.Notice("rs hash: ", rs.Hash)
			log.Notice("id: ", rs.ID)
			ID=uint64(rs.ID);
		}
	}
	return hex.EncodeToString(hasher.Hash(routers).Bytes()),ID
}


func (this *GrpcPeerManager)  DeleteNode(hash string) error{

	if this.LocalAddr.Hash == hash {
		// delete local node and stop all server
		this.LocalNode.StopServer()

	} else{
		// delete the specific node
		for _,pers := range this.peersPool.GetPeers(){
			if pers.PeerAddr.Hash == hash{
				this.peersPool.DeletePeer(pers)
			}
		}
		//TODO update node id
		hasher := crypto.NewKeccak256Hash("keccak256Hanser")
		routers := this.peersPool.ToRoutingTableWithout(hash)
		hash = hex.EncodeToString(hasher.Hash(routers).Bytes())

		if hash == this.LocalAddr.Hash {
			log.Critical("THIS NODE WAS BEEN CLOSED...")
			return nil
		}

		for _,per :=range this.peersPool.GetPeers(){
			if per.PeerAddr.Hash == hash{
				deleteList := this.peersPool.DeletePeer(per)
				this.configs.DelNodesAndPersist(deleteList)
			}else{
				for _,router := range routers.Routers{
					if router.Hash == per.PeerAddr.Hash{
						//TODO CHECK here the jsonrpc port
						per.PeerAddr = pb.NewPeerAddr(router.IP,int(router.Port),0,int(router.ID))
					}
				}
			}

		}
		return nil


	}
	return nil
}