package common

import (
	pb "hyperchain/service/common/protos"
	"github.com/op/go-logging"
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
	ds        *DispatchServer
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
	logger   *logging.Logger
}

func NewService(namespace, id string, stream pb.Dispatcher_RegisterServer, ds *DispatchServer) Service {
	return &serviceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
		logger:    logging.MustGetLogger("service"),
		ds:        ds,
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
	for {
		msg, err := si.stream.Recv()
		if err != nil {
			si.logger.Error(err)
			//TODO: handle broken stream
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			//ds.handleRegister(msg, stream)
			//No register event should be here
		case pb.Type_DISPATCH:
			si.ds.HandleDispatch(si.namespace, msg)
		case pb.Type_ADMIN:
				//TODO: other types todo
		case pb.Type_RESPONSE:
		}
	}
}
