package client

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperchain/hyperchain/common"
	pb "github.com/hyperchain/hyperchain/common/service/protos"
	"github.com/hyperchain/hyperchain/common/service/util"
	"github.com/op/go-logging"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	CONSENTER     = "consenter"
	APISERVER     = "apiserver"
	EXECUTOR      = "executor"
	NETWORK       = "network"
	ADMINISTRATOR = "administrator"
)

const maxReceiveBufferSize = 1024

// ServiceClient used to send messages to eventhub or receive message
// from the event hub.
type ServiceClient struct {
	host   string
	port   int
	sid    string // service id
	domain string // for admin is address, others is namespace

	msgRecv chan *pb.IMessage //received messages from server
	client  pb.Dispatcher_RegisterClient

	logger     *logging.Logger
	h          Handler
	closed     int32
	ctx        context.Context
	cf         context.CancelFunc
	rspAuxMap  map[uint64]chan *pb.IMessage
	rspAuxLock sync.RWMutex
	syncReqId  *util.ID
	rspAuxPool sync.Pool

	sync.RWMutex
}

func New(port int, host, sid, domain string) (*ServiceClient, error) {
	if len(host) == 0 || port < 0 {
		return nil, fmt.Errorf("Invalid host or port, %s:%d ", host, port)
	}
	logName := domain
	if strings.Contains(domain, ":") {
		logName = common.DEFAULT_LOG
	}
	sc := &ServiceClient{
		host:    host,
		port:    port,
		msgRecv: make(chan *pb.IMessage, maxReceiveBufferSize),
		logger:  common.GetLogger(logName, "service_client"),

		sid:    sid,
		domain: domain,
		closed: 0,

		rspAuxMap: make(map[uint64]chan *pb.IMessage),
		syncReqId: util.NewId(uint64(time.Now().Nanosecond())),
		rspAuxPool: sync.Pool{
			New: func() interface{} {
				return make(chan *pb.IMessage, 1)
			},
		},
	}
	sc.ctx, sc.cf = context.WithCancel(context.Background())

	return sc, nil
}

func (sc *ServiceClient) stream() pb.Dispatcher_RegisterClient {
	sc.RLock()
	defer sc.RUnlock()
	return sc.client
}

func (sc *ServiceClient) setStream(client pb.Dispatcher_RegisterClient) {
	sc.Lock()
	defer sc.Unlock()
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
	sc.logger.Debugf("%s connect successful", sc.string())
	sc.setStream(stream)
	sc.listenProcessMsg()
	return nil
}

func (sc *ServiceClient) reconnect() error {
	maxRetryT := 10 //TODO: make these params configurable
	sleepT := 1
	for i := 0; i < maxRetryT; i++ {
		err := sc.Connect()
		if err == nil {
			_, err = sc.Register(0, getFrom(sc.sid), sc.getRegMessage(sc.sid)) //TODO: Fix id
			if err != nil {
				sc.logger.Error(err)
				return err
			}
			return nil
		} else {
			d, _ := time.ParseDuration(fmt.Sprintf("%ds", sleepT))
			sc.logger.Noticef("Reconnect error %v, sleep %v then try to reconnect", err, d)
			time.Sleep(d)
			sleepT *= 2
		}
	}
	return fmt.Errorf("Recoonect error, exceed retry times: %d ", maxRetryT)
}

func (sc *ServiceClient) Register(cid uint64, serviceType pb.FROM, rm *pb.RegisterMessage) (*pb.IMessage, error) {
	payload, err := proto.Marshal(rm)
	if err != nil {
		return nil, err
	}

	if rsp, err := sc.SyncSend(&pb.IMessage{
		Type:    pb.Type_REGISTER,
		From:    serviceType,
		Payload: payload,
		Cid:     cid,
	}); err != nil {
		return nil, err
	} else {
		if rsp.Type == pb.Type_RESPONSE && rsp.Ok == true {
			sc.logger.Noticef("%s register successful", sc.string())
			return rsp, nil
		} else {
			return nil, fmt.Errorf("%s register failed", sc.string())
		}
	}
}

func (sc *ServiceClient) Close() {
	sc.Lock()
	defer sc.Unlock()
	atomic.StoreInt32(&sc.closed, 1)
	sc.cf()
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
	if sc.client == nil {
		sc.logger.Noticef("service client is nil, try to reconnect")
		sc.reconnect()
	}
	sc.Lock()
	err := sc.client.Send(msg)
	sc.Unlock()
	if err != nil {
		err = sc.reconnect()
		if err != nil {
			sc.Close()
			return err
		}
		sc.Lock()
		defer sc.Unlock()
		return sc.client.Send(msg)
	}
	return nil
}

func (sc *ServiceClient) SyncSend(msg *pb.IMessage) (*pb.IMessage, error) {
	if msg.Type != pb.Type_SYNC_REQUEST && msg.Type != pb.Type_REGISTER {
		return nil, fmt.Errorf("Invalid syncsend request type: %v ", msg.Type)
	}
	msg.Id = sc.syncReqId.IncAndGet()
	rspCh := sc.rspAuxPool.Get().(chan *pb.IMessage)

	sc.rspAuxLock.Lock()
	sc.rspAuxMap[msg.Id] = rspCh
	sc.rspAuxLock.Unlock()
	//TODO: add recovery if stream == nil or stream is closed
	sc.Lock()
	if err := sc.client.Send(msg); err != nil {
		sc.rspAuxLock.Lock()
		delete(sc.rspAuxMap, msg.Id)
		sc.rspAuxLock.Unlock()
		sc.Unlock()
		return nil, err
	}
	//TODO: add timeout detection
	sc.Unlock()
	rsp := <-rspCh
	return rsp, nil
}

//AddHandler add self defined message handler.
func (sc *ServiceClient) AddHandler(h Handler) {
	sc.h = h
}

func (sc *ServiceClient) listenProcessMsg() {
	go func() {
		for {
			if atomic.LoadInt32(&sc.closed) == 1 {
				return
			}
			msg, err := sc.stream().Recv() //TODO: how to get latest client, if reconnect happened

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
			//sc.logger.Debugf("receive msg %v ", msg)
			if msg != nil {
				if msg.Type == pb.Type_RESPONSE {
					if msg.Id == 0 {
						sc.logger.Errorf("Response id is not true")
					}

					sc.rspAuxLock.RLock()
					if ch, ok := sc.rspAuxMap[msg.Id]; ok {
						ch <- msg
					} else {
						sc.logger.Errorf("No response channel found for %d", msg.Id)
					}

					sc.rspAuxLock.RUnlock()
				} else {
					sc.msgRecv <- msg
				}
			}
		}
	}()

	sc.logger.Debug("Start Message processing goroutine")

	go func() {
		for {
			select {
			case msg := <-sc.msgRecv:
				//sc.logger.Infof("handle receive msg: %v ", msg)
				if sc.h == nil {
					sc.logger.Debugf("No handler to handle message: %v", msg)
				} else {
					sc.h.Handle(sc.client, msg)
				}
			case <-sc.ctx.Done():
				return
			}
		}
	}()
}

//string service client description
func (sc *ServiceClient) string() string {
	return fmt.Sprintf("ServiceClient[namespace: %s, serviceId: %s]", sc.domain, sc.sid)
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
	case ADMINISTRATOR:
		return pb.FROM_ADMINISTRATOR

	}
	return -1
}

func (sc *ServiceClient) getRegMessage(sid string) *pb.RegisterMessage {
	switch sid {
	case CONSENTER:
		fallthrough
	case NETWORK:
		fallthrough
	case EXECUTOR:
		fallthrough
	case APISERVER:
		return &pb.RegisterMessage{
			Namespace: sc.domain,
		}
	case ADMINISTRATOR:
		return &pb.RegisterMessage{
			Address: sc.domain,
		}

	}
	return nil
}
