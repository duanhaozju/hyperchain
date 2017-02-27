//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/core/crypto/primitives"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/persist"
	"hyperchain/p2p/transport"
	"hyperchain/p2p/transport/ecdh"
	"hyperchain/recovery"
	"math"
	"strconv"
	"time"
	"sync"
)

var MAX_PEER_NUM = 4

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
	CM *admittance.CAManager
	//isValidate peer
	IsVP bool
}

func NewGrpcManager(conf *common.Config) *GRPCPeerManager {
	var newgRPCManager GRPCPeerManager
	config := peerComm.NewConfigReader(conf.GetString("global.configs.peers"))
	// configs
	newgRPCManager.configs = config
	newgRPCManager.LocalAddr = pb.NewPeerAddr(config.GetLocalIP(), config.GetLocalGRPCPort(), config.GetLocalJsonRPCPort(), config.GetLocalID())
	MAX_PEER_NUM = newgRPCManager.configs.GetMaxPeerNumber()
	if (newgRPCManager.configs.GetLocalID()-1) > MAX_PEER_NUM{
		panic("the node id should not large than max peer number")
	}
	//get the maxpeer from config
	newgRPCManager.IsOriginal = config.IsOrigin()
	newgRPCManager.IsVP = config.IsVP()
	newgRPCManager.Introducer = pb.NewPeerAddr(config.GetIntroducerIP(), config.GetIntroducerPort(), config.GetIntroducerJSONRPCPort(), config.GetIntroducerID())
	return &newgRPCManager
}

// Start start the Normal local listen server
func (this *GRPCPeerManager) Start(aliveChain chan int, eventMux *event.TypeMux, cm *admittance.CAManager) {
	if this.LocalAddr.ID == 0 || this.configs == nil {
		log.Errorf("ID %d", this.LocalAddr.ID)
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
	bol, _ := persist.GetBool("onceOnline")
	if bol {
		//log.Critical("reconnect")
		//TODO func reconnect
		this.reconnectToPeers(aliveChain)
		this.IsOnline = true
	} else {
		//log.Critical("connect")
		//log.Critical("isvp", this.IsVP, "isorigin", this.IsOriginal)
		if this.IsOriginal {
			// load the waiting connecting node information
			log.Debug("start node as oirgin mode")
			this.connectToPeers(aliveChain)
			this.IsOnline = true
			//aliveChain <- 0
		} else if !this.IsOriginal {
			// start attend routinei
			log.Debug("connect to introducer")
			// IMPORT should aliveChain <- 1 frist
			aliveChain <- 1
			go this.LocalNode.attendNoticeProcess(this.LocalNode.N)
			this.connectToIntroducer(*this.Introducer)
		}
		//else if !this.IsVP {
		//	log.Debug("connect to vp")
		//	this.vpConnect()
		//	aliveChain <- 2
		//}
	}
	log.Notice("┌────────────────────────────┐")
	log.Notice("│  All NODES WERE CONNECTED  |")
	log.Notice("└────────────────────────────┘")
	persist.PutBool("onceOnline", true)
}

func (this *GRPCPeerManager) ConnectToOthers() {
	//TODO更新路由表之后进行连接
	allPeersWithTemp := this.peersPool.GetPeersWithTemp()
	payload, _ := proto.Marshal(this.LocalAddr.ToPeerAddress())
	newNodeMessage := &pb.Message{
		MessageType:  pb.Message_ATTEND_NOTIFY,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         this.LocalAddr.ToPeerAddress(),
	}
	log.Critical("Connect to Others", payload)
	for _, peer := range allPeersWithTemp {
		//review 返回值不做处理
		//retMessage, err := peer.Chat(newNodeMessage)
		//if err != nil {
		//	log.Error("notice other node Attend Failed", err)
		//}
		retMessage, err := peer.Client.Chat(context.Background(), newNodeMessage)
		log.Debug("reconnect return :", retMessage)
		if err != nil {
			log.Error("cannot establish a connection", err)
			return
		}
		if retMessage.MessageType == pb.Message_ATTEND_NOTIFY_RESPONSE {
			remoteECert := retMessage.Signature.ECert
			if remoteECert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				//return nil,errors.New("remote ecert is nil")
			}
			remoteRCert := retMessage.Signature.RCert
			if remoteRCert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				//return nil,errors.New("Remote ECert is nil %v")
			}

			verifyRcert, rcertErr := peer.CM.VerifyRCert(string(remoteRCert))
			if !verifyRcert || rcertErr != nil {
				peer.TEM.SetIsVerified(false, retMessage.From.Hash)
			} else {
				peer.TEM.SetIsVerified(true, retMessage.From.Hash)
			}

			//peer.TEM.SetSignPublicKey(signpuByte, retMessage.From.Hash)

			log.Critical("put peer into temp invoke: CONNECT TO OTHERS",peer.PeerAddr.ID)
			this.peersPool.PutPeerToTemp(*peer.PeerAddr, peer)
			}
	}
}

func (this *GRPCPeerManager) connectToIntroducer(introducerAddress pb.PeerAddr) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	introducerPeer, payload, err := NewPeerIntroducer(&introducerAddress, this.LocalAddr, this.TEM, this.CM)
	if err != nil {
		log.Error("Node: ", introducerAddress.IP, ":", introducerAddress.Port, " can not connect to introducer!\n")
		return
	}
	var routers pb.Routers
	unmarErr := proto.Unmarshal(payload, &routers)
	if unmarErr != nil {
		log.Error("routing table unmarshal err ", unmarErr)
	}

	//将介绍人的信息放入路由表中
	this.peersPool.PutPeer(*introducerPeer.PeerAddr, introducerPeer)

	this.peersPool.MergeFromRoutersToTemp(routers)

	localAddressPayload, err := proto.Marshal(this.LocalNode.localAddr.ToPeerAddress())
	if err != nil {
		log.Error("cannot marshal the interducer Address payload")
	}

	for _, p := range this.peersPool.GetPeersWithTemp() {
		//review this.LocalNode.attendChan <- 1
		attend_message := &pb.Message{
			MessageType:  pb.Message_ATTEND,
			Payload:      localAddressPayload,
			MsgTimeStamp: time.Now().UnixNano(),
			From:         this.LocalNode.localAddr.ToPeerAddress(),
		}
		SignCert(attend_message, this.CM)
		log.Debug("SEND ATTEND TO",p.PeerAddr.ID)
		//retMessage, err := p.Chat(attend_message)
		retMessage, err := p.Client.Chat(context.Background(), attend_message)
		//log.Debug("reconnect return :", retMessage)
		if err != nil {
			log.Error("cannot establish a connection", err)
			return
		}
		if retMessage.MessageType == pb.Message_ATTEND_RESPONSE {
			log.Debug("get a new ATTTEND RESPNSE from ID",retMessage.From.ID)
			remoteECert := retMessage.Signature.ECert
			if remoteECert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				return
			}
			ecert, err := primitives.ParseCertificate(string(remoteECert))
			if err != nil {
				log.Error("cannot parse certificate")
				return
			}
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpuByte := ecdh256.Marshal(*signpub)
			//verify the rcert and set the status

			remoteRCert := retMessage.Signature.RCert
			if remoteRCert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				return
			}

			verifyRcert, rcertErr := this.CM.VerifyRCert(string(remoteRCert))
			if !verifyRcert || rcertErr != nil {
				this.TEM.SetIsVerified(false, retMessage.From.Hash)
			} else {
				this.TEM.SetIsVerified(true, retMessage.From.Hash)
			}

			this.TEM.SetSignPublicKey(signpuByte, retMessage.From.Hash)
			genErr := this.TEM.GenerateSecret(signpuByte, retMessage.From.Hash)
			if genErr != nil {
				log.Errorf("generate the share secret key, from node id: %d, error info %s ", retMessage.From.ID, genErr)
			}
			//else {
			//	log.Critical("put peer into temp invoke: ATTENDRESPONSE",p.PeerAddr.ID)
			//	this.peersPool.PutPeerToTemp(*p.PeerAddr, p)
			//}
		}else{
			log.Error("ATTEND FAILED! TO NODE ID:",p.PeerAddr.ID)
		}
		if err != nil {
			log.Error(err)
		}
	}
	this.LocalNode.N = len(this.GetAllPeersWithTemp())
}

// passed
func (this *GRPCPeerManager) connectToPeers(alive chan int) {
	peerStatus := make(map[int]bool)
	for i := 1; i <= MAX_PEER_NUM; i++ {
		peerStatus[i] = false
	}
	if this.LocalAddr.ID > MAX_PEER_NUM {
		panic("LocalAddr.ID large than the max peer num")
	}
	peerStatus[this.LocalAddr.ID] = true
	N := MAX_PEER_NUM
	F := int(math.Floor(float64(MAX_PEER_NUM - 1) / 3.0))
	MaxNum := N - F - 1
	if this.configs.IsOrigin() {
		MaxNum = N - 1
	}
	// connect other peers
	_index := 1
	for range time.Tick(200 * time.Millisecond) {
		log.Debugf("current alive node num is %d, at least connect to %d", this.peersPool.GetAliveNodeNum(), MaxNum)
		log.Debugf("current alive num %d", this.peersPool.GetAliveNodeNum())
		if this.peersPool.GetAliveNodeNum() >= MaxNum {
			alive <- 0
		}
		if this.peersPool.GetAliveNodeNum() >= MAX_PEER_NUM-1 {
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
	if this.LocalAddr.ID > MAX_PEER_NUM {
		panic("LocalAddr.ID large than the max peer num")
	}
	N := MAX_PEER_NUM
	F := int(math.Floor(float64(MAX_PEER_NUM - 1) / 3.0))
	MaxNum := N - F
	var connected Stack
	var unconnected Stack
	for i := 1; i <= MAX_PEER_NUM; i++ {
		if i!=this.LocalAddr.ID{
			unconnected.Push(i)
		}
	}
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func(wgp *sync.WaitGroup ){
		flag := false
		for range time.Tick(200*time.Millisecond){
			_index := unconnected.Pop().(int)
			peerAddress := pb.NewPeerAddr(this.configs.GetIP(_index), this.configs.GetPort(_index), this.configs.GetRPCPort(_index), this.configs.GetID(_index))
			//log.Debugf("peeraddress to connect %v", peerAddress)
			if peer, connectErr := this.reconnectToPeer(peerAddress, this.configs.GetID(_index)); connectErr != nil {
				// cannot connect to other peer

				unconnected.Push(_index)
				log.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
			} else {
				connected.Push(_index)
				this.peersPool.PutPeer(*peerAddress, peer)
				log.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
			}
			if this.peersPool.GetAliveNodeNum() >= MaxNum && !flag {
				flag = true
				wgp.Done()
				alive <- 0
			}
			if unconnected.Length() == 0 {
				break;
			}
		}
	}(wg)
	wg.Wait()

}

func (this *GRPCPeerManager) vpConnect() {
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
				log.Debugf("putpeer %v,%v", this.peersPool.GetPeers(), peerAddress)
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
func (this *GRPCPeerManager) GetVPPeers() []*Peer {
	return this.peersPool.GetVPPeers()
}
func (this *GRPCPeerManager) GetNVPPeers() []*Peer {
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
		log.Error("cannot marshal the payload", err)
	}
	testUnmarshal := new(pb.PeerAddress)
	err = proto.Unmarshal(payload, testUnmarshal)
	if err != nil {
		log.Error("self unmarshal failed!", err)
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
	if this.CM.GetIsUsed() {
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

	if this.IsOnline {
		toUpdateAddress := new(pb.PeerAddress)
		err := proto.Unmarshal(payload, toUpdateAddress)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("Attend Notify address",toUpdateAddress.ID,toUpdateAddress.IP,toUpdateAddress.Port)
		log.Debug("updateRoutingTable")
		newPeer, retMessage, err := NewPeerAttendNotify(pb.RecoverPeerAddr(toUpdateAddress), this.LocalAddr, this.TEM, this.CM, this.LocalNode.IsPrimary)
		if err != nil {
			log.Error("recover address failed", err)
			return
		}

		log.Debugf("NEWPEER id: %d", newPeer.PeerAddr.ID)

		if retMessage.MessageType == pb.Message_ATTEND_NOTIFY_RESPONSE {
			remoteECert := retMessage.Signature.ECert
			if remoteECert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				return
			}
			ecert, err := primitives.ParseCertificate(string(remoteECert))
			if err != nil {
				log.Error("cannot parse certificate")
				return
			}
			signpub := ecert.PublicKey.(*(ecdsa.PublicKey))
			ecdh256 := ecdh.NewEllipticECDH(elliptic.P256())
			signpuByte := ecdh256.Marshal(*signpub)
			//verify the rcert and set the status

			remoteRCert := retMessage.Signature.RCert
			if remoteRCert == nil {
				log.Errorf("Remote ECert is nil %v", retMessage.From)
				return
			}

			verifyRcert, rcertErr := this.CM.VerifyRCert(string(remoteRCert))
			if !verifyRcert || rcertErr != nil {
				this.TEM.SetIsVerified(false, retMessage.From.Hash)
			} else {
				this.TEM.SetIsVerified(true, retMessage.From.Hash)
			}

			this.TEM.SetSignPublicKey(signpuByte, retMessage.From.Hash)
			genErr := this.TEM.GenerateSecret(signpuByte, retMessage.From.Hash)
			if genErr != nil {
				log.Errorf("generate the share secret key, from node id: %d, error info %s ", retMessage.From.ID, genErr)
			} else {
				this.peersPool.MergeTempPeers(newPeer)
				log.Debug("add new peer into peerspool, new peer id",newPeer.PeerAddr.ID)
				this.LocalNode.N += 1
				this.configs.AddNodesAndPersist(this.peersPool.GetPeersAddrMap())
			}


		}

	} else {
		log.Error(" new node shouldn't call update routing table")
	}
}


func (this *GRPCPeerManager) SetOnline(){
	log.Debug("As a new attend node, save the introducer info into peer addr")
	//新节点
	//新节点在一开始的时候就已经将介绍人的节点列表加入了所以这里不需要处理
	//the new attend node
	this.IsOnline = true
	this.peersPool.MergeTempPeersForNewNode()
	log.Debugf(" NEW PEER SET ONLINE",this.LocalNode.localAddr.ID)
	this.LocalNode.N = this.peersPool.GetAliveNodeNum()
	this.configs.AddNodesAndPersist(this.peersPool.GetPeersAddrMap())
}

/*********************************
 * delete LocalNode part
 ********************************/
func (this *GRPCPeerManager) GetLocalNodeHash() string {
	return this.LocalAddr.Hash
}

func (this *GRPCPeerManager) GetRouterHashifDelete(hash string) (string, uint64,uint64) {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	log.Debug("GetRouterHashifDelete")
	//routers := this.peersPool.ToRoutingTableWithout(hash)
	//var WantDeleteID uint64
	//for _, peer := range this.peersPool.GetPeers() {
	//	if peer.PeerAddr.Hash == hash{
	//		WantDeleteID = uint64(peer.PeerAddr.ID)
	//	}
	//}
	//log.Critical("WantDeleteID: ",WantDeleteID)
	//if uint64(WantDeleteID) < uint64(this.LocalAddr.ID){
	//	return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(this.LocalNode.GetNodeID() - 1),uint64(WantDeleteID)
	//}
	//return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(this.LocalNode.GetNodeID()),uint64(WantDeleteID)


	routers := this.peersPool.ToRoutingTableWithout(hash)
	var DeleteID uint64
	for _, peer := range this.peersPool.GetPeers(){
		log.Warning("range HASH, ",peer.PeerAddr.Hash," wanted hash",hash)
		if peer.PeerAddr.Hash == hash{
			log.Debug("RS hash: ", peer.PeerAddr.Hash)
			DeleteID = uint64(peer.PeerAddr.ID)
		}
	}
	if uint64(DeleteID) < uint64(this.LocalAddr.ID){
		return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(this.LocalAddr.ID -1) ,uint64(DeleteID)
	}
	log.Debug("DELETE ID",DeleteID)

	return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(this.LocalAddr.ID) ,uint64(DeleteID)

}

func (this *GRPCPeerManager) DeleteNode(hash string) error {

	if this.LocalAddr.Hash == hash {
		// delete local node and stop all server
		log.Critical("Stop Server")
		this.LocalNode.StopServer()
		this.peersPool.Clear()
		panic("THIS NODE HAS BEEN QUITTED")
	} else {
		// delete the specific node
		//TODO update node id
		var deleteID int
		for _, per := range this.peersPool.GetPeers() {
			if per.PeerAddr.Hash == hash {
				deleteID = per.PeerAddr.ID
				deleteList := this.peersPool.DeletePeer(per)
				log.Debug("Delete node and persist")
				this.configs.DelNodesAndPersist(deleteList)

			}
		}
		for _, per := range this.peersPool.GetPeers() {
			if per.PeerAddr.ID > deleteID{
				per.PeerAddr.ID--
			}

		}
		if this.LocalAddr.ID > deleteID{
				this.LocalAddr.ID--
		}
		return nil

	}
	return nil
}
