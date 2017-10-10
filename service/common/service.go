package common

import (
	"fmt"
	"github.com/op/go-logging"
	pb "hyperchain/service/common/protos"
)

//Service interface to be implemented by component.
type Service interface {
	Namespace() string
	Id() string // service identifier.
	Send(syn bool, msg *pb.Message) error
	Close()
	Serve() error
	isHealth() bool
}

type serviceImpl struct {
	ds        *DispatchServer
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
	r         chan *pb.Message
	logger    *logging.Logger
}

func NewService(namespace, id string, stream pb.Dispatcher_RegisterServer, ds *DispatchServer) Service {
	return &serviceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
		logger:    logging.MustGetLogger("service"),
		ds:        ds,
		r:         make(chan *pb.Message),
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
func (si *serviceImpl) Send(sync bool, msg *pb.Message) error {
	if !sync {
		if si.stream == nil {
			return fmt.Errorf("[%s:%s]stream is empty, wait for this component to reconnect", si.namespace, si.id)
		}
		return si.stream.Send(msg)
	} else {
		//TODO: syn send operation
	}
	return nil
}

func (si *serviceImpl) Close() {
	//TODO: close service
}

//Serve handle logic impl here.
func (si *serviceImpl) Serve() error {
	for {
		msg, err := si.stream.Recv()
		if err != nil {
			si.logger.Error(err)
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			si.logger.Errorf("No register message should be here! msg: %v", msg)
		case pb.Type_DISPATCH:
			si.ds.HandleDispatch(si.namespace, msg)
		case pb.Type_ADMIN:
			si.ds.handleAdmin(si.namespace, msg)
		case pb.Type_RESPONSE:
			si.r <- msg
		}
	}
}

func (si *serviceImpl) isHealth() bool {
	//TODO: more to check
	return si.stream != nil
}
