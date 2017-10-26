package jsonrpc

import (
	"github.com/stretchr/testify/mock"
	"hyperchain/common"
)

type MockjsonCodecImpl struct {
	WriteData interface{}
	mock.Mock
}

func (mc *MockjsonCodecImpl) CheckHttpHeaders(namespace string, method string) common.RPCError {
	args := mc.Called(namespace, method)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(common.RPCError)
}

func (mc *MockjsonCodecImpl) ReadRawRequest(options CodecOption) ([]*common.RPCRequest, bool, common.RPCError) {
	args := mc.Called(options)
	if args.Get(2) == nil {
		return args.Get(0).([]*common.RPCRequest), args.Bool(1), nil
	}
	return args.Get(0).([]*common.RPCRequest), args.Bool(1), args.Get(2).(common.RPCError)
}

func (mc *MockjsonCodecImpl) CreateResponse(id interface{}, namespace string, reply interface{}) interface{} {
	args := mc.Called(id, namespace, reply)
	return args.Get(0)
}

func (mc *MockjsonCodecImpl) CreateErrorResponse(id interface{}, namespace string, err common.RPCError) interface{} {
	args := mc.Called(id, namespace, err)
	return args.Get(0)
}

func (mc *MockjsonCodecImpl) CreateErrorResponseWithInfo(id interface{}, namespace string, err common.RPCError, info interface{}) interface{} {
	args := mc.Called(id, namespace, err, info)
	return args.Get(0)
}

func (mc *MockjsonCodecImpl) CreateNotification(subid common.ID, service, method, namespace string, event interface{}) interface{} {
	args := mc.Called(subid, service, method, namespace, event)
	return args.Get(0)
}

func (mc *MockjsonCodecImpl) GetAuthInfo() (string, string) {
	args := mc.Called()
	return args.String(0), args.String(1)
}

func (mc *MockjsonCodecImpl) Write(data interface{}) error {
	args := mc.Called(data)
	mc.WriteData = data
	return args.Error(0)
}

func (mc *MockjsonCodecImpl) WriteNotify(data interface{}) error {
	args := mc.Called(data)
	return args.Error(0)
}

func (mc *MockjsonCodecImpl) Close() {}

func (mc *MockjsonCodecImpl) Closed() <-chan interface{} {
	args := mc.Called()
	return args.Get(0).(chan interface{})
}
