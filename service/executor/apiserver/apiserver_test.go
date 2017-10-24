package apiserver

import (
	"testing"
	"hyperchain/service/executor/manager"
	"hyperchain/common"
)

func TestAPIServerImpl_Restart(t *testing.T) {
	conf := common.NewConfig("../namespaces/global/config/namespace.toml")
	conf.MergeConfig("../global.toml")
	conf.Set(common.P2P_TLS_CA, conf.GetString(common.P2P_TLS_CA))

	conf.Set(common.P2P_TLS_CERT, conf.GetString(common.P2P_TLS_CERT))
	conf.Set(common.P2P_TLS_CERT_PRIV, conf.GetString(common.P2P_TLS_CERT_PRIV))
	// init logger
	conf.Set(common.LOG_DUMP_FILE, false)
	conf.Set(common.ENCRYPTION_CHECK_ENABLE_T, false)
	ecMgr := manager.GetExecutorMgr(conf, make(chan bool), make(chan bool))
	ecMgr.Start()
	api := GetAPIServer(ecMgr, conf)
	go api.Start()
	c := make(chan bool)
	<-c
}
