package jsonrpc

import (
	"reflect"
	"sync"
	"gopkg.in/fatih/set.v0"
)

// callback is a method callback which was registered in the server
type callback struct {
	rcvr        reflect.Value  // receiver of method
	method      reflect.Method // callback
	argTypes    []reflect.Type // input argument types
	hasCtx      bool           // method's first argument is a context (not included in argTypes)
	errPos      int            // err return idx, of -1 when method cannot return error
	//isSubscribe bool           // indication if the callback is a subscription
}


// service represents a registered object
type service struct {
	name          string        // name for service
	rcvr          reflect.Value // receiver of methods for the service
	typ           reflect.Type  // receiver type
	callbacks     callbacks     // registered handlers
	subscriptions subscriptions // available subscriptions/notifications
}

// serverRequest is an incoming request
type serverRequest struct {
	id            interface{}
	svcname       string
	rcvr          reflect.Value
	callb         *callback
	args          []reflect.Value
	//isUnsubscribe bool
	err           RPCError
}

type serviceRegistry map[string]*service       // collection of services
type callbacks map[string]*callback            // collection of RPC callbacks
type subscriptions map[string]*callback        // collection of subscription callbacks
type subscriptionRegistry map[string]*callback // collection of subscription callbacks

// Server represents a RPC server
type Server struct {
	services       serviceRegistry // 这是一个map，key为hpc，value为hpc的所有方法
	muSubcriptions sync.Mutex // protects subscriptions
	subscriptions  subscriptionRegistry

	run      int32
	codecsMu sync.Mutex
	codecs   *set.Set
}

// rpcRequest represents a raw incoming RPC request
type rpcRequest struct {
	service  string
	method   string
	id       interface{}
	isPubSub bool
	params   interface{}
}

// RPCError implements RPC error, is add support for error codec over regular go errors
type RPCError interface {
	// RPC error code
	Code() int
	// Error message
	Error() string
}

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	// Read next request
	ReadRequestHeaders() ([]rpcRequest, bool, RPCError)
	// Parse request argument to the given types
	ParseRequestArguments([]reflect.Type, interface{}) ([]reflect.Value, RPCError)
	// Assemble success response, expects response id and payload
	CreateResponse(interface{}, interface{}) interface{}
	// Assemble error response, expects response id and error
	CreateErrorResponse(interface{}, RPCError) interface{}
	// Assemble error response with extra information about the error through info
	CreateErrorResponseWithInfo(id interface{}, err RPCError, info interface{}) interface{}
	// Create notification response
	//CreateNotification(string, interface{}) interface{}
	// Write msg to client.
	Write(interface{}) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
