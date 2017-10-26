//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"github.com/hyperchain/hyperchain/common"
	"testing"
)

var (
	configPath = "../configuration/global.toml"
	defaultNS  = common.DEFAULT_NAMESPACE
	rpc        RPCServer
)

func initial() {
	// init conf
	config := common.NewConfig(configPath)
	config.Set(common.P2P_TLS_CA, "./test/tls/tlsca.ca")
	config.Set(common.P2P_TLS_CERT, "./test/tls/tls_peer1.cert")
	config.Set(common.P2P_TLS_CERT_PRIV, "./test/tls/tls_peer1.priv")
	// init logger
	config.Set(common.LOG_DUMP_FILE, false)
	common.InitHyperLogger(defaultNS, config)

	rpc = GetRPCServer(nil, config)
}

func TestRPCServerImpl_Start(t *testing.T) {
	initial()
	err := rpc.Start()
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	err = rpc.Stop()
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestRPCServerImpl_Restart(t *testing.T) {
	initial()
	err := rpc.Restart()
	if err != nil {
		t.Error(err)
	}
}
