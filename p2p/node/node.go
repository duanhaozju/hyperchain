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
	address            pb.PeerAddress
	gRPCServer         *grpc.Server
	higherEventManager *event.TypeMux
}

var globalNode Node

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int, isTest bool,hEventManager *event.TypeMux) *Node {
	if isTest {
		log.Println("Unit test: start local node, port", port)
		var TestNode Node
		TestNode.address.Ip = peerComm.GetIpLocalIpAddr()
		TestNode.address.Port = int32(port)
		TestNode.higherEventManager = hEventManager
		TestNode.startServer()
		return &TestNode
	}
	log.Println("start local node, port", port)
	if globalNode.address.Ip != "" && globalNode.address.Port != 0 {
		return &globalNode
	} else {
		globalNode.address.Ip = peerComm.GetIpLocalIpAddr()
		globalNode.address.Port = int32(port)
		globalNode.higherEventManager = hEventManager
		globalNode.startServer()
		return &globalNode
	}

}
func GetNodeAddr() pb.PeerAddress {
	return globalNode.address
}

// Chat Implements the ServerSide Function
func (this *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error) {
	var response pb.Message
	response.From = &this.address
	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO :{
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("Hi")
		 //REVIEW No Need to add the peer to pool because during the init, this local node will dial the peer automatically
		 //REVIEW This no need to call hello event handler
		return &response, nil
	}
	case pb.Message_CONSUS:{
		response.MessageType = pb.Message_RESPONSE
		response.Payload = []byte("Consensus broadcast has already received!")
		//post payload to high layer
		this.higherEventManager.Post(event.ConsensusEvent{
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
func (this *Node)startServer() {
	log.Println("Starting the grpc listening server")
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(int(this.address.Port)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		log.Fatal("PLS RESTART THE SERVER NODE")
	}
	this.gRPCServer = grpc.NewServer()
	pb.RegisterChatServer(this.gRPCServer, this)
	log.Println("listening rpc request...")
	go this.gRPCServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (this *Node)StopServer() {
	this.gRPCServer.GracefulStop()
}