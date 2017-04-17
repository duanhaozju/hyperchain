//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"context"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/core/vm/jcee/protos"
	"sync/atomic"
	"hyperchain/core/vm"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
)

type ContractExecutor interface {
	//Execute execute the contract cmd.
	Execute(tx *pb.Request) (*pb.Response, error)
	//Start start the contract executor.
	Start() error
	//Stop stop the contract executor.
	Stop() error
	//
	Run(vm.VmContext, []byte) ([]byte, error)

}

type contractExecutorImpl struct {
	client  pb.ContractClient
	conn    *grpc.ClientConn
	address string
	logger  *logging.Logger
	close   *int32
}

func NewContractExecutor() ContractExecutor {
	Jvm := &contractExecutorImpl{
		address: "localhost:50051",
		logger:  logging.MustGetLogger("contract"),
	}
	return Jvm
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
func (cei *contractExecutorImpl) Run(ctx vm.VmContext, in []byte) ([]byte, error) {
	request := cei.parse(ctx, in)
	response, err := cei.Execute(request)
	if err != nil {
		return nil, err
	} else {
		return response.Result, nil
	}
}

func (cei *contractExecutorImpl) parse(ctx vm.VmContext, in []byte) *pb.Request {
	var args types.InvokeArgs
	if err := proto.Unmarshal(in, &args); err != nil {
		return nil
	}
	return &pb.Request{
		Context:  &pb.RequestContext{
			Cid:         ctx.Address().Hex(),
			Namespace:   ctx.GetEnv().Namespace(),
			Txid:        ctx.GetEnv().TransactionHash().Hex(),
		},
		Method:   args.MethodName,
		Args:     args.Args,
	}
}
