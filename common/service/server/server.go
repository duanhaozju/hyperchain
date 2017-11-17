package server

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	pb "github.com/hyperchain/hyperchain/common/protos"
	"github.com/hyperchain/hyperchain/common/service"
	"github.com/op/go-logging"
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

func (is *InternalServer) AdminRegister() chan struct{} {
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
			if s != nil && msg.From == pb.FROM_ADMINISTRATOR {
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
	is.sr.Register(s)
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
	if msg.From == pb.FROM_ADMINISTRATOR {
		// the admin stream register
		service := NewRemoteService(rm.Namespace, adminId(&rm), stream)
		is.logger.Debugf("admin addr %v", rm.Address)
		restart := is.sr.AddAdminService(service)
		var payload []byte
		if restart {
			payload = []byte("restart")
		}

		if err := service.Send(&pb.IMessage{
			Id:      msg.Id,
			Type:    pb.Type_RESPONSE,
			Ok:      true,
			Payload: payload,
		}); err != nil {
			is.logger.Error(err)
		}
		is.logger.Debug("Send admin register ok response!")
		go service.Serve()
		return service
	} else {
		// normal stream register

		if len(rm.Namespace) == 0 {
			is.logger.Error("namespace error, no namespace specified, using global instead")
			rm.Namespace = "global"
		}
		service := NewRemoteService(rm.Namespace, serviceId(msg), stream)
		is.sr.Register(service)
		if err := service.Send(&pb.IMessage{
			Id:   msg.Id,
			Type: pb.Type_RESPONSE,
			Ok:   true,
		}); err != nil {
			is.logger.Error(err)
		}
		is.logger.Debug("Send register ok response!")
		return service
	}
}

func (is *InternalServer) dispatchResponse(msg *pb.IMessage) {

}

//serviceId generate service id
func serviceId(msg *pb.IMessage) string {
	id := fmt.Sprintf("%s-%d", msg.From, msg.Cid)
	return id
}

func adminId(msg *pb.RegisterMessage) string {
	return msg.Address
}
