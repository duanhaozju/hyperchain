package api

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/ipc"
	hrpc "hyperchain/rpc"
	hm "hyperchain/service/hypexec/controller"
	"sync"
)

var (
	once      sync.Once
	log       *logging.Logger // package-level logger
	apiserver APIServer
)

type APIServer interface {
	Start() error
	Stop() error
}

type internalRPCServer interface {
	start() error
	stop() error
	restart() error
	getPort() int
	setPort(port int) error
}

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	// Check http header
	CheckHttpHeaders(namespace string, method string) common.RPCError
	// Read next request
	ReadRawRequest(options hrpc.CodecOption) ([]*common.RPCRequest, bool, common.RPCError)
	// Assemble success response, expects response id and payload
	CreateResponse(id interface{}, namespace string, reply interface{}) interface{}
	// Assemble error response, expects response id and error
	CreateErrorResponse(id interface{}, namespace string, err common.RPCError) interface{}
	// Assemble error response with extra information about the error through info
	CreateErrorResponseWithInfo(id interface{}, namespace string, err common.RPCError, info interface{}) interface{}
	// Create notification response
	CreateNotification(subid common.ID, service, method, namespace string, event interface{}) interface{}
	// GatAuthInfo read authentication info (token and method) from http header
	GetAuthInfo() (string, string)
	// Write msg to client.
	Write(interface{}) error
	// Write notify msg to client.
	WriteNotify(interface{}) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}

// receiver implements handling RPC request in channel.
type receiver interface {
	handleChannelReq(codec ServerCodec, rq *common.RPCRequest) interface{}
}

type APIServerImpl struct {
	httpServer internalRPCServer
	wsServer   internalRPCServer
}

func GetAPIServer(er hm.ExecutorController, config *common.Config) APIServer {
	log = common.GetLogger(common.DEFAULT_LOG, "executor_api")
	once.Do(func() {
		apiserver = NewAPIServer(er, config)
	})
	return apiserver
}

func NewAPIServer(er hm.ExecutorController, config *common.Config) *APIServerImpl {
	apiserver := &APIServerImpl{}
	apiserver.httpServer = GetHttpServer(er, config)
	apiserver.wsServer = GetWSServer(er, config)

	ipc.RegisterFunc("serviceEx", apiserver.Command)
	return apiserver
}

func (asi *APIServerImpl) Start() error {
	// start http server
	if err := asi.httpServer.start(); err != nil {
		log.Error(err)
		return err
	}

	// start websocket server
	//TODO start WebSocket
	if err := asi.wsServer.start(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (asi *APIServerImpl) Stop() error {
	// stop http server
	if err := asi.httpServer.stop(); err != nil {
		return err
	}

	// stop websocket server
	if err := asi.wsServer.stop(); err != nil {
		return err
	}
	return nil
}

func (asi *APIServerImpl) Restart() error {
	// restart http server
	if err := asi.httpServer.restart(); err != nil {
		return err
	}

	// restart websocket server
	if err := asi.wsServer.restart(); err != nil {
		return err
	}
	return nil
}
