package errors

import "fmt"

type ConfigNotFoundError struct {
	message string
}

func ConfigNotFoundErr(str string, v ...interface{}) *ConfigNotFoundError {
	return &ConfigNotFoundError{fmt.Sprintf(str, v...)}
}

func (self *ConfigNotFoundError) Error() string {
	return self.message
}

type StartNamespaceError struct {
	message string
}

func StartNamespaceErr(str string, v ...interface{}) *StopNamespaceError {
	return &StopNamespaceError{fmt.Sprintf(str, v...)}
}

func (self *StartNamespaceError) Error() string {
	return self.message
}

type StopNamespaceError struct {
	message string
}

func StopNamespaceErr(str string, v ...interface{}) *StopNamespaceError {
	return &StopNamespaceError{fmt.Sprintf(str, v...)}
}

func (self *StopNamespaceError) Error() string {
	return self.message
}
