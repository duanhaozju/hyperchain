//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package p2p

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hyperchain/event"
	"hyperchain/p2p/peerComm"
	pb "hyperchain/p2p/peermessage"
	"hyperchain/recovery"
	"net"
	"strconv"
	"time"
	//"hyperchain/membersrvc"

	"hyperchain/membersrvc"
	"sync"
)


type Node struct {
	address            *pb.PeerAddress
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
	//common information
	IsPrimary          bool
	DelayTable         map[uint64]int64
	DelayTableMutex    sync.RWMutex
	PeerPool           *PeersPool

}

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int64, hEventManager *event.TypeMux, nodeID uint64, peerspool *PeersPool) *Node {
	var newNode Node
	newNode.address = peerComm.ExtractAddress(peerComm.GetLocalIp(), port, nodeID)
	newNode.higherEventManager = hEventManager
	newNode.DelayTable = make(map[uint64]int64)
	newNode.PeerPool = peerspool


	log.Debug("节点启动")
	log.Debug("本地节点hash", newNode.address.Hash)
	log.Debug("本地节点ip", newNode.address.IP)
	log.Debug("本地节点port", newNode.address.Port)

	return &newNode

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
	log.Debug("MSG TYPE:", msg.MessageType)
	go func() {
		this.DelayTableMutex.Lock()
		this.DelayTable[msg.From.ID] = time.Now().UnixNano() - msg.MsgTimeStamp
		this.DelayTableMutex.Unlock()
	}()
	switch msg.MessageType {
	case pb.Message_HELLO:
		{
			log.Debug("=================================")
			log.Debug("negotiate the keys")
			log.Debug("local addr is :", this.address.ID, this.address.IP, this.address.Port)
			log.Debug("remote addr is :", msg.From.ID, msg.From.IP, msg.From.Port)
			log.Debug("=================================")
			response.MessageType = pb.Message_HELLO_RESPONSE
			remotePublicKey := msg.Payload
			genErr := this.PeerPool.TEM.GenerateSecret(remotePublicKey, msg.From.Hash)
			if genErr != nil {
				log.Error("gen sec error", genErr)
			}
			log.Warning("Now Say Hello from:", msg.From.ID, this.PeerPool.TEM.GetSecret(msg.From.Hash))
			//every times get the public key is same
			transportPublicKey := this.PeerPool.TEM.GetLocalPublicKey()
			//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
			response.Payload = transportPublicKey
			//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
			//REVIEW This no need to call hello event handler
			//判断是否需要反向建立链接
			go this.reconnect(msg)
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
	case pb.Message_CONSUS:
		{
			log.Debug("<<<< GOT A CONSUS MESSAGE >>>>")

			log.Debug("×××××× Node Decode MSG ××××××")
			log.Debug("Node need to decode msg: ", hex.EncodeToString(msg.Payload))
			transferData,err := this.PeerPool.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil{
				log.Error("cannot decode the message",err)
				return nil,err
			}
			//log.Debug("Node解密后信息", hex.EncodeToString(transferData))
			//log.Debug("Node解密后信息2", string(transferData))
			response.Payload = []byte("GOT_A_CONSENSUS_MESSAGE")
			if string(transferData) == "TEST" {
				response.Payload = []byte("GOT_A_TEST_CONSENSUS_MESSAGE")
			}
			log.Debug("From Node ", msg.From.ID)
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
			transferData,err := this.PeerPool.TEM.DecWithSecret(msg.Payload, msg.From.Hash)
			if err != nil{
				log.Error("cannot decode the message",err);
				return nil,err
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
		log.Warning("Unkown Message type!")
	}
	// 返回信息加密
	if msg.MessageType != pb.Message_HELLO && msg.MessageType != pb.Message_HELLO_RESPONSE && msg.MessageType != pb.Message_RECONNECT_RESPONSE && msg.MessageType != pb.Message_RECONNECT {
		var err error;
		response.Payload,err = this.PeerPool.TEM.EncWithSecret(response.Payload, msg.From.Hash)
		if err != nil {
			log.Error("encode error",err)
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

func (this *Node)reconnect(msg *pb.Message) {
	if this.PeerPool.GetPeer(*msg.From) == nil{
		//this situation needn't reconnect
		return
	}
	peer,err :=NewPeerByIpAndPort(msg.From.IP,msg.From.Port,msg.From.ID,this.PeerPool.TEM,this.GetNodeAddr(),this.PeerPool)

	if err != nil {
		log.Warning("Cannot reverse connect to remote peer",msg.From)
	}
	this.PeerPool.PutPeer(*msg.From,peer)
}
