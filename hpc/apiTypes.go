//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/hyperdb"
	"time"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

var Apis []API

func GetAPIs(eventMux *event.TypeMux, pm *manager.ProtocolManager, ratelimitEnable bool, txPeak int64 , txRate time.Duration, contractPeak int64, contractRate time.Duration, stateType string) []API{

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
	}

	Apis = []API{
		{
			Namespace: "tx",
			Version: "0.4",
			Service: NewPublicTransactionAPI(eventMux, pm, db, ratelimitEnable, txPeak, txRate),
			Public: true,
		},
		{
			Namespace: "node",
			Version: "0.4",
			Service: NewPublicNodeAPI(pm),
			Public: true,
		},
		{
			Namespace: "block",
			Version: "0.4",
			Service: NewPublicBlockAPI(db),
			Public: true,
		},
		{
			Namespace: "account",
			Version: "0.4",
			Service: NewPublicAccountAPI(pm, db, stateType),
			Public: true,
		},
		{
			Namespace: "contract",
			Version: "0.4",
			Service: NewPublicContractAPI(eventMux, pm, db, ratelimitEnable, contractPeak, contractRate, stateType),
			Public: true,
		},
	}

	return Apis
}

func GetApiObjectByNamespace(name string) API{
	for _,api := range Apis {
		if api.Namespace == name {
			return api
		}
	}
	return API{}
}
