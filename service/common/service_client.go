package common

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/service/common/protos"
	"sync"
	"time"
)

// ServiceClient used to send messages to eventhub or receive message
// from the event hub.
type ServiceClient struct {
	host   string
	port   int
	msgs   chan *pb.Message
	slock  sync.RWMutex
	client pb.Dispatcher_RegisterClient
	logger *logging.Logger
	h      Handler
	sid    string // service id
	ns     string // namespace
}

func New(port int, host, sid, ns string) (*ServiceClient, error) {
	if len(host) == 0 || port < 0 {
		return nil, fmt.Errorf("Invalid host or port, %s:%d", host, port)
	}
	return &ServiceClient{
		host:   host,
		port:   port,
		msgs:   make(chan *pb.Message, 1024),
		logger: logging.MustGetLogger("service_client"), // TODO: replace this logger with hyperlogger
		sid:    sid,
		ns:     ns,
	}, nil
}

func (sc *ServiceClient) stream() pb.Dispatcher_RegisterClient {
	sc.slock.RLock()
	defer sc.slock.RUnlock()
	return sc.client
}

func (sc *ServiceClient) setStream(client pb.Dispatcher_RegisterClient) {
	sc.slock.Lock()
	defer sc.slock.Unlock()
	sc.client = client
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
	sc.logger.Debug("service client connect successful")
	sc.setStream(stream)
	return nil
}

func (sc *ServiceClient) reconnect() error {
	maxRetryT := 10
	sleepT := 1
	for i := 0; i < maxRetryT; i++ {
		err := sc.Connect()
		if err == nil {
			err = sc.Register(getFrom(sc.sid), &pb.RegisterMessage{
				Namespace: sc.ns,
			})
			if err != nil {
				sc.logger.Error(err)
				return err
			}
			return nil
		} else {
			d, _ := time.ParseDuration(fmt.Sprintf("%ds", sleepT))
			sc.logger.Debugf("Sleep %v then try to reconnect", d)
			time.Sleep(d)
			sleepT *= 2
		}
	}
	return nil
}

func (sc *ServiceClient) Register(serviceType pb.FROM, rm *pb.RegisterMessage) error {
	sc.slock.RLock()
	defer sc.slock.RUnlock()
	payload, err := proto.Marshal(rm)
	if err != nil {
		return err
	}
	if err = sc.stream().Send(&pb.Message{
		Type:    pb.Type_REGISTER,
		From:    serviceType,
		Payload: payload,
	}); err != nil {
		return err
	}

	sc.logger.Debug("try to wait the register response")

	//timeout detection
	msg, err := sc.stream().Recv()
	if err != nil {
		return err
	}

	sc.logger.Debug("register successful")

	if msg.Type == pb.Type_RESPONSE && msg.Ok == true {
		sc.logger.Infof("Service %v in namespace %v register successful", serviceType, rm.Namespace)
		return nil
	} else {
		return fmt.Errorf("Service %v in namespace %v register failed", serviceType, rm.Namespace)
	}
}

//Send msg asynchronous
func (sc *ServiceClient) Send(msg *pb.Message) error {
	if sc.stream == nil {
		sc.reconnect()
	}
	err := sc.stream().Send(msg)
	if err != nil {
		sc.reconnect()
		return sc.stream().Send(msg)
	}
	return nil
}

//AddHandler add self defined message handler.
func (sc *ServiceClient) AddHandler(h Handler) {
	sc.h = h
}

func (sc *ServiceClient) ProcessMessagesUntilExit() {
	exit := make(chan struct{})
	go func() {
		for {
			msg, err := sc.stream().Recv()
			if err != nil {
				sc.logger.Error(err)
				err = sc.reconnect()
				if err != nil {
					sc.logger.Errorf("Reconnect failed, %v", err)
					exit <- struct{}{}
					return
				}
			}
			if msg != nil {
				//sc.logger.Debugf("Receive message: %v", msg)
				sc.msgs <- msg
			}
		}
	}()

	sc.logger.Debug("Start Message processing go routine")
	for {
		select {
		case msg := <-sc.msgs:
			if sc.h == nil {
				sc.logger.Debug("No handler to handle message: %v")
			} else {
				sc.h.Handle(msg)
			}
		case <-exit:
			return
		}
	}
}

func getFrom(sid string) pb.FROM {
	switch sid {
	case CONSENTER:
		return pb.FROM_CONSENSUS
	case NETWORK:
		return pb.FROM_NETWORK
	case EXECUTOR:
		return pb.FROM_EXECUTOR
	case APISERVER:
		return pb.FROM_APISERVER
	}
	return -1
}
