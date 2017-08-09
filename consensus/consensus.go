//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package consensus provide consensus service for blockchian
//the consensus algorithm is pluggable, the default implementation is PBFT.
package consensus

// Consenter provides functions related to consensus to be invoked by outer services
// Every consensus algorithm needs to implement this interface.
type Consenter interface {
	//RecvMsg called serially with incoming messages from gRPC.
	RecvMsg(msgPayload []byte) error

	//RecvMsg called if local service pass message to consensus module.
	RecvLocal(msg interface{}) error

	//Start start the consensus service
	Start()

	//Close close the consenter service
	Close()
}
