package server

import (
	"github.com/gogo/protobuf/proto"
	pb "hyperchain/common/protos"
	"hyperchain/common/service"
	"hyperchain/manager/event"
)

func (is *InternalServer) DispatchConsensusMsg(namespace string, msg *pb.IMessage) {
	is.logger.Debugf("dispatch consensus message: %v for namespace: %s", msg, namespace)

	if !is.sr.ContainsNamespace(namespace) {
		is.logger.Errorf("no namespace found [%s]", namespace)
		return
	}

	switch msg.Event {
	case pb.Event_InformPrimaryEvent:
		service := is.sr.Namespace(namespace).Service(service.NETWORK)
		if service != nil {
			service.Send(msg)
		} else {
			//is.logger.Errorf("No service found for [%s]", service.NETWORK)
		}
	}
}

func (is *InternalServer) DispatchExecutorMsg(namespace string, msg *pb.IMessage) {

	is.logger.Debugf("dispatch executor message: %v for namespace: %s", msg, namespace)

	if !is.sr.ContainsNamespace(namespace) {
		is.logger.Errorf("no namespace found [%s]", namespace)
		return
	}
	is.logger.Debugf("receive event: %v", msg)
	switch msg.Event {
	//TODO: check whether we have more other messages
	case pb.Event_ExecutorToConsensusEvent:
		e := &event.ExecutorToConsensusEvent{}
		if err := proto.Unmarshal(msg.Payload, e); err != nil {
			is.logger.Error(err)
			return
		}
		is.sr.Namespace(namespace).Service(service.EVENTHUB).Send(e) //send this message to event hub

	case pb.Event_ExecutorToP2PEvent:
		e := &event.ExecutorToP2PEvent{}
		if err := proto.Unmarshal(msg.Payload, e); err != nil {
			is.logger.Error(err)
			return
		}
		is.sr.Namespace(namespace).Service(service.EVENTHUB).Send(e)
	}
}

func (is *InternalServer) DispatchNetworkMsg(namespace string, msg *pb.IMessage) {

}

func (is *InternalServer) DispatchAPIServerMsg(namespace string, msg *pb.IMessage) {

}
