package exception


type ExceptionError interface {
	ErrorCode()  int
	Error() string
	SubType()  string
}
