//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"github.com/op/go-logging"
	"hyperchain/core/contract/jcee/go"
	pb "hyperchain/core/contract/jcee/protos"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"time"
	"strconv"
)

var logger *logging.Logger
const (
	address     = "localhost:50051"
	defaultName = "world"
)

func init() {
	logger = logging.MustGetLogger("test")
}

func startServer()  {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50052))
	if err != nil {
		//log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterLedgerServer(grpcServer, jcee.NewLedgerProxy())
	grpcServer.Serve(lis)
}

func main() {
	go startServer()
	exe := jcee.NewContractExecutor()
	exe.Start()
	for  i := 0; i < 10; i ++{
		time.Sleep(3 * time.Second)
		request := &pb.Request{
			Txid:"tx000000" + strconv.Itoa(i),
			Cid:"msc001",
			Method:"invoke",
			Args:[][]byte{[]byte("test"), []byte("wangxiaoyi")},
		}
		response, err := exe.Execute(request)

		if err!= nil {
			logger.Error(err)
		}
		logger.Info(response)
	}

	x := make(chan bool)
	<- x
}
