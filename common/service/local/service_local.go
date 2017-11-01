package local

import (
	//"fmt"
	pb "hyperchain/common/protos"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"fmt"
	"hyperchain/common/service"
    "hyperchain/common"
)

type localServiceImpl struct {
	namespace string
	id        string
	r         chan *pb.IMessage
	hub       *manager.EventHub
}

func NewLocalService(namespace, id string, hub *manager.EventHub) service.Service {
	return &localServiceImpl{
		id:        id,
		namespace: namespace,
		r:         make(chan *pb.IMessage),
		hub:       hub,
	}
}

func (lsi *localServiceImpl) Namespace() string {
	return lsi.namespace
}

func (lsi *localServiceImpl) Id() string {
	return lsi.id
}

func (lsi *localServiceImpl) Send(se service.ServiceEvent) error {
    logger := common.GetLogger("global", "localservice")
    //logger.Criticalf("send %v", se)
	switch e := se.(type) {
	case *event.ExecutorToConsensusEvent:
	    //logger.Criticalf("Send event: %v", e)
		lsi.hub.DispatchExecutorToConsensus(*e)
		return nil
	case *event.ExecutorToP2PEvent:
		lsi.hub.DispatchExecutorToP2P(*e)
		return nil
	default:
	    logger.Criticalf("Send default %v", e)
		return fmt.Errorf("no event handler found for %v", se)
	}
}

func (lsi *localServiceImpl) Close() {

}

func (lsi *localServiceImpl) Serve() error {
	return nil
}

func (lsi *localServiceImpl) IsHealth() bool {
	return true
}

func (lsi *localServiceImpl) Response() chan *pb.IMessage {
	return lsi.r
}

func (lsi *localServiceImpl) SyncSend(se service.ServiceEvent) (*pb.IMessage, error)  {

	return nil, nil
}