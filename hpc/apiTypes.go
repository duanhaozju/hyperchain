package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/hyperdb"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

func GetAPIs(eventMux *event.TypeMux, pm *manager.ProtocolManager) []API{

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
	}

	return []API{
		{
			Namespace: "tx",
			Version: "0.4",
			Service: NewPublicTransactionAPI(eventMux, pm, db),
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
			Namespace: "acc",
			Version: "0.4",
			Service: NewPublicAccountAPI(pm, db),
			Public: true,
		},
		{
			Namespace: "contract",
			Version: "0.4",
			Service: NewPublicContractAPI(eventMux, pm, db),
			Public: true,
		},
	}
}
