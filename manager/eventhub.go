//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/consensus"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"time"
	"hyperchain/admittance"
	"hyperchain/consensus/pbft"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("eventhub")
}

type EventHub struct {
	namespace       string
	executor        *executor.Executor
	PeerManager     p2p.PeerManager
	nodeInfo        p2p.PeerInfos // node info ,store node status,ip,port
	consenter       consensus.Consenter
	AccountManager  *accounts.AccountManager
	eventMux        *event.TypeMux
	validateSub     event.Subscription
	commitSub       event.Subscription
	consensusSub    event.Subscription
	viewChangeSub   event.Subscription
	invalidSub      event.Subscription
	syncChainSub    event.Subscription
	syncBlockSub    event.Subscription
	syncReplicaSub  event.Subscription
	peerMaintainSub event.Subscription
	executorSub     event.Subscription
	initType        int
}
type NodeManager struct {
	peerManager p2p.PeerManager
}

var eventMuxAll *event.TypeMux

func NewEventHub(namespace string, executor *executor.Executor, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
	am *accounts.AccountManager) *EventHub {
	manager := &EventHub{
		namespace:          namespace,
		executor:           executor,
		eventMux:            eventMux,
		consenter:           consenter,
		PeerManager:         peerManager,
		AccountManager:      am,
	}
	manager.nodeInfo = make(p2p.PeerInfos, 0, 1000)
	eventMuxAll = eventMux
	return manager
}
func (hub *EventHub) GetEventObject() *event.TypeMux {
	return hub.eventMux
}
func GetEventObject() *event.TypeMux {
	return eventMuxAll
}

// start listen new block msg and consensus msg
func (hub *EventHub) Start(c chan int, cm *admittance.CAManager) {
	hub.consensusSub = hub.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{}, event.NewTxEvent{},
		event.NegoRoutersEvent{})
	hub.validateSub = hub.eventMux.Subscribe(event.ValidationEvent{})
	hub.commitSub = hub.eventMux.Subscribe(event.CommitEvent{})
	hub.syncChainSub = hub.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{}, event.ReceiveSyncBlockEvent{})
	hub.invalidSub = hub.eventMux.Subscribe(event.InvalidTxsEvent{})
	hub.viewChangeSub = hub.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	hub.peerMaintainSub = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.RecvNewPeerEvent{},
		event.DelPeerEvent{}, event.BroadcastDelPeerEvent{}, event.RecvDelPeerEvent{})
	hub.executorSub = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})
	hub.syncReplicaSub = hub.eventMux.Subscribe(event.ReplicaInfoEvent{})

	go hub.listenValidateEvent()
	go hub.listenCommitEvent()
	go hub.listenConsensusEvent()
	go hub.ListenSynchronizationEvent()
	go hub.listenExecutorEvent()
	go hub.listenInvalidTxEvent()
	go hub.listenViewChangeEvent()
	go hub.listenReplicaInfoEvent()
	go hub.listenpeerMaintainEvent()

	go hub.PeerManager.Start(c, hub.eventMux, cm)
	hub.initType = <- c
	if hub.initType == 0 {
		// start in normal mode
		hub.PassRouters()
		hub.NegotiateView()
	}
}

func (hub *EventHub) ListenSynchronizationEvent() {
	for obj := range hub.syncChainSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.SendCheckpointSyncEvent:
			hub.executor.SendSyncRequest(ev)

		case event.StateUpdateEvent:
			hub.executor.ReceiveSyncRequest(ev)

		case event.ReceiveSyncBlockEvent:
			hub.executor.ReceiveSyncBlocks(ev)
		}
	}
}

// listen validate msg
func (hub *EventHub) listenValidateEvent() {
	for obj := range hub.validateSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ValidationEvent:
			log.Debugf("[Namespace = %s] message middleware: [mit]", hub.namespace)
			hub.executor.Validate(ev, hub.PeerManager)
		}
	}
}

// listen commit msg
func (hub *EventHub) listenCommitEvent() {
	for obj := range hub.commitSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.CommitEvent:
			log.Debugf("[Namespace = %s] message middleware: [commit]", hub.namespace)
			hub.executor.CommitBlock(ev, hub.PeerManager)
		}
	}
}

func (hub *EventHub) listenInvalidTxEvent() {
	for obj := range hub.invalidSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.InvalidTxsEvent:
			log.Debugf("[Namespace = %s] message middleware: [invalid tx]", hub.namespace)
			hub.executor.StoreInvalidTransaction(ev)
		}
	}
}

func (hub *EventHub) listenViewChangeEvent() {
	for obj := range hub.viewChangeSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			log.Debugf("[Namespace = %s] message middleware: [vc reset]", hub.namespace)
			hub.executor.Rollback(ev)
		case event.InformPrimaryEvent:
			log.Debugf("[Namespace = %s] message middleware: [inform primary]", hub.namespace)
			hub.PeerManager.SetPrimary(ev.Primary)
		}
	}
}

func (hub *EventHub) listenReplicaInfoEvent() {
	for obj := range hub.syncReplicaSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ReplicaInfoEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive replica info]", hub.namespace)
			hub.executor.ReceiveReplicaInfo(ev)
		}
	}
}

func (hub *EventHub) listenConsensusEvent() {
	for obj := range hub.consensusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.BroadcastConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [broadcast consensus]", hub.namespace)
			go hub.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:
			log.Debugf("[Namespace = %s] message middleware: [tx unicast]", hub.namespace)
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go hub.PeerManager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
		case event.NewTxEvent:
			log.Debugf("[Namespace = %s] message middleware: [new tx]", hub.namespace)
			if ev.Simulate == true {
				tx := &types.Transaction{}
				proto.Unmarshal(ev.Payload, tx)
				hub.executor.RunInSandBox(tx)
			} else {
				go hub.sendMsg(ev.Payload)
			}
		case event.ConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive consensus]", hub.namespace)
			hub.consenter.RecvMsg(ev.Payload)
		case event.NegoRoutersEvent:
			log.Debugf("[Namespace = %s] message middleware: [negotiate routers]", hub.namespace)
			hub.PeerManager.UpdateAllRoutingTable(ev.Payload)
		}
	}
}

func (hub *EventHub) listenpeerMaintainEvent() {

	for obj := range hub.peerMaintainSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.NewPeerEvent:
			log.Debug("NewPeerEvent")
			// a new peer required to join the network and past the local CA validation
			// payload is the new peer's address information
			msg := &protos.AddNodeMessage{
				Payload: ev.Payload,
			}
			e := &pbft.LocalEvent{
				Service:   pbft.NODE_MGR_SERVICE,
				EventType: pbft.NODE_MGR_ADD_NODE_EVENT,
				Event:     msg,
			}
			hub.consenter.RecvLocal(e)
		case event.BroadcastNewPeerEvent:
			log.Debug("BroadcastNewPeerEvent")
			// receive this event from consensus module
			// broadcast the local CA validition result to other replica
			peers := hub.PeerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			hub.PeerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_NEWPEER)
		case event.RecvNewPeerEvent:
			log.Debug("RecvNewPeerEvent")
			// receive from replica for a new peer CA validation
			// deliver it to consensus module
			hub.consenter.RecvMsg(ev.Payload)
		case event.DelPeerEvent:
			// a peer submit a request to exit the alliance
			log.Debug("DelPeerEvent")
			payload := ev.Payload
			routerHash, id, del := hub.PeerManager.GetRouterHashifDelete(string(payload))
			msg := &protos.DelNodeMessage{
				DelPayload: payload,
				RouterHash: routerHash,
				Id:         id,
				Del:		del,
			}
			e := &pbft.LocalEvent{
				Service:   pbft.NODE_MGR_SERVICE,
				EventType: pbft.NODE_MGR_DEL_NODE_EVENT,
				Event:     msg,
			}
			hub.consenter.RecvLocal(e)
		case event.BroadcastDelPeerEvent:
			log.Debug("BroadcastDelPeerEvent")
			// receive this event from consensus module
			// broadcast to other replica
			peers := hub.PeerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			hub.PeerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_DELPEER)
		case event.RecvDelPeerEvent:
			log.Debug("RecvNewPeerEvent")
			// receive from replica for a peer exit request submission
			// deliver it to consensus module
			hub.consenter.RecvMsg(ev.Payload)
		case event.UpdateRoutingTableEvent:
			log.Debug("UpdateRoutingTableEvent")
			// a new peer's join chain request has been accepted
			// update local routing table
			// TODO notify consensus module to add flag
			if ev.Type == true {
				// add a peer
				hub.PeerManager.UpdateRoutingTable(ev.Payload)
				hub.PassRouters()
			} else {
				// remove a peer
				hub.PeerManager.DeleteNode(string(ev.Payload))
				hub.PassRouters()
			}
		case event.AlreadyInChainEvent:
			log.Debug("AlreadyInChainEvent")
			// send negotiate event
			if hub.initType == 1 {
				hub.PeerManager.SetOnline()
				payload := hub.PeerManager.GetLocalAddressPayload()
				msg := &protos.NewNodeMessage{
					Payload: payload,
				}
				e := &pbft.LocalEvent{
					Service:   pbft.NODE_MGR_SERVICE,
					EventType: pbft.NODE_MGR_NEW_NODE_EVENT,
					Event:     msg,
				}
				hub.consenter.RecvLocal(e)
				hub.PassRouters()
				hub.NegotiateView()
			}
		}
	}
}

func (hub *EventHub) listenExecutorEvent() {
	for obj := range hub.executorSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ExecutorToConsensusEvent:
			hub.dispatchExecutorToConsensus(ev)
		case event.ExecutorToP2PEvent:
			hub.dispatchExecutorToP2P(ev)
		}
	}
}

func (hub *EventHub) sendMsg(payload []byte) {
	msg := &protos.Message{
		Type:    protos.Message_TRANSACTION,
		Payload: payload,
		//Payload: payLoad,
		Timestamp: time.Now().UnixNano(),
		Id:        0,
	}
	msgSend, err := proto.Marshal(msg)
	if err != nil {
		log.Notice("sendMsg marshal message failed")
		return
	}
	hub.consenter.RecvMsg(msgSend)

}

// Broadcast consensus msg to a batch of peers not knowing about it
func (hub *EventHub) BroadcastConsensus(payload []byte) {
	hub.PeerManager.BroadcastPeers(payload)

}

func (hub *EventHub) GetNodeInfo() p2p.PeerInfos {
	hub.nodeInfo = hub.PeerManager.GetPeerInfo()
	log.Info("nodeInfo is ", hub.nodeInfo)
	return hub.nodeInfo
}

func (hub *EventHub) PassRouters() {

	router := hub.PeerManager.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	hub.consenter.RecvLocal(msg)

}

func (hub *EventHub) NegotiateView() {

	negoView := &protos.Message{
		Type:      protos.Message_NEGOTIATE_VIEW,
		Timestamp: time.Now().UnixNano(),
		Payload:   nil,
		Id:        0,
	}
	msg, err := proto.Marshal(negoView)
	if err != nil {
		log.Notice("nego view start")
	}
	hub.eventMux.Post(event.ConsensusEvent{
		Payload: msg,
	})
}

func (hub *EventHub) dispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_REMOVE_CACHE:
		log.Debugf("[Namespace = %s] message middleware: [remove cache]", hub.namespace)
		hub.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_VC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [vc done]", hub.namespace)
		e := &pbft.LocalEvent{
			Service:   pbft.VIEW_CHANGE_SERVICE,
			EventType: pbft.VIEW_CHANGE_VC_RESET_DONE_EVENT,
			Event:     ev.Payload,
		}
		hub.consenter.RecvLocal(e)
	case executor.NOTIFY_VALIDATION_RES:
		log.Debugf("[Namespace = %s] message middleware: [validation result]", hub.namespace)
		e := &pbft.LocalEvent{
			Service:   pbft.CORE_PBFT_SERVICE,
			EventType: pbft.CORE_VALIDATED_TXS_EVENT,
			Event:     ev.Payload,
		}
		hub.consenter.RecvLocal(e)
	case executor.NOTIFY_SYNC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [sync done]", hub.namespace)
		hub.consenter.RecvMsg(ev.Payload.([]byte))
	}
}
func (hub *EventHub) dispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	switch ev.Type {
	case executor.NOTIFY_BROADCAST_DEMAND:
		log.Debugf("[Namespace = %s] message middleware: [broadcast demand]", hub.namespace)
		hub.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCCHECKPOINT)
	case executor.NOTIFY_UNICAST_INVALID:
		log.Debugf("[Namespace = %s] message middleware: [unicast invalid tx]", hub.namespace)
		peerId := ev.Peers[0]
		if peerId == uint64(hub.PeerManager.GetNodeId()) {
			hub.executor.StoreInvalidTransaction(event.InvalidTxsEvent{
				Payload: ev.Payload,
			})
		} else {
			hub.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_INVALIDRESP)
		}
	case executor.NOTIFY_BROADCAST_SINGLE:
		log.Debugf("[Namespace = %s] message middleware: [broadcast single]", hub.namespace)
		hub.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCSINGLE)
	case executor.NOTIFY_UNICAST_BLOCK:
		log.Debugf("[Namespace = %s] message middleware: [unicast block]", hub.namespace)
		hub.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCBLOCK)
	case executor.NOTIFY_SYNC_REPLICA:
		log.Debugf("[Namespace = %s] message middleware: [sync replica]", hub.namespace)
		chain := &types.Chain{}
		proto.Unmarshal(ev.Payload, chain)
		addr := hub.PeerManager.GetLocalNode().GetNodeAddr()
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain:    chain,
			Ip:       []byte(addr.IP),
			Port:     int32(addr.Port),
			Namespace:[]byte(hub.namespace),
		})
		peers := hub.PeerManager.GetVPPeers()
		var peerIds = make([]uint64, len(peers))
		for idx, peer := range peers {
			peerIds[idx] = uint64(peer.PeerAddr.ID)
		}
		hub.PeerManager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
		hub.eventMux.Post(event.ReplicaInfoEvent{
			Payload: payload,
		})
	}
}
