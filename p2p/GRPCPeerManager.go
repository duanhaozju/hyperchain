//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/admittance"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"math"
	"strconv"
	"time"
	"hyperchain/core/crypto/primitives"
	"hyperchain/p2p/persist"
)

const MAX_PEER_NUM = 4

// gRPC peer manager struct, which to manage the gRPC peers
type GRPCPeerManager struct {
	LocalAddr  *pb.PeerAddr
	LocalNode  *Node
	peersPool  PeersPool
	TEM        transport.TransportEncryptManager
	configs    peerComm.Config
	IsOriginal bool
	IsOnline   bool
	//interducer information
	Introducer *pb.PeerAddr
	//CERT Manager
	CM         *admittance.CAManager
	//isValidate peer
	IsVP bool
}

func NewGrpcManager(conf *common.Config) *GRPCPeerManager {
	var newgRPCManager GRPCPeerManager
	config := peerComm.NewConfigReader(conf.GetString("global.configs.peers"))
	// configs
	newgRPCManager.configs = config
	newgRPCManager.LocalAddr = pb.NewPeerAddr(config.GetLocalIP(), config.GetLocalGRPCPort(), config.GetLocalJsonRPCPort(), config.GetLocalID())
	//get the maxpeer from config
	newgRPCManager.IsOriginal = config.IsOrigin()
	newgRPCManager.IsVP =config.IsVP()
	newgRPCManager.Introducer = pb.NewPeerAddr(config.GetIntroducerIP(), config.GetIntroducerPort(), config.GetIntroducerJSONRPCPort(), config.GetIntroducerID())
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GRPCPeerManager) Start(aliveChain chan int, eventMux *event.TypeMux, cm *admittance.CAManager) {
	if this.LocalAddr.ID == 0 || this.configs == nil {
		panic("the PeerManager hasn't initlized")
	}
	//cert manager
	this.CM = cm
	//handshakemanager
	//HSM only instanced once, so peersPool and Node Hsm are same instance
	this.TEM = transport.NewHandShakeMangerNew(this.CM)
	this.peersPool = NewPeerPoolIml(this.TEM, this.LocalAddr, this.CM)
	this.LocalNode = NewNode(this.LocalAddr, eventMux, this.TEM, this.peersPool, this.CM)
	this.LocalNode.StartServer()
	this.LocalNode.N = this.configs.GetMaxPeerNumber()
	// connect to peer
	bol,_ := persist.GetBool("onceOnline")
	if bol {
		log.Critical("reconnect")
		//TODO func reconnect
		this.reconnectToPeers(aliveChain)
		this.IsOnline = true
	}else {
		log.Critical("connect")
		if this.IsVP && this.IsOriginal {
			// load the waiting connecting node information
			log.Debug("start node as oirgin mode")
			this.connectToPeers(aliveChain)
			this.IsOnline = true
		} else if this.IsVP && !this.IsOriginal {
			// start attend routine
			go this.LocalNode.attendNoticeProcess(this.LocalNode.N)
			//review connect to introducer
			this.connectToIntroducer(*this.Introducer)
			aliveChain <- 1
		} else if !this.IsOriginal {
			this.vpConnect()
			aliveChain <- 2
		}
	}
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  |")
	log.Notice("└────────────────────────────┘")
	persist.PutBool("onceOnline",true)
}

func (this *GRPCPeerManager) ConnectToOthers() {
	//TODO更新路由表之后进行连接
	allPeersWithTemp := this.peersPool.GetPeersWithTemp()
	payload, _ := proto.Marshal(this.LocalAddr.ToPeerAddress())
	newNodeMessage := pb.Message{
		MessageType:  pb.Message_ATTEND,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         this.LocalAddr.ToPeerAddress(),
	}
	log.Critical("Connect to Others",payload)
	for _, peer := range allPeersWithTemp {
		//review 返回值不做处理
		_, err := peer.Chat(newNodeMessage)
		if err != nil {
			log.Error("notice other node Attend Failed", err)
		}
	}
}

func (this *GRPCPeerManager) connectToIntroducer(introducerAddress pb.PeerAddr) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	peer, err := NewPeer(&introducerAddress, this.LocalAddr, this.TEM, this.CM)
	if err != nil {
		log.Error("Node: ", introducerAddress.IP, ":", introducerAddress.Port, " can not connect to introducer!\n")
		return
	}
	//发送introduce 信息,取得路由表
	payload, _ := proto.Marshal(this.LocalAddr.ToPeerAddress())
	introduce_message := pb.Message{
		MessageType:  pb.Message_INTRODUCE,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         this.LocalAddr.ToPeerAddress(),
	}
	SignCert(introduce_message,this.CM)

	retMsg, sendErr := peer.Chat(introduce_message)

	if sendErr != nil {
		log.Errorf("get routing table error, %v",sendErr)
		return
	}

	var routers pb.Routers
	unmarErr := proto.Unmarshal(retMsg.Payload, &routers)
	if unmarErr != nil {
		log.Error("routing table unmarshal err ", unmarErr)
	}

	//将介绍人的信息放入路由表中
	this.peersPool.PutPeer(*peer.PeerAddr, peer)

	log.Warning("merge the routing table and connect to :", routers)

	this.peersPool.MergeFromRoutersToTemp(routers)

	for _, p := range this.peersPool.GetPeersWithTemp() {
		//review this.LocalNode.attendChan <- 1
		attend_message := pb.Message{
			MessageType:  pb.Message_ATTEND,
			Payload:      payload,
			MsgTimeStamp: time.Now().UnixNano(),
			From:         this.LocalNode.localAddr.ToPeerAddress(),
		}
		_, err := p.Chat(attend_message)
		if err != nil {
			log.Error(err)
		}
	}
	this.LocalNode.N = len(this.GetAllPeersWithTemp())
}

// passed
func (this *GRPCPeerManager) connectToPeers(alive chan int) {
	peerStatus := make(map[int]bool)
	for i := 1; i <= MAX_PEER_NUM; i++ {peerStatus[i] = false}
	if this.LocalAddr.ID>MAX_PEER_NUM {panic("LocalAddr.ID large than the max peer num")}
	peerStatus[this.LocalAddr.ID] = true
	N := MAX_PEER_NUM
	F := int(math.Floor((MAX_PEER_NUM - 1) / 3))
	MaxNum :=  N - F - 1
	if this.configs.IsOrigin() {MaxNum = N - 1}
	// connect other peers
	_index := 1
	for range time.Tick(200 * time.Millisecond) {
		log.Debugf("current alive node num is %d, at least connect to %d",this.peersPool.GetAliveNodeNum(),MaxNum)
		log.Debugf("current alive num %d",this.peersPool.GetAliveNodeNum())
		if this.peersPool.GetAliveNodeNum() >= MaxNum{
			alive <- 0
		}
		if this.peersPool.GetAliveNodeNum() >= MAX_PEER_NUM - 1{
			break
		}
		if peerStatus[_index] {
			_index++
			continue
		}
		if _index > MAX_PEER_NUM {
			_index = 1
		}

		log.Debug("status map", _index, peerStatus[_index])
		//if this node is not online, connect it
		peerAddress := pb.NewPeerAddr(this.configs.GetIP(_index), this.configs.GetPort(_index), this.configs.GetRPCPort(_index), this.configs.GetID(_index))
		log.Debugf("peeraddress to connect %v", peerAddress)
		if peer, connectErr := this.connectToPeer(peerAddress, this.configs.GetID(_index)); connectErr != nil {
			// cannot connect to other peer
			log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
			continue
		} else {
			// add  peer to peer pool
			this.peersPool.PutPeer(*peerAddress, peer)
			peerStatus[_index] = true
			log.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
		}
	}
}

func (this *GRPCPeerManager) reconnectToPeers(alive chan int) {
	peerStatus := make(map[int]bool)
	for i := 1; i <= MAX_PEER_NUM; i++ {peerStatus[i] = false}
	if this.LocalAddr.ID>MAX_PEER_NUM {panic("LocalAddr.ID large than the max peer num")}
	peerStatus[this.LocalAddr.ID] = true
	N := MAX_PEER_NUM
	F := int(math.Floor((MAX_PEER_NUM - 1) / 3))
	MaxNum :=  N - F - 1
	if this.configs.IsOrigin() {MaxNum = N - 1}
	// connect other peers
	_index := 1
	for range time.Tick(200 * time.Millisecond) {
		log.Debugf("current alive node num is %d, at least connect to %d",this.peersPool.GetAliveNodeNum(),MaxNum)
		log.Debugf("current alive num %d",this.peersPool.GetAliveNodeNum())
		if this.peersPool.GetAliveNodeNum() >= MaxNum{
			alive <- 0
		}
		if this.peersPool.GetAliveNodeNum() >= MAX_PEER_NUM - 1{
			break
		}
		if peerStatus[_index] {
			_index++
			continue
		}
		if _index > MAX_PEER_NUM {
			_index = 1
		}

		log.Debug("status map", _index, peerStatus[_index])
		//if this node is not online, connect it
		peerAddress := pb.NewPeerAddr(this.configs.GetIP(_index), this.configs.GetPort(_index), this.configs.GetRPCPort(_index), this.configs.GetID(_index))
		log.Debugf("peeraddress to connect %v", peerAddress)
		if peer, connectErr := this.reconnectToPeer(peerAddress, this.configs.GetID(_index)); connectErr != nil {
			// cannot connect to other peer
			log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
			continue
		} else {
			// add  peer to peer pool
			this.peersPool.PutPeer(*peerAddress, peer)
			peerStatus[_index] = true
			log.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
		}
	}
}

func (this *GRPCPeerManager) vpConnect(){
	var peerStatus map[int]bool
	peerStatus = make(map[int]bool)
	//这里需要进行
	for i := 1; i <= this.configs.GetMaxPeerNumber(); i++ {
		if i == this.LocalAddr.ID {
			peerStatus[i] = true
		} else {
			peerStatus[i] = false
		}
	}
	MaxNum := this.configs.GetMaxPeerNumber()

	// connect other peers
	//TODO RETRY CONNECT 重试连接(未实现)
	for this.peersPool.GetAliveNodeNum() < MaxNum {
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
			peerAddress := pb.NewPeerAddr(peerIp, peerPort, this.configs.GetLocalGRPCPort(), _index)
			log.Info(peerAddress)
			peer, connectErr := this.connectToPeer(peerAddress, _index)
			if connectErr != nil {
				// cannot connect to other peer
				log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
				continue
			} else {
				// add  peer to peer pool
				this.peersPool.PutPeer(*peerAddress, peer)
				log.Debugf("putpeer %v,%v",this.peersPool.GetPeers(),peerAddress)
				//this.TEM.[peer.Addr.Hash]=peer.TEM
				peerStatus[_index] = true
				log.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
				log.Debug("nodes number:", this.peersPool.GetAliveNodeNum())
			}

		}
	}

}
//connect to peer by ip address and port (why int32? because of protobuf limit)
func (this *GRPCPeerManager) connectToPeer(peerAddress *pb.PeerAddr, nid int) (*Peer, error) {
	//if this node is not online, connect it
	var peer *Peer
	var peerErr error
	peer, peerErr = NewPeer(peerAddress, this.LocalAddr, this.TEM, this.CM)
	if peerErr != nil {
		// cannot connect to other peer
		log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n")
		return nil, peerErr
	} else {
		//log.Critical("连接到节点", nid,isReconnect)
		return peer, nil
	}

}

//connect to peer by ip address and port (why int32? because of protobuf limit)
func (this *GRPCPeerManager) reconnectToPeer(peerAddress *pb.PeerAddr, nid int) (*Peer, error) {
	//if this node is not online, connect it
	var peer *Peer
	var peerErr error
	peer, peerErr = NewPeerReconnect(peerAddress, this.LocalAddr, this.TEM, this.CM)
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
func (this *GRPCPeerManager) GetAllPeers() []*Peer {
	return this.peersPool.GetPeers()
}

func (this *GRPCPeerManager) GetAllPeersWithTemp() []*Peer {
	return this.peersPool.GetPeersWithTemp()
}
func (this *GRPCPeerManager) GetVPPeers() []*Peer{
	return this.peersPool.GetVPPeers()
}
func (this *GRPCPeerManager) GetNVPPeers() []*Peer{
	return this.peersPool.GetNVPPeers()

}
// BroadcastPeers Broadcast Massage to connected peers
func (this *GRPCPeerManager) BroadcastPeers(payLoad []byte) {
	//log.Warning("P2P broadcast")
	if !this.IsOnline {
		log.Warning("IS NOT Online")
		return
	}
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         this.LocalNode.GetNodeAddr().ToPeerAddress(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(this, broadCastMessage, this.peersPool.GetPeers())
}
// BroadcastPeers Broadcast Massage to connected VP peers
func (this *GRPCPeerManager) BroadcastVPPeers(payLoad []byte) {
	//log.Warning("P2P broadcast")
	if !this.IsOnline {
		log.Warning("IS NOT Online")
		return
	}
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         this.LocalNode.GetNodeAddr().ToPeerAddress(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(this, broadCastMessage, this.peersPool.GetVPPeers())
}
// BroadcastPeers Broadcast Massage to connected NVP peers
func (this *GRPCPeerManager) BroadcastNVPPeers(payLoad []byte) {
	//log.Warning("P2P broadcast")
	if !this.IsOnline {
		log.Warning("IS NOT Online")
		return
	}
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_CONSUS,
		From:         this.LocalNode.GetNodeAddr().ToPeerAddress(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(this, broadCastMessage, this.peersPool.GetNVPPeers())
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(grpcPeerManager *GRPCPeerManager, broadCastMessage pb.Message, peers []*Peer) {
	for _, peer := range peers {
		//REVIEW 这里没有返回值,不知道本次通信是否成功
		//REVIEW 其实这里不需要处理返回值，需要将其go起来
		//REVIEW Chat 方法必须要传实例，否则将会重复加密，请一定要注意！！
		//REVIEW Chat Function must give a message instance, not a point, if not the encrypt will break the payload!
		go func(p2 *Peer) {
			start := time.Now().UnixNano()
			_, err := p2.Chat(broadCastMessage)
			if err == nil {
				grpcPeerManager.LocalNode.DelayChan <- UpdateTable{updateID: p2.LocalAddr.ID, updateTime: time.Now().UnixNano() - start}
			} else {
				log.Error("chat failed", err)
			}
		}(peer)

	}
}

func (this *GRPCPeerManager) GetLocalAddressPayload() (payload []byte) {
	payload, err := proto.Marshal(this.LocalAddr.ToPeerAddress())
	if err != nil {
		log.Error("cannot marshal the payload",err)
	}
	testUnmarshal := new(pb.PeerAddress)
	err = proto.Unmarshal(payload,testUnmarshal)
	if err != nil{
		log.Error("self unmarshal failed!",err)
	}
	return
}

// SendMsgToPeers Send msg to specific peer peerlist
func (this *GRPCPeerManager) SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType) {

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
			//convert the uint64 to int
			//because the unicast node is not confirm so, here use double loop
			if p.PeerAddr.ID == int(NodeID) {
				start := time.Now().UnixNano()
				_, err := p.Chat(syncMessage)
				if err != nil {
					log.Error("Broadcast failed,Node", p.LocalAddr.ID)
				} else {
					this.LocalNode.DelayChan <- UpdateTable{updateID: p.LocalAddr.ID, updateTime: time.Now().UnixNano() - start}

					//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
				}
			}
		}

	}

}

func (this *GRPCPeerManager) GetPeerInfo() PeerInfos {
	peers := this.peersPool.GetPeers()
	localNodeAddr := this.LocalNode.GetNodeAddr()

	var keepAliveMessage = pb.Message{
		MessageType:  pb.Message_KEEPALIVE,
		From:         localNodeAddr.ToPeerAddress(),
		Payload:      []byte("Query Status"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	if this.CM.GetIsUsed(){
		var pri interface{}
		pri = this.CM.GetECertPrivKey()
		ecdsaEncry := primitives.NewEcdsaEncrypto("ecdsa")
		sign, err := ecdsaEncry.Sign(keepAliveMessage.Payload, pri)
		if err == nil {
			if keepAliveMessage.Signature == nil {
				payloadSign := pb.Signature{
					Signature: sign,
				}
				keepAliveMessage.Signature = &payloadSign
			}
			keepAliveMessage.Signature.Signature = sign
		}
	}
	var perinfos PeerInfos
	for _, per := range peers {
		var perinfo PeerInfo
		log.Debug("rage the peer")
		perinfo.IP = per.PeerAddr.IP
		perinfo.Port = per.PeerAddr.Port
		perinfo.RPCPort = per.PeerAddr.RPCPort

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
		RPCPort:   this.LocalAddr.RPCPort,
		Status:    ALIVE,
		IsPrimary: this.LocalNode.IsPrimary,
		Delay:     this.LocalNode.delayTable[this.LocalAddr.ID],
	}
	perinfos = append(perinfos, self_info)
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (this *GRPCPeerManager) GetNodeId() int {
	return this.LocalAddr.ID
}

func (this *GRPCPeerManager) SetPrimary(_id uint64) error {
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

func (this *GRPCPeerManager) GetLocalNode() *Node {
	return this.LocalNode
}

func (this *GRPCPeerManager) UpdateRoutingTable(payload []byte) {
	//这里的payload 应该是前面传输过去的 address,里面应该只有一个

	toUpdateAddress := new(pb.PeerAddress)
	err := proto.Unmarshal(payload, toUpdateAddress)
	if err != nil {
		log.Error(err)
		return
	}
	log.Critical("updateRoutingTable")
	newPeer, err := NewPeer(pb.RecoverPeerAddr(toUpdateAddress), this.LocalAddr, this.TEM, this.CM)
	if err != nil {
		log.Error(err)
		return

	}

	log.Debugf("NEWPEER id: %d", newPeer.PeerAddr.ID)
	//新消息
	//payload, _ = proto.Marshal(this.LocalAddr.ToPeerAddress())
	payload = []byte("false")
	if this.LocalNode.IsPrimary{
		payload = []byte("true")
	}

	attendResponseMsg := pb.Message{
		MessageType:  pb.Message_ATTEND_RESPONSE,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         this.LocalAddr.ToPeerAddress(),
	}

	if this.IsOnline {
		log.Debug("Notify the new attend node, this node is aleady merge the node info!")
		this.peersPool.MergeTempPeers(newPeer)
		//通知新节点进行接洽
		_,err :=newPeer.Chat(attendResponseMsg)
		if err != nil {
			log.Errorf("Cannot notify the new attend node, id: %d",newPeer.PeerAddr.ID)
		}
		this.LocalNode.N += 1
		this.configs.AddNodesAndPersist(this.peersPool.GetPeersAddrMap())
	} else {
		log.Debug("As a new attend node, save the introducer info into peer addr")
		//新节点
		//新节点在一开始的时候就已经将介绍人的节点列表加入了所以这里不需要处理
		//the new attend node
		this.IsOnline = true
		this.peersPool.MergeTempPeersForNewNode()
		log.Criticalf("new node peers : %v ,temp: %v",this.peersPool.GetPeers(),this.peersPool.GetPeersWithTemp())
		this.LocalNode.N = this.peersPool.GetAliveNodeNum()
		this.configs.AddNodesAndPersist(this.peersPool.GetPeersAddrMap())
	}
}

/*********************************
 * delete LocalNode part
 ********************************/
func (this *GRPCPeerManager) GetLocalNodeHash() string {
	return this.LocalAddr.Hash
}

func (this *GRPCPeerManager) GetRouterHashifDelete(hash string) (string, uint64) {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	routers := this.peersPool.ToRoutingTableWithout(hash)
	hash = hex.EncodeToString(hasher.Hash(routers).Bytes())

	var ID uint64
	localHash := this.LocalAddr.Hash
	for _, rs := range routers.Routers {
		log.Debug("RS hash: ", rs.Hash)
		if rs.Hash == localHash {
			log.Notice("rs hash: ", rs.Hash)
			log.Notice("id: ", rs.ID)
			ID = uint64(rs.ID)
		}
	}
	return hex.EncodeToString(hasher.Hash(routers).Bytes()), ID
}

func (this *GRPCPeerManager) DeleteNode(hash string) error {

	if this.LocalAddr.Hash == hash {
		// delete local node and stop all server
		this.LocalNode.StopServer()

	} else {
		// delete the specific node
		for _, pers := range this.peersPool.GetPeers() {
			if pers.PeerAddr.Hash == hash {
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

		for _, per := range this.peersPool.GetPeers() {
			if per.PeerAddr.Hash == hash {
				deleteList := this.peersPool.DeletePeer(per)
				this.configs.DelNodesAndPersist(deleteList)
			} else {
				for _, router := range routers.Routers {
					if router.Hash == per.PeerAddr.Hash {
						//TODO CHECK here the jsonrpc port
						per.PeerAddr = pb.NewPeerAddr(router.IP, int(router.Port), 0, int(router.ID))
					}
				}
			}

		}
		return nil

	}
	return nil
}
