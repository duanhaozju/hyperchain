package dispatcher

import (
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
	"fmt"
)

//DispatchServer handleDispatch service
type DispatchServer struct {
	port   int
	host   string
	sr     common.ServiceRegistry
	logger *logging.Logger
}

func NewDispatchServer(port int, host string) (*DispatchServer, error) {
	ds := &DispatchServer{
		port:   port,
		host:   host,
		//logger: hcomm.GetLogger("system", "dispatcher"),
		sr:     common.NewServiceRegistry(),
		logger: logging.MustGetLogger("dispatcher"),
	}

	return ds, nil
}

func (ds *DispatchServer) Addr() string  {
	return fmt.Sprintf("%s:%d", ds.host, ds.port)
}

//Register receive a new connection
func (ds *DispatchServer) Register(stream pb.Dispatcher_RegisterServer) error {

	ds.logger.Infof("Receive new service connection!")
	for {
		msg, err := stream.Recv()
		if err != nil {
			ds.logger.Error(err)
			//TODO: handle broken stream
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			ds.handleRegister(msg, stream)
		case pb.Type_DISPATCH:
			ds.handleDispatch(msg)
		case pb.Type_ADMIN:
			//TODO: other types todo
		case pb.Type_RESPONSE:
		}
	}

	return nil
}

//handleDispatch handleDispatch messages
func (ds *DispatchServer) handleDispatch(msg *pb.Message) {
	switch msg.From {
	case pb.FROM_APISERVER:
		ds.dispatchAPIServerMsg(msg)
	case pb.FROM_CONSENSUS:
		ds.dispatchConsensusMsg(msg)
	case pb.FROM_EXECUTOR:
		ds.dispatchExecutorMsg(msg)
	case pb.FROM_NETWORK:
		ds.dispatchNetworkMsg(msg)
	default:
		ds.logger.Errorf("Undefined message: %v", msg)
	}
}

func (ds *DispatchServer) handleAdmin(msg *pb.Message) {
	//TODO: handle admin messages
}

//handleRegister parse msg and register this stream
func (ds *DispatchServer) handleRegister(msg *pb.Message, stream pb.Dispatcher_RegisterServer) {
	ds.logger.Debugf("handle register msg: %v", msg)
	rm := pb.RegisterMessage{}
	err := proto.Unmarshal(msg.Payload, &rm)
	if err != nil {
		ds.logger.Errorf("unmarshal register message error: %v", err)
		return
	}

	if len(rm.Namespace) == 0 {
		ds.logger.Error("namespace error, no namespace specified, using global instead")
		rm.Namespace = "global"
	}

	service := common.NewService(rm.Namespace, serviceId(msg), stream)
	ds.sr.Register(service)
	ds.logger.Debug("Send register ok response!")
	if err := stream.Send(&pb.Message{
		Type:pb.Type_RESPONSE,
		Ok:true,
	}); err != nil {
		ds.logger.Error(err)
	}

	go service.Serve()
}

//serviceId generate service id
func serviceId(msg *pb.Message) string {
	switch msg.From {
	case pb.FROM_CONSENSUS:
		return common.CONSENTER
	case pb.FROM_APISERVER:
		return common.APISERVER
	case pb.FROM_NETWORK:
		return common.NETWORK
	case pb.FROM_EXECUTOR:
		return common.EXECUTOR
	default:
		return ""
	}
}
