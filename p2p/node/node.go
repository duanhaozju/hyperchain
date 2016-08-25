// gRPC Server
// author: Chen Quan
// date: 2016-08-24
// last modified:2016-08-24

package Server

import (
	pb "hyperchain-alpha/p2p/peermessage"
	"golang.org/x/net/context"
	"net"
	"log"
	"google.golang.org/grpc"

	"strconv"
	"hyperchain-alpha/p2p/peerComm"
)

type Node struct {
	address pb.PeerAddress
	grpcServer *grpc.Server
}
var globalChatServer Node
const DefaultgRpcPort = 8001

// NewChatServer return a NewChatServer which can offer a gRPC server single instance mode
func NewNode(port int) *Node {
	if globalChatServer.address.Ip != "" && globalChatServer.address.Port !=0 {
		return &globalChatServer
	}else{
		globalChatServer.address.Ip = peerComm.GetIpLocalIpAddr()
		globalChatServer.address.Port = int32(port)
		globalChatServer.startServer()
		return &globalChatServer
	}

}
func GetNodeAddr() pb.PeerAddress{
	return globalChatServer.address
}

// Chat Implements the ServerSide Function
func (chatServer *Node) Chat(ctx context.Context, msg *pb.Message) (*pb.Message, error){
	MeAddress := pb.PeerAddress{
		Ip:peerComm.GetIpLocalIpAddr(),
		Port:8001,
	}
	var response pb.Message
	response.From = &MeAddress
	//handle the message
	switch msg.MessageType {
	case pb.Message_HELLO :{
		// TODO response a response type message
		response.MessageType = pb.Message_RESPONSE
		response.Payload =[]byte("Hi")
		// TODO save the peer information to peer pool
	}
	case pb.Message_CONSUS:{
		//TODO Post to high layer event manager
		//TODO and return a response type message
		//TODO get the payload inner message and post higher layer
	}
	}
	return &response,nil
}

// StartServer start the gRPC server
func (chatServer *Node)startServer(){
	this := chatServer
	log.Println("Starting the grpc listening server")
	lis, err := net.Listen("tcp",":"+strconv.Itoa(int(this.address.Port)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		//TODO here should be handled
	}
	this.grpcServer = grpc.NewServer()
	pb.RegisterChatServer(this.grpcServer,this)
	log.Println("listening rpc request...")
	go this.grpcServer.Serve(lis)
}

//StopServer stops the gRPC server gracefully. It stops the server to accept new
// connections and RPCs and blocks until all the pending RPCs are finished.
func (chatServer *Node)StopServer(){
	this := chatServer
	this.grpcServer.GracefulStop()
}