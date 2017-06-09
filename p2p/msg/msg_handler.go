package msg

import (
	pb "hyperchain/p2p/message"
	"fmt"
)

type MsgHandler interface{
	BidiHandler
	SingleHandler
}

type BidiHandler interface {
	Recive() <-chan *pb.Message
	Process()
	TearDown()
}

type SingleHandler interface {
	Execute(msg *pb.Message) (*pb.Message,error)
}

