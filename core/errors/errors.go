// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package errors

import (
	"errors"
	"fmt"
)

// ValueTransferError refers a balance transfer error.
type ValueTransferError struct {
	message string
}

func ValueTransferErr(str string, v ...interface{}) *ValueTransferError {
	return &ValueTransferError{fmt.Sprintf(str, v...)}
}

func (self *ValueTransferError) Error() string {
	return self.message
}
func IsValueTransferErr(e error) bool {
	_, ok := e.(*ValueTransferError)
	return ok
}

// SignatureError refers an invalid signature error.
type SignatureError struct {
	message string
}

func SignatureErr(str string, v ...interface{}) *SignatureError {
	return &SignatureError{fmt.Sprintf(str, v...)}
}
func (self *SignatureError) Error() string {
	return self.message
}
func IsSignatureErr(e error) bool {
	_, ok := e.(*SignatureError)
	return ok
}

// ExecContractError refers a virtual machine execution failed error.
type ExecContractError struct {
	message string
	errType int
}

func ExecContractErr(t int, str string, v ...interface{}) *ExecContractError {
	return &ExecContractError{
		message: fmt.Sprintf(str, v...),
		errType: t,
	}
}
func (self *ExecContractError) Error() string {
	return self.message
}

func (self *ExecContractError) GetType() int {
	return self.errType
}

func IsExecContractErr(e error) bool {
	_, ok := e.(*ExecContractError)
	return ok
}

// InvalidInvokePermissionError refers a permission error.
// Always occurs when the caller doesn't have enough permission.
type InvalidInvokePermissionError struct {
	message string
}

func InvalidInvokePermissionErr(str string, v ...interface{}) *InvalidInvokePermissionError {
	return &InvalidInvokePermissionError{fmt.Sprintf(str, v...)}
}
func (self *InvalidInvokePermissionError) Error() string {
	return self.message
}

func IsInvalidInvokePermissionErr(e error) bool {
	_, ok := e.(*InvalidInvokePermissionError)
	return ok
}

/*
	Constant errors
*/
var (
	InvalidParamsErr   = errors.New("invalid params")
	NoDefinedCaseErr   = errors.New("no defined case")
	EmptyPointerErr    = errors.New("nil pointer")
	TxIdLenErr         = errors.New("tx's id length does not match.")
	MarshalFailedErr   = errors.New("marshal failed")
	NotEnoughReplyErr  = errors.New("not enough reply")
	BlockIntegrityErr  = errors.New("block in not integrated")
	SeizeLockFailedErr = errors.New("failed to seize the lock")
	TimeoutErr         = errors.New("timeout")
)
