//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"github.com/spf13/viper"
	"hyperchain/api"
	"hyperchain/common"
	"strings"
	"os"
)

var (
	ErrInvalidNs         = errors.New("namespace/nsmgr: invalid namespace")
	ErrCannotNewNs       = errors.New("namespace/nsmgr: can not new namespace")
	ErrRegistered        = errors.New("namespace/nsmgr: namespace has been registered")
	ErrNsClosed          = errors.New("namespace/nsmgr: namespace closed")
	ErrNodeNotFound      = errors.New("namespace/node: nod not found")
	ErrIllegalNodeConfig = errors.New("namespace/node: illegal node config")
	ErrNonExistConfig    = errors.New("namespace/nsmgr: namespace config file doesn't exist")
)

//constructConfigFromDir read all info needed by
func (nr *nsManagerImpl) constructConfigFromDir(path string) (*common.Config, error) {
	var conf *common.Config
	nsConfigPath := path + "/global.yaml"

	if strings.HasSuffix(path, "/"+DEFAULT_NAMESPACE+"/config") {
		_, err := nr.conf.MergeConfig(nsConfigPath)
		if err != nil {
			logger.Errorf("Merge config: %s error %v", nsConfigPath, err)
			panic(err)
		}
		conf = nr.conf
	} else {
		if _, err := os.Stat(nsConfigPath); os.IsNotExist(err) {
			logger.Error("namespace config file doesn't exist!")
			return nil, ErrNonExistConfig
		}
		conf = common.NewConfig(nsConfigPath)
	}
	// init peer configurations
	peerConfigPath := conf.GetString("global.configs.peers")
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		logger.Errorf("err %v", err)
	}
	conf.Set(common.C_NODE_ID, peerViper.GetInt("self.node_id"))
	conf.Set(common.C_HTTP_PORT, peerViper.GetInt("self.jsonrpc_port"))
	conf.Set(common.C_REST_PORT, peerViper.GetInt("self.restful_port"))
	conf.Set(common.C_GRPC_PORT, peerViper.GetInt("self.grpc_port"))
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, nsConfigPath)
	conf.Set(common.C_JVM_PORT, peerViper.GetInt("self.jvm_port"))
	conf.Set(common.C_LEDGER_PORT, peerViper.GetInt("self.ledger_port"))

	if strings.HasSuffix(path, "/"+DEFAULT_NAMESPACE+"/config") {
		nr.conf.Set(common.C_HTTP_PORT, peerViper.GetInt("self.jsonrpc_port"))
		nr.conf.Set(common.C_REST_PORT, peerViper.GetInt("self.restful_port"))
	}

	return conf, nil
}

func (ns *namespaceImpl) GetApis(namespace string) map[string]*api.API {
	return map[string]*api.API{
		"tx": {
			Srvname: "tx",
			Version: "1.4",
			Service: api.NewPublicTransactionAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"node": {
			Srvname: "node",
			Version: "1.4",
			Service: api.NewPublicNodeAPI(namespace, ns.eh),
			Public:  true,
		},
		"block": {
			Srvname: "block",
			Version: "1.4",
			Service: api.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"account": {
			Srvname: "account",
			Version: "1.4",
			Service: api.NewPublicAccountAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"contract": {
			Srvname: "contract",
			Version: "1.4",
			Service: api.NewPublicContractAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"cert": {
			Srvname: "cert",
			Version: "1.4",
			Service: api.NewCertAPI(namespace, ns.caMgr),
			Public:  true,
		},
	}
}
