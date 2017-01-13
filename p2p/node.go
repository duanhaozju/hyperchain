//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"encoding/hex"
	"errors"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/event"
	"hyperchain/membersrvc"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/p2p/transport"
	"hyperchain/recovery"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	address            *pb.PeerAddress
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary       bool
	delayTable      map[uint64]int64
	delayTableMutex sync.RWMutex
	DelayChan       chan UpdateTable
	sentEvent       bool
	attendChan      chan int
	PeerPool        *PeersPool
	N               int
	DelayTable      map[uint64]int64
	DelayTableMutex sync.Mutex
	TEM             transport.TransportEncryptManager
}

type UpdateTable struct {
	updateID   uint64
	updateTime int64
}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int64, hEventManager *event.TypeMux, nodeID uint64, TEM transport.TransportEncryptManager, peersPool *PeersPool) *Node {
	var newNode Node
	newNode.address = peerComm.ExtractAddress(peerComm.GetLocalIp(), port, nodeID)
	newNode.TEM = TEM
	newNode.higherEventManager = hEventManager
	newNode.DelayTable = make(map[uint64]int64)
	newNode.PeerPool = peersPool
	newNode.attendChan = make(chan int)
	newNode.sentEvent = false
	newNode.delayTable = make(map[uint64]int64)
	newNode.DelayChan = make(chan UpdateTable)
	//listen the update
	go newNode.UpdateDelayTableThread()

	log.Debug("节点启动")
	log.Debug("本地节点hash", newNode.address.Hash)
	log.Debug("本地节点ip", newNode.address.IP)
	log.Debug("本地节点port", newNode.address.Port)
	return &newNode

}

func (this *Node) UpdateDelayTableThread() {
	for v := range this.DelayChan {
		if v.updateID > 0 {
			this.delayTableMutex.Lock()
			this.delayTable[v.updateID] = v.updateTime
			this.delayTableMutex.Unlock()
		}

	}
}

//新节点需要监听相应的attend类型
func (this *Node) attendNoticeProcess(N int) {
	f := int(math.Floor(float64((N - 1) / 3)))
	num := 0
	for {
		select {
		case attendFlag := <-this.attendChan:
			{
				log.Debug("连接到一个节点...!!!!!! N:", N, "f", f, "num", num)
				if attendFlag == 1 {
					num += 1
					if num >= (N-f) && !this.sentEvent {
						//TODO 修改向上post的消息类型
						log.Debug("新节点已经连接到chain上>>>>><<<<<<<<<<")
						this.higherEventManager.Post(event.AlreadyInChainEvent{})
						this.sentEvent = true
						num = 0

					}
				} else {
					log.Warning("非法链接...!!!!!! N:", N, "f", f, "num", num)
				}
			}
		}
	}
}

func (this *Node) GetNodeAddr() *pb.PeerAddress {
	return this.address
}

// GetNodeID which init by new function
func (this *Node) GetNodeHash() string {
	return this.address.Hash
}
func (this *Node) GetNodeID() uint64 {
	return this.address.ID
}

// Chat Implements the ServerSide Function
func (this *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	var response pb.Message
	response.MessageType = pb.Message_RESPONSE
	response.MsgTimeStamp = time.Now().UnixNano()
	response.From = this.address
	//handle the message
	log.Debug("消息类型", msg.MessageType)
	switch msg.MessageType {
	case pb.Message_HELLO:
		{
			log.Debug("=================================")
			log.Debug("negotiating key")
			log.Debug("local addr is ", this.address.ID, this.address.IP, this.address.Port)
			log.Debug("remote addr is", msg.From.ID, msg.From.IP, msg.From.Port)
			log.Debug("=================================")
			response.MessageType = pb.Message_HELLO_RESPONSE
			//review 协商密钥
			remotePublicKey := msg.Payload
			genErr := this.PeerPool.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Notice("remote addr hash：", msg.From.Hash)
			log.Notice("negotiated key is ", this.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := this.PeerPool.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			response.Payload = transportPublicKey
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接

			return &response, nil
		}
	case pb.Message_HELLO_RESPONSE:
		{
			log.Warning("Invalidate HELLO_RESPONSE message")
		}
	case pb.Message_RECONNECT:
		{
			log.Warning("A Node is Reconnecting")
			response.MessageType = pb.Message_RECONNECT_RESPONSE
			remotePublicKey := msg.Payload
			genErr := this.PeerPool.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Warning("reconnect remote id:", msg.From.ID, this.PeerPool.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := this.PeerPool.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接需要重新建立新链接
			go this.reconnect(msg)
			response.Payload = transportPublicKey

		}
	case pb.Message_RECONNECT_RESPONSE:
		{
			log.Debug("=================================")
			log.Debug("negotiate keys ")
			log.Debug("", this.address.ID, this.address.IP, this.address.Port)
			log.Debug("remote addr is ", msg.From.ID, msg.From.IP, msg.From.Port)
			log.Debug("=================================")
			response.MessageType = pb.Message_HELLO_RESPONSE
			remotePublicKey := msg.Payload
			genErr := this.PeerPool.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Warning("Message_HELLO_RESPONSE remote id:", msg.From.ID, this.PeerPool.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := this.PeerPool.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			response.Payload = transportPublicKey
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接

			return &response, nil

		}
	case pb.Message_INTRODUCE:
		{
			//返回路由表信息
			response.MessageType = pb.Message_INTRODUCE_RESPONSE
			routers := this.PeerPool.ToRoutingTable()
			response.Payload, _ = proto.Marshal(&routers)
		}
	case pb.Message_INTRODUCE_RESPONSE:
		{
			log.Warning("节点已经接受请求", msg.From)
			//this.higherEventManager.Post(event.)

		}
	case pb.Message_ATTEND:
		{
			log.Debug("Message_ATTEND############")
			//新节点全部连接上之后通知
			this.higherEventManager.Post(event.NewPeerEvent{
				Payload: msg.Payload,
			})
			//response
			//response.MessageType = pb.Message_ATTEND_RESPNSE
		}
	case pb.Message_ATTEND_RESPNSE:
		{
			//这里需要进行判断是并且进行更新
			this.attendChan <- 1
		}
	case pb.Message_CONSUS:
		{
			log.Debug("<<<< GOT A CONSUS MESSAGE >>>>")

			log.Debug("×××××× Node Decode MSG ××××××")
			log.Debug("Node need to decode msg: ", hex.EncodeToString(msg.Payload))
			transferData, err := this.PeerPool.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil {
				log.Error("cannot decode the message", err)
				return nil, err
			}
			//log.Debug("Node解密后信息", hex.EncodeToString(transferData))
			//log.Debug("Node解密后信息2", string(transferData))
			response.Payload = []byte("GOT_A_CONSENSUS_MESSAGE")
			if string(transferData) == "TEST" {
				response.Payload = []byte("GOT_A_TEST_CONSENSUS_MESSAGE")
			}
			log.Debug("from Node", msg.From.ID)
			log.Debug(hex.EncodeToString(transferData))

			go this.higherEventManager.Post(
				event.ConsensusEvent{
					Payload: transferData,
				})
		}
	case pb.Message_SYNCMSG:
		{
			// package the response msg
			response.MessageType = pb.Message_RESPONSE
			transferData, err := this.PeerPool.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil {
				log.Error("cannot decode the message", err)
				return nil, err
			}
			response.Payload = []byte("got a sync msg")
			log.Debug("<<<< GOT A Unicast MESSAGE >>>>")
			var SyncMsg recovery.Message
			unMarshalErr := proto.Unmarshal(transferData, &SyncMsg)
			if unMarshalErr != nil {
				response.Payload = []byte("Sync message Unmarshal error")
				log.Error("sync UnMarshal error!")
			}
			switch SyncMsg.MessageType {
			case recovery.Message_SYNCBLOCK:
				{

					go this.higherEventManager.Post(event.ReceiveSyncBlockEvent{
						Payload: SyncMsg.Payload,
					})

				}
			case recovery.Message_SYNCCHECKPOINT:
				{
					go this.higherEventManager.Post(event.StateUpdateEvent{
						Payload: SyncMsg.Payload,
					})

				}
			case recovery.Message_SYNCSINGLE:
				{
					go this.higherEventManager.Post(event.StateUpdateEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_RELAYTX:
				{
					//log.Warning("Message_RELAYTX: ")
					go this.higherEventManager.Post(event.ConsensusEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_INVALIDRESP:
				{
					go this.higherEventManager.Post(event.RespInvalidTxsEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_SYNCREPLICA:
				{
					go this.higherEventManager.Post(event.ReplicaStatusEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_BROADCAST_NEWPEER:
				{
					log.Debug("receive Message_BROADCAST_NEWPEER")
					go this.higherEventManager.Post(event.RecvNewPeerEvent{
						Payload: SyncMsg.Payload,
					})
				}
			case recovery.Message_BROADCAST_DELPEER:
				{
					log.Debug("receive Message_BROADCAST_DELPEER")
					go this.higherEventManager.Post(event.RecvDelPeerEvent{
						Payload: SyncMsg.Payload,
					})
				}
			}
		}
	case pb.Message_KEEPALIVE:
		{
			//客户端会发来keepAlive请求,返回response即可
			// client may send a keep alive request, just response A response type message,if node is not ready, send a pending status message
			response.MessageType = pb.Message_RESPONSE
			response.Payload = []byte("RESPONSE FROM SERVER")
		}
	case pb.Message_RESPONSE:
		{
			// client couldn't send a response message to server, so server should never receive a response type message
			log.Warning("Client Send a Response Message to Server, this is not allowed!")

		}
	case pb.Message_PENDING:
		{
			log.Warning("Got a PADDING Message")
		}
	default:
		log.Warning(msg.MessageType)
		log.Warning("Unkown Message type!")
	}
	if msg.MessageType != pb.Message_HELLO && msg.MessageType != pb.Message_HELLO_RESPONSE && msg.MessageType != pb.Message_RECONNECT_RESPONSE && msg.MessageType != pb.Message_RECONNECT {
		var err error

		response.Payload, err = this.PeerPool.TEM.EncWithSecret(response.Payload, msg.From.Hash)
		if err != nil {
			log.Error("encode error", err)
		}

	}
	return &response, nil
}

// StartServer start the gRPC server
func (this *Node) StartServer() {
	log.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(int(this.address.Port)))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	opts := membersrvc.GetGrpcServerOpts()
	this.gRPCServer = grpc.NewServer(opts...)
	//this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(this.gRPCServer, this)
	log.Info("Listening gRPC request...")
	go this.gRPCServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (this *Node) StopServer() {
	this.gRPCServer.GracefulStop()

}

func (this *Node) reconnect(msg *pb.Message) {
	opts := membersrvc.GetGrpcClientOpts()
	conn, err := grpc.Dial(msg.From.IP+":"+strconv.Itoa(int(msg.From.Port)), opts...)
	//conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		errors.New("Cannot establish a connection!")
		log.Error("err:", err)
	}

	Client := pb.NewChatClient(conn)
	if _, ok := this.PeerPool.peers[msg.From.Hash]; ok {
		log.Warning("This remote Node already existed, and try to reconnect...")
	} else {
		return
	}
	this.PeerPool.peers[msg.From.Hash].Client = Client
	this.PeerPool.peers[msg.From.Hash].Connection = conn
	//esatblish the connect and regenerate the secrate
	helloMessage := pb.Message{
		MessageType:  pb.Message_RECONNECT_RESPONSE,
		Payload:      this.PeerPool.TEM.GetLocalPublicKey(),
		From:         this.address,
		MsgTimeStamp: time.Now().UnixNano(),
	}

	retMessage, err2 := this.PeerPool.peers[msg.From.Hash].Client.Chat(context.Background(), &helloMessage)
	if err2 != nil {
		log.Error("cannot establish a connection", err2)
	} else {
		//review 取得对方的秘钥
		if retMessage.MessageType == pb.Message_HELLO_RESPONSE {
			remotePublicKey := retMessage.Payload
			genErr := this.PeerPool.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("genErr", err)
			}

			if err != nil {
				log.Error("cannot decrypt the nodeidinfo!")
				errors.New("Decrypt ERROR")
			}
			log.Critical("reconnect to reverse ID :", msg.From.ID, this.PeerPool.TEM.GetSecret(msg.From.Hash))
			log.Critical(this.PeerPool.TEM.GetSecret(msg.From.Hash))
		}
	}
}
