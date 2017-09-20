//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package consensus provide consensus service for blockchian
//the consensus algorithm is pluggable, the default implementation is PBFT.
package consensus

// Consenter provides functions related to consensus to be invoked by outer services
// Every consensus algorithm needs to implement this interface.
type Consenter interface {
	// RecvMsg is called serially with incoming messages from gRPC.
	RecvMsg(msgPayload []byte) error

	// RecvMsg is called if local service pass message to consensus module.
	RecvLocal(msg interface{}) error

	// Start starts the consensus service
	Start()

	// Close closes the consensus service
	Close()

	// GetStatus returns the current status of consensus service
	GetStatus() (normal bool, full bool)
}
