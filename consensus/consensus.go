package consensus


// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(msg []byte) error // Called serially with incoming messages from gRPC
}