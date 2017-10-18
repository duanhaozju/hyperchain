package service

import (
	pb "hyperchain/common/protos"
)

type Handler interface {
	Handle(msg *pb.Message)
	//deal and response
}

type handlerImpl struct {
}

func (*handlerImpl) Handle(msg *pb.Message) {
	//call executor method


}