//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package consensus provide consensus service for blockchian
//the consensus algorithm is pluggable, the default implementation is RBFT.
package consensus

// This file defines the Consenter interface, which manages all
// operations related to a certain consenter.

// Consenter provides functions related to consensus invoked by outer services.
// Every consensus algorithm needs to implement this interface.
type Consenter interface {
	// RecvMsg is called serially with incoming messages from gRPC.
	RecvMsg(msgPayload []byte) error

	// RecvLocal is called if local service sends message to consensus module.
	RecvLocal(msg interface{}) error

	// Start starts the consensus service
	Start()

	// Close closes the consensus service
	Stop()

	// GetStatus returns the current status of consensus service,
	// normal means this system is working well, and full
	// means this node can't process transactions any more. If this node
	// is abnormal or full, it can't work in that moment.
	GetStatus() (normal bool, full bool)
}
