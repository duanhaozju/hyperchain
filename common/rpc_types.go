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
	Namespace string
	Id        interface{}
	//Reply []reflect.Value
	Reply     interface{}
	Error     RPCError
}