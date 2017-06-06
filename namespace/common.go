//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"github.com/spf13/viper"
	"hyperchain/api"
	"hyperchain/common"
	"strings"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version
	Service interface{} // receiver instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}

var Apis map[string]*API

//constructConfigFromDir read all info needed by
func (nr *nsManagerImpl) constructConfigFromDir(path string) *common.Config {
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
	//TODO: Refactor these codes later
	nodeID := peerViper.GetInt("self.node_id")
	grpcPort := peerViper.GetInt("self.grpc_port")
	jsonrpcPort := peerViper.GetInt("self.jsonrpc_port")
	restfulPort := peerViper.GetInt("self.restful_port")
	wsPort := peerViper.GetInt("self.websocket_port")

	conf.Set(common.C_NODE_ID, nodeID)
	conf.Set(common.C_HTTP_PORT, jsonrpcPort)
	conf.Set(common.C_REST_PORT, restfulPort)
	conf.Set(common.C_WEBSOCKET_PORT, wsPort)
	conf.Set(common.C_GRPC_PORT, grpcPort)
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, nsConfigPath)

	if strings.HasSuffix(path, "/"+DEFAULT_NAMESPACE+"/config") {
		nr.conf.Set(common.C_HTTP_PORT, jsonrpcPort)
		nr.conf.Set(common.C_REST_PORT, restfulPort)
		nr.conf.Set(common.C_WEBSOCKET_PORT, wsPort)
	}

	return conf
}

func (ns *namespaceImpl) GetApis(namespace string) map[string]*api.API {

	subch := common.GetSubChan()

	return map[string]*api.API{
		"tx": {
			Srvname: "tx",
			Version: "0.4",
			Service: api.NewPublicTransactionAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"node": {
			Srvname: "node",
			Version: "0.4",
			Service: api.NewPublicNodeAPI(ns.eh),
			Public:  true,
		},
		"block": {
			Srvname: "block",
			Version: "0.4",
			Service: api.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"account": {
			Srvname: "account",
			Version: "0.4",
			Service: api.NewPublicAccountAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"contract": {
			Srvname: "contract",
			Version: "0.4",
			Service: api.NewPublicContractAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"cert": {
			Srvname: "cert",
			Version: "0.4",
			Service: api.NewCertAPI(namespace, ns.caMgr),
			Public:  true,
		},
		"sub": {
			Srvname: "sub",
			Version: "0.4",
			Service: api.NewFilterAPI(namespace, ns.eh, ns.conf, subch),
		},
	}
}
