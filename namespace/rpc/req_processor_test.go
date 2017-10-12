package rpc

import (
	"context"
	"fmt"
	"hyperchain/api"
	"hyperchain/common"
	"testing"
)

var (
	ns         = "global"
	configPath = "../../configuration/global.toml"
)

type ServiceAPI struct{}

func NewPublicServiceAPI() *ServiceAPI {
	return &ServiceAPI{}
}

func (api *ServiceAPI) Hello() (string, error) {
	fmt.Println("welcome to use hyperchain")
	return "welcome to use hyperchain", nil
}

func getApis() map[string]*api.API {
	return map[string]*api.API{
		"test": {
			Srvname: "test",
			Version: "1.5",
			Service: NewPublicServiceAPI(),
			Public:  true,
		},
	}
}

func TestJsonRpcProcessorImpl_ProcessRequest(t *testing.T) {

	// init conf
	config := common.NewConfig(configPath)
	// init logger
	config.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(ns, config)

	jrpi := NewJsonRpcProcessorImpl(ns, getApis())
	jrpi.Start()
	req := &common.RPCRequest{
		Service:   "test",
		Method:    "hello",
		Namespace: ns,
		Id:        1,
		IsPubSub:  false,
		Params:    nil,
		Ctx:       context.Background(),
	}
	resp := jrpi.ProcessRequest(req)
	if resp.Error != nil {
		t.Error(resp.Error)
	}
}
