package service

import (
	pb "hyperchain/common/protos"
	"hyperchain/manager/event"
)

type localServiceImpl struct {
	eventMux  *event.TypeMux // used to post message to event hub
	namespace string
	id        string
	r         chan *pb.IMessage
}

func NewLocalService(namespace, id string, em *event.TypeMux) Service {
	return &localServiceImpl{
		id:        id,
		namespace: namespace,
		eventMux:  em,
		r:         make(chan *pb.IMessage),
	}
}

func (lsi *localServiceImpl) Namespace() string {
	return lsi.namespace
}

func (lsi *localServiceImpl) Id() string {
	return lsi.id
}

func (lsi *localServiceImpl) Send(event interface{}) error {
	return lsi.eventMux.Post(event)
}

func (lsi *localServiceImpl) Close() {

}

func (lsi *localServiceImpl) Serve() error {
	return nil
}

func (lsi *localServiceImpl) isHealth() bool {
	return true
}

func (lsi *localServiceImpl) Response() chan *pb.IMessage {
	return lsi.r
}
