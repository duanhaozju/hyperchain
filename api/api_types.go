//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"hyperchain/admittance"
	"hyperchain/common"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

var Apis []API

func GetAPIs(eventMux *event.TypeMux, pm *manager.ProtocolManager, cm *admittance.CAManager,config *common.Config) []API {

	db, err := hyperdb.GetDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
	}

	Apis = []API{
		{
			Namespace: "tx",
			Version:   "0.4",
			Service:   NewPublicTransactionAPI("Global", eventMux, pm, db, config),
			Public:    true,
		},
		{
			Namespace: "node",
			Version:   "0.4",
			Service:   NewPublicNodeAPI(pm),
			Public:    true,
		},
		{
			Namespace: "block",
			Version:   "0.4",
			Service:   NewPublicBlockAPI("Global", db),
			Public:    true,
		},
		{
			Namespace: "account",
			Version:   "0.4",
			Service:   NewPublicAccountAPI("Global", pm, db, config),
			Public:    true,
		},
		{
			Namespace: "contract",
			Version:   "0.4",
			Service:   NewPublicContractAPI("Global", eventMux, pm, db, config),
			Public:    true,
		},
		{
			Namespace: "cert",
			Version:   "0.4",
			Service:   NewPublicCertAPI(cm),
			Public:    true,
		},
	}

	return Apis
}

func GetApiObjectByNamespace(name string) API {
	for _, api := range Apis {
		if api.Namespace == name {
			return api
		}
	}
	return API{}
}
