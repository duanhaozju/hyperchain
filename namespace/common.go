//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"github.com/spf13/viper"
	"hyperchain/api"
	"hyperchain/common"
	"os"
)

var (
	ErrNoSuchNamespace   = errors.New("namespace/nsmgr: no such namespace found")
	ErrInvalidNs         = errors.New("namespace/nsmgr: invalid namespace")
	ErrCannotNewNs       = errors.New("namespace/nsmgr: can not new namespace")
	ErrRegistered        = errors.New("namespace/nsmgr: namespace has been registered")
	ErrNsClosed          = errors.New("namespace/nsmgr: namespace closed")
	ErrNodeNotFound      = errors.New("namespace/node: nod not found")
	ErrIllegalNodeConfig = errors.New("namespace/node: illegal node config")
	ErrNonExistConfig    = errors.New("namespace/nsmgr: namespace config file doesn't exist")
)

//constructConfigFromDir read all info needed by
func (nr *nsManagerImpl) constructConfigFromDir(namespace, path string) (*common.Config, error) {
	var conf *common.Config
	nsConfigPath := path + "/namespace.toml"
	if _, err := os.Stat(nsConfigPath); os.IsNotExist(err) {
		logger.Error("namespace config file doesn't exist!")
		return nil, ErrNonExistConfig
	}
	conf = common.NewConfig(nsConfigPath)
	// init peer configurations
	peerConfigPath := common.GetPath(namespace, conf.GetString(common.PEER_CONFIG_PATH))
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		logger.Errorf("err %v", err)
	}
	// global part
	conf.Set(common.P2P_PORT, nr.conf.GetInt(common.P2P_PORT))
	conf.Set(common.JVM_PORT, nr.conf.GetInt(common.JVM_PORT))
	// ns part
	conf.Set(common.C_NODE_ID, peerViper.GetInt("self.id"))
	return conf, nil
}

func (ns *namespaceImpl) GetApis(namespace string) map[string]*api.API {
	return map[string]*api.API{
		"tx": {
			Srvname: "tx",
			Version: "1.5",
			Service: api.NewPublicTransactionAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"node": {
			Srvname: "node",
			Version: "1.5",
			Service: api.NewPublicNodeAPI(namespace, ns.eh),
			Public:  true,
		},
		"block": {
			Srvname: "block",
			Version: "1.5",
			Service: api.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"account": {
			Srvname: "account",
			Version: "1.5",
			Service: api.NewPublicAccountAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"contract": {
			Srvname: "contract",
			Version: "1.5",
			Service: api.NewPublicContractAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"cert": {
			Srvname: "cert",
			Version: "1.5",
			Service: api.NewCertAPI(namespace, ns.caMgr),
			Public:  true,
		},
		"sub": {
			Srvname: "sub",
			Version: "1.5",
			Service: api.NewFilterAPI(namespace, ns.eh, ns.conf),
		},
		"archive": {
			Srvname: "archive",
			Version: "1.5",
			Service: api.NewPublicArchiveAPI(namespace, ns.eh, ns.conf),
		},
	}
}
