package consensus

import (
	pb "github.com/hyperledger/fabric/protos"
)


// Consenter is used to receive messages from the network
// Every consensus plugin needs to implement this interface
type Consenter interface {
	RecvMsg(msg *pb.Message, peerID uint64) error // Called serially with incoming messages from gRPC
}

// Message queue outside consensus
type MessageQueue interface {
	GetMessageQueue() (msgQ chan *pb.Message)
}

type PeerID interface {
	GetPeerID() (peerID uint64)
}

type Stack interface {
	MessageQueue
	PeerID
}