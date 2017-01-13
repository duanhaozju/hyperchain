//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package consensus

// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(msgPayload []byte) error // Called serially with incoming messages from gRPC
	RecvLocal(msg interface{}) error //Called if local pass message to consensus module
	Close()
}
