package exception


// Exception type of the module
const (
	exceptionSubType_ViewChange      = "viewchange"
	exceptionSubType_CutdownBlock    = "cutdownBlock"
)

// Error code
const (
	// definition format: <ExceptionCode> + <Module> + <SubType>
	exceptionCode_Executor_Viewchange int =  -1 * iota
	// etc ...
)

type ExecutorViewchangeError struct {
	Msg  	string
}

func (ev ExecutorViewchangeError) ErrorCode() int {return exceptionCode_Executor_Viewchange}
func (ev ExecutorViewchangeError) SubType() string {return exceptionSubType_ViewChange}
func (ev ExecutorViewchangeError) Error() string {
	return ev.Msg
}


