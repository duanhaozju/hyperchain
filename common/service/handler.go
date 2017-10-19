package service

import (
	pb "hyperchain/common/protos"
)

type Handler interface {
	Handle(msg *pb.IMessage)
}