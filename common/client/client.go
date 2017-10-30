package client

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	pb "hyperchain/common/protos"
	"sync"
	"sync/atomic"
	"time"
    "hyperchain/core/ledger/chain"
    "hyperchain/manager/event"
)

const (
    CONSENTER = "consenter"
    APISERVER = "apiserver"
    EXECUTOR  = "executor"
    NETWORK   = "network"
    EVENTHUB  = "eventhub"
)

// ServiceClient used to send messages to eventhub or receive message
// from the event hub.
type ServiceClient struct {
	host string
	port int
	sid  string // service id
	ns   string // namespace

	msgRecv		chan *pb.IMessage //received messages from server
	msgSend		chan *pb.IMessage //send message to server
	slock  sync.RWMutex
	client pb.Dispatcher_RegisterClient

	logger *logging.Logger
	h      Handler
	close  chan struct{}
	closed int32
}

func New(port int, host, sid, ns string) (*ServiceClient, error) {
	if len(host) == 0 || port < 0 {
		return nil, fmt.Errorf("Invalid host or port, %s:%d ", host, port)
	}
	return &ServiceClient{
		host:   host,
		port:   port,
		msgRecv:   make(chan *pb.IMessage, 1024),
		msgSend:   make(chan *pb.IMessage, 1024),
		logger: logging.MustGetLogger("service_client"),
		// TODO: replace this logger with hyperlogger ?
		sid:    sid,
		ns:     ns,
		close:  make(chan struct{}, 2),
		closed: 0,
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
	sc.logger.Debug("%s connect successful", sc.string())
	sc.setStream(stream)
	return nil
}

func (sc *ServiceClient) reconnect() error {
	maxRetryT := 10 //TODO: make these params configurable
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
			sc.logger.Debugf("Reconnect error %v, sleep %v then try to reconnect", err, d)
			time.Sleep(d)
			sleepT *= 2
		}
	}
	return fmt.Errorf("Recoonect error, exceed retry times: %d ", maxRetryT)
}

func (sc *ServiceClient) Register(serviceType pb.FROM, rm *pb.RegisterMessage) error {
	sc.slock.RLock()
	defer sc.slock.RUnlock()
	payload, err := proto.Marshal(rm)
	if err != nil {
		return err
	}
	if err = sc.stream().Send(&pb.IMessage{
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

	sc.logger.Debugf("%s connect successful", sc.string())

	if msg.Type == pb.Type_RESPONSE && msg.Ok == true {
		sc.logger.Infof("%s register successful", sc.string())
		sc.listenProcessMsg()
		return nil
	} else {
		return fmt.Errorf("%s register failed", sc.string())
	}
}

func (sc *ServiceClient) Close() {
	atomic.StoreInt32(&sc.closed, 1)
	sc.close <- struct{}{} //close receive goroutine
	sc.close <- struct{}{} //close message process goroutine
	if sc.client != nil {
		if err := sc.client.CloseSend(); err != nil {
			sc.logger.Error(err)
		}
	}
	sc.logger.Criticalf("%s closed!", sc.string())
}

func (sc *ServiceClient) isClosed() bool {
	return atomic.LoadInt32(&sc.closed) == 1
}

//Send msg asynchronous
func (sc *ServiceClient) Send(msg *pb.IMessage) error {
	//TODO: Add msg format check
	if sc.stream == nil {
		sc.reconnect()
	}
	err := sc.stream().Send(msg)
	if err != nil {
		err = sc.reconnect()
		if err != nil {
			sc.Close()
			return err
		}
		return sc.stream().Send(msg)
	}
	return nil
}

//AddHandler add self defined message handler.
func (sc *ServiceClient) AddHandler(h Handler) {
	sc.h = h
}

func (sc *ServiceClient) listenProcessMsg() {
	go func() {
		for {
			msg, err := sc.stream().Recv()
			if err != nil {
				sc.logger.Error(err)
				if !sc.isClosed() {
					err = sc.reconnect()
					if err != nil {
						sc.logger.Errorf("Reconnect failed, %v", err)
						sc.Close()
						return
					}
				}
			}
			if msg != nil {
				sc.msgRecv <- msg
            }
		}
	}()

	sc.logger.Debug("Start Message processing go routine")

	go func() {
		for {
			select {
			case msg := <-sc.msgRecv:
				if sc.h == nil {
					sc.logger.Debugf("No handler to handle message: %v", msg)
				} else {
                    if msg.Type == pb.Type_SYNC_REQUEST {
                        e := &event.MemChainEvent{}
                        err := proto.Unmarshal(msg.Payload, e)
                        if err != nil {
                            sc.logger.Criticalf("MemChainEvent unmarshal err: %v", err)
                        }
                        go func() {
                            m:= chain.GetMemChain(e.Namespace, e.Checkpoint)
                            payload, err := proto.Marshal(m)
                            if err != nil {
                                sc.logger.Error(err)
                                return
                            }
                            msg := &pb.IMessage{
                                Type:  pb.Type_RESPONSE,
                                From:  pb.FROM_EXECUTOR,
                                Ok: true,
                                Payload: payload,
                            }
                            sc.client.Send(msg)
                        }()
                    } else {
                        sc.h.Handle(msg)
                    }
				}
			case <-sc.close:
				return
			}
		}
	}()
}

//string service client description
func (sc *ServiceClient) string() string {
	return fmt.Sprintf("ServiceClient[namespace: %s, serviceId: %s]", sc.ns, sc.sid)
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

