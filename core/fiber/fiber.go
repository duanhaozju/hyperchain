package fiber

//Fiber response for transferring data between distributed components.
type Fiber interface {
	//Start start the fiber system
	Start() error
	//Stop the fiber system
	Stop()
	//Send info to the remote peer
	Send(msg interface{}) error
}
