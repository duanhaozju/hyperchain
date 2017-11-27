package api

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/namespace"
	"github.com/hyperchain/hyperchain/rpc"
	"github.com/hyperchain/hyperchain/service/executor/controller"
	"testing"
)

func TestAPIServerImpl_Restart(t *testing.T) {
	conf := common.NewConfig("/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/namespaces/global/config/namespace.toml")
	conf.MergeConfig("/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/global.toml")
	conf.Set(common.P2P_TLS_CA, "/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/tls/tlsca.ca")

	conf.Set(common.P2P_TLS_CERT, "/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/tls/tls_peer1.cert")
	conf.Set(common.P2P_TLS_CERT_PRIV, "/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/tls/tls_peer1.priv")
	// init logger
	conf.Set(common.LOG_DUMP_FILE, false)
	conf.Set(common.ENCRYPTION_CHECK_ENABLE_T, false)
	conf.Set(common.EXECUTOR_EMBEDDED, true)
	conf.Set(namespace.NS_CONFIG_DIR_ROOT, "/Users/guowei/go/src/github.com/hyperchain/hyperchain/configuration/namespaces")
	conf.Set(common.JSON_RPC_PORT_EXECUTOR, "9091")
	ecMgr := controller.GetExecutorCtl(conf, make(chan bool), make(chan bool))
	ecMgr.Start()
	api := jsonrpc.GetRPCServer(ecMgr, conf, true)
	go api.Start()
	c := make(chan bool)
	<-c
}
