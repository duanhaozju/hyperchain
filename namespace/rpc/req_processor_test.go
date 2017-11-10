package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/common"

	"github.com/stretchr/testify/assert"
)

var (
	namespace = "test"
)

// Normal API used to test normal flow.
type NormalAPI struct{}

func NewNormalAPI() *NormalAPI {
	return &NormalAPI{}
}

func getCommonApis() map[string]*api.API {
	return map[string]*api.API{
		"test": {
			Svcname: "test",
			Version: "1.5",
			Service: NewNormalAPI(),
			Public:  true,
		},
	}
}

func (api *NormalAPI) Hello() (string, error) {
	fmt.Println("Hello, hyperchain.")
	return "Hello, hyperchain.", nil
}

func (api *NormalAPI) Name(name string) (string, error) {
	cmpName := fmt.Sprintf("Company name: %s.", name)
	fmt.Print(cmpName)
	return cmpName, nil
}

func (api *NormalAPI) Company(name string, age int) string {
	cmpInfo := fmt.Sprintf("Company name: %s, age: %d.", name, age)
	fmt.Print(cmpInfo)
	return cmpInfo
}

func (api *NormalAPI) Block(ctx context.Context, isVerbose bool) (common.ID, error) {
	fmt.Printf("Subscribe block, isVerbose: %v\n", isVerbose)
	return "Subscribe block.", nil
}

func getEmptyNameApis() map[string]*api.API {
	return map[string]*api.API{
		"test": {
			Svcname: "",
			Version: "1.5",
			Service: NewNormalAPI(),
			Public:  true,
		},
	}
}

// unExport API whose receiver is unexported.
type unExportAPI struct{}

func NewUnExportAPI() *unExportAPI {
	return &unExportAPI{}
}

func getUnexportApis() map[string]*api.API {
	return map[string]*api.API{
		"test": {
			Svcname: "test",
			Version: "1.5",
			Service: NewUnExportAPI(),
			Public:  true,
		},
	}
}

// ExportAPI whose receiver is exported.
type ExportAPI struct{}

func NewExportAPI() *ExportAPI {
	return &ExportAPI{}
}

func getExportApis() map[string]*api.API {
	return map[string]*api.API{
		"test": {
			Svcname: "test",
			Version: "1.5",
			Service: NewExportAPI(),
			Public:  true,
		},
	}
}

func TestJsonRpcProcessorImpl_ProcessRequest(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")
	ast := assert.New(t)

	jrpi := NewJsonRpcProcessorImpl(namespace, nil)
	err := jrpi.Start()
	ast.Error(ErrNoApis, err, "Failed start processor as we provide mockParams nil api.")

	jrpi = NewJsonRpcProcessorImpl(namespace, getEmptyNameApis())
	err = jrpi.Start()
	ast.Error(ErrNoServiceName, err, "Failed start processor as we provide mockParams api with empty name.")

	jrpi = NewJsonRpcProcessorImpl(namespace, getUnexportApis())
	err = jrpi.Start()
	ast.Error(ErrNotExported, err, "Failed start processor as we provide mockParams api with unexport receiver.")

	jrpi = NewJsonRpcProcessorImpl(namespace, getExportApis())
	err = jrpi.Start()
	ast.Error(ErrNoSuitable, err, "Failed start processor as we provide mockParams api with no callbacks.")
}

func TestCommonCase(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")

	mockParams := []byte(`["hyperchain", 1]`)
	msg := json.RawMessage(mockParams)

	commonReq := &common.RPCRequest{
		Service:   "test",
		Method:    "company",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  false,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi := NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	jrpi.registerAllAPIService()
	resp := jrpi.ProcessRequest(commonReq)
	if resp.Error != nil {
		t.Error(resp.Error)
	}
}

func TestNonExistentService(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")
	ast := assert.New(t)
	methodNotFoundErr := &common.MethodNotFoundError{}

	req := &common.RPCRequest{
		Service:   "mock_test",
		Method:    "hello",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  false,
		Params:    nil,
		Ctx:       context.Background(),
	}
	jrpi := NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp := jrpi.ProcessRequest(req)
	ast.Equal(methodNotFoundErr.Code(), resp.Error.Code(),
		"We should get a methodNotFoundErr.")
}

func TestInvalidParams(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")
	ast := assert.New(t)
	invalidParamsErr := &common.InvalidParamsError{}

	req := &common.RPCRequest{
		Service:   "test",
		Method:    "name",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  false,
		Params:    "mock_params",
		Ctx:       context.Background(),
	}
	jrpi := NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp := jrpi.ProcessRequest(req)
	ast.Equal(invalidParamsErr.Code(), resp.Error.Code(),
		"We should get a invalidParamsErr")

	mockParams := []byte(`["hyperchain", 1]`)
	msg := json.RawMessage(mockParams)

	req = &common.RPCRequest{
		Service:   "test",
		Method:    "name",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  false,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi = NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp = jrpi.ProcessRequest(req)
	ast.Equal(invalidParamsErr.Code(), resp.Error.Code(),
		"We should get a invalidParamsEr.r")

	mockParams = []byte(`["hyperchain"]`)
	msg = json.RawMessage(mockParams)

	req = &common.RPCRequest{
		Service:   "test",
		Method:    "company",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  false,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi = NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp = jrpi.ProcessRequest(req)
	ast.Equal(invalidParamsErr.Code(), resp.Error.Code(),
		"We should get a invalidParamsEr.r")
}

func TestSubscription(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")
	ast := assert.New(t)
	methodNotFoundErr := &common.MethodNotFoundError{}

	mockParams := []byte(`["block", true]`)
	msg := json.RawMessage(mockParams)

	req := &common.RPCRequest{
		Service:   "test",
		Method:    "block",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  true,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi := NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp := jrpi.ProcessRequest(req)
	ast.Nil(resp.Error)

	invalidReq := &common.RPCRequest{
		Service:   "test",
		Method:    "subscribe",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  true,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi = NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp = jrpi.ProcessRequest(invalidReq)
	ast.Equal(resp.Error.Code(), methodNotFoundErr.Code(),
		"We should get a methodNotFoundErr.")
}

func TestUnsubscription(t *testing.T) {
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "rpc", "DEBUG")
	ast := assert.New(t)

	mockParams := []byte(`["block"]`)
	msg := json.RawMessage(mockParams)

	req := &common.RPCRequest{
		Service:   "test",
		Method:    "block_unsubscribe",
		Namespace: namespace,
		Id:        1,
		IsPubSub:  true,
		Params:    msg,
		Ctx:       context.Background(),
	}
	jrpi := NewJsonRpcProcessorImpl(namespace, getCommonApis())
	jrpi.Start()
	resp := jrpi.ProcessRequest(req)
	ast.Nil(resp.Error)
}
