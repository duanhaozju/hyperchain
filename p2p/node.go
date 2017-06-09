//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/manager/event"
	"hyperchain/admittance"
	pb "hyperchain/p2p/message"
	"hyperchain/p2p/transport"
	"net"
	"strconv"
	"sync"
	"time"
	pc "hyperchain/p2p/common"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/p2p/network"
	"hyperchain/p2p/msg"
)

type Node struct {
	localAddr          *pb.PeerAddr
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary       bool
	delayTable      map[int]int64
	delayTableMutex sync.RWMutex
	DelayChan       chan UpdateTable
	attendChan      chan int
	PeersPool       *PeersPool
	N               int
	DelayTableMutex sync.Mutex
	TM              *transport.TransportManager
	CM              *admittance.CAManager
	config          pc.Config
	namespace       string
	logger          *logging.Logger
	// release 1.3
	server *network.Server
	router MsgRouter
}

type UpdateTable struct {
	updateID   int
	updateTime int64
}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode() *Node {
	return Node{
		server : network.Server{},
	}
}

func(node *Node)Start(){
	node.server.RegisterSlot(pb.Message_HELLO,msg.NewHelloHandler(node.router.BlackHole()))
	node.server.RegisterSlot(pb.Message_KEEPALIVE,msg.NewKeepAliveHandler(node.router.BlackHole()))
	node.server.StartServer()
}

//新节点需要监听相应的attend类型
func (n *Node) attendNoticeProcess(N int) {
	n.logger.Critical("AttendProcess N:", N)
	// fix the N as N-1
	// temp
	// isPrimaryConnectFlag := false
	f := (N - 1) / 3
	num := 0
	for {
		flag := <-n.attendChan
		n.logger.Debug("attend flag: ", flag, " num: ", num)
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
			n.logger.Debug("new node has online ,post already in chain event")
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
	case pb.Message_PENDING:
	case pb.Message_RECONNECT:
	case pb.Message_RECONNECT_RESPONSE:
	case pb.Message_RESPONSE:
	case pb.Message_SESSION:
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
	case pb.Message_SESSION:{
		node.logger.Infof("Get a session msg from %d (ip: %s, port: %d) ",msg.From.ID,msg.From.IP,msg.From.Port)
		transferData, err := node.TM.Decrypt(msg.Payload, pb.RecoverPeerAddr(msg.From))
		if err != nil {
			node.logger.Errorf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err)
			return response, errors.New(fmt.Sprintf("cannot decrypt the message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		go node.higherEventManager.Post(event.SessionEvent{
			Message:  transferData,
		})
		payload := []byte("GOT_A_SESSION_MESSAGE")
		rpayload,err := node.TM.Encrypt(payload,pb.RecoverPeerAddr(msg.From))
		if err != nil{
			return nil, errors.New(fmt.Sprintf("Sync message encrypt error message(%d -> %d),%v", msg.From.ID, node.localAddr.ID, err))
		}
		response.Payload = rpayload
	}
	case pb.Message_KEEPALIVE:{
		node.logger.Debugf("Get a keep alive msg from %d (ip: %s, port: %d) ",msg.From.ID,msg.From.IP,msg.From.Port)
		//客户端会发来keepAlive请求,返回response即可
		// client may send a keep alive request, just response A response type message,if node is not ready, send a pending status message
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("RESPONSE FROM SERVER")
	}
	case pb.Message_PENDING:{
		node.logger.Warning("Got a PADDING Message")
	}
	default:
		node.logger.Warningf("ignore a unknown message : %v, content", msg.MessageType, string(msg.Payload))
		return response, errors.New(fmt.Sprintf("ignore a unknown message : %v, content: %s", msg.MessageType, string(msg.Payload)))
	}

	signMsg, err := node.TM.SignMsg(response)
	if err != nil {
		return response, errors.New(fmt.Sprintf("Cannot complate sign the response (peer %d)", msg.From.ID))
	}
	return &signMsg, nil
}

// StartServer start the gRPC server
func (node *Node) StartServer() error {
	node.logger.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(node.localAddr.Port))
	if err != nil {
		node.logger.Fatalf("Failed to listen: %v", err)
		node.logger.Fatal("PLEASE RESTART THE SERVER NODE!")
		return err
	}
	opts := node.CM.GetGrpcServerOpts()
	node.gRPCServer = grpc.NewServer(opts...)
	pb.RegisterChatServer(node.gRPCServer, node)
	node.logger.Info("Listening gRPC request...")
	go node.gRPCServer.Serve(lis)
	return nil
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (node *Node) StopServer() {
	//node.gRPCServer.GracefulStop()
	node.gRPCServer.Stop()
}

func (node *Node) reverseConnect(msg *pb.Message) error {
	//REVIEW FIX get local config reverse connect address
	reNodeID := int(msg.From.ID)
	reAddr := pb.NewPeerAddr(node.config.GetIP(reNodeID), node.config.GetPort(reNodeID), node.config.GetRPCPort(reNodeID), reNodeID)
	peer, err := node.PeersPool.GetPeerByHash(msg.From.Hash)
	if err != nil {
		node.logger.Criticalf("cannot get the old peer ID: %d",msg.From.ID)
		peer := NewPeer(reAddr,node.localAddr,node.TM,node.CM,node.namespace)
		_,err := peer.Connect(node.TM.GetLocalPublicKey(),pb.Message_RECONNECT,true,peer.ReconnectHandler)
		if err != nil{
			node.logger.Criticalf("cannot create a new peer again, please chekc the peer (id: %d)",msg.From.ID)
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
