package service

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	e "hyperchain/manager/event"
	"sync"
)

//InternalServer handle internal service connections
type InternalServer struct {
	port   int
	host   string
	sr     serviceRegistry
	logger *logging.Logger
}

func NewInternalServer(port int, host string) (*InternalServer, error) {
	ds := &InternalServer{
		port:   port,
		host:   host,
		sr:     NewServiceRegistry(),
		logger: logging.MustGetLogger("dispatcher"),
	}

	return ds, nil
}

func (is *InternalServer) Addr() string {
	return fmt.Sprintf("%s:%d", is.host, is.port)
}

//Register receive a new connection
func (is *InternalServer) Register(stream pb.Dispatcher_RegisterServer) error {
	is.logger.Infof("Receive new service connection!")

	var s Service
	var lock sync.RWMutex
	for {
		msg, err := stream.Recv()
		if err != nil {
			is.logger.Error(err)
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			lock.Lock()
			s = is.handleRegister(msg, stream)
			lock.Unlock()
		default:
			is.logger.Errorf("Message undefined %v", msg)
		}

		lock.RLock()
		if s != nil && s.isHealth() {
			lock.RUnlock()
			err := s.Serve()
			if err != nil {
				is.logger.Error(err)
				return err
			}
			return nil
		} else {
			is.logger.Errorf("Service register error, msg %v", msg)
		}
		lock.RUnlock()

	}
	return nil
}

func (is *InternalServer) RegisterLocal(s Service)  {
	is.sr.Register(s)
}

//handleDispatch handleDispatch messages
func (is *InternalServer) HandleDispatch(namespace string, msg *pb.IMessage) {
	is.logger.Debugf("try to handle dispatch message: %v for namespace: %s", msg, namespace)
	switch msg.From {
	case pb.FROM_APISERVER:
		is.dispatchAPIServerMsg(namespace, msg)
	case pb.FROM_CONSENSUS:
		is.dispatchConsensusMsg(namespace, msg)
	case pb.FROM_EXECUTOR:
		is.dispatchExecutorMsg(namespace, msg)
	case pb.FROM_NETWORK:
		is.dispatchNetworkMsg(namespace, msg)
	default:
		is.logger.Errorf("Undefined message: %v", msg)
	}
}

func (is *InternalServer) HandleAdmin(namespace string, msg *pb.IMessage) {
	//TODO: handle admin messages
}

//handleRegister parse msg and register this stream
func (is *InternalServer) handleRegister(msg *pb.IMessage, stream pb.Dispatcher_RegisterServer) Service {
	is.logger.Debugf("handle register msg: %v", msg)
	rm := pb.RegisterMessage{}
	err := proto.Unmarshal(msg.Payload, &rm)
	if err != nil {
		is.logger.Errorf("unmarshal register message error: %v", err)
		return nil
	}

	if len(rm.Namespace) == 0 {
		is.logger.Error("namespace error, no namespace specified, using global instead")
		rm.Namespace = "global"
	}

	service := NewRemoteService(rm.Namespace, serviceId(msg), stream, is)
	is.sr.Register(service)
	is.logger.Debug("Send register ok response!")
	if err := stream.Send(&pb.IMessage{
		Type: pb.Type_RESPONSE,
		Ok:   true,
	}); err != nil {
		is.logger.Error(err)
	}

	v := &e.ValidationEvent{
	    Digest: "xcc",
	    IsPrimary: true,
	    View: 111,
	    SeqNo: 222,
    }
	mv, err := proto.Marshal(v)
    if err := stream.Send(&pb.IMessage{
        Type: pb.Type_DISPATCH,
        Event:pb.Event_ValidationEvent,
        Payload: mv,
    }); err != nil {
        is.logger.Error(err)
    }
	return service
}

//serviceId generate service id
func serviceId(msg *pb.IMessage) string {
	switch msg.From {
	case pb.FROM_CONSENSUS:
		return CONSENTER
	case pb.FROM_APISERVER:
		return APISERVER
	case pb.FROM_NETWORK:
		return NETWORK
	case pb.FROM_EXECUTOR:
		return EXECUTOR
	default:
		return ""
	}
}
