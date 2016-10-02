package core

import "fmt"

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

type SignatureErr struct {
	message string
}

func SignatureError(str string, v ...interface{}) *SignatureErr {
	return &SignatureErr{fmt.Sprintf(str, v...)}
}
func (self *SignatureErr) Error() string {
	return self.message
}
func IsSignatureError(e error) bool {
	_, ok := e.(*SignatureErr)
	return ok
}
