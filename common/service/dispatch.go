package service

import (
	pb "hyperchain/common/protos"
)

func (ds *DispatchServer) dispatchConsensusMsg(namespace string, msg *pb.Message) {
	ds.logger.Debugf("dispatch consensus message: %v for namespace: %s", msg, namespace)

	if !ds.sr.ContainsNamespace(namespace) {
		ds.logger.Errorf("no namespace found [%s]", namespace)
		return
	}

	switch msg.Event {
	case pb.Event_InformPrimaryEvent:
		service := ds.sr.Namespace(namespace).Service(NETWORK)
		if service != nil {
			service.Send(false, msg)
		}else {
			ds.logger.Errorf("No service found for [%s]", NETWORK)
		}
	}
}

func (ds *DispatchServer) dispatchExecutorMsg(namespace string, msg *pb.Message) {

}

func (ds *DispatchServer) dispatchNetworkMsg(namespace string, msg *pb.Message) {

}

func (ds *DispatchServer) dispatchAPIServerMsg(namespace string, msg *pb.Message) {

}
