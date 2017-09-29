package common

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/service/common/protos"
)

// ServiceClient used to send messages to eventhub or receive message
// from the event hub.
type ServiceClient struct {
	host   string
	port   int
	msgs   chan *pb.Message
	stream pb.Dispatcher_RegisterClient
	logger *logging.Logger
}

func New(host string, port int) (*ServiceClient, error) {
	if len(host) == 0 || port < 0 {
		return nil, fmt.Errorf("Invalid host or port, %s:%d", host, port)
	}
	return &ServiceClient{
		host:   host,
		port:   port,
		msgs:   make(chan *pb.Message, 1024),
		logger: logging.MustGetLogger("service_client"), // TODO: replace this logger with hyperlogger
	}, nil
}

//Connect connect to dispatch server.
func (sc *ServiceClient) Connect() error {
	connString := fmt.Sprintf("%s:%d", sc.host, sc.port)

	conn, err := grpc.Dial(connString, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pb.NewDispatcherClient(conn)
	stream, err := client.Register(context.Background())
	if err != nil {
		return err
	}

	sc.stream = stream
	return nil
}

func (sc *ServiceClient) Register(serviceType pb.FROM, rm *pb.RegisterMessage) error {
	payload, err := proto.Marshal(rm)
	if err != nil {
		return err
	}
	if err = sc.stream.Send(&pb.Message{
		Type:    pb.Type_REGISTER,
		From:    serviceType,
		Payload: payload,
	}); err != nil {
		return err
	}

	//timeout detection
	msg, err := sc.stream.Recv()
	if err != nil {
		return err
	}

	if msg.Type == pb.Type_RESPONSE && msg.Ok == true {
		sc.logger.Infof("Service %v in namespace %v register successful", serviceType, rm.Namespace)
		return nil
	} else {
		return fmt.Errorf("Service %v in namespace %v register failed", serviceType, rm.Namespace)
	}
}

//Send msg asynchronous
func (sc *ServiceClient) Send(msg *pb.Message) error {
	return sc.stream.Send(msg)
}

func (sc *ServiceClient) SyncSend(msg *pb.Message) (*pb.Message, error) {
	err := sc.stream.Send(msg)
	if err != nil {
		return nil, err
	}
	//TODO: should receive notification when the coresponding response is matched

	return nil, nil
}
