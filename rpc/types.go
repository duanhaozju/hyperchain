//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"gopkg.in/fatih/set.v0"
	admin "hyperchain/api/admin"
	"hyperchain/common"
	"hyperchain/namespace"
	"sync"
)

// Server represents a RPC server
type Server struct {
	run          int32
	codecsMu     sync.Mutex
	codecs       *set.Set
	namespaceMgr namespace.NamespaceManager
	admin        *admin.Administrator
	reqMgrMu     sync.Mutex
	requestMgr   map[string]*requestManager
}

// ServerCodec implements reading, parsing and writing RPC messages for the server side of
// a RPC session. Implementations must be go-routine safe since the codec can be called in
// multiple go-routines concurrently.
type ServerCodec interface {
	// Check http header
	CheckHttpHeaders(namespace string, method string) common.RPCError
	// Read next request
	ReadRequestHeaders(options CodecOption) ([]*common.RPCRequest, bool, common.RPCError)
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