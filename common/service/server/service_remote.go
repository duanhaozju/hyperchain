package server

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common/service"
	pb "github.com/hyperchain/hyperchain/common/service/protos"
	"github.com/hyperchain/hyperchain/common/service/util"
	"github.com/op/go-logging"
	"sync"
	"time"
)

//remoteServiceImpl represent a remote service.
type remoteServiceImpl struct {
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
	r         chan *pb.IMessage
	logger    *logging.Logger
	syncReqId *util.ID

	msg chan *pb.IMessage // cache to store remote peer's messages.

	rspAuxMap  map[uint64]chan *pb.IMessage
	rspAuxLock sync.RWMutex
	rspAuxPool sync.Pool
	close      chan struct{}

	sync.Mutex
}

func NewRemoteService(namespace, id string, stream pb.Dispatcher_RegisterServer) service.Service {
	return &remoteServiceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
		logger:    logging.MustGetLogger("service"),
		r:         make(chan *pb.IMessage),
		msg:       make(chan *pb.IMessage),
		syncReqId: util.NewId(uint64(time.Now().UnixNano())),
		rspAuxMap: make(map[uint64]chan *pb.IMessage),
		rspAuxPool: sync.Pool{
			New: func() interface{} {
				return make(chan *pb.IMessage, 1)
			},
		},
	}
}

func (rsi remoteServiceImpl) Namespace() string {
	return rsi.namespace
}

// Id service identifier.
func (rsi *remoteServiceImpl) Id() string {
	return rsi.id
}

// Send sync send msg.
func (rsi *remoteServiceImpl) Send(event service.ServiceEvent) error {
	if msg, ok := event.(*pb.IMessage); !ok {
		return fmt.Errorf("send message type error, %v need pb.IMessage ", event)
	} else {
		if msg.Type == pb.Type_SYNC_REQUEST {
			return fmt.Errorf("Sync type event should be sent by SyncSend ")
		}

		if rsi.stream == nil {
			return fmt.Errorf("[%s:%s]stream is empty, wait for this component to reconnect", rsi.namespace, rsi.id)
		}
		rsi.Lock()
		defer rsi.Unlock()
		return rsi.stream.Send(msg)
	}
}

func (rsi *remoteServiceImpl) Close() {
	rsi.close <- struct{}{}
}

//Serve handle logic impl here.
func (rsi *remoteServiceImpl) Serve() error {
	//dispatch responses
	go rsi.dispatchResponse()
	for {
		msg, err := rsi.stream.Recv()
		if err != nil {
			return err
		}
		switch msg.Type {
		case pb.Type_REGISTER:
			rsi.logger.Errorf("No register message should be here! msg: %v", msg)
		case pb.Type_NORMAL:
			rsi.msg <- msg
		case pb.Type_SYNC_REQUEST:
			rsi.msg <- msg
		case pb.Type_RESPONSE:
			rsi.r <- msg
		default:
			rsi.logger.Errorf("Invalid message type, %v", msg.Type)
		}
	}
}

func (rsi *remoteServiceImpl) IsHealth() bool {
	//TODO: more to check
	return rsi.stream != nil
}

func (rsi *remoteServiceImpl) SyncSend(se service.ServiceEvent) (*pb.IMessage, error) {
	if msg, ok := se.(*pb.IMessage); !ok {
		return nil, fmt.Errorf("send message type error, %v need pb.IMessage ", se)
	} else {
		if rsi.stream == nil {
			return nil, fmt.Errorf("[%s:%s]stream is empty, wait for this component to reconnect", rsi.namespace, rsi.id)
		}
		//return rsi.stream.Send(se)
		if msg.Type != pb.Type_SYNC_REQUEST {
			return nil, fmt.Errorf("Invalid syncsend request type: %v ", msg.Type)
		}
		msg.Id = rsi.syncReqId.IncAndGet()
		rspCh := rsi.rspAuxPool.Get().(chan *pb.IMessage)

		rsi.rspAuxLock.Lock()
		rsi.rspAuxMap[msg.Id] = rspCh
		rsi.rspAuxLock.Unlock()
		//TODO: add recovery if stream == nil or stream is closed
		rsi.Lock()
		defer rsi.Unlock()
		if err := rsi.stream.Send(msg); err != nil {
			rsi.rspAuxLock.Lock()
			delete(rsi.rspAuxMap, msg.Id)
			rsi.rspAuxLock.Unlock()
			return nil, err
		}
		//TODO: add timeout detection
		rsp := <-rspCh
		return rsp, nil
	}
	return nil, nil
}

func (rsi *remoteServiceImpl) dispatchResponse() {
	for {
		select {
		case r := <-rsi.r:
			if r.Id == 0 {
				rsi.logger.Errorf("Response id is not true")
			}

			rsi.rspAuxLock.RLock()
			if ch, ok := rsi.rspAuxMap[r.Id]; ok {
				ch <- r
			} else {
				rsi.logger.Errorf("No response channel found for %d", r.Id)
			}

			rsi.rspAuxLock.RUnlock()
		case <-rsi.close:
			rsi.logger.Infof("%s dispatch response thread is close", rsi.String())
		}
	}
}

func (rsi *remoteServiceImpl) String() string {
	return fmt.Sprintf("RemoteService [%s:%s] ", rsi.namespace, rsi.id)
}

func (rsi *remoteServiceImpl) Receive() chan *pb.IMessage {
	return rsi.msg
}
