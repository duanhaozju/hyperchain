package main

import (
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

var port *string
var host *string

type jvmServer struct {
}

var count = 0

func (js *jvmServer) HeartBeat(c context.Context, req *pb.Request) (*pb.Response, error) {
	count++
	fmt.Printf("Process request %d: %v\n", count, req)
	return &pb.Response{Ok: true}, nil
}

func (js *jvmServer) Register(stream pb.Contract_RegisterServer) error {
	reqs := make(chan *pb.Message, 1000)
	go func() {
		js.startProcessThread(stream, reqs)
	}()

	for { // close judge
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println(err)
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		reqs <- req
	}
	return nil
}

func (js *jvmServer) startProcessThread(stream pb.Contract_RegisterServer, reqs chan *pb.Message) {
	var i int = 0
	for {
		i++
		select {
		case req := <-reqs:
			fmt.Printf("Process request %d: %v\n", i, req)
			//Handle in another thread
			rsp := &pb.Response{Ok: true}
			payload, _ := proto.Marshal(rsp)

			stream.Send(&pb.Message{
				Type:    pb.Message_RESPONSE,
				Payload: payload,
			})
		}
	}
}

func NewJvmServer() *jvmServer {
	return &jvmServer{}
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

	pb.RegisterContractServer(grpcServer, jvmServer)

	fmt.Printf("start jvm server, listen on addr %v\n", addr)
	grpcServer.Serve(lis)
	fmt.Println("stop, server!")
}
