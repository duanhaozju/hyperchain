//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package namespace

import (
	"errors"
	"os"

	"github.com/hyperchain/hyperchain/api"
	"github.com/hyperchain/hyperchain/common"
)

var (
	// ErrInvalidNs returns when no namespace instance found.
	ErrInvalidNs = errors.New("namespace/nsmgr: invalid namespace")

	// ErrCannotNewNs returns when get namespace failed.
	ErrCannotNewNs = errors.New("namespace/nsmgr: can not newed namespace")

	// ErrRegistered returns when register a registered namespace.
	ErrRegistered = errors.New("namespace/nsmgr: namespace has been registered")

	// ErrNonExistConfig returns when specified config file doesn't exist.
	ErrNonExistConfig = errors.New("namespace/nsmgr: namespace config file doesn't exist")

	// ErrNoRootDir when namespace root dir is not set in global.toml.
	ErrNoRootDir = errors.New("namespace/nsmgr: namespace config root dir is not set in config")

	// ErrNodeNotFound returns when cannot find node by id.
	ErrNodeNotFound = errors.New("namespace/node: nod not found")

	// ErrIllegalNodeConfig returns when add an illegal node config.
	ErrIllegalNodeConfig = errors.New("namespace/node: illegal node config")

	// ErrBinNotFound returns when specified hyperjvm bin is not existed.
	ErrBinNotFound = errors.New("namespace/hyperjvm: hyperjvm bin is not found")
)

// constructConfigFromDir constructs namespace's config from specified config path.
func (nr *nsManagerImpl) constructConfigFromDir(namespace, path string) (*common.Config, error) {
	var conf *common.Config
	// init namespace configuration(namespace.toml)
	nsConfigPath := path + "/namespace.toml"
	if _, err := os.Stat(nsConfigPath); os.IsNotExist(err) {
		logger.Error("Namespace config file doesn't exist!")
		return nil, ErrNonExistConfig
	}
	conf = common.NewConfig(nsConfigPath)
	// global part
	conf.Set(common.NAMESPACE, namespace)
	conf.Set(common.P2P_PORT, nr.conf.GetInt(common.P2P_PORT))
	conf.Set(common.JVM_PORT, nr.conf.GetInt(common.JVM_PORT))
	conf.Set(common.C_JVM_START, nr.conf.GetBool(common.C_JVM_START))
	// ns part, merge peer configuration(peerconfig.toml) to get node id.
	peerConfigPath := common.GetPath(namespace, conf.GetString(common.PEER_CONFIG_PATH))
	conf.MergeConfig(peerConfigPath)
	return conf, nil
}

// GetApis returns the RPC api of specified namespace.
func (ns *namespaceImpl) GetApis(namespace string) map[string]*api.API {
	return map[string]*api.API{
		"tx": {
			Svcname: "tx",
			Version: "1.5",
			Service: api.NewPublicTransactionAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"node": {
			Svcname: "node",
			Version: "1.5",
			Service: api.NewPublicNodeAPI(namespace, ns.eh),
			Public:  true,
		},
		"block": {
			Svcname: "block",
			Version: "1.5",
			Service: api.NewPublicBlockAPI(namespace),
			Public:  true,
		},
		"account": {
			Svcname: "account",
			Version: "1.5",
			Service: api.NewPublicAccountAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"contract": {
			Svcname: "contract",
			Version: "1.5",
			Service: api.NewPublicContractAPI(namespace, ns.eh, ns.conf),
			Public:  true,
		},
		"cert": {
			Svcname: "cert",
			Version: "1.5",
			Service: api.NewCertAPI(namespace, ns.caMgr),
			Public:  true,
		},
		"sub": {
			Svcname: "sub",
			Version: "1.5",
			Service: api.NewFilterAPI(namespace, ns.eh, ns.conf),
		},
		"archive": {
			Svcname: "archive",
			Version: "1.5",
			Service: api.NewPublicArchiveAPI(namespace, ns.eh, ns.conf),
		},
	}
}
