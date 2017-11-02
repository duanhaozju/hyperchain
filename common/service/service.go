package service

import (
	pb "hyperchain/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	//Namespace return namespace this service reside.
	Namespace() string

	//Id service identifier.
	Id() string

	//Send async send method.
	Send(se ServiceEvent) error

	//SyncSend synchronous send method.
	SyncSend(se ServiceEvent) (*pb.IMessage, error)

	//Close close the service.
	Close()

	//Serve start the service goroutine.
	Serve() error

	//IsHealth judge whether the service is health.
	IsHealth() bool
}

//common service event
type ServiceEvent interface{}
