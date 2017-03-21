//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "hyperchain/core/contract/protos"
)

var logger *logging.Logger
const (
	address     = "localhost:50051"
	defaultName = "world"
)

func init() {
	logger = logging.MustGetLogger("test")
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		logger.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewContractClient(conn)
	//r, err := c.Execute(context.Background(), &pb.InvokeRequest{Method: "TTT"})
	r, err := c.HeartBeat(context.Background(), &pb.Request{Method:"ping"})
	if err != nil {
		logger.Fatal(err)
		return
	}
	logger.Noticef("result %v", r)
}
