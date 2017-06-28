package jsonrpc

import (
	"hyperchain/namespace"
	"sync"
	"github.com/op/go-logging"
	"hyperchain/common"
)

var (
	once         sync.Once
        log 	     *logging.Logger // package-level logger
	rpcs 	     RPCServer
)

type RPCServer interface {
	//Start start the rpc service.
	Start() error
	//Stop the rpc service.
	Stop() error
	//Restart the rpc service.
	Restart() error
}

type RPCServerImpl struct {
	httpServer  RPCServer
	wsServer    RPCServer
}

func GetRPCServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) RPCServer {
	log = common.GetLogger(common.DEFAULT_LOG, "jsonrpc")
	once.Do(func() {
		rpcs = newRPCServer(nr, stopHp, restartHp)
	})
	return rpcs
}

func newRPCServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) *RPCServerImpl{
	rsi := &RPCServerImpl {}
	rsi.httpServer = GetHttpServer(nr, stopHp, restartHp)
	rsi.wsServer = GetWSServer(nr, stopHp, restartHp)
	return rsi
}

// Start startup all rpc server.
func (rsi *RPCServerImpl) Start() error{

	// start http server
	if err := rsi.httpServer.Start(); err != nil {
		log.Error(err)
		return err
	}

	// start websocket server
	if err := rsi.wsServer.Start(); err != nil {
		log.Error(err)
		rsi.httpServer.Stop()
		return err
	}
	return nil
}

// Stop terminates all rpc server.
func (rsi *RPCServerImpl) Stop() error {

	// stop http server
	if err := rsi.httpServer.Stop(); err != nil {
		return err
	}

	// stop websocket server
	if err := rsi.wsServer.Stop(); err != nil {
		return err
	}
	return nil
}

// Restart all rpc server
func (rsi *RPCServerImpl) Restart() error {

	if err := rsi.Stop(); err != nil {
		return err
	}
	if err := rsi.Start(); err != nil {
		return err
	}
	return nil
}