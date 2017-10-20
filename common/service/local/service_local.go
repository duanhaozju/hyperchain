package local

import (
	//"fmt"
	pb "hyperchain/common/protos"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"fmt"
	"hyperchain/common/service"
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
	switch e := se.(type) {
	case event.ExecutorToConsensusEvent:
		lsi.hub.DispatchExecutorToConsensus(e)
		return nil
	case event.ExecutorToP2PEvent:
		lsi.hub.DispatchExecutorToP2P(e)
		return nil
	default:
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
