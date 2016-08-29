package consensus

import "hyperchain/consensus/events"

// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(e events.Event) error // Called serially with incoming messages from gRPC
}