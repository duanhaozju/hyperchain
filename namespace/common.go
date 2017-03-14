//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"hyperchain/common"
	"github.com/spf13/viper"
)

const (
	NS_CONFIG_DIR_ROOT = "global.nsConfigRootPath"
)

//ConstructConfigFromDir read all info needed by
func ConstructConfigFromDir(path string) *common.Config {
	nsConfigPath := path + "/global.yaml"
	conf := common.NewConfig(nsConfigPath)
	// init peer configurations
	peerConfigPath := conf.GetString("global.configs.peers")
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		logger.Errorf("err %v", err)
	}
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
