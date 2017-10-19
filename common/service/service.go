package service

import (
	pb "hyperchain/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string // service identifier.
	Send(msg interface{}) error
	Close()
	Serve() error
	isHealth() bool
	Response() chan *pb.IMessage
}