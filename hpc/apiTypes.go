package hpc

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

func GetAPIs() []API{
	return []API{
		{
			Namespace: "tx",
			Version: "0.4",
			Service: NewPublicTransactionAPI(),
			Public: true,
		},
		{
			Namespace: "node",
			Version: "0.4",
			Service: NewPublicNodeAPI(),
			Public: true,
		},
		{
			Namespace: "block",
			Version: "0.4",
			Service: NewPublicBlockAPI(),
			Public: true,
		},
		{
			Namespace: "acot",
			Version: "0.4",
			Service: NewPublicAccountAPI(),
			Public: true,
		},
	}
}