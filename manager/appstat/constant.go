package appstat

const (
	Normal     = true
	Exception  = false
)

// Exception module name
const (
	// definition format: <ExceptionModule> + <Module>
	ExceptionModule_P2P      = "p2p"
	ExceptionModule_Consenus = "consensus"
	ExceptionModule_Executor = "executor"
	ExceptionModule_Jsonrpc  = "jsonrpc"
	// etc ...
)

// Exception type of the module
const (
	ExceptionSubType_ViewChange   = "viewchange"
	ExceptionSubType_Transactions = "test"
)

// Error code
const (
	// definition format: <ExceptionCode> + <Module> + <SubType>
	ExceptionCode_Executor_Viewchange int =  -1 * (iota + 1)
	ExceptionCode_Jsonrpc_Transactions
	// etc ...
)
