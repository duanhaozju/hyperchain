package exception

// Exception type of the module
const (
	exceptionSubType_Transactions      = "test"
)

// Error code
const (
	// definition format: <ExceptionCode> + <Module> + <SubType>
	exceptionCode_Jsonrpc_Transactions int =  -1 * iota
	// etc ...
)

type JsonrpcTestError struct {
	Msg  	string
}

func (ev JsonrpcTestError) ErrorCode() int {return exceptionCode_Jsonrpc_Transactions}
func (ev JsonrpcTestError) SubType() string {return exceptionSubType_Transactions}
func (ev JsonrpcTestError) Error() string {
	return ev.Msg
}