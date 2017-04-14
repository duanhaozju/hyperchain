//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"fmt"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"hyperchain/core/vm/jcee/go/client"
	lg "hyperchain/core/vm/jcee/go/ledger"
	pb "hyperchain/core/vm/jcee/protos"
	exec "hyperchain/core/executor"
	"net"
	"strconv"
	"time"
)

var logger *logging.Logger

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func init() {
	logger = logging.MustGetLogger("test")
}

func startServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50052))
	if err != nil {
		//log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	ledger := lg.NewLedgerProxy()
	executor := exec.NewExecutor("global", nil, nil)
	executor.Start()
	ledger.Register("global", executor.FetchStateDb())
	pb.RegisterLedgerServer(grpcServer, lg.NewLedgerProxy())
	grpcServer.Serve(lis)
}

func main() {
	go startServer()
	exe := jcee.NewContractExecutor()
	exe.Start()
	testNum := 10 * 10000
	t1 := time.Now()
	for i := 0; i < testNum; i++ {
		//time.Sleep(3 * time.Second)
		request := &pb.Request{
			Txid:   "tx000000" + strconv.Itoa(i),
			Cid:    "msc001",
			Method: "invoke",
			Args:   [][]byte{[]byte("test"), []byte("wangxiaoyi")},
		}
		response, err := exe.Execute(request)
		//_, err := exe.Execute(request)

		if err != nil {
			logger.Error(err)
		}
		logger.Info(response)
	}
	t2 := time.Now()

	//logger.Critical((testNum * 1.0) / t2.Sub(t1).Seconds())
	a := (float64(1.0 * testNum)) / t2.Sub(t1).Seconds()
	logger.Critical(a)

	x := make(chan bool)
	<-x
}
