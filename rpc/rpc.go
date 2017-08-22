package jsonrpc

import (
	"hyperchain/namespace"
	"sync"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/ipc"
)

var (
	once         sync.Once
        log 	     *logging.Logger // package-level logger
	rpcs 	     RPCServer
)

type RPCServer interface {
	//Start start the rpc service. It will use default port of config.
	Start() error
	//Stop the rpc service.
	Stop() error
	//Restart the rpc service.
	Restart() error
	//Start http server. It will use specified port.
	StartHttpServer(port int) error
	//Stop http server
	StopHttpServer() error
	//Restart http server
	RestartHttpServer() error
	//Start websocket server. It will use specified port.
	StartWSServer(port int) error
	//Stop websocket server
	StopWSServer() error
	//Restart websocket server
	RestartWSServer() error

}

type internalRPCServer interface {
	// Start start the rpc service.
	start() error
	// Stop the rpc service.
	stop() error
	// Restart the rpc service.
	restart() error
	// Get service listening port.
	getPort() int
	// Set service port.
	setPort(port int) error
}

type RPCServerImpl struct {
	httpServer  internalRPCServer
	wsServer    internalRPCServer
}

func GetRPCServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) RPCServer {
	log = common.GetLogger(common.DEFAULT_LOG, "jsonrpc")
	once.Do(func() {
		rpcs = newRPCServer(nr, stopHp, restartHp)
	})
	return rpcs
}

func newRPCServer(nr namespace.NamespaceManager, stopHp chan bool, restartHp chan bool) *RPCServerImpl{
	rsi := &RPCServerImpl{}
	rsi.httpServer = GetHttpServer(nr, stopHp, restartHp)
	rsi.wsServer   = GetWSServer(nr, stopHp, restartHp)

	ipc.RegisterFunc("service", rsi.Command)
	return rsi
}

// Start startup all rpc server.
func (rsi *RPCServerImpl) Start() error{

	// start http server
	if err := rsi.httpServer.start(); err != nil {
		log.Error(err)
		return err
	}

	// start websocket server
	if err := rsi.wsServer.start(); err != nil {
		log.Error(err)
		rsi.httpServer.stop()
		return err
	}
	return nil
}

// Stop terminates all rpc server.
func (rsi *RPCServerImpl) Stop() error {

	// stop http server
	if err := rsi.httpServer.stop(); err != nil {
		return err
	}

	// stop websocket server
	if err := rsi.wsServer.stop(); err != nil {
		return err
	}
	return nil
}

// Restart all rpc server
func (rsi *RPCServerImpl) Restart() error {

	// restart http server
	if err := rsi.httpServer.restart(); err != nil {
		return err
	}

	// restart websocket server
	if err := rsi.wsServer.restart(); err != nil {
		return err
	}
	return nil
}

// StartHttpServer starts http service.
func (rsi *RPCServerImpl) StartHttpServer(port int) error {
	rsi.httpServer.setPort(port)
	return rsi.httpServer.start()
}

// StopHttpServer stops http service.
func (rsi *RPCServerImpl) StopHttpServer() error {
	return rsi.httpServer.stop()
}

// RestartHttpServer restarts http service.
func (rsi *RPCServerImpl) RestartHttpServer() error {
	return rsi.httpServer.restart()
}

// StartWSServer starts websocket service.
func (rsi *RPCServerImpl) StartWSServer(port int) error {
	rsi.wsServer.setPort(port)
	return rsi.wsServer.start()
}

// StopWSServer stops websocket service.
func (rsi *RPCServerImpl) StopWSServer() error {
	return rsi.wsServer.stop()
}

// RestartWSServer restarts websocket service.
func (rsi *RPCServerImpl) RestartWSServer() error {
	return rsi.wsServer.restart()
}