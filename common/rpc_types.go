//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	"reflect"
	"sync"
	"gopkg.in/fatih/set.v0"
)

// rpcRequest represents a raw incoming RPC request
type RPCRequest struct {
	Service   string
	Method    string
	Namespace string
	Id        interface{}
	IsPubSub  bool
	Params    interface{}
}

// rpcResponse represents a raw incoming RPC request
type RPCResponse struct {
	Reply []reflect.Value
	Error RPCError
}

//// callback is a method callback which was registered in the server
//type callback struct {
//	rcvr     reflect.Value  // receiver of method
//	method   reflect.Method // callback
//	argTypes []reflect.Type // input argument types
//	hasCtx   bool           // method's first argument is a context (not included in argTypes)
//	errPos   int            // err return idx, of -1 when method cannot return error
//				// isSubscribe bool           // indication if the callback is a subscription
//}
//
//// service represents a registered object
//type service struct {
//	name          string        // name for service
//	rcvr          reflect.Value // receiver of methods for the service
//	typ           reflect.Type  // receiver type
//	callbacks     callbacks     // registered handlers
//	//subscriptions subscriptions // available subscriptions/notifications
//}
//
//// serverRequest is an incoming request
//type serverRequest struct {
//	id      interface{}
//	svcname string
//	rcvr    reflect.Value
//	callb   *callback
//	args    []reflect.Value
//	//isUnsubscribe bool
//	err RPCError
//}
//
//type serviceRegistry map[string]*service       // collection of services
//type callbacks map[string]*callback            // collection of RPC callbacks
////type subscriptions map[string]*callback        // collection of subscription callbacks
////type subscriptionRegistry map[string]*callback // collection of subscription callbacks
//
//
// RPCError implements RPC error, is add support for error codec over regular go errors
type RPCError interface {
	// RPC error code
	Code() int
	// Error message
	Error() string
}

type Process interface {
	Process()
}