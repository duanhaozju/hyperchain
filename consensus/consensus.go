package consensus

import "hyperchain/event"

// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(msgPayload []byte) error // Called serially with incoming messages from gRPC
	RecvValidatedResult(validateBatch event.ValidatedTxs) error //called to pass validated batch result
	Close()
}