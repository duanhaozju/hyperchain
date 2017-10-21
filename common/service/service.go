package service

import (
	pb "hyperchain/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string // service identifier.
	Send(se ServiceEvent) error
	Close()
	Serve() error
	IsHealth() bool
	Response() chan *pb.IMessage
}

type ServiceEvent interface {}
