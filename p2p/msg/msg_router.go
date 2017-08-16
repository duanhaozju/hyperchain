package msg

import (
	"hyperchain/manager/event"
	pb "hyperchain/p2p/message"
)

type MsgRouter interface {
	Register(namespace string, eventMux *event.TypeMux) error
	DeRegister(namespace string) error
	BlackHole() chan <- *pb.Message
	// Notice this method should run in go routine
	Distribute() error
}
