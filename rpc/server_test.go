package jsonrpc

import (
	"context"
	"github.com/hyperchain/hyperchain/common"
	ns "github.com/hyperchain/hyperchain/namespace/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	rpcReqs []*common.RPCRequest
	rpcReq  *common.RPCRequest
	config  *common.Config

	successJSONRes *JSONResponse
	errorJSONRes   *JSONResponse
	successRPCRes  *common.RPCResponse
	rpcError       common.RPCError
)

func initialData() {
	log = common.GetLogger(defaultNS, "jsonrpc")
	config = common.NewConfig(configPath)
	rpcReqs = make([]*common.RPCRequest, 0)
	rpcReq = &common.RPCRequest{
		Service:   "test",
		Method:    "hello",
		Namespace: defaultNS,
		Id:        1,
		IsPubSub:  false,
		Params:    nil,
		Ctx:       context.Background(),
	}
	successJSONRes = &JSONResponse{
		Version:   JSONRPCVersion,
		Namespace: defaultNS,
		Id:        1,
		Code:      0,
		Message:   "SUCCESS",
		Result:    "response data",
	}
	rpcError = &common.MethodNotFoundError{Service: "admin", Method: "hello"}
	errorJSONRes = &JSONResponse{
		Version:   JSONRPCVersion,
		Namespace: defaultNS,
		Id:        1,
		Code:      rpcError.Code(),
		Message:   rpcError.Error(),
	}

	successRPCRes = &common.RPCResponse{
		Namespace: defaultNS,
		Id:        1,
		Reply:     "response data",
		Error:     nil,
		IsPubSub:  false,
		IsUnsub:   false,
	}
}

func TestServer_ServeSingleRequest(t *testing.T) {
	ast := assert.New(t)
	initialData()

	mockCodec := &MockjsonCodecImpl{}

	// mock: ReadRawRequest returns a batch request with one request
	rpcReqs = append(rpcReqs, rpcReq)
	mockCodec.On("ReadRawRequest", OptionMethodInvocation).Return(rpcReqs, true, nil)

	// mock: CheckHttpHeaders will pass check
	mockCodec.On("CheckHttpHeaders", defaultNS, "hello").Return(nil)

	// mock: CreateResponse creates and returns a successful response
	mockCodec.On("CreateResponse", 1, defaultNS, "response data").Return(successJSONRes)

	// mock: Write will write successful response successJSONRes to client
	mockCodec.On("Write", successJSONRes).Return(nil)
	mockCodec.On("Close")

	mockNSMgr := &ns.MockNSMgr{}

	// mock: ProcessRequest will execute request and then return successful response successRPCRes
	mockNSMgr.On("ProcessRequest", defaultNS, rpcReq).Return(successRPCRes)

	s := NewServer(mockNSMgr, config)

	if err := s.ServeSingleRequest(mockCodec, OptionMethodInvocation); err != nil {
		t.Fatalf("ServeSingleRequest error: %v", err)
	}

	ast.Equal(successJSONRes, mockCodec.WriteData, "server doesn't write correct successful data to client for this request")
}

func TestNoRequestFound(t *testing.T) {
	ast := assert.New(t)
	initialData()

	rpcErr := &common.InvalidRequestError{Message: "no request found"}
	errorRPCRes := &common.RPCResponse{
		Namespace: defaultNS,
		Id:        1,
		Reply:     nil,
		Error:     rpcErr,
		IsPubSub:  false,
		IsUnsub:   false,
	}
	errorJSONRes := &JSONResponse{Version: JSONRPCVersion, Namespace: defaultNS, Id: 1, Code: rpcErr.Code(), Message: rpcErr.Error()}

	mockCodec := &MockjsonCodecImpl{}

	// mock: ReadRawRequest returns a batch request with no request
	mockCodec.On("ReadRawRequest", OptionMethodInvocation).Return(rpcReqs, true, nil)

	// mock: CheckHttpHeaders will pass check
	mockCodec.On("CheckHttpHeaders", defaultNS, "hello").Return(nil)

	// mock: CreateResponse creates and returns a error response
	mockCodec.On("CreateErrorResponse", nil, "", rpcErr).Return(errorJSONRes)

	// mock: Write will write error response errorJSONRes to client
	mockCodec.On("Write", errorJSONRes).Return(nil)
	mockCodec.On("Close")

	mockNSMgr := &ns.MockNSMgr{}

	// mock: ProcessRequest will execute request and then return error response errorRPCRes
	mockNSMgr.On("ProcessRequest", defaultNS, rpcReq).Return(errorRPCRes)

	s := NewServer(mockNSMgr, config)

	if err := s.ServeSingleRequest(mockCodec, OptionMethodInvocation); err != nil {
		t.Fatalf("ServeSingleRequest error: %v", err)
	}

	ast.Equal(errorJSONRes, mockCodec.WriteData, "server doesn't write common.InvalidRequestError to client for this request")
}

func TestCheckHttpHeaderError(t *testing.T) {
	ast := assert.New(t)
	initialData()

	rpcErr := &common.CertError{Message: "cert error"}
	errorRPCRes := &common.RPCResponse{
		Namespace: defaultNS,
		Id:        1,
		Reply:     nil,
		Error:     rpcErr,
		IsPubSub:  false,
		IsUnsub:   false,
	}
	errorJSONRes := &JSONResponse{Version: JSONRPCVersion, Namespace: defaultNS, Id: 1, Code: rpcErr.Code(), Message: rpcErr.Error()}

	mockCodec := &MockjsonCodecImpl{}

	// mock: ReadRawRequest returns a batch request with one request
	rpcReqs = append(rpcReqs, rpcReq)
	mockCodec.On("ReadRawRequest", OptionMethodInvocation).Return(rpcReqs, true, nil)

	// mock: CheckHttpHeaders will don't pass check and return error
	mockCodec.On("CheckHttpHeaders", defaultNS, "hello").Return(rpcErr)

	// mock: CreateResponse creates and returns a error response
	mockCodec.On("CreateErrorResponse", 1, defaultNS, rpcErr).Return(errorJSONRes)

	// mock: Write will write error response errorJSONRes to client
	mockCodec.On("Write", errorJSONRes).Return(nil)
	mockCodec.On("Close")

	mockNSMgr := &ns.MockNSMgr{}

	// mock: ProcessRequest will execute request and then return error response errorRPCRes
	mockNSMgr.On("ProcessRequest", defaultNS, rpcReq).Return(errorRPCRes)

	s := NewServer(mockNSMgr, config)

	if err := s.ServeSingleRequest(mockCodec, OptionMethodInvocation); err != nil {
		t.Fatalf("ServeSingleRequest error: %v", err)
	}

	ast.Equal(errorJSONRes, mockCodec.WriteData, "server doesn't write common.CertError to client for this request")
}

func TestHandleAdmin(t *testing.T) {

	ast := assert.New(t)
	initialData()

	mockNSMgr := &ns.MockNSMgr{}
	mockCodec := &MockjsonCodecImpl{}

	adminReq := &common.RPCRequest{
		Service:   "admin",
		Method:    "hello",
		Namespace: defaultNS,
		Id:        1,
		IsPubSub:  false,
		Params:    nil,
		Ctx:       context.Background(),
	}
	rpcReqs = append(rpcReqs, adminReq)
	// mock: ReadRawRequest returns a batch request with one request
	mockCodec.On("ReadRawRequest", OptionMethodInvocation).Return(rpcReqs, true, nil)

	// mock: CheckHttpHeaders will pass check
	mockCodec.On("CheckHttpHeaders", defaultNS, "hello").Return(nil)
	mockCodec.On("GetAuthInfo").Return("", "")

	// mock: CreateResponse creates and returns a successful response
	mockCodec.On("CreateResponse", 1, defaultNS, "response data").Return(successJSONRes)
	mockCodec.On("CreateErrorResponse", 1, defaultNS, rpcError).Return(errorJSONRes)

	// mock: Write will write successful response successJSONRes to client
	mockCodec.On("Write", errorJSONRes).Return(nil)
	mockCodec.On("Close")

	s := NewServer(mockNSMgr, config)

	if err := s.ServeSingleRequest(mockCodec, OptionMethodInvocation); err != nil {
		t.Fatalf("ServeSingleRequest error: %v", err)
	}

	ast.Equal(errorJSONRes, mockCodec.WriteData, "server doesn't write correct successful data to client for this request")
}
