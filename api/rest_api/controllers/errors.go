package controllers

type JSONError interface {
	Code() int
	Error() string
}

type callbackError struct {
	message string
}

func (e *callbackError) Code() int {
	return -32000
}

func (e *callbackError) Error() string {
	return e.message
}

type invalidParamsError struct {
	message string
}

func (e *invalidParamsError) Code() int {
	return -32602
}

func (e *invalidParamsError) Error() string {
	return e.message
}

//type contractError struct {
//	message string
//}
//
//func (e *contractError) Code() int {
//	return -32602
//}
//
//func (e *contractError) Error() string {
//	return e.message
//}
