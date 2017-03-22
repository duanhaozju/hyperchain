//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/event"
	"hyperchain/admittance"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"net"
	"strconv"
	"sync"
	"time"
	"hyperchain/p2p/peerComm"
	"fmt"
)

type Node struct {
	localAddr          *pb.PeerAddr
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary          bool
	delayTable         map[int]int64
	delayTableMutex    sync.RWMutex
	DelayChan          chan UpdateTable
	attendChan         chan int
	PeersPool          *PeersPool
	N                  int
	DelayTableMutex    sync.Mutex
	TM                 *transport.TransportManager
	CM                 *admittance.CAManager
	config             peerComm.Config
}

type UpdateTable struct {
	updateID   int
	updateTime int64
}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(localAddr *pb.PeerAddr, hEventManager *event.TypeMux, TM *transport.TransportManager, peersPool *PeersPool, cm *admittance.CAManager, config peerComm.Config) *Node {
	var newNode Node
	newNode.localAddr = localAddr
	newNode.TM = TM
	newNode.CM = cm
	newNode.higherEventManager = hEventManager
	newNode.PeersPool = peersPool
	newNode.attendChan = make(chan int, 1000)
	newNode.delayTable = make(map[int]int64)
	newNode.DelayChan = make(chan UpdateTable)
	newNode.config = config
	//listen the update
	go newNode.UpdateDelayTableThread()

	log.Debug("NODE START...")
	log.Debugf("LOCAL NODE INFO:\nID: %d\nIP: %s\nPORT: %d\nHASH: %s", localAddr.ID, localAddr.IP, localAddr.Port, localAddr.Hash)
	return &newNode
}

//监听节点状态更新线程
func (node *Node) UpdateDelayTableThread() {
	for v := range node.DelayChan {
		if v.updateID > 0 {
			node.delayTableMutex.Lock()
			node.delayTable[v.updateID] = v.updateTime
			node.delayTableMutex.Unlock()
		}

	}
}

//新节点需要监听相应的attend类型
func (n *Node) attendNoticeProcess(N int) {
	log.Critical("AttendProcess N:", N)
	// fix the N as N-1
	// temp
	// isPrimaryConnectFlag := false
	f := (N - 1) / 3
	num := 0
	for {
		flag := <-n.attendChan
		log.Debug("attend flag: ", flag, " num: ", num)
		switch flag {
		case 1: {
			num++
		}
		case 2:{
			//isPrimaryConnectFlag =true
			num++
		}
		}
		if num >= (N - f) {
			log.Debug("new node has online ,post already in chain event")
			go n.higherEventManager.Post(event.AlreadyInChainEvent{})
		}

		if num == N {
			break
		}

	}

}

func (node *Node) GetNodeAddr() *pb.PeerAddr {
	return node.localAddr
}

// GetNodeID which init by new function
func (this *Node) GetNodeHash() string {
	return this.localAddr.Hash
}
func (node *Node) GetNodeID() int {
	return node.localAddr.ID
}

func (node *Node)printMsg(msg *pb.Message) {
	switch msg.MessageType {
	case pb.Message_HELLO:
	case pb.Message_HELLO_RESPONSE:
	case pb.Message_ATTEND_NOTIFY:
	case pb.Message_ATTEND_NOTIFY_RESPONSE:
	case pb.Message_ATTEND_RESPONSE:
	case pb.Message_HELLOREVERSE:
	case pb.Message_HELLOREVERSE_RESPONSE:
	case pb.Message_INTRODUCE:
	case pb.Message_INTRODUCE_RESPONSE:
	case pb.Message_KEEPALIVE:
	case pb.Message_CONSUS:
	case pb.Message_PENDING:
	case pb.Message_RECONNECT:
	case pb.Message_RECONNECT_RESPONSE:
	case pb.Message_SYNCMSG:
	case pb.Message_RESPONSE:
	}
}

// Chat Implements the ServerSide Function
func (node *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	response := &pb.Message{
		MessageType:pb.Message_RESPONSE,
		MsgTimeStamp:time.Now().UnixNano(),
		From:node.localAddr.ToPeerAddress(),
		Payload:[]byte("empty"),
	}
	//verify the msg cert first
	if f, e := node.TM.VerifyMsg(msg); !f || e != nil {
		return response, errors.New(fmt.Sprintf("cannot verify ecert or cert signture (verify node %d, err :%v)", node.localAddr.ID,e))
	}
	// TODO pre handle
	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO:{
		response.MessageType = pb.Message_HELLO_RESPONSE
		// hello msg will not accept the nvp peer connect
		// so if a nvp peer connect to node, here will return a error
		if f, e := node.TM.VerifyRCert(msg); !f || e != nil {
			return response, errors.New(fmt.Sprintf("NVP Peer is not allow to send a hello message,peer id: %d", msg.From.ID))
		}
		err := node.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			return response, errors.New(fmt.Sprintf("Cannot complate nego share secret (peer %d)", msg.From.ID))
		}
		response.Payload = node.TM.GetLocalPublicKey()

	}
	case pb.Message_HELLOREVERSE:{
		response.MessageType = pb.Message_HELLOREVERSE_RESPONSE
		// hello reverse msg  accept the nvp peer connect
		// so if a nvp peer connect to node, this should set the nvp peer into nvp peers pool
		if f, e := node.TM.VerifyRCert(msg); !f || e != nil {
			//TODO set the peer into nvo peers pool
			return response, errors.New(fmt.Sprintf("NVP Peer is not allow to send a hello message,peer id: %d", msg.From.ID))
		}
		err := node.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			return response, errors.New(fmt.Sprintf("Cannot complate nego share secret (peer %d)", msg.From.ID))
		}
		//every times get the public key is same
		response.Payload = node.TM.GetLocalPublicKey()
	}
	case pb.Message_RECONNECT:{
		response.MessageType = pb.Message_RECONNECT_RESPONSE
		// hello reverse msg  accept the nvp peer connect
		// so if a nvp peer connect to node, this should set the nvp peer into nvp peers pool
		if f, e := node.TM.VerifyRCert(msg); !f || e != nil {
			//TODO set the peer into nvo peers pool
			return response, errors.New(fmt.Sprintf("NVP Peer is not allow to send a hello message,peer id: %d", msg.From.ID))
		}
		err := node.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			return response, errors.New(fmt.Sprintf("Cannot complate nego share secret (peer %d)", msg.From.ID))
		}
		//every times get the public key is same
		response.Payload = node.TM.GetLocalPublicKey()
		go node.reverseConnect(msg)

	}
	case pb.Message_INTRODUCE:{
		//TODO 验证签名
		//返回路由表信息
		response.MessageType = pb.Message_INTRODUCE_RESPONSE
		routers := node.PeersPool.ToRoutingTable()
		response.Payload, _ = proto.Marshal(&routers)
	}
	case pb.Message_ATTEND:{
		//新节点全部连接上之后通知
		go node.higherEventManager.Post(event.NewPeerEvent{
			Payload: msg.Payload,
		})
		response.MessageType = pb.Message_ATTEND_RESPONSE
		// hello reverse msg  accept the nvp peer connect
		// so if a nvp peer connect to node, this should set the nvp peer into nvp peers pool
		if f, e := node.TM.VerifyRCert(msg); !f || e != nil {
			//TODO set the peer into nvo peers pool
			return response, errors.New(fmt.Sprintf("NVP Peer is not allow to send a hello message,peer id: %d", msg.From.ID))
		}
		err := node.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			return response, errors.New(fmt.Sprintf("Cannot complate nego share secret (peer %d)", msg.From.ID))
		}
		//every times get the public key is same
		response.Payload = node.TM.GetLocalPublicKey()

	}
	case pb.Message_ATTEND_NOTIFY:{
		// here need to judge if update
		//if primary
		if string(msg.Payload) == "true" {
			node.attendChan <- 2
		} else {
			node.attendChan <- 1
		}
		response.MessageType = pb.Message_ATTEND_NOTIFY_RESPONSE
		// hello reverse msg  accept the nvp peer connect
		// so if a nvp peer connect to node, this should set the nvp peer into nvp peers pool
		if f, e := node.TM.VerifyRCert(msg); !f || e != nil {
			//TODO set the peer into nvo peers pool
			return response, errors.New(fmt.Sprintf("NVP Peer is not allow to send a hello message,peer id: %d", msg.From.ID))
		}
		err := node.TM.NegoShareSecret(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			return response, errors.New(fmt.Sprintf("Cannot complate nego share secret (peer %d)", msg.From.ID))
		}
		//every times get the public key is same
		response.Payload = node.TM.GetLocalPublicKey()

	}
	case pb.Message_CONSUS:{
		log.Infof("Get a consus msg from %d (ip: %s, port: %d) ",msg.From.ID,msg.From.IP,msg.From.Port)
		transferData, err := node.TM.Decrypt(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			log.Errorf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err)
			return response, errors.New(fmt.Sprintf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		go node.higherEventManager.Post(event.ConsensusEvent{
			Payload: transferData,
		})
		payload := []byte("GOT_A_CONSENSUS_MESSAGE")
		rpayload,err := node.TM.Encrypt(payload,pb.RecoverPeerAddr(msg.From))
		if err != nil{
			return nil, errors.New(fmt.Sprintf("Sync message encrypt error message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		response.Payload = rpayload
	}
	case pb.Message_SYNCMSG:{
		// package the response msg
		if err := node.handleSyncMsg(msg); err != nil {
			response.Payload = []byte("Sync message Unmarshal error")
			return nil, errors.New(fmt.Sprintf("Sync message Unmarshal error message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		payload := []byte("GOT_A_SYNC_MESSAGE")
		rpayload,err := node.TM.Encrypt(payload,pb.RecoverPeerAddr(msg.From))
		if err != nil{
			return nil, errors.New(fmt.Sprintf("Sync message encrypt error message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		response.Payload = rpayload

	}
	case pb.Message_KEEPALIVE:{
		log.Debugf("Get a keep alive msg from %d (ip: %s, port: %d) ",msg.From.ID,msg.From.IP,msg.From.Port)
		//客户端会发来keepAlive请求,返回response即可
		// client may send a keep alive request, just response A response type message,if node is not ready, send a pending status message
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("RESPONSE FROM SERVER")
	}
	case pb.Message_PENDING:{
		log.Warning("Got a PADDING Message")
	}
	default:
		log.Warningf("ignore a unknown message : %v, content", msg.MessageType, string(msg.Payload))
		return response, errors.New(fmt.Sprintf("ignore a unknown message : %v, content: %s", msg.MessageType, string(msg.Payload)))
	}

	signMsg, err := node.TM.SignMsg(response)
	if err != nil {
		return response, errors.New(fmt.Sprintf("Cannot complate sign the response (peer %d)", msg.From.ID))
	}
	return &signMsg, nil
}

func (node *Node)handleSyncMsg(msg *pb.Message) error {
	log.Infof("Get async msg from %d (ip: %s, port: %d) ",msg.From.ID,msg.From.IP,msg.From.Port)
	//todo should handle in pre handle method
	transferData, err := node.TM.Decrypt(msg.Payload, pb.RecoverPeerAddr(msg.From))
	if err != nil {
		log.Errorf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err)
		return  errors.New(fmt.Sprintf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
	}
	var syncMsg recovery.Message
	unMarshalErr := proto.Unmarshal(transferData, &syncMsg)
	if unMarshalErr != nil {
		log.Error("sync UnMarshal error!")
		return unMarshalErr
	}
	switch syncMsg.MessageType {
	case recovery.Message_SYNCBLOCK:
		{

			go node.higherEventManager.Post(event.SyncBlockReceiveEvent{
				Payload: syncMsg.Payload,
			})

		}
	case recovery.Message_SYNCCHECKPOINT:
		{
			go node.higherEventManager.Post(event.SyncBlockReqEvent{
				Payload: syncMsg.Payload,
			})

		}
	case recovery.Message_SYNCSINGLE:
		{
			go node.higherEventManager.Post(event.SyncBlockReqEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_RELAYTX:
		{
			//log.Warning("Message_RELAYTX: ")
			go node.higherEventManager.Post(event.ConsensusEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_INVALIDRESP:
		{
			go node.higherEventManager.Post(event.InvalidTxsEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_SYNCREPLICA:
		{
			go node.higherEventManager.Post(event.ReplicaInfoEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_BROADCAST_NEWPEER:
		{
			log.Debug("receive Message_BROADCAST_NEWPEER")
			go node.higherEventManager.Post(event.RecvNewPeerEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_BROADCAST_DELPEER:
		{
			log.Debug("receive Message_BROADCAST_DELPEER")
			go node.higherEventManager.Post(event.RecvDelPeerEvent{
				Payload: syncMsg.Payload,
			})
		}
	case recovery.Message_VERIFIED_BLOCK:
		{
			log.Debug("receive Message_BROADCAST_DELPEER")
			go node.higherEventManager.Post(event.ReceiveVerifiedBlock{
				Payload: syncMsg.Payload,
			})
		}
	}
	return nil
}

// StartServer start the gRPC server
func (node *Node) StartServer() {
	log.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(node.localAddr.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	opts := node.CM.GetGrpcServerOpts()
	node.gRPCServer = grpc.NewServer(opts...)
	//this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(node.gRPCServer, node)
	log.Info("Listening gRPC request...")
	go node.gRPCServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (node *Node) StopServer() {
	node.gRPCServer.GracefulStop()

}

func (node *Node) reverseConnect(msg *pb.Message) error {
	//REVIEW FIX get local config reverse connect address
	reNodeID := int(msg.From.ID)
	reAddr := pb.NewPeerAddr(node.config.GetIP(reNodeID), node.config.GetPort(reNodeID), node.config.GetRPCPort(reNodeID), reNodeID)
	peer, err := node.PeersPool.GetPeerByHash(msg.From.Hash)
	if err != nil {
		log.Criticalf("cannot get the old peer ID: %d",msg.From.ID)
		peer := NewPeer(reAddr,node.localAddr,node.TM,node.CM)
		_,err := peer.Connect(node.TM.GetLocalPublicKey(),pb.Message_RECONNECT,true,peer.ReconnectHandler)
		if err != nil{
			log.Criticalf("cannot create a new peer again, please chekc the peer (id: %d)",msg.From.ID)
		}
	} else {
		go peer.eventMux.Post(RecoveryEvent{
			addr:reAddr,
			recoveryTimeout:node.CM.RecoveryTimeout,
			recoveryTimes:0,
		})
	}
	return nil
}
