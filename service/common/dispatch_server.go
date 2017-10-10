package common

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	pb "hyperchain/service/common/protos"
	"sync"
)

//DispatchServer handleDispatch service
type DispatchServer struct {
	port   int
	host   string
	sr     ServiceRegistry
	logger *logging.Logger
}

func NewDispatchServer(port int, host string) (*DispatchServer, error) {
	ds := &DispatchServer{
		port:   port,
		host:   host,
		sr:     NewServiceRegistry(),
		logger: logging.MustGetLogger("dispatcher"),
	}

	return ds, nil
}

func (ds *DispatchServer) Addr() string {
	return fmt.Sprintf("%s:%d", ds.host, ds.port)
}

//Register receive a new connection
func (ds *DispatchServer) Register(stream pb.Dispatcher_RegisterServer) error {
	ds.logger.Infof("Receive new service connection!")

	var s Service
	var lock sync.RWMutex
	for {
		msg, err := stream.Recv()
		if err != nil {
			ds.logger.Error(err)
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			lock.Lock()
			s = ds.handleRegister(msg, stream)
			lock.Unlock()
		default:
			ds.logger.Errorf("Message undefined %v", msg)
		}

		lock.RLock()
		if s != nil && s.isHealth() {
			lock.RUnlock()
			err := s.Serve()
			if err != nil {
				ds.logger.Error(err)
				return err
			}
			return nil
		}else {
			ds.logger.Errorf("Service register error, msg %v", msg)
		}
		lock.RUnlock()

	}
	return nil
}

//handleDispatch handleDispatch messages
func (ds *DispatchServer) HandleDispatch(namespace string, msg *pb.Message) {
	//ds.logger.Debugf("try to handle dispatch message: %v for namespace: %s", msg, namespace)

	switch msg.From {
	case pb.FROM_APISERVER:
		ds.dispatchAPIServerMsg(namespace, msg)
	case pb.FROM_CONSENSUS:
		ds.dispatchConsensusMsg(namespace, msg)
	case pb.FROM_EXECUTOR:
		ds.dispatchExecutorMsg(namespace, msg)
	case pb.FROM_NETWORK:
		ds.dispatchNetworkMsg(namespace, msg)
	default:
		ds.logger.Errorf("Undefined message: %v", msg)
	}
}

func (ds *DispatchServer) handleAdmin(namespace string, msg *pb.Message) {
	//TODO: handle admin messages
}

//handleRegister parse msg and register this stream
func (ds *DispatchServer) handleRegister(msg *pb.Message, stream pb.Dispatcher_RegisterServer) Service {
	ds.logger.Debugf("handle register msg: %v", msg)
	rm := pb.RegisterMessage{}
	err := proto.Unmarshal(msg.Payload, &rm)
	if err != nil {
		ds.logger.Errorf("unmarshal register message error: %v", err)
		return nil
	}

	if len(rm.Namespace) == 0 {
		ds.logger.Error("namespace error, no namespace specified, using global instead")
		rm.Namespace = "global"
	}

	service := NewService(rm.Namespace, serviceId(msg), stream, ds)
	ds.sr.Register(service)
	ds.logger.Debug("Send register ok response!")
	if err := stream.Send(&pb.Message{
		Type: pb.Type_RESPONSE,
		Ok:   true,
	}); err != nil {
		ds.logger.Error(err)
	}
	return service
}

//serviceId generate service id
func serviceId(msg *pb.Message) string {
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
