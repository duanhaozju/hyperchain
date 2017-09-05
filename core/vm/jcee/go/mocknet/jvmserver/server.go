package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"io"
	"log"
	"net"
)

var port *string
var host *string

type jvmServer struct {
	req    chan *pb.Message
	stream pb.Contract_RegisterServer
}

var count = 0

func (js *jvmServer) HeartBeat(c context.Context, req *pb.Request) (*pb.Response, error) {
	count++
	fmt.Printf("Process request %d: %v\n", count, req)
	return &pb.Response{Ok: true}, nil
}

func (js *jvmServer) Register(stream pb.Contract_RegisterServer) error {
	js.stream = stream
	for { // close judge
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		js.req <- req
	}
	return nil
}

func (js *jvmServer) startProcessThread() {
	var i int = 0
	for {
		i++
		select {
		case req := <-js.req:
			fmt.Printf("Process request %d: %v\n", i, req)
			//Handle in another thread
			rsp := &pb.Response{Ok: true}
			payload, _ := proto.Marshal(rsp)

			js.stream.Send(&pb.Message{
				Type:    pb.Message_RESPONSE,
				Payload: payload,
			})
		}
	}
}

func NewJvmServer() *jvmServer {
	return &jvmServer{
		req: make(chan (*pb.Message), 1000),
	}
}

func main() {

	port = flag.String("p", "50051", "server port")
	host = flag.String("h", "0.0.0.0", "server host")

	flag.Parse()
	addr := fmt.Sprintf("%s:%s", *host, *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	jvmServer := NewJvmServer()
	go jvmServer.startProcessThread()

	pb.RegisterContractServer(grpcServer, jvmServer)

	fmt.Printf("start jvm server, listen on addr %v\n", addr)
	grpcServer.Serve(lis)
	fmt.Println("stop, server!")
}
