//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"github.com/spf13/viper"
	"hyperchain/api"
	"hyperchain/common"
	"strings"
)

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

	conf.Set(common.C_NODE_ID, nodeID)
	conf.Set(common.C_HTTP_PORT, jsonrpcPort)
	conf.Set(common.C_REST_PORT, restfulPort)
	conf.Set(common.C_GRPC_PORT, grpcPort)
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, nsConfigPath)

	return conf
}

func (ns *namespaceImpl) GetApis(namespace string) []hpc.API {
	return []hpc.API{
		{
			Srvname: "tx",
			Version: "0.4",
			Service: hpc.NewPublicTransactionAPI(namespace, ns.eventMux, ns.eh, ns.conf),
			Public:  true,
		},
		{
			Srvname: "node",
			Version: "0.4",
			Service: hpc.NewPublicNodeAPI(ns.eh),
			Public:  true,
		},
		{
			Srvname: "block",
			Version: "0.4",
			Service: hpc.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		{
			Srvname: "account",
			Version: "0.4",
			Service: hpc.NewPublicAccountAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		{
			Srvname: "contract",
			Version: "0.4",
			Service: hpc.NewPublicContractAPI(namespace, ns.eventMux, ns.eh, ns.conf),
			Public:  true,
		},
		{
			Srvname: "cert",
			Version: "0.4",
			Service: hpc.NewPublicCertAPI(ns.caMgr),
			Public:  true,
		},
	}

}
