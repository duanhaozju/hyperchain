package service

import (
	pb "hyperchain/common/protos"
)

func (is *InternalServer) dispatchConsensusMsg(namespace string, msg *pb.IMessage) {
	is.logger.Debugf("dispatch consensus message: %v for namespace: %s", msg, namespace)

	if !is.sr.ContainsNamespace(namespace) {
		is.logger.Errorf("no namespace found [%s]", namespace)
		return
	}

	switch msg.Event {
	case pb.Event_InformPrimaryEvent:
		service := is.sr.Namespace(namespace).Service(NETWORK)
		if service != nil {
			service.Send(msg)
		} else {
			is.logger.Errorf("No service found for [%s]", NETWORK)
		}
	}
}

func (is *InternalServer) dispatchExecutorMsg(namespace string, msg *pb.IMessage) {

	is.logger.Debugf("dispatch executor message: %v for namespace: %s", msg, namespace)

	if !is.sr.ContainsNamespace(namespace) {
		is.logger.Errorf("no namespace found [%s]", namespace)
		return
	}

	switch msg.Event {

	}
	//TODO: 1.unmarshal specific event
	//TODO: 2.find correct service
	//TODO: 3.send the message
}

func (is *InternalServer) dispatchNetworkMsg(namespace string, msg *pb.IMessage) {

}

func (is *InternalServer) dispatchAPIServerMsg(namespace string, msg *pb.IMessage) {

}
