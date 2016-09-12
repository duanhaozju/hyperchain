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
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/Server")
}
type Node struct {
	address		pb.PeerAddress
	gRPCServer	*grpc.Server
	NodeID		string
	higherEventManager *event.TypeMux
}

var globalNode Node
var DESKEY = []byte("sfe023f_sefiel#fi32lf3e!")

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int, isTest bool,hEventManager *event.TypeMux,nodeID int) *Node {
	if isTest {
		log.Info("Unit test: start local node, port", port)
		var TestNode Node
		TestNode.address.Ip = peerComm.GetLocalIp()
		TestNode.address.Port = int32(port)
		TestNode.higherEventManager = hEventManager
		TestNode.NodeID = strconv.Itoa(nodeID)
		TestNode.startServer()
		return &TestNode
	}
	if globalNode.address.Ip != "" && globalNode.address.Port != 0 {
		return &globalNode
	} else {
		globalNode.address.Ip = peerComm.GetLocalIp()
		globalNode.address.Port = int32(port)
		globalNode.NodeID = strconv.Itoa(nodeID)
		globalNode.higherEventManager = hEventManager
		globalNode.startServer()
		return &globalNode
	}

}
func GetNodeAddr() pb.PeerAddress {
	return globalNode.address
}
// GetNodeID which init by new function
func GetNodeID() string{
	return globalNode.NodeID
}
// Chat Implements the ServerSide Function
func (this *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	var response pb.Message
	response.From = &this.address
	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO :{
		response.MessageType = pb.Message_RESPONSE

		result, err := transport.TripleDesEncrypt([]byte(this.NodeID), DESKEY)
		if err!=nil{
			log.Error(err)
			log.Fatal("TripleDesEncrypt Failed!")

		}
		//REVIEW NODEID IS Encrypted, in peer handler function must decrypt it !!
		response.Payload = result
		 //REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
		 //REVIEW This no need to call hello event handler
		return &response, nil
	}
	case pb.Message_CONSUS:{
		response.MessageType = pb.Message_RESPONSE
		result, err := transport.TripleDesEncrypt([]byte("Consensus has received, response from " + strconv.Itoa(int(GetNodeAddr().Port))), DESKEY)
		if err!=nil{
			log.Fatal("TripleDesEncrypt Failed!")
		}
		response.Payload =result
		log.Debug("<<<< GOT A CONSUS MESSAGE >>>>")
		origData, err := transport.TripleDesDecrypt(msg.Payload, DESKEY)
		//log.Notice(string(origData))
		if err != nil {
			panic(err)
		}
		go this.higherEventManager.Post(event.ConsensusEvent{
			Payload:origData,
		})

		return &response, nil

	}
	case pb.Message_SYNCMSG:{
		// package the response msg
		response.MessageType = pb.Message_RESPONSE
		enResult, err := transport.TripleDesEncrypt([]byte("got a sync msg"), DESKEY)
		if err!=nil{
			log.Fatal("TripleDesEncrypt Failed!")
		}
		response.Payload = enResult


		log.Debug("<<<< GOT A SYNC MESSAGE >>>>")
		origData, err := transport.TripleDesDecrypt(msg.Payload, DESKEY)
		if err != nil {
			panic(err)
		}
		var SyncMsg recovery.Message
		unMarshalErr := proto.Unmarshal(origData,&SyncMsg)
		if unMarshalErr != nil{
			log.Error("sync UnMarshal error!")
		}
		switch SyncMsg.MessageType {
		case recovery.Message_SYNCBLOCK:{
			log.Error("post ReceiveSyncBlockEvent")
			go this.higherEventManager.Post(event.ReceiveSyncBlockEvent{
				Payload:SyncMsg.Payload,
			})

		}
		case recovery.Message_SYNCCHECKPOINT:{
			log.Error("post StateUpdateEvent")
			go this.higherEventManager.Post(event.StateUpdateEvent{
				Payload:SyncMsg.Payload,
			})

		}
		}
		go this.higherEventManager.Post(event.ConsensusEvent{
			Payload:origData,
		})


	}
	case pb.Message_KEEPALIVE:{
		//客户端会发来keepAlive请求,返回response即可
		// client may send a keep alive request, just response A response type message
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("RESPONSE FROM SERVER")

		return &response, nil

	}
	case pb.Message_RESPONSE:{
		// client couldn't send a response message to server, so server should never receive a response type message
		log.Info("Client Send a Response Message to Server, this is not allowed!")
		return &response, nil
	}
	default:
		return &response, nil
	}
	return &response, nil

}

// StartServer start the gRPC server
func (this *Node)startServer() {
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