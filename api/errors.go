package hpc

type leveldbNotFoundError struct {
	message  string
}

// unable to decode supplied params, or an invalid number of parameters
type invalidParamsError struct {
	message string
}

func (e *invalidParamsError) Code() int {
	return -32602
}

func (e *invalidParamsError) Error() string {
	return e.message
}

// logic error, callback returned an error
type callbackError struct {
	message string
}

func (e *callbackError) Code() int {
	return -32000
}

func (e *callbackError) Error() string {
	return e.message
}

func (e *leveldbNotFoundError) Code() int {
	return -32001
}

func (e *leveldbNotFoundError) Error() string {
	return "Not found " + e.message
}

type outofBalanceError struct {
	message  string
}

func (e *outofBalanceError) Code() int {
	return -32002
}

func (e *outofBalanceError) Error() string {
	return e.message
}

type signatureInvalidError struct {
	message  string
}

func (e *signatureInvalidError) Code() int {
	return -32003
}

func (e *signatureInvalidError) Error() string {
	return e.message
}

type contractDeployError struct {
	message  string
}

func (e *contractDeployError) Code() int {
	return -32004
}

func (e *contractDeployError) Error() string {
	return e.message
}

type contractInvokeError struct {
	message  string
}

func (e *contractInvokeError) Code() int {
	return -32005
}

func (e *contractInvokeError) Error() string {
	return e.message
}

type systemTooBusyError struct {
	message  string
}

func (e *systemTooBusyError) Code() int {
	return -32006
}

func (e *systemTooBusyError) Error() string {
	return e.message
}

type repeadedTxError struct {
	message  string
}

func (e *repeadedTxError) Code() int {
	return -32007
}

func (e *repeadedTxError) Error() string {
	return e.message
}

//type marshalError struct {
//	message  string
//}
//
//func (e *marshalError) Code() int {
//	return -32008
//}
//
//func (e *marshalError) Error() string {
//	return e.message
//}
//
//type nullPointError struct {
//	message  string
//}
//
//func (e *nullPointError) Code() int {
//	return -32009
//}
//
//func (e *nullPointError) Error() string {
//	return e.message
//}