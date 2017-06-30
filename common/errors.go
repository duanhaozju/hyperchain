package common

import "fmt"

const (
	serviceMethodSeparator = "_"
)

// JSON-RPC specified error
const (
	Specified_InvalidMessageError int    = -32700
	Specified_CallbackError       int    = -32000
	Specified_ShutdownError       int    = -32000
)
const (
	Specified_InvalidRequestError int    =  -32600 - iota
	Specified_MethodNotFoundError
	Specified_InvalidParamsError
)

// JSON-RPC custom error
const (
	CUSTOM_UnauthorizedError     int   = -32098
	CUSTOM_CertError             int   = -32099
)
const (
	CUSTOM_LeveldbNotFoundError  int   = -32001 - iota
	CUSTOM_OutofBalanceError
	CUSTOM_SignatureInvalidError
	CUSTOM_ContractDeployError
	CUSTOM_ContractInvokeError
	CUSTOM_SystemTooBusyError
	CUSTOM_RepeadedTxError
	CUSTOM_ContractPermissionError
	CUSTOM_AccountNotExistError
	CUSTOM_NamespaceNotFoundError
	CUSTOM_NoBlockGeneratedError
	CUSTOM_SubNotExistError
	CUSTOM_SnapshotError
)


// RPCError implements RPC error, is add support for error codec over regular go errors
type RPCError interface {
	// RPC error code
	Code() int
	// Error message
	Error() string
}

// CORE ERRORS
// received message isn't a valid request
type InvalidRequestError struct {
	Message string
}

func (e *InvalidRequestError) Code() int {return Specified_InvalidRequestError}
func (e *InvalidRequestError) Error() string {return e.Message}

// request is for an unknown service
type MethodNotFoundError struct {
	Service string
	Method  string
}

func (e *MethodNotFoundError) Code() int {return Specified_MethodNotFoundError}
func (e *MethodNotFoundError) Error() string {return fmt.Sprintf("The method %s%s%s does not exist/is not available",
										e.Service, serviceMethodSeparator, e.Method)}

// unable to decode supplied params, or an invalid number of parameters
type InvalidParamsError struct {
	Message string
}

func (e *InvalidParamsError) Code() int {return Specified_InvalidParamsError}
func (e *InvalidParamsError) Error() string {return e.Message}

// received message is invalid
type InvalidMessageError struct {
	Message string
}

func (e *InvalidMessageError) Code() int {return Specified_InvalidMessageError}
func (e *InvalidMessageError) Error() string {return e.Message}

// logic error, callback returned an error
type CallbackError struct {
	Message string
}

func (e *CallbackError) Code() int {return Specified_CallbackError}
func (e *CallbackError) Error() string {return e.Message}

// issued when a request is received after the server is issued to stop.
type ShutdownError struct {}

func (e *ShutdownError) Code() int {return Specified_ShutdownError}
func (e *ShutdownError) Error() string {return "server is shutting down"}

// JSONRPC CUSTOM ERRORS
type JSONRPCError interface {
	Code() int
	Error() string
}

type LeveldbNotFoundError struct {
	Message string
}

func (e *LeveldbNotFoundError) Code() int {return CUSTOM_LeveldbNotFoundError}
func (e *LeveldbNotFoundError) Error() string {return "Not found " + e.Message}

type OutofBalanceError struct {
	Message string
}

func (e *OutofBalanceError) Code() int {return CUSTOM_OutofBalanceError}
func (e *OutofBalanceError) Error() string {return e.Message}

type SignatureInvalidError struct {
	Message string
}

func (e *SignatureInvalidError) Code() int {return CUSTOM_SignatureInvalidError}
func (e *SignatureInvalidError) Error() string {return e.Message}

type ContractDeployError struct {
	Message string
}

func (e *ContractDeployError) Code() int {return CUSTOM_ContractDeployError}
func (e *ContractDeployError) Error() string {return e.Message}

type ContractInvokeError struct {
	Message string
}

func (e *ContractInvokeError) Code() int {return CUSTOM_ContractInvokeError}
func (e *ContractInvokeError) Error() string {return e.Message}

type SystemTooBusyError struct {
	Message string
}

func (e *SystemTooBusyError) Code() int {return CUSTOM_SystemTooBusyError}
func (e *SystemTooBusyError) Error() string {return e.Message}

type RepeadedTxError struct {
	Message string
}

func (e *RepeadedTxError) Code() int {return CUSTOM_RepeadedTxError}
func (e *RepeadedTxError) Error() string {return e.Message}

type ContractPermissionError struct {
	Message string
}

func (e *ContractPermissionError) Code() int {return CUSTOM_ContractPermissionError}
func (e *ContractPermissionError) Error() string {return fmt.Sprintf("The contract invocation permission not enough '%s'", e.Message)}

type AccountNotExistError struct {
	Message string
}

func (e *AccountNotExistError) Code() int {return CUSTOM_AccountNotExistError}
func (e *AccountNotExistError) Error() string {return fmt.Sprintf("The account dose not exist '%s'", e.Message)}

type NamespaceNotFound struct {
	Name string
}

func (e *NamespaceNotFound) Code() int {return CUSTOM_NamespaceNotFoundError}
func (e *NamespaceNotFound) Error() string {return fmt.Sprintf("The namespace '%s' does not exist", e.Name)}

type NoBlockGeneratedError struct {
	Message string
}

func (e *NoBlockGeneratedError) Code() int {return CUSTOM_NoBlockGeneratedError}
func (e *NoBlockGeneratedError) Error() string {return fmt.Sprintf(e.Message)}

type SubNotExistError struct {
	Message string
}

func (e *SubNotExistError) Code() int {return CUSTOM_SubNotExistError}
func (e *SubNotExistError) Error() string {return e.Message}

type SnapshotErr struct {
	Message string
}

func (e *SnapshotErr) Code() int {return CUSTOM_SnapshotError}
func (e *SnapshotErr) Error() string {return e.Message}

type UnauthorizedError struct {}

func (e *UnauthorizedError) Code() int {return CUSTOM_UnauthorizedError}
func (e *UnauthorizedError) Error() string {return "Unauthorized, Please check your cert"}

type CertError struct {
	Message string
}

func (e *CertError) Code() int {return CUSTOM_CertError}
func (e *CertError) Error() string {return e.Message}
