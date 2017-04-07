//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"context"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/core/contract/jcee/protos"
	"io"
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
	stream  pb.Contract_DataPipelineClient
	address string
	logger  *logging.Logger
	ledger  *LedgerProxy
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
	atomic.StoreInt32(cei.close, 0)
	conn, err := grpc.Dial(cei.address, grpc.WithInsecure())
	if err != nil {
		cei.logger.Fatalf("did not connect: %v", err)
		return err
	}
	cei.client = pb.NewContractClient(conn)

	stream, err := cei.client.DataPipeline(context.Background())
	stream.Send(&pb.Response{Ok: true})
	if err != nil {
		cei.logger.Error(err)
		return err
	}
	cei.stream = stream
	go cei.dataPipeLine()
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

func (cei *contractExecutorImpl) dataPipeLine() {
	firstIter := true
	for cei.isActive() {
		cmd, err := cei.stream.Recv()
		if err == io.EOF {
			//TODO: add reconnect or restart the jcee logic here
			return
		}
		if err != nil {
			cei.logger.Fatalf("Failed to receive a note : %v", err)
		}
		cei.logger.Noticef("Got message %v", cmd)

		if firstIter {
			firstIter = false
		} else {
			data, err := cei.ledger.ProcessCommand(cmd)
			r := &pb.Response{Ok:false, Result:[]byte("")}
			if err == nil {
				r.Ok = true
				r.Result = data
			}
			err = cei.stream.Send(r)
			if err != nil {
				cei.logger.Errorf("send command resutlt to jcee error, %v", err)
				//TODO: add reconnect or restart the jcee logic here
			}
		}
	}
}
