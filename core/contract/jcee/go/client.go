//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jcee

import (
	"context"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/core/contract/jcee/protos"
	"io"
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
	close   bool
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
	return cei.conn.Close()
}

func (cei *contractExecutorImpl) isActive() bool {
	return true
}

func (cei *contractExecutorImpl) dataPipeLine() {
	firstIter := true
	for cei.isActive() {
		cmd, err := cei.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			cei.logger.Fatalf("Failed to receive a note : %v", err)
		}
		cei.logger.Noticef("Got message %v", cmd.Name)

		if firstIter {
			firstIter = false
			cei.logger.Noticef("Got message %v", cmd.Name)
		} else {
			data, err := cei.ledger.ProcessCommand(cmd)
			if err != nil {
				r := &pb.Response{Ok: true, Result: data}
				cei.stream.Send(r)
			} else {
				r := &pb.Response{Ok: false, Result: []byte("")}
				cei.stream.Send(r)
			}
		}

	}
}
