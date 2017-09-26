package common

import (
	pb "hyperchain/service/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string // service identifier.
	Send(syn bool, msg *pb.Message)
	Close()
	Serve()
}

type serviceImpl struct {
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
}

func NewService(namespace, id string, stream pb.Dispatcher_RegisterServer) Service {
	return &serviceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
	}
}

func (si serviceImpl) Namespace() string {
	return si.namespace
}

// Id service identifier.
func (si *serviceImpl) Id() string {
	return si.id
}

// Send sync send msg.
func (si *serviceImpl) Send(sync bool, msg *pb.Message) {
	if !sync {
		si.stream.Send(msg)
	} else {
		//TODO: syn send operation
	}
}

func (si *serviceImpl) Close() {
	//TODO: close service
}

//Serve handle logic impl here.
func (si *serviceImpl) Serve() {

}
