//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jvm

import (
	"context"
	"fmt"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"hyperchain/common"
	"hyperchain/core/vm"
	pb "hyperchain/core/vm/jcee/protos"
	"sync/atomic"
	"errors"
	"time"
)

type ContractExecutor interface {
	// Start start the contract executor.
	Start() error
	// Stop stop the contract executor.
	Stop() error
	// Run invoke contract, use `Execute` internally
	Run(vm.VmContext, []byte) ([]byte, error)
	// Ping send ping package for healthy assurance
	Ping() (*pb.Response, error)
}

var (
	JVMServerErr = errors.New("jvm server execute error")
	CodeNotMatchErr = errors.New("execution code not match with ledger error")
	ContextTypeErr = errors.New("mismatch when do context type convert")
)

type contractExecutorImpl struct {
	address    string
	logger     *logging.Logger

	client     pb.ContractClient
	conn       *grpc.ClientConn

	close      int32
	maintainer *ConnMaintainer
}

type executeRs struct {
	Response *pb.Response
	Err error
}

func NewContractExecutor(conf *common.Config, namespace string) ContractExecutor {
	address := fmt.Sprintf("localhost:%d", conf.Get(common.JVM_PORT))
	Jvm := &contractExecutorImpl{
		address:    address,
		logger:     common.GetLogger(namespace, "jvm"),
	}
	return Jvm
}


func (cei *contractExecutorImpl) Start() error {
	atomic.StoreInt32(&cei.close, 0)
	cei.maintainer = NewConnMaintainer(cei, cei.logger)
	if err := cei.maintainer.conn(); err != nil {
	}
	go cei.maintainer.Serve()
	return nil
}

func (cei *contractExecutorImpl) Stop() error {
	atomic.StoreInt32(&cei.close, 1)
	cei.maintainer.Exit()
	return cei.conn.Close()
}

func (cei *contractExecutorImpl) isActive() bool {
	return atomic.LoadInt32(&cei.close) == 0
}

func (cei *contractExecutorImpl) Run(ctx vm.VmContext, in []byte) ([]byte, error) {
	context, ok := ctx.(*Context)
	if !ok {
		return nil, ContextTypeErr
	}
	request := Parse(context, in)
	response, err := cei.execute(request)

	if err != nil {
		return nil, err
	} else if response.Ok == false {
		return nil, errors.New(string(response.Result))
	} else if !hexMatch(response.CodeHash, context.GetCodeHash().Hex()) {
		return nil, CodeNotMatchErr
	} else {
		return response.Result, nil
	}
}

func (cei *contractExecutorImpl) Ping() (*pb.Response, error){
	return cei.heartbeat()
}

func (cei *contractExecutorImpl) Address() string {
	return cei.address
}

// execute send invocation message to jvm server.
func (cei *contractExecutorImpl) execute(tx *pb.Request) (*pb.Response, error) {
	if cei.client == nil {
		return nil, errors.New("no client establish")
	}
	er := make(chan *executeRs)
	go func() {
		r, err := cei.client.Execute(context.Background(), tx)
		er <- &executeRs{r, err}
	}()
	select {
	case <- time.Tick(5 * time.Second):
		return &pb.Response{Ok:false, Result:[]byte("hyperjvm execute timeout")}, errors.New("execute timeout")
	case rs := <- er:
		return rs.Response, rs.Err
	}
}

// heartbeat send health chech info to jvm server.
func (cei *contractExecutorImpl) heartbeat() (*pb.Response, error) {
	if cei.client == nil {
		return nil, errors.New("no client establish")
	}
	return cei.client.HeartBeat(context.Background(), &pb.Request{}, grpc.FailFast(true))
}

