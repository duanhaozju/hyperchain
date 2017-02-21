package hpc


type JSONRPCError interface{
	Code() int
	Error() string
}

// unable to decode supplied params, or an invalid number of parameters
type InvalidParamsError struct {
	Message string
}

func (e *InvalidParamsError) Code() int {
	return -32602
}

func (e *InvalidParamsError) Error() string {
	return e.Message
}

// logic error, callback returned an error
type CallbackError struct {
	Message string
}

func (e *CallbackError) Code() int {
	return -32000
}

func (e *CallbackError) Error() string {
	return e.Message
}

type LeveldbNotFoundError struct {
	Message string
}

func (e *LeveldbNotFoundError) Code() int {
	return -32001
}

func (e *LeveldbNotFoundError) Error() string {
	return "Not found " + e.Message
}

type OutofBalanceError struct {
	Message string
}

func (e *OutofBalanceError) Code() int {
	return -32002
}

func (e *OutofBalanceError) Error() string {
	return e.Message
}

type SignatureInvalidError struct {
	Message string
}

func (e *SignatureInvalidError) Code() int {
	return -32003
}

func (e *SignatureInvalidError) Error() string {
	return e.Message
}

type ContractDeployError struct {
	Message string
}

func (e *ContractDeployError) Code() int {
	return -32004
}

func (e *ContractDeployError) Error() string {
	return e.Message
}

type ContractInvokeError struct {
	Message string
}

func (e *ContractInvokeError) Code() int {
	return -32005
}

func (e *ContractInvokeError) Error() string {
	return e.Message
}

type SystemTooBusyError struct {
	Message string
}

func (e *SystemTooBusyError) Code() int {
	return -32006
}

func (e *SystemTooBusyError) Error() string {
	return e.Message
}

type RepeadedTxError struct {
	Message string
}

func (e *RepeadedTxError) Code() int {
	return -32007
}

func (e *RepeadedTxError) Error() string {
	return e.Message
}

//type marshalError struct {
//	Message  string
//}
//
//func (e *marshalError) Code() int {
//	return -32008
//}
//
//func (e *marshalError) Error() string {
//	return e.Message
//}
//
//type nullPointError struct {
//	Message  string
//}
//
//func (e *nullPointError) Code() int {
//	return -32009
//}
//
//func (e *nullPointError) Error() string {
//	return e.Message
//}

type CertError struct {
	Message string
}

func (e *CertError) Code() int {
	return -32099
}

func (e *CertError) Error() string {
	return e.Message
}
