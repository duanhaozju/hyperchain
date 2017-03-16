//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"gopkg.in/fatih/set.v0"
	"reflect"
	"sync"
	"hyperchain/common"
	"hyperchain/namespace"
)

// Server represents a RPC server
type Server struct {
	//services       serviceRegistry // 这是一个map，key为hpc，value为hpc的所有方法
	muSubcriptions sync.Mutex      // protects subscriptions
	//subscriptions  subscriptionRegistry

	run      int32
	codecsMu sync.Mutex
	codecs   *set.Set
	namespaceMgr namespace.NamespaceManager
}

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	// Check http header
	CheckHttpHeaders(namespace string) common.RPCError
	// Read next request
	ReadRequestHeaders() ([]common.RPCRequest, bool, common.RPCError)
	// Parse request argument to the given types
	ParseRequestArguments([]reflect.Type, interface{}) ([]reflect.Value, common.RPCError)
	// Assemble success response, expects response id and payload
	CreateResponse(interface{}, string, interface{}) interface{}
	// Assemble error response, expects response id and error
	CreateErrorResponse(interface{}, string, common.RPCError) interface{}
	// Assemble error response with extra information about the error through info
	CreateErrorResponseWithInfo(id interface{}, name string, err common.RPCError, info interface{}) interface{}
	// Write msg to client.
	Write(interface{}) error
	// Close underlying data stream
	Close()
	// Closed when underlying connection is closed
	Closed() <-chan interface{}
}
