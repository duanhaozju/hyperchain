package service

import (
	pb "hyperchain/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string // service identifier.
	Send(se ServiceEvent) error
	SyncSend(se ServiceEvent) (*pb.IMessage, error)
	Close()
	Serve() error
	IsHealth() bool
	Response() chan *pb.IMessage //TODO: this method will be deprecated
}

type ServiceEvent interface{}
