package dispatcher

import (
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	hcomm "hyperchain/common"
	"hyperchain/service/common"
	pb "hyperchain/service/common/protos"
)

//DispatchServer dispatch service
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
		logger: hcomm.GetLogger("system", "dispatcher"),
	}
	return ds, nil
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
		case pb.Message_REGISTER:
			ds.handleRegister(msg, stream)
		case pb.Message_ADMIN:
			//TODO: other types todo
		case pb.Message_INVOKE:
			ds.handleInvoke(msg)
		case pb.Message_RESPONSE:
		}
	}

	return nil
}

//dispatch dispatch messages
func (ds *DispatchServer) dispatch() {

}

//handleRegister parse msg and register this stream
func (ds *DispatchServer) handleRegister(msg *pb.Message, stream pb.Dispatcher_RegisterServer) {
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
	go service.Serve()
}

//handleInvoke handle message invocation
func (ds *DispatchServer) handleInvoke(msg *pb.Message) {
	//dispatch message
}

//serviceId generate service id
func serviceId(msg *pb.Message) string {
	switch msg.From {
	case pb.Message_CONSENSUS:
		return common.CONSENTER
	case pb.Message_APISERVER:
		return common.APISERVER
	case pb.Message_NETWORK:
		return common.NETWORK
	case pb.Message_EXECUTOR:
		return common.EXECUTOR
	default:
		return ""
	}
}
