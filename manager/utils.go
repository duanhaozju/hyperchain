package manager

import (
	"hyperchain/consensus/pbft"
	"hyperchain/recovery"
	"hyperchain/event"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
)

func (hub *EventHub) invokePbftLocal(serviceType, eventType int, content interface{}) {
	e := &pbft.LocalEvent{
		Service:   serviceType,
		EventType: eventType,
		Event:     content,
	}
	hub.consenter.RecvLocal(e)
}


func (hub *EventHub) dispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_REMOVE_CACHE:
		log.Debugf("[Namespace = %s] message middleware: [remove cache]", hub.namespace)
		hub.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_VC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [vc done]", hub.namespace)
		hub.invokePbftLocal(pbft.VIEW_CHANGE_SERVICE, pbft.VIEW_CHANGE_VC_RESET_DONE_EVENT, ev.Payload)
	case executor.NOTIFY_VALIDATION_RES:
		log.Debugf("[Namespace = %s] message middleware: [validation result]", hub.namespace)
		hub.invokePbftLocal(pbft.CORE_PBFT_SERVICE, pbft.CORE_VALIDATED_TXS_EVENT, ev.Payload)
	case executor.NOTIFY_SYNC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [sync done]", hub.namespace)
		hub.consenter.RecvMsg(ev.Payload.([]byte))
	}
}
func (hub *EventHub) dispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	switch ev.Type {
	case executor.NOTIFY_BROADCAST_DEMAND:
		log.Debugf("[Namespace = %s] message middleware: [broadcast demand]", hub.namespace)
		hub.peerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCCHECKPOINT)
	case executor.NOTIFY_UNICAST_INVALID:
		log.Debugf("[Namespace = %s] message middleware: [unicast invalid tx]", hub.namespace)
		peerId := ev.Peers[0]
		if peerId == uint64(hub.peerManager.GetNodeId()) {
			hub.executor.StoreInvalidTransaction(event.InvalidTxsEvent{
				Payload: ev.Payload,
			})
		} else {
			hub.peerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_INVALIDRESP)
		}
	case executor.NOTIFY_BROADCAST_SINGLE:
		log.Debugf("[Namespace = %s] message middleware: [broadcast single]", hub.namespace)
		hub.peerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCSINGLE)
	case executor.NOTIFY_UNICAST_BLOCK:
		log.Debugf("[Namespace = %s] message middleware: [unicast block]", hub.namespace)
		hub.peerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCBLOCK)
	case executor.NOTIFY_SYNC_REPLICA:
		log.Debugf("[Namespace = %s] message middleware: [sync replica]", hub.namespace)
		chain := &types.Chain{}
		proto.Unmarshal(ev.Payload, chain)
		addr := hub.peerManager.GetLocalNode().GetNodeAddr()
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain:     chain,
			Ip:        []byte(addr.IP),
			Port:      int32(addr.Port),
			Namespace: []byte(hub.namespace),
		})
		peers := hub.peerManager.GetVPPeers()
		var peerIds = make([]uint64, len(peers))
		for idx, peer := range peers {
			peerIds[idx] = uint64(peer.PeerAddr.ID)
		}
		hub.peerManager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
		hub.eventMux.Post(event.ReplicaInfoEvent{
			Payload: payload,
		})
	}
}
