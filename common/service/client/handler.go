package client

import (
	pb "github.com/hyperchain/hyperchain/common/service/protos"
)

type Handler interface {
	Handle(client pb.Dispatcher_RegisterClient, msg *pb.IMessage)
}
