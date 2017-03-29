//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/manager/event"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/persist"
	"hyperchain/p2p/transport"
	"math"
	"strconv"
	"sync"
	"time"
)

// gRPC peer manager struct, which to manage the gRPC peers
type GRPCPeerManager struct {
	MaxPeerNum int
	LocalAddr  *pb.PeerAddr
	LocalNode  *Node
	peersPool  *PeersPool
	nvpPool    *PeersPool
	TM         *transport.TransportManager
	configs    peerComm.Config
	IsOriginal bool
	IsOnline   bool
	//interducer information
	Introducer *pb.PeerAddr
	//CERT Manager
	CM *admittance.CAManager
	//isValidate peer
	IsVP      bool
	namespace string
	logger    *logging.Logger
	// to ensure the init type
	aliveChain chan int
	//eventMux
	eventMux *event.TypeMux
}

func NewGrpcManager(conf *common.Config, eventMux *event.TypeMux, cm *admittance.CAManager) *GRPCPeerManager {
	namespace := conf.GetString(common.NAMESPACE)
	logger := common.GetLogger(namespace, "grpcmgr")
	logger.Critical("GRPC: new instance for namespace:", namespace)
	var grpcmgr GRPCPeerManager
	config := peerComm.NewConfigReader(conf.GetString("global.configs.peers"))
	if config == nil || config.LocalID() == 0 {
		logger.Errorf("ID %d", config.LocalID())
		panic("the PeerManager hasn't initlized")
	}
	localAddr := pb.NewPeerAddr(config.LocalIP(), config.LocalGRPCPort(), config.LocalJsonRPCPort(), config.LocalID())
	pb.NewPeerAddr(config.IntroIP(), config.IntroPort(), config.IntroJSONRPCPort(), config.IntroID())
	if config.LocalID() <= 0 || (config.LocalID()-1) > config.MaxNum() {
		panic("the node id should not less than zero")
	}
	if config.IsOrigin() && config.LocalID() > config.MaxNum() {
		panic("LocalAddr.ID shouldn't large than the max peer num")
	} else if !config.IsOrigin() && config.LocalID()-1 > config.MaxNum() {
		panic("LocalAddr.ID - 1 shouldn't large than the max peer num")
	}
	return &GRPCPeerManager{
		configs:    config,
		LocalAddr:  localAddr,
		MaxPeerNum: config.MaxNum(),
		IsOriginal: config.IsOrigin(),
		IsVP:       config.IsVP(),
		namespace:  namespace,
		logger:     logger,
		aliveChain: make(chan int, 1),
		CM:         cm,
		eventMux:   eventMux,
	}

	return &grpcmgr
}

// Start start the Normal local listen server
func (grpcmgr *GRPCPeerManager) Start() {
	//cert manager
	//handshakemanager
	//HSM only instanced once, so peersPool and Node Hsm are same instance
	tm, err := transport.NewTransportManager(grpcmgr.CM, grpcmgr.namespace)
	if err != nil {
		panic(err)
	}
	grpcmgr.TM = tm
	grpcmgr.peersPool = NewPeersPool(grpcmgr.TM, grpcmgr.LocalAddr, grpcmgr.CM, grpcmgr.namespace)
	grpcmgr.LocalNode = NewNode(grpcmgr.LocalAddr, grpcmgr.eventMux, grpcmgr.TM, grpcmgr.peersPool, grpcmgr.CM, grpcmgr.configs, grpcmgr.namespace)
	grpcmgr.LocalNode.StartServer()
	grpcmgr.LocalNode.N = grpcmgr.configs.MaxNum()

	// connect to peer
	rec, _ := persist.GetBool("onceOnline", grpcmgr.namespace)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	if !rec && grpcmgr.IsOriginal {
		//creator
		grpcmgr.logger.Debug("start node as oirgin mode")
		grpcmgr.create(wg)
		grpcmgr.IsOnline = true
	} else if !rec && !grpcmgr.IsOriginal {
		//newnode
		grpcmgr.logger.Debug("connect to introducer")
		// IMPORT should aliveChain <- 1 frist
		grpcmgr.aliveChain <- 1
		go grpcmgr.LocalNode.attendNoticeProcess(grpcmgr.LocalNode.N)
		grpcmgr.connectIntro(*grpcmgr.Introducer)
	} else {
		//reconnect
		grpcmgr.logger.Debug("reconnect to peers")
		grpcmgr.reconnect(wg)
		grpcmgr.IsOnline = true
	}
	wg.Wait()
	grpcmgr.logger.Notice("┌────────────────────────────┐")
	grpcmgr.logger.Notice("│  All NODES WERE CONNECTED  |")
	grpcmgr.logger.Notice("└────────────────────────────┘")
	persist.PutBool("onceOnline", true, grpcmgr.namespace)
}

func (grpcmgr GRPCPeerManager) GetInitType() <-chan int {
	return grpcmgr.aliveChain
}

// create other peers
// Hyperchain creator create the origin block chain
// who use create method create others
func (grpcmgr *GRPCPeerManager) create(owg *sync.WaitGroup) {
	N := grpcmgr.configs.MaxNum()
	F := int(math.Floor(float64(N-1) / 3.0))
	MaxNum := N - F - 1
	if grpcmgr.configs.IsOrigin() {
		MaxNum = N - 1
	}
	var connected Stack
	var unconnected Stack
	for _, p := range grpcmgr.allPeerAddr() {
		unconnected.Push(p)
	}
	// this flag is prepare for the outer wg done ahead
	flag := false
	for range time.Tick(200 * time.Millisecond) {
		addr := unconnected.Pop().(*pb.PeerAddr)
		if peer, connectErr := grpcmgr.connect(addr); connectErr != nil {
			// cannot connect to other peer
			unconnected.UnShift(addr)
			grpcmgr.logger.Error("Node: ", addr.IP, ":", addr.Port, " can not connect!\n", connectErr)
		} else {
			//notify self event process to keep alive
			connected.Push(addr)
			grpcmgr.peersPool.PutPeer(*addr, peer)
			grpcmgr.logger.Debug("Peer Node ID:", addr.ID, "has connected!")
		}
		if grpcmgr.peersPool.GetAliveNodeNum() >= MaxNum && !flag {
			// notify the higer layer to start up the consensus module.
			// this channel may block so go it
			flag = true
			go func(alive chan int) {
				alive <- 0
			}(grpcmgr.aliveChain)
			owg.Done()
		}
		if unconnected.Length() == 0 {
			break
		}
	}
	// if the loop process break with false flag, make the outer wg done
	if !flag {
		owg.Done()
	}

}

//reconnect to already exist peer
func (grpcmgr *GRPCPeerManager) reconnect(owg *sync.WaitGroup) {
	N := grpcmgr.configs.MaxNum()
	F := int(math.Floor(float64(N-1) / 3.0))
	MaxNum := N - F
	var connected Stack
	var unconnected Stack
	for _, p := range grpcmgr.allPeerAddr() {
		unconnected.Push(p)
	}
	flag := false
	for range time.Tick(200 * time.Millisecond) {
		addr := unconnected.Pop().(*pb.PeerAddr)
		if peer, connectErr := grpcmgr.reconnectToPeer(addr); connectErr != nil {
			unconnected.UnShift(addr)
			grpcmgr.logger.Error("Node: ", addr.IP, ":", addr.Port, " can not connect!\n", connectErr)
		} else {
			connected.Push(addr)
			grpcmgr.peersPool.PutPeer(*addr, peer)
			grpcmgr.logger.Debug("Peer Node ID:", addr.ID, "has connected!")
		}
		if grpcmgr.peersPool.GetAliveNodeNum() >= MaxNum && !flag {
			flag = true
			go func(alive chan int) {
				alive <- 0
			}(grpcmgr.aliveChain)
			owg.Done()
		}
		if unconnected.Length() == 0 {
			break
		}
	}
	if !flag {
		owg.Done()
	}

}

// get all peers info from config file, without self info
func (grpcmgr *GRPCPeerManager) allPeerAddr() []*pb.PeerAddr {
	peers := grpcmgr.configs.Peers()
	plist := make([]*pb.PeerAddr, 0)
	for _, p := range peers {
		if p.ID == grpcmgr.LocalNode.GetNodeID() {
			continue
		}
		addr := pb.NewPeerAddr(p.Address, p.Port, p.Port, p.ID)
		plist = append(plist, addr)
	}
	return plist
}

//newMsg create a new peer msg
//TODO change new peer msg into peer.go
func (grpcmgr *GRPCPeerManager) newMsg(payload []byte, msgType pb.Message_MsgType) *pb.Message {
	msg := &pb.Message{
		MessageType:  msgType,
		Payload:      payload,
		MsgTimeStamp: time.Now().UnixNano(),
		From:         grpcmgr.LocalAddr.ToPeerAddress(),
	}
	signMsg, err := grpcmgr.TM.SignMsg(msg)
	if err != nil {
		grpcmgr.logger.Warningf("sign msg failed, err : %v", err)
	}
	return &signMsg
}

func (grpcmgr *GRPCPeerManager) connectIntro(introAddr pb.PeerAddr) {
	//连接介绍人,并且将其路由表取回,然后进行存储
	introPeer := NewPeer(&introAddr, grpcmgr.LocalAddr, grpcmgr.TM, grpcmgr.CM, grpcmgr.namespace)
	payload, err := proto.Marshal(grpcmgr.LocalAddr.ToPeerAddress())
	if err != nil {
		grpcmgr.logger.Errorf("cannot connect to intro peer,err %v", err)
		return
	}
	routerb, err := introPeer.Connect(payload, pb.Message_INTRODUCE, true, introPeer.IntroHandler)
	if err != nil {
		grpcmgr.logger.Errorf("cannot connect to intro peer,err %v", err)
		return
	}
	routers := routerb.(pb.Routers)

	//将介绍人的信息放入路由表中
	grpcmgr.peersPool.PutPeer(*introPeer.PeerAddr, introPeer)

	grpcmgr.peersPool.MergeFromRoutersToTemp(routers, introPeer.PeerAddr)

	grpcmgr.LocalNode.N = len(grpcmgr.GetAllPeersWithTemp())
}

//connect to peer by ip address and port (why int32? because of protobuf limit)
func (grpcmgr *GRPCPeerManager) connect(peerAddress *pb.PeerAddr) (*Peer, error) {
	//if this node is not online, connect it
	peer := NewPeer(peerAddress, grpcmgr.LocalAddr, grpcmgr.TM, grpcmgr.CM, grpcmgr.namespace)
	_, err := peer.Connect(grpcmgr.TM.GetLocalPublicKey(), pb.Message_HELLO, true, peer.HelloHandler)
	if err != nil {
		// cannot connect to other peer
		grpcmgr.logger.Error("Node ", peerAddress.ID, ":", peerAddress.IP, ":", peerAddress.Port, " can not connect!\nerr: ", err)
		return nil, err
	} else {
		return peer, nil
	}

}

//connect to peer by ip address and port (why int32? because of protobuf limit)
func (grpcmgr *GRPCPeerManager) reconnectToPeer(peerAddress *pb.PeerAddr) (*Peer, error) {
	peer := NewPeer(peerAddress, grpcmgr.LocalAddr, grpcmgr.TM, grpcmgr.CM, grpcmgr.namespace)
	_, err := peer.Connect(grpcmgr.TM.GetLocalPublicKey(), pb.Message_RECONNECT, true, peer.ReconnectHandler)
	if err != nil {
		// cannot connect to other peer
		grpcmgr.logger.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\nerr: ", err)
		return nil, err
	} else {
		return peer, nil
	}

}

func (grpcmgr *GRPCPeerManager) GetRouters() []byte {
	routers := grpcmgr.peersPool.ToRoutingTable()
	payload, err := proto.Marshal(&routers)
	if err != nil {
		grpcmgr.logger.Error("marshal router info failed")
	}
	return payload

}

// GetAllPeers get all connected peer in the peer pool
func (grpcmgr *GRPCPeerManager) GetAllPeers() []*Peer {
	return grpcmgr.peersPool.GetPeers()
}

// GetAllPeers get all connected peer in the peer pool
func (grpcmgr *GRPCPeerManager) GetVPPeers() []*Peer {
	return grpcmgr.peersPool.GetPeers()
}
func (grpcmgr *GRPCPeerManager) GetAllPeersWithTemp() []*Peer {
	return grpcmgr.peersPool.GetPeersWithTemp()
}

// BroadcastPeers Broadcast Massage to connected peers
func (grpcmgr *GRPCPeerManager) BroadcastPeers(payLoad []byte) {
	//log.Warning("P2P broadcast")
	if !grpcmgr.IsOnline {
		grpcmgr.logger.Warningf("this node IS NOT Online ID: %d", grpcmgr.LocalAddr.ID)
		return
	}
	var broadCastMessage = pb.Message{
		MessageType:  pb.Message_SESSION,
		From:         grpcmgr.LocalNode.GetNodeAddr().ToPeerAddress(),
		Payload:      payLoad,
		MsgTimeStamp: time.Now().UnixNano(),
	}
	//log.Warning("call broadcast")
	go broadcast(grpcmgr, broadCastMessage, grpcmgr.peersPool.GetPeers())
}

// inner the broadcast method which serve BroadcastPeers function
func broadcast(grpcmgr *GRPCPeerManager, broadCastMessage pb.Message, peers []*Peer) {
	for _, peer := range peers {
		grpcmgr.logger.Infof("broadcast msg to peer %d->%d", peer.LocalAddr.ID, peer.PeerAddr.ID)
		//REVIEW 这里没有返回值,不知道本次通信是否成功
		//REVIEW 其实这里不需要处理返回值，需要将其go起来
		//REVIEW Chat 方法必须要传实例，否则将会重复加密，请一定要注意！！
		//REVIEW Chat Function must give a message instance, not a point, if not the encrypt will break the payload!
		go func(p2 *Peer) {
			start := time.Now().UnixNano()
			_, err := p2.Chat(broadCastMessage)
			if err == nil {
				grpcmgr.LocalNode.DelayChan <- UpdateTable{updateID: p2.LocalAddr.ID, updateTime: time.Now().UnixNano() - start}
			} else {
				grpcmgr.logger.Warning(grpcmgr.LocalNode.localAddr.ID, ">>", p2.PeerAddr.ID, ": chat failed", err)
			}
		}(peer)

	}
}

func (grpcmgr *GRPCPeerManager) GetLocalAddressPayload() (payload []byte) {
	payload, err := proto.Marshal(grpcmgr.LocalAddr.ToPeerAddress())
	if err != nil {
		grpcmgr.logger.Error("cannot marshal the payload", err)
	}
	testUnmarshal := new(pb.PeerAddress)
	err = proto.Unmarshal(payload, testUnmarshal)
	if err != nil {
		grpcmgr.logger.Error("self unmarshal failed!", err)
	}
	return
}

// SendMsgToPeers Send msg to specific peer peerlist
func (grpcmgr *GRPCPeerManager) SendMsgToPeers(payLoad []byte, peerList []uint64) {
	grpcmgr.logger.Debug("need send message to ", peerList)
	syncMessage := grpcmgr.newMsg(payLoad, pb.Message_SESSION)

	// broadcast to special peers
	for _, NodeID := range peerList {
		peers := grpcmgr.peersPool.GetPeers()
		for _, p := range peers {
			//convert the uint64 to int
			//because the unicast node is not confirm so, here use double loop
			if p.PeerAddr.ID == int(NodeID) {
				start := time.Now().UnixNano()
				_, err := p.Chat(*syncMessage)
				if err != nil {
					grpcmgr.logger.Debug("Broadcast failed,Node", p.LocalAddr.ID)
				} else {
					grpcmgr.LocalNode.DelayChan <- UpdateTable{updateID: p.LocalAddr.ID, updateTime: time.Now().UnixNano() - start}

					//this.eventManager.PostEvent(pb.Message_RESPONSE,*resMsg)
				}
			}
		}

	}

}

func (grpcmgr *GRPCPeerManager) GetPeerInfo() PeerInfos {
	peers := grpcmgr.peersPool.GetPeers()
	localNodeAddr := grpcmgr.LocalNode.GetNodeAddr()

	var keepAliveMessage = pb.Message{
		MessageType:  pb.Message_KEEPALIVE,
		From:         localNodeAddr.ToPeerAddress(),
		Payload:      []byte("Query Status"),
		MsgTimeStamp: time.Now().UnixNano(),
	}
	signMsg, err := grpcmgr.TM.SignMsg(&keepAliveMessage)
	if err != nil {
		grpcmgr.logger.Warning("sign the keep alive msg failed, please check the cert file and priv file.")
	}
	var perinfos PeerInfos
	for _, per := range peers {
		var perinfo PeerInfo
		grpcmgr.logger.Debug("rage the peer")
		perinfo.IP = per.PeerAddr.IP
		perinfo.Port = per.PeerAddr.Port
		perinfo.RPCPort = per.PeerAddr.RPCPort

		retMsg, err := per.Client.Chat(context.Background(), &signMsg)
		if err != nil {
			perinfo.Status = STOP
		} else if retMsg.MessageType == pb.Message_RESPONSE {
			perinfo.Status = ALIVE
		} else if retMsg.MessageType == pb.Message_PENDING {
			perinfo.Status = PENDING
		}
		perinfo.IsPrimary = per.IsPrimary
		grpcmgr.LocalNode.delayTableMutex.RLock()
		perinfo.Delay = grpcmgr.LocalNode.delayTable[per.PeerAddr.ID]
		grpcmgr.LocalNode.delayTableMutex.RUnlock()
		perinfo.ID = per.PeerAddr.ID
		perinfos = append(perinfos, perinfo)
	}
	var self_info = PeerInfo{
		IP:        grpcmgr.LocalAddr.IP,
		Port:      grpcmgr.LocalAddr.Port,
		ID:        grpcmgr.LocalAddr.ID,
		RPCPort:   grpcmgr.LocalAddr.RPCPort,
		Status:    ALIVE,
		IsPrimary: grpcmgr.LocalNode.IsPrimary,
		Delay:     grpcmgr.LocalNode.delayTable[grpcmgr.LocalAddr.ID],
	}
	perinfos = append(perinfos, self_info)
	return perinfos
}

// GetNodeId GetLocalNodeIdHash string
func (grpcmgr *GRPCPeerManager) GetNodeId() int {
	return grpcmgr.LocalAddr.ID
}

func (grpcmgr *GRPCPeerManager) SetPrimary(_id uint64) error {
	__id := strconv.FormatUint(_id, 10)
	id, _err := strconv.Atoi(__id)
	if _err != nil {
		grpcmgr.logger.Error("convert err", _err)
	}
	peers := grpcmgr.peersPool.GetPeers()
	grpcmgr.LocalNode.IsPrimary = false
	for _, per := range peers {
		per.IsPrimary = false
	}

	for _, per := range peers {
		if per.PeerAddr.ID == id {
			per.IsPrimary = true
		}
		if grpcmgr.LocalAddr.ID == id {
			grpcmgr.LocalNode.IsPrimary = true
		}
	}
	return nil
}

func (grpcmgr *GRPCPeerManager) GetLocalNode() *Node {
	return grpcmgr.LocalNode
}

func (grpcmgr *GRPCPeerManager) UpdateAllRoutingTable(routerPayload []byte) {
	toUpdateRouter := new(pb.Routers)
	err := proto.Unmarshal(routerPayload, toUpdateRouter)
	if err != nil {
		grpcmgr.logger.Error(err)
		return
	}
	grpcmgr.logger.Info("Update ALL Router Table")

	for _, r := range toUpdateRouter.Routers {

		if r.Hash == grpcmgr.LocalAddr.Hash {
			continue
		}

		hash := r.Hash
		if _, ok := grpcmgr.peersPool.GetPeerByHash(hash); ok != nil {
			//update hash table
			// add node handler
			peerAddress := pb.NewPeerAddr(r.IP, int(r.Port), int(r.RPCPort), int(r.ID))
			grpcmgr.logger.Debugf("peeraddress to connect %v", peerAddress)
			if peer, connectErr := grpcmgr.reconnectToPeer(peerAddress); connectErr != nil {
				// cannot connect to other peer
				grpcmgr.logger.Error("Node: ", peerAddress.IP, ":", peerAddress.Port, " can not connect!\n", connectErr)
				//TODO retry
				continue
			} else {
				// add  peer to peer pool
				grpcmgr.peersPool.PutPeer(*peerAddress, peer)
				grpcmgr.logger.Debug("Peer Node ID:", peerAddress.ID, "has connected!")
			}

		} else {
			grpcmgr.logger.Debug("this node already in , skip this ,node id:", r.ID)
			//skip this router item
		}

	}

	//delete node handler

	for _, r := range toUpdateRouter.Routers {
		flag := false
		for _, p := range grpcmgr.peersPool.GetPeers() {
			if p.PeerAddr.Hash == r.Hash {
				flag = true
			}
		}

		if !flag {
			grpcmgr.logger.Debugf("delete node (%d)\n", r.ID)
			grpcmgr.peersPool.DeletePeerByHash(r.Hash)
		}
	}

}
func (grpcmgr *GRPCPeerManager) UpdateRoutingTable(payload []byte) {

	if !grpcmgr.IsOnline {
		grpcmgr.logger.Warning(" new node shouldn't call update routing table")
		return
	}
	toUpdateAddress := new(pb.PeerAddress)
	err := proto.Unmarshal(payload, toUpdateAddress)
	if err != nil {
		grpcmgr.logger.Error(err)
		return
	}
	grpcmgr.logger.Debugf("Attend Notify address", toUpdateAddress.ID, toUpdateAddress.IP, toUpdateAddress.Port)
	grpcmgr.logger.Debug("updateRoutingTable")
	newPeer := NewPeer(pb.RecoverPeerAddr(toUpdateAddress), grpcmgr.LocalAddr, grpcmgr.TM, grpcmgr.CM, grpcmgr.namespace)
	if grpcmgr.LocalNode.IsPrimary {
		payload = []byte("true")
	} else {
		payload = []byte("false")
	}
	_, err = newPeer.Connect(payload, pb.Message_ATTEND_NOTIFY, true, newPeer.AttendHandler)
	if err != nil {
	} else {
		grpcmgr.peersPool.MergeTempPeers(newPeer)
		grpcmgr.logger.Debug("add new peer into peerspool, new peer id", newPeer.PeerAddr.ID)
		grpcmgr.LocalNode.N += 1
		grpcmgr.configs.AddNodesAndPersist(grpcmgr.peersPool.GetPeersAddrMap())
	}

}

func (grpcmgr *GRPCPeerManager) SetOnline() {
	grpcmgr.logger.Debug("As a new attend node, save the introducer info into peer addr")
	//新节点
	//新节点在一开始的时候就已经将介绍人的节点列表加入了所以这里不需要处理
	//the new attend node
	grpcmgr.IsOnline = true
	grpcmgr.peersPool.MergeTempPeersForNewNode()
	grpcmgr.logger.Debugf(" NEW PEER SET ONLINE", grpcmgr.LocalNode.localAddr.ID)
	grpcmgr.LocalNode.N = grpcmgr.peersPool.GetAliveNodeNum()
	grpcmgr.configs.AddNodesAndPersist(grpcmgr.peersPool.GetPeersAddrMap())
}

/*********************************
 * delete LocalNode part
 ********************************/
func (grpcmgr *GRPCPeerManager) GetLocalNodeHash() string {
	return grpcmgr.LocalAddr.Hash
}

func (grpcmgr *GRPCPeerManager) GetRouterHashifDelete(hash string) (string, uint64, uint64) {
	hasher := crypto.NewKeccak256Hash("keccak256Hanser")
	grpcmgr.logger.Debug("GetRouterHashifDelete")
	routers := grpcmgr.peersPool.ToRoutingTableWithout(hash)
	var DeleteID uint64
	for _, peer := range grpcmgr.peersPool.GetPeers() {
		if peer.PeerAddr.Hash == hash {
			DeleteID = uint64(peer.PeerAddr.ID)
		}
	}
	if uint64(DeleteID) < uint64(grpcmgr.LocalAddr.ID) {
		return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(grpcmgr.LocalAddr.ID - 1), uint64(DeleteID)
	}

	return hex.EncodeToString(hasher.Hash(routers).Bytes()), uint64(grpcmgr.LocalAddr.ID), uint64(DeleteID)

}

func (grpcmgr *GRPCPeerManager) DeleteNode(hash string) error {

	if grpcmgr.LocalAddr.Hash == hash {
		// delete local node and stop all server
		grpcmgr.logger.Critical("Stop Server")
		grpcmgr.LocalNode.StopServer()
		grpcmgr.peersPool.Clear()
		panic("THIS NODE HAS BEEN QUITTED")
	} else {
		// delete the specific node
		//TODO update node id
		var deleteID int
		for _, per := range grpcmgr.peersPool.GetPeers() {
			if per.PeerAddr.Hash == hash {
				deleteID = per.PeerAddr.ID
				deleteList := grpcmgr.peersPool.DeletePeer(per)
				grpcmgr.logger.Debug("Delete node and persist")
				grpcmgr.configs.DelNodesAndPersist(deleteList)

			}
		}
		for _, per := range grpcmgr.peersPool.GetPeers() {
			if per.PeerAddr.ID > deleteID {
				per.PeerAddr.ID--
			}

		}
		if grpcmgr.LocalAddr.ID > deleteID {
			grpcmgr.LocalAddr.ID--
		}
		return nil

	}
	return nil
}

func (grpcmgr *GRPCPeerManager) Stop() {

}
