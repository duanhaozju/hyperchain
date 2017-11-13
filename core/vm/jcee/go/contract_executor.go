//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jvm

import (
	"errors"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/hyperchain/hyperchain/core/vm/jcee/go/jvm"
	pb "github.com/hyperchain/hyperchain/core/vm/jcee/protos"
	"github.com/op/go-logging"
	"sync/atomic"
	"time"
)

type ContractExecutor interface {
	// Start start the contract executor.
	Start() error
	// Stop stop the contract executor.
	Stop() error
	// Run invoke contract, use `Execute` internally
	Run(vm.VmContext, []byte) ([]byte, error)
	// Clear context and make some stats
	Finalize()
	// Ping send ping package for healthy assurance
	Ping() (*pb.Response, error)
}

var (
	JVMServerErr    = errors.New("jvm server execute error")
	CodeNotMatchErr = errors.New("execution code not match with ledger error")
	ContextTypeErr  = errors.New("mismatch when do context type convert")
)

type contractExecutorImpl struct {
	logger     *logging.Logger
	client     *jvm.Client
	close      int32
	maintainer *ConnMaintainer
}

type executeRs struct {
	Response *pb.Response
	Err      error
}

func NewContractExecutor(conf *common.Config, namespace string) ContractExecutor {
	Jvm := &contractExecutorImpl{
		logger: common.GetLogger(namespace, "jvm"),
		client: jvm.NewClient(conf),
	}
	return Jvm
}

func (cei *contractExecutorImpl) Start() error {
	atomic.StoreInt32(&cei.close, 0)
	cei.maintainer = NewConnMaintainer(cei, cei.logger)
	if err := cei.maintainer.conn(); err != nil {
		cei.logger.Error(err)
	}
	go cei.maintainer.Serve()
	return nil
}

func (cei *contractExecutorImpl) Stop() error {
	atomic.StoreInt32(&cei.close, 1)
	cei.maintainer.Exit()
	return cei.client.Close()
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

func (cei *contractExecutorImpl) Finalize() {

}

func (cei *contractExecutorImpl) Ping() (*pb.Response, error) {
	return cei.heartbeat()
}

func (cei *contractExecutorImpl) Address() string {
	return cei.client.Addr()
}

// execute send invocation message to jvm server.
func (cei *contractExecutorImpl) execute(tx *pb.Request) (*pb.Response, error) {
	if cei.client == nil {
		return nil, errors.New("no client establish")
	}
	er := make(chan *executeRs)
	go func() {
		r, err := cei.client.SyncExecute(tx)
		er <- &executeRs{r, err}
	}()
	select {
	case <-time.Tick(5 * time.Second):
		return &pb.Response{Ok: false, Result: []byte("hyperjvm execute timeout")}, errors.New("execute timeout")
	case rs := <-er:
		return rs.Response, rs.Err
	}
}

// heartbeat send health chech info to jvm server.
func (cei *contractExecutorImpl) heartbeat() (*pb.Response, error) {
	if cei.client == nil {
		return nil, errors.New("no client establish")
	}
	return cei.client.HeartBeat()
}
