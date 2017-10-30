package server

import (
	"fmt"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
)

//remoteServiceImpl represent a remote service.
type remoteServiceImpl struct {
	ds        *InternalServer
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
	r         chan *pb.IMessage
	logger    *logging.Logger
}

func NewRemoteService(namespace, id string, stream pb.Dispatcher_RegisterServer, ds *InternalServer) service.Service {
	return &remoteServiceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
		logger:    logging.MustGetLogger("service"),
		ds:        ds,
		r:         make(chan *pb.IMessage),
	}
}

func (rsi remoteServiceImpl) Namespace() string {
	return rsi.namespace
}

// Id service identifier.
func (rsi *remoteServiceImpl) Id() string {
	return rsi.id
}

// Send sync send msg.
func (rsi *remoteServiceImpl) Send(event service.ServiceEvent) error {
	if msg, ok := event.(*pb.IMessage); !ok {
		return fmt.Errorf("send message type error, %v need pb.IMessage ", event)
	} else {
		if rsi.stream == nil {
			return fmt.Errorf("[%s:%s]stream is empty, wait for this component to reconnect", rsi.namespace, rsi.id)
		}
		return rsi.stream.Send(msg)
	}
}

func (rsi *remoteServiceImpl) Close() {
	//TODO: close service
}

//Serve handle logic impl here.
func (rsi *remoteServiceImpl) Serve() error {
	for {
		msg, err := rsi.stream.Recv()
		if err != nil {
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			rsi.logger.Errorf("No register message should be here! msg: %v", msg)
		case pb.Type_DISPATCH:
			rsi.ds.HandleDispatch(rsi.namespace, msg)
		case pb.Type_ADMIN:
			rsi.ds.HandleAdmin(rsi.namespace, msg)
		case pb.Type_RESPONSE:
			rsi.r <- msg
		}
		rsi.logger.Debugf("%s, %s service serve", rsi.namespace, rsi.id)
	}
}

func (rsi *remoteServiceImpl) IsHealth() bool {
	//TODO: more to check
	return rsi.stream != nil
}

func (rsi *remoteServiceImpl) Response() chan *pb.IMessage {
	return rsi.r
}
