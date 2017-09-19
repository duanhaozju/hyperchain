package msg

import (
	pb "hyperchain/p2p/message"
)

type MsgHandler interface {
	BidiHandler
	SingleHandler
}

type BidiHandler interface {
	Receive() chan<- interface{}
	Process()
	Teardown()
}

type SingleHandler interface {
	Execute(msg *pb.Message) (*pb.Message, error)
}
