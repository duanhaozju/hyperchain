package common

import (
	pb "hyperchain/service/common/protos"
)

type Handler interface {
	Handle(msg *pb.Message)
}