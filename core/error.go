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
