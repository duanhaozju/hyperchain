package hpc

import (
	"hyperchain/event"
	"hyperchain/manager"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

var TxAPI *PublicTransactionAPI
var NodeAPI *PublicNodeAPI
var BlockAPI *PublicBlockAPI
var AccountAPI *PublicAccountAPI
var ContractAPI *PublicContractAPI

func GetAPIs(eventMux *event.TypeMux, pm *manager.ProtocolManager) []API{

	TxAPI = NewPublicTransactionAPI(eventMux,pm)
	NodeAPI = NewPublicNodeAPI(pm)
	BlockAPI = NewPublicBlockAPI()
	AccountAPI = NewPublicAccountAPI(pm)
	ContractAPI = NewPublicContractAPI()

	return []API{
		{
			Namespace: "tx",
			Version: "0.4",
			Service: TxAPI,
			Public: true,
		},
		{
			Namespace: "node",
			Version: "0.4",
			Service: NodeAPI,
			Public: true,
		},
		{
			Namespace: "block",
			Version: "0.4",
			Service: BlockAPI,
			Public: true,
		},
		{
			Namespace: "acot",
			Version: "0.4",
			Service: AccountAPI,
			Public: true,
		},
		{
			Namespace: "contract",
			Version: "0.4",
			Service: ContractAPI,
			Public: true,
		},
	}
}

func GetTxAPI() *PublicTransactionAPI{
	return TxAPI
}

func GetBlockAPI() *PublicBlockAPI{
	return BlockAPI
}

func GetNodeAPI() *PublicNodeAPI{
	return NodeAPI
}

func GetAccountAPI() *PublicAccountAPI{
	return AccountAPI
}

func GetContractAPI() *PublicContractAPI{
	return ContractAPI
}