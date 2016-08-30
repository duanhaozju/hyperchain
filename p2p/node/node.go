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
	"log"
	"google.golang.org/grpc"
	"strconv"
	"hyperchain/p2p/peerComm"
	"hyperchain/event"
)

type Node struct {
	address    pb.PeerAddress
	grpcServer *grpc.Server
	higerEventManager *event.TypeMux
}

var globalChatServer Node

const DEFAULT_GRPC_PORT = 8001

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int, isTest bool,hEventManager *event.TypeMux) *Node {
	if isTest {
		log.Println("Unit test: start local node, port", port)
		var TestNode Node
		TestNode.address.Ip = peerComm.GetIpLocalIpAddr()
		TestNode.address.Port = int32(port)
		TestNode.higerEventManager = hEventManager
		TestNode.startServer()
		return &TestNode
	}
	log.Println("start local node, port", port)
	if globalChatServer.address.Ip != "" && globalChatServer.address.Port != 0 {
		return &globalChatServer
	} else {
		globalChatServer.address.Ip = peerComm.GetIpLocalIpAddr()
		globalChatServer.address.Port = int32(port)
		globalChatServer.higerEventManager = hEventManager
		globalChatServer.startServer()
		return &globalChatServer
	}

}
func GetNodeAddr() pb.PeerAddress {
	return globalChatServer.address
}

// Chat Implements the ServerSide Function
func (this *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	MeAddress := pb.PeerAddress{
		Ip:peerComm.GetIpLocalIpAddr(),
		Port:8001,
	}
	var response pb.Message
	response.From = &MeAddress
	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO :{
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("Hi")
		 //REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
		return &response, nil
	}
	case pb.Message_CONSUS:{
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("Consensus broadcast has already received!")
		//post payload to high layer
		this.higerEventManager.Post(event.ConsensusEvent{
			Payload:msg.Payload,
		})
		return &response, nil

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
		log.Println("Client Send a Response Message to Server, this is not allowed!")
		return &response, nil
	}
	default:
		return &response, nil
	}

}

// StartServer start the gRPC server
func (chatServer *Node)startServer() {
	this := chatServer
	log.Println("Starting the grpc listening server")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(int(this.address.Port)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		log.Fatal("PLS RESTART THE SERVER NODE")
	}
	this.grpcServer = grpc.NewServer()
	pb.RegisterChatServer(this.grpcServer, this)
	log.Println("listening rpc request...")
	go this.grpcServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (chatServer *Node)StopServer() {
	this := chatServer
	this.grpcServer.GracefulStop()
}