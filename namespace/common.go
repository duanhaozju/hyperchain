//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"github.com/spf13/viper"
	"hyperchain/api"
	"hyperchain/common"
	"strings"
	"fmt"
)

var (
	ErrInvalidNs         = errors.New("namespace/nsmgr: invalid namespace")
	ErrCannotNewNs       = errors.New("namespace/nsmgr: can not new namespace")
	ErrNsClosed          = errors.New("namespace/nsmgr: namespace closed")
	ErrNodeNotFound      = errors.New("namespace/node: nod not found")
	ErrIllegalNodeConfig = errors.New("namespace/node: illegal node config")
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
	peerConfigPath := conf.GetString("global.configs.peers13")
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		logger.Errorf("err %v", err)
	}
	// global part
	//conf.Set(common.C_HTTP_PORT,  nr.conf.GetInt("global.jsonrpc_port"))
	nr.conf.Set(common.C_HTTP_PORT, nr.conf.GetInt("global.jsonrpc_port"))
	//fmt.Println(conf.GetInt(common.C_HTTP_PORT))
	//conf.Set(common.C_REST_PORT,  nr.conf.GetInt("global.restful_port"))
	nr.conf.Set(common.C_REST_PORT, nr.conf.GetInt("global.restful_port"))
	//fmt.Println(conf.GetInt(common.C_REST_PORT))
	nr.conf.Set(common.C_GRPC_PORT,  nr.conf.GetInt("global.p2p.port"))
	//fmt.Println(conf.GetInt(common.C_GRPC_PORT))
	conf.Set(common.C_JVM_PORT,   nr.conf.GetInt("global.jvm_port"))
	fmt.Println(conf.GetInt(common.C_JVM_PORT))
	conf.Set(common.C_LEDGER_PORT,nr.conf.GetInt("global.ledger_port"))
	fmt.Println(conf.GetInt(common.C_LEDGER_PORT))
	// ns part
	conf.Set(common.C_NODE_ID, peerViper.GetInt("self.id"))
	fmt.Println(conf.GetInt(common.C_NODE_ID))
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, nsConfigPath)


	if strings.HasSuffix(path, "/"+DEFAULT_NAMESPACE+"/config") {
		nr.conf.Set(common.C_HTTP_PORT, peerViper.GetInt("self.jsonrpc_port"))
		nr.conf.Set(common.C_REST_PORT, peerViper.GetInt("self.restful_port"))
	}

	return conf
}

func (ns *namespaceImpl) GetApis(namespace string) map[string]*api.API {
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
			Service: api.NewPublicNodeAPI(namespace, ns.eh),
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
	}
}
