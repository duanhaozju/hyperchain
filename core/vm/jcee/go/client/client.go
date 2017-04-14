//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"context"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"sync/atomic"
)

type ContractExecutor interface {
	//Execute execute the contract cmd.
	Execute(tx *pb.Request) (*pb.Response, error)
	//Start start the contract executor.
	Start() error
	//Stop stop the contract executor.
	Stop() error
}

type contractExecutorImpl struct {
	client  pb.ContractClient
	conn    *grpc.ClientConn
	address string
	logger  *logging.Logger
	close   *int32
}

func NewContractExecutor() ContractExecutor {
	cei := &contractExecutorImpl{address: "localhost:50051"}
	cei.logger = logging.MustGetLogger("contract")
	return cei
}

func (cei *contractExecutorImpl) Execute(tx *pb.Request) (*pb.Response, error) {
	return cei.client.Execute(context.Background(), tx)
}

func (cei *contractExecutorImpl) Start() error {
	cei.close = new(int32)
	atomic.StoreInt32(cei.close, 0)
	conn, err := grpc.Dial(cei.address, grpc.WithInsecure())
	if err != nil {
		cei.logger.Fatalf("did not connect: %v", err)
		return err
	}
	cei.client = pb.NewContractClient(conn)
	cei.conn = conn
	return nil
}

func (cei *contractExecutorImpl) Stop() error {
	atomic.StoreInt32(cei.close, 1)
	return cei.conn.Close()
}

func (cei *contractExecutorImpl) isActive() bool {
	return atomic.LoadInt32(cei.close) == 0
}