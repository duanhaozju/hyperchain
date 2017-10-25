package server

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"sync"
)

//InternalServer handle internal service connections
type InternalServer struct {
	port          int
	host          string
	sr            service.ServiceRegistry
	logger        *logging.Logger
	adminRegister chan struct{}
}

func NewInternalServer(port int, host string) (*InternalServer, error) {
	ds := &InternalServer{
		port:          port,
		host:          host,
		sr:            service.NewServiceRegistry(),
		logger:        logging.MustGetLogger("dispatcher"),
		adminRegister: make(chan struct{}, 100),
	}

	return ds, nil
}

func (is *InternalServer) AdminRegister() chan struct{}  {
	return is.adminRegister
}

func (is *InternalServer) Addr() string {
	return fmt.Sprintf("%s:%d", is.host, is.port)
}

func (is *InternalServer) ServerRegistry() service.ServiceRegistry {
	return is.sr
}

//Register receive a new connection
func (is *InternalServer) Register(stream pb.Dispatcher_RegisterServer) error {
	is.logger.Infof("Receive new service connection!")

	var s service.Service
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
			if s != nil && msg.Event == pb.Event_AdminRegisterEvent {
				is.adminRegister <- struct{}{}
			}
			lock.Unlock()
		default:
			is.logger.Errorf("Message undefined %v", msg)
		}

		lock.RLock()
		if s != nil && s.IsHealth() {
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

func (is *InternalServer) RegisterLocal(s service.Service) {
	is.logger.Error(is.sr == nil)
	is.sr.Register(s)
}

//handleDispatch handleDispatch messages
func (is *InternalServer) HandleDispatch(namespace string, msg *pb.IMessage) {
	is.logger.Debugf("try to handle dispatch message: %v for namespace: %s", msg, namespace)
	switch msg.From {
	case pb.FROM_APISERVER:
		is.DispatchAPIServerMsg(namespace, msg)
	case pb.FROM_CONSENSUS:
		is.DispatchConsensusMsg(namespace, msg)
	case pb.FROM_EXECUTOR:
		is.DispatchExecutorMsg(namespace, msg)
	case pb.FROM_NETWORK:
		is.DispatchNetworkMsg(namespace, msg)
	default:
		is.logger.Errorf("Undefined message: %v", msg)
	}
}

func (is *InternalServer) HandleAdmin(namespace string, msg *pb.IMessage) {
	//TODO: handle admin messages
}

//handleRegister parse msg and register this stream
func (is *InternalServer) handleRegister(msg *pb.IMessage, stream pb.Dispatcher_RegisterServer) service.Service {
	is.logger.Debugf("handle register msg: %v", msg)
	rm := pb.RegisterMessage{}
	err := proto.Unmarshal(msg.Payload, &rm)
	if err != nil {
		is.logger.Errorf("unmarshal register message error: %v", err)
		return nil
	}
	if msg.Event == pb.Event_AdminRegisterEvent {
		// the admin stream register
		service := NewRemoteService(rm.Namespace, adminId(&rm), stream, is)
		is.sr.AddAdminService(service)
		is.logger.Debug("Send admin register ok response!")

		if err := stream.Send(&pb.IMessage{
			Type: pb.Type_RESPONSE,
			Ok:   true,
		}); err != nil {
			is.logger.Error(err)
		}
		return service
	} else {
		// normal stream register

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
		return service
	}
}

//serviceId generate service id
func serviceId(msg *pb.IMessage) string {
	switch msg.From {
	case pb.FROM_CONSENSUS:
		return service.CONSENTER
	case pb.FROM_APISERVER:
		return service.APISERVER
	case pb.FROM_NETWORK:
		return service.NETWORK
	case pb.FROM_EXECUTOR:
		return service.EXECUTOR
	default:
		return ""
	}
}

func adminId(msg *pb.RegisterMessage) string {
	return fmt.Sprintf("%s:%s", msg.Address, msg.Port)
}
