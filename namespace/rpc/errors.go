package rpc

import "fmt"

const (
	JSONRPCVersion         = "2.0"
	serviceMethodSeparator = "_"
)

// RPCError implements RPC error, is add support for error codec over regular go errors
type RPCError interface {
	// RPC error code
	Code() int
	// Error message
	Error() string
}

// CORE ERRORS
// request is for an unknown service
type methodNotFoundError struct {
	service string
	method  string
}

func (e *methodNotFoundError) Code() int {
	return -32601
}

func (e *methodNotFoundError) Error() string {
	return fmt.Sprintf("The method %s%s%s does not exist/is not available", e.service, serviceMethodSeparator, e.method)
}

// received message isn't a valid request
type invalidRequestError struct {
	message string
}

func (e *invalidRequestError) Code() int {
	return -32600
}

func (e *invalidRequestError) Error() string {
	return e.message
}

// received message is invalid
type invalidMessageError struct {
	message string
}

func (e *invalidMessageError) Code() int {
	return -32700
}

func (e *invalidMessageError) Error() string {
	return e.message
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

// issued when a request is received after the server is issued to stop.
type shutdownError struct {
}

func (e *shutdownError) Code() int {
	return -32000
}

func (e *shutdownError) Error() string {
	return "server is shutting down"
}

type UnauthorizedError struct {
}

func (e *UnauthorizedError) Code() int {
	return -32099
}

func (e *UnauthorizedError) Error() string {
	return "Unauthorized, Please check your cert"
}



// JSONRPC ERRORS
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
