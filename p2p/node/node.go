// gRPC Server
// author: Chen Quan
// date: 2016-08-24
// last modified:2016-08-24
// change log:  1. add a header comment of this file
//		2. modified the hello message handler, DO NOT save the peer into the peer pool

package Server

import (
	pb "hyperchain/p2p/peermessage"
	"golang.org/x/net/context"
	"net"
	"google.golang.org/grpc"
	"strconv"
	"hyperchain/p2p/peerComm"
	"hyperchain/event"
	"github.com/op/go-logging"
	"hyperchain/p2p/transport"
	"github.com/golang/protobuf/proto"
	"hyperchain/recovery"
	"encoding/hex"
	"time"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/Server")
}
type Node struct {
	address		*pb.PeerAddress
	gRPCServer	*grpc.Server
	NodeID		string
	higherEventManager *event.TypeMux
	//common information
	Cname	     string
	TEM transport.TransportEncryptManager
}



// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int, hEventManager *event.TypeMux,nodeID int,Cname string,TEM transport.TransportEncryptManager) *Node {
	var newNode Node
	newNode.address = peerComm.ExtractAddress(peerComm.GetLocalIp(),port,int32(nodeID))
	newNode.Cname = Cname
	newNode.TEM = TEM
	newNode.NodeID = strconv.Itoa(nodeID)
	newNode.higherEventManager = hEventManager


	return &newNode

}
func (this *Node)GetNodeAddr() *pb.PeerAddress {
	return this.address
}
// GetNodeID which init by new function
func (this *Node)GetNodeHash() string{
	return this.address.Hash
}
func (this *Node)GetNodeID() string{
	return this.NodeID
}
// Chat Implements the ServerSide Function
func (this *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	var response pb.Message
	response.MessageType = pb.Message_RESPONSE
	response.MsgTimeStamp = time.Now().UnixNano()
	response.From = this.address
	//handle the message
	//review decrypt
	log.Debug("消息类型",msg.MessageType)

	switch msg.MessageType {
	case pb.Message_HELLO :{
		log.Debug("=================================")
		log.Debug("协商秘钥")
		log.Debug("本地地址为",this.address.ID,this.address.Ip,this.address.Port)
		log.Debug("远端地址为",msg.From.ID,msg.From.Ip,msg.From.Port)
		log.Debug("=================================")
		response.MessageType = pb.Message_HELLO_RESPONSE
		//review 协商密钥
		remotePublicKey := msg.Payload
		this.TEM.GenerateSecret(remotePublicKey,msg.From.Hash)
		//log.Notice("协商秘钥为：",this.TEM.GetSecret(msg.From.Hash))
		transportPublicKey := this.TEM.GetLocalPublicKey()
		//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
		response.Payload = transportPublicKey
		//REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
		//REVIEW This no need to call hello event handler
		return &response, nil
	}
	case pb.Message_HELLO_RESPONSE :{
		log.Warning("Invalidate HELLO_RESPONSE message")
	}
	case pb.Message_CONSUS:{
		log.Debug("<<<< GOT A CONSUS MESSAGE >>>>")

		log.Debug("××××××Node解密信息××××××")
		log.Debug("Node待解密信息",hex.EncodeToString(msg.Payload))
		transferData := this.TEM.DecWithSecret(msg.Payload,msg.From.Hash)
		log.Debug("Node解密后信息",hex.EncodeToString(transferData))
		log.Debug("Node解密后信息2",string(transferData))
		response.Payload =  []byte("GOT_A_CONSENSUS_MESSAGE")
		if string(transferData) == "TEST"{
			response.Payload = []byte("GOT_A_TEST_CONSENSUS_MESSAGE")
		}
		log.Debug("来自节点",msg.From.ID)
		log.Debug(hex.EncodeToString(transferData))

		go this.higherEventManager.Post(event.ConsensusEvent{
			Payload:transferData,
		})
	}
	case pb.Message_SYNCMSG:{
		// package the response msg
		response.MessageType = pb.Message_RESPONSE

		response.Payload = []byte("got a sync msg")
		log.Debug("<<<< GOT A SYNC MESSAGE >>>>")
		var SyncMsg recovery.Message
		unMarshalErr := proto.Unmarshal(msg.Payload,&SyncMsg)
		if unMarshalErr != nil{
			response.Payload = []byte("Sync message Unmarshal error")
			log.Error("sync UnMarshal error!")
		}
		switch SyncMsg.MessageType {
		case recovery.Message_SYNCBLOCK:{

			go this.higherEventManager.Post(event.ReceiveSyncBlockEvent{
				Payload:SyncMsg.Payload,
			})

		}
		case recovery.Message_SYNCCHECKPOINT:{
			go this.higherEventManager.Post(event.StateUpdateEvent{
				Payload:SyncMsg.Payload,
			})

		}
		case recovery.Message_RELAYTX:{

			go this.higherEventManager.Post(event.ConsensusEvent{
				Payload:SyncMsg.Payload,
			})

		}
		}

	}
	case pb.Message_KEEPALIVE:{
		//客户端会发来keepAlive请求,返回response即可
		// client may send a keep alive request, just response A response type message,if node is not ready, send a pending status message
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("RESPONSE FROM SERVER")
	}
	case pb.Message_RESPONSE:{
		// client couldn't send a response message to server, so server should never receive a response type message
		log.Warning("Client Send a Response Message to Server, this is not allowed!")

	}
	case pb.Message_PENDING:{
		log.Warning("Got a PADDING Message")
	}
	default:
		log.Warning("Unkown Message type!")
	}
	// 返回信息加密
	if msg.MessageType !=pb.Message_HELLO && msg.MessageType != pb.Message_HELLO_RESPONSE{
		response.Payload = this.TEM.EncWithSecret(response.Payload,msg.From.Hash)
	}
	return &response, nil
}

// StartServer start the gRPC server
func (this *Node)StartServer() {
	log.Info("Starting the grpc listening server...")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(int(this.address.Port)))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
		log.Fatal("PLEASE RESTART THE SERVER NODE!")
	}
	this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(this.gRPCServer, this)
	log.Info("Listening gRPC request...")
	go this.gRPCServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (this *Node)StopServer() {
	this.gRPCServer.GracefulStop()

}