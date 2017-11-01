package server

import (
	"fmt"
	"github.com/op/go-logging"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/common/service/util"
	"sync"
	"time"
)

//remoteServiceImpl represent a remote service.
type remoteServiceImpl struct {
	ds        *InternalServer
	namespace string
	id        string
	stream    pb.Dispatcher_RegisterServer
	r         chan *pb.IMessage
	logger    *logging.Logger
	syncReqId *util.ID

	rspAuxMap  map[uint64]chan *pb.IMessage
	rspAuxLock sync.RWMutex
	rspAuxPool sync.Pool
	close      chan struct{}
}

func NewRemoteService(namespace, id string, stream pb.Dispatcher_RegisterServer, ds *InternalServer) service.Service {
	return &remoteServiceImpl{
		namespace: namespace,
		id:        id,
		stream:    stream,
		logger:    logging.MustGetLogger("service"),
		ds:        ds,
		r:         make(chan *pb.IMessage),
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
		case pb.Type_DISPATCH:
			rsi.ds.HandleDispatch(rsi.namespace, msg)
		case pb.Type_ADMIN:
			rsi.ds.HandleAdmin(rsi.namespace, msg)
		case pb.Type_SYNC_REQUEST:
			rsi.ds.HandleSyncRequest(rsi.namespace, msg)
		case pb.Type_RESPONSE:
			rsi.r <- msg
		}
		//rsi.logger.Debugf("%s, %s service serve", rsi.namespace, rsi.id)
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
