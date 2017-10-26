package jsonrpc

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/ipc"
	"sync"
	"hyperchain/common/interface"
)

var (
	once sync.Once
	log  *logging.Logger // package-level logger
	rpcs RPCServer
)

// RPCServer wraps all external server operations.
type RPCServer interface {
	//Start start the rpc service. It will startup all supported external service.
	Start() error

	//Stop the rpc service. It will stop all supported external service.
	Stop() error

	//Restart the rpc service. It will restart all supported external service.
	Restart() error
}

type internalRPCServer interface {
	start() error
	stop() error
	restart() error
	getPort() int
	setPort(port int) error
}

type RPCServerImpl struct {
	httpServer internalRPCServer
	wsServer   internalRPCServer
}

// GetRPCServer creates and returns a new RPCServerImpl instance implements RPCServer interface.
func GetRPCServer(nsMgrProcessor intfc.NsMgrProcessor, config *common.Config, forExe bool) RPCServer {
	//TODO: implements singleton
	log = common.GetLogger(common.DEFAULT_LOG, "jsonrpc")
	rpcs = newRPCServer(nsMgrProcessor, config, forExe)
	return rpcs
}

func newRPCServer(nsMgrProcessor intfc.NsMgrProcessor, config *common.Config, forExe bool) *RPCServerImpl {
	rsi := &RPCServerImpl{}
	rsi.httpServer = GetHttpServer(nsMgrProcessor, config, forExe)
	//rsi.wsServer = GetWSServer(nsMgrProcessor, config, forExe)

	ipc.RegisterFunc("service", rsi.Command)
	return rsi
}

// Start startups all rpc server. It will startup all supported external service
// including http/https, websocket.
func (rsi *RPCServerImpl) Start() error {

	// start http server
	if err := rsi.httpServer.start(); err != nil {
		log.Error(err)
		return err
	}

	// start websocket server
	//if err := rsi.wsServer.start(); err != nil {
	//	log.Error(err)
	//	return err
	//}
	return nil
}

// Stop terminates all rpc server. It will stop all supported external service
// including http/https, websocket.
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

// Restart restarts all rpc server. It will restart all supported external service
// including http/https, websocket.
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
