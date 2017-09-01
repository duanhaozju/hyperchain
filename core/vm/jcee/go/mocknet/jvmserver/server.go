package main

import (
	"fmt"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"io"
	"log"
	"net"
)

type jvmServer struct {
	req    chan *pb.Request
	stream pb.Contract_StreamExecuteServer
}
var count = 0
func (js *jvmServer) Execute(c context.Context, req *pb.Request) (*pb.Response, error) {
	count ++
	fmt.Printf("Process request %d: %v\n", count, req)
	return &pb.Response{Ok: true}, nil
}

func (js *jvmServer) HeartBeat(c context.Context, req *pb.Request) (*pb.Response, error) {

	return nil, nil
}

func (js *jvmServer) StreamExecute(stream pb.Contract_StreamExecuteServer) error {
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
		i ++
		select {
		case req := <-js.req:
			fmt.Printf("Process request %d: %v\n", i, req)
			//Handle in another thread
			js.stream.Send(&pb.Response{
				Ok: true,
			})
		}
	}
}

func NewJvmServer() *jvmServer {
	return &jvmServer{
		req: make(chan (*pb.Request), 1000),
	}
}

func main() {
	addr := "localhost:50051"
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
