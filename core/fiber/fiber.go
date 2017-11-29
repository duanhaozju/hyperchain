package fiber

//Fiber response for transferring data between distributed components.
type Fiber interface {
	Start() error
	Stop()
}
