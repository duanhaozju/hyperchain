package apiserver

import (
	"testing"
	"hyperchain/service/executor/manager"
	"hyperchain/common"
	"hyperchain/rpc"
	"hyperchain/namespace"
)

func TestAPIServerImpl_Restart(t *testing.T) {
	conf := common.NewConfig("/Users/guowei/go/src/hyperchain/configuration/namespaces/global/config/namespace.toml")
	conf.MergeConfig("/Users/guowei/go/src/hyperchain/configuration/global.toml")
	conf.Set(common.P2P_TLS_CA, "/Users/guowei/go/src/hyperchain/configuration/tls/tlsca.ca")

	conf.Set(common.P2P_TLS_CERT, "/Users/guowei/go/src/hyperchain/configuration/tls/tls_peer1.cert")
	conf.Set(common.P2P_TLS_CERT_PRIV, "/Users/guowei/go/src/hyperchain/configuration/tls/tls_peer1.priv")
	// init logger
	conf.Set(common.LOG_DUMP_FILE, false)
	conf.Set(common.ENCRYPTION_CHECK_ENABLE_T, false)
	conf.Set(common.EXECUTOR_EMBEDDED, true)
	conf.Set(namespace.NS_CONFIG_DIR_ROOT, "/Users/guowei/go/src/hyperchain/configuration/namespaces")
	ecMgr := manager.GetExecutorMgr(conf, make(chan bool), make(chan bool))
	ecMgr.Start("global")
	api := jsonrpc.GetRPCServer(ecMgr, conf, true)
	go api.Start()
	c := make(chan bool)
	<-c
}
