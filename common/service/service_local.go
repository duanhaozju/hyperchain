package service

import (
	pb "hyperchain/common/protos"
	"hyperchain/manager/event"
)

type localServiceImpl struct {
	eventMux  *event.TypeMux // used to post message to event hub
	namespace string
	id        string
	r         chan *pb.Message
}

func NewLocalService(namespace, id string, em *event.TypeMux) Service {
	return &localServiceImpl{
		id:        id,
		namespace: namespace,
		eventMux:  em,
		r:         make(chan *pb.Message),
	}
}

func (lsi *localServiceImpl) Namespace() string {
	return lsi.namespace
}

func (lsi *localServiceImpl) Id() string {
	return lsi.id
}

func (lsi *localServiceImpl) Send(event interface{}) error {
	lsi.eventMux.Post(event)
	return nil
}

func (lsi *localServiceImpl) Close() {

}

func (lsi *localServiceImpl) Serve() error {
	return nil
}

func (lsi *localServiceImpl) isHealth() bool {
	return true
}

func (lsi *localServiceImpl) Response() chan *pb.Message {
	return lsi.r
}
