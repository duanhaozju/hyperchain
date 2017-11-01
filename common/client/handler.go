package client

import (
	pb "hyperchain/common/protos"
)

type Handler interface {
	Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage)
}