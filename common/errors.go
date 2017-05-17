package common

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
type MethodNotFoundError struct {
	Service string
	Method  string
}

func (e *MethodNotFoundError) Code() int {
	return -32601
}

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("The method %s%s%s does not exist/is not available", e.Service, serviceMethodSeparator, e.Method)
}

// received message isn't a valid request
type InvalidRequestError struct {
	Message string
}

func (e *InvalidRequestError) Code() int {
	return -32600
}

func (e *InvalidRequestError) Error() string {
	return e.Message
}

// received message is invalid
type InvalidMessageError struct {
	Message string
}

func (e *InvalidMessageError) Code() int {
	return -32700
}

func (e *InvalidMessageError) Error() string {
	return e.Message
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

// issued when a request is received after the server is issued to stop.
type ShutdownError struct {
}

func (e *ShutdownError) Code() int {
	return -32000
}

func (e *ShutdownError) Error() string {
	return "server is shutting down"
}




// JSONRPC ERRORS
type JSONRPCError interface{
	Code() int
	Error() string
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

type ContractPermissionError struct {
	Message    string
}

func (e *ContractPermissionError) Code() int {
	return -32008
}

func (e *ContractPermissionError) Error() string {
	return fmt.Sprintf("The contract invocation permission not enough '%s'", e.Message)
}

type AccountNotExistError struct {
	Message    string
}

func (e *AccountNotExistError) Code() int {
	return -32009
}

func (e *AccountNotExistError) Error() string {
	return fmt.Sprintf("The account dose not exist '%s'", e.Message)
}

type NamespaceNotFound struct {
	Name    string
}

func (e *NamespaceNotFound) Code() int {
	return -32010
}

func (e *NamespaceNotFound) Error() string {
	return fmt.Sprintf("The namespace '%s' does not exist", e.Name)
}

type NoBlockGeneratedError struct {
	Message    string
}

func (e *NoBlockGeneratedError) Code() int {
	return -32011
}

func (e *NoBlockGeneratedError) Error() string {
	return fmt.Sprintf(e.Message)
}

type UnauthorizedError struct {
}

func (e *UnauthorizedError) Code() int {
	return -32098
}

func (e *UnauthorizedError) Error() string {
	return "Unauthorized, Please check your cert"
}

type CertError struct {
	Message string
}

func (e *CertError) Code() int {
	return -32099
}

func (e *CertError) Error() string {
	return e.Message
}
