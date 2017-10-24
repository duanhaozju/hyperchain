//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"testing"
	"hyperchain/common"
	"hyperchain/namespace"
	"time"
)

var (
	configPath = "../configuration/global.toml"
	defaultNS    = common.DEFAULT_NAMESPACE
	rpc			RPCServer
)

func initial() {
	// init conf
	config := common.NewConfig(configPath)
	config.Set(common.P2P_TLS_CA, "./test/"+config.GetString(common.P2P_TLS_CA))
	config.Set(common.P2P_TLS_CERT, "./test/"+config.GetString(common.P2P_TLS_CERT))
	config.Set(common.P2P_TLS_CERT_PRIV, "./test/"+config.GetString(common.P2P_TLS_CERT_PRIV))
	// init logger
	config.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(defaultNS, config)

	namespace := namespace.GetNamespaceManager(config, make(chan bool), make(chan bool))
	rpc = GetRPCServer(namespace, config)
}

func TestRPCServerImpl_Start(t *testing.T) {
	initial()
	err := rpc.Start()
	if err != nil {
		//t.Error(err)
		t.Logf("err=%v",err)
	}
	time.Sleep(100*time.Second)
}

func TestRPCServerImpl_Stop(t *testing.T) {
	initial()
	err := rpc.Start()
	if err != nil {
		t.Error(err)
	}

	err = rpc.Stop()
	if err != nil {
		t.Error(err)
	}
}

func TestRPCServerImpl_Restart(t *testing.T) {
	initial()
	err := rpc.Restart()
	if err != nil {
		t.Error(err)
	}
}

