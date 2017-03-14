//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import "golang.org/x/net/context"


// rpcRequest represents a raw incoming RPC request
type RPCRequest struct {
	Service   string
	Method    string
	Namespace string
	Id        interface{}
	IsPubSub  bool
	Params    interface{}
	Ctx context.Context
}

// rpcResponse represents a raw incoming RPC request
type RPCResponse struct {
	Id    interface{}
	//Reply []reflect.Value
	Reply interface{}
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
// serverRequest is an incoming request
//type serverRequest struct {
//	id      interface{}
//	svcname string
//	rcvr    reflect.Value
//	callb   *callback
//	args    []reflect.Value
//	//isUnsubscribe bool
//	err RPCError
//}

//type serviceRegistry map[string]*service       // collection of services
//type callbacks map[string]*callback            // collection of RPC callbacks
////type subscriptions map[string]*callback        // collection of subscription callbacks
////type subscriptionRegistry map[string]*callback // collection of subscription callbacks
//
//