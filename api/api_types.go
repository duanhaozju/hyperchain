//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/admittance"
	"hyperchain/common"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

var Apis []API

// todo 该方法挪到 rpcmanger
func GetAPIs(eventMux *event.TypeMux, pm *manager.EventHub, cm *admittance.CAManager,config *common.Config) []API {


	Apis = []API{
		{
			Srvname: "tx",
			Version:   "0.4",
			Service:   NewPublicTransactionAPI("Global", eventMux, pm, config),
			Public:    true,
		},
		{
			Srvname: "node",
			Version:   "0.4",
			Service:   NewPublicNodeAPI(pm),
			Public:    true,
		},
		{
			Srvname: "block",
			Version:   "0.4",
			Service:   NewPublicBlockAPI("Global"),
			Public:    true,
		},
		{
			Srvname: "account",
			Version:   "0.4",
			Service:   NewPublicAccountAPI("Global", pm, config),
			Public:    true,
		},
		{
			Srvname: "contract",
			Version:   "0.4",
			Service:   NewPublicContractAPI("Global", eventMux, pm, config),
			Public:    true,
		},
		{
			Srvname: "cert",
			Version:   "0.4",
			Service:   NewPublicCertAPI(cm),
			Public:    true,
		},
	}

	return Apis
}

func GetApiObjectByNamespace(name string) API {
	for _, api := range Apis {
		if api.Srvname == name {
			return api
		}
	}
	return API{}
}
