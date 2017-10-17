package service

import (
	pb "hyperchain/common/protos"
)

func (ds *InternalServer) dispatchConsensusMsg(namespace string, msg *pb.Message) {
	ds.logger.Debugf("dispatch consensus message: %v for namespace: %s", msg, namespace)

	if !ds.sr.ContainsNamespace(namespace) {
		ds.logger.Errorf("no namespace found [%s]", namespace)
		return
	}

	switch msg.Event {
	case pb.Event_InformPrimaryEvent:
		service := ds.sr.Namespace(namespace).Service(NETWORK)
		if service != nil {
			service.Send(msg)
		} else {
			ds.logger.Errorf("No service found for [%s]", NETWORK)
		}
	}
}

func (ds *InternalServer) dispatchExecutorMsg(namespace string, msg *pb.Message) {
	//TODO: 1.unmarshal specific event
	//TODO: 2.find correct service
	//TODO: 3.send the message
}

func (ds *InternalServer) dispatchNetworkMsg(namespace string, msg *pb.Message) {

}

func (ds *InternalServer) dispatchAPIServerMsg(namespace string, msg *pb.Message) {

}
