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

func (lsi *localServiceImpl) Namespace() string {
	return lsi.namespace
}

func (lsi *localServiceImpl) Id() string {
	return lsi.id
}

func (lsi *localServiceImpl) Send(event interface{}) {
	lsi.eventMux.Post(event)
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
