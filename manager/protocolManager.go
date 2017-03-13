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
func (pm *EventHub) GetEventObject() *event.TypeMux {
	return pm.eventMux
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

func (self *EventHub) ListenSynchronizationEvent() {
	for obj := range self.syncChainSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.SendCheckpointSyncEvent:
			self.executor.SendSyncRequest(ev)

		case event.StateUpdateEvent:
			self.executor.ReceiveSyncRequest(ev)

		case event.ReceiveSyncBlockEvent:
			self.executor.ReceiveSyncBlocks(ev)
		}
	}
}

// listen validate msg
func (self *EventHub) listenValidateEvent() {
	for obj := range self.validateSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ValidationEvent:
			log.Debugf("[Namespace = %s] message middleware: [mit]", self.namespace)
			self.executor.Validate(ev, self.PeerManager)
		}
	}
}

// listen commit msg
func (self *EventHub) listenCommitEvent() {
	for obj := range self.commitSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.CommitEvent:
			log.Debugf("[Namespace = %s] message middleware: [commit]", self.namespace)
			self.executor.CommitBlock(ev, self.PeerManager)
		}
	}
}

func (self *EventHub) listenInvalidTxEvent() {
	for obj := range self.invalidSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.InvalidTxsEvent:
			log.Debugf("[Namespace = %s] message middleware: [invalid tx]", self.namespace)
			self.executor.StoreInvalidTransaction(ev)
		}
	}
}

func (self *EventHub) listenViewChangeEvent() {
	for obj := range self.viewChangeSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			log.Debugf("[Namespace = %s] message middleware: [vc reset]", self.namespace)
			self.executor.Rollback(ev)
		case event.InformPrimaryEvent:
			log.Debugf("[Namespace = %s] message middleware: [inform primary]", self.namespace)
			self.PeerManager.SetPrimary(ev.Primary)
		}
	}
}

func (self *EventHub) listenReplicaInfoEvent() {
	for obj := range self.syncReplicaSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ReplicaInfoEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive replica info]", self.namespace)
			self.executor.ReceiveReplicaInfo(ev)
		}
	}
}

func (self *EventHub) listenConsensusEvent() {
	for obj := range self.consensusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.BroadcastConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [broadcast consensus]", self.namespace)
			go self.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:
			log.Debugf("[Namespace = %s] message middleware: [tx unicast]", self.namespace)
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go self.PeerManager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
		case event.NewTxEvent:
			log.Debugf("[Namespace = %s] message middleware: [new tx]", self.namespace)
			if ev.Simulate == true {
				tx := &types.Transaction{}
				proto.Unmarshal(ev.Payload, tx)
				self.executor.RunInSandBox(tx)
			} else {
				go self.sendMsg(ev.Payload)
			}
		case event.ConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive consensus]", self.namespace)
			self.consenter.RecvMsg(ev.Payload)
		case event.NegoRoutersEvent:
			log.Debugf("[Namespace = %s] message middleware: [negotiate routers]", self.namespace)
			self.PeerManager.UpdateAllRoutingTable(ev.Payload)
		}
	}
}

func (self *EventHub) listenpeerMaintainEvent() {

	for obj := range self.peerMaintainSub.Chan() {
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
			self.consenter.RecvLocal(e)
		case event.BroadcastNewPeerEvent:
			log.Debug("BroadcastNewPeerEvent")
			// receive this event from consensus module
			// broadcast the local CA validition result to other replica
			peers := self.PeerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			self.PeerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_NEWPEER)
		case event.RecvNewPeerEvent:
			log.Debug("RecvNewPeerEvent")
			// receive from replica for a new peer CA validation
			// deliver it to consensus module
			self.consenter.RecvMsg(ev.Payload)
		case event.DelPeerEvent:
			// a peer submit a request to exit the alliance
			log.Debug("DelPeerEvent")
			payload := ev.Payload
			routerHash, id, del := self.PeerManager.GetRouterHashifDelete(string(payload))
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
			self.consenter.RecvLocal(e)
		case event.BroadcastDelPeerEvent:
			log.Debug("BroadcastDelPeerEvent")
			// receive this event from consensus module
			// broadcast to other replica
			peers := self.PeerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			self.PeerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_DELPEER)
		case event.RecvDelPeerEvent:
			log.Debug("RecvNewPeerEvent")
			// receive from replica for a peer exit request submission
			// deliver it to consensus module
			self.consenter.RecvMsg(ev.Payload)
		case event.UpdateRoutingTableEvent:
			log.Debug("UpdateRoutingTableEvent")
			// a new peer's join chain request has been accepted
			// update local routing table
			// TODO notify consensus module to add flag
			if ev.Type == true {
				// add a peer
				self.PeerManager.UpdateRoutingTable(ev.Payload)
				self.PassRouters()
			} else {
				// remove a peer
				self.PeerManager.DeleteNode(string(ev.Payload))
				self.PassRouters()
			}
		case event.AlreadyInChainEvent:
			log.Debug("AlreadyInChainEvent")
			// send negotiate event
			if self.initType == 1 {
				self.PeerManager.SetOnline()
				payload :=self.PeerManager.GetLocalAddressPayload()
				msg := &protos.NewNodeMessage{
					Payload: payload,
				}
				e := &pbft.LocalEvent{
					Service:   pbft.NODE_MGR_SERVICE,
					EventType: pbft.NODE_MGR_NEW_NODE_EVENT,
					Event:     msg,
				}
				self.consenter.RecvLocal(e)
				self.PassRouters()
				self.NegotiateView()
			}
		}
	}
}

func (self *EventHub) listenExecutorEvent() {
	for obj := range self.executorSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ExecutorToConsensusEvent:
			self.dispatchExecutorToConsensus(ev)
		case event.ExecutorToP2PEvent:
			self.dispatchExecutorToP2P(ev)
		}
	}
}

func (self *EventHub) sendMsg(payload []byte) {
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
	self.consenter.RecvMsg(msgSend)

}

// Broadcast consensus msg to a batch of peers not knowing about it
func (self *EventHub) BroadcastConsensus(payload []byte) {
	self.PeerManager.BroadcastPeers(payload)

}

func (self *EventHub) GetNodeInfo() p2p.PeerInfos {
	self.nodeInfo = self.PeerManager.GetPeerInfo()
	log.Info("nodeInfo is ", self.nodeInfo)
	return self.nodeInfo
}

func (self *EventHub) PassRouters() {

	router := self.PeerManager.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	self.consenter.RecvLocal(msg)

}

func (self *EventHub) NegotiateView() {

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
	self.eventMux.Post(event.ConsensusEvent{
		Payload: msg,
	})
}

func (self *EventHub) dispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_REMOVE_CACHE:
		log.Debugf("[Namespace = %s] message middleware: [remove cache]", self.namespace)
		self.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_VC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [vc done]", self.namespace)
		self.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_VALIDATION_RES:
		log.Debugf("[Namespace = %s] message middleware: [validation result]", self.namespace)
		self.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_SYNC_DONE:
		log.Debugf("[Namespace = %s] message middleware: [sync done]", self.namespace)
		self.consenter.RecvMsg(ev.Payload.([]byte))
	}
}
func (self *EventHub) dispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	switch ev.Type {
	case executor.NOTIFY_BROADCAST_DEMAND:
		log.Debugf("[Namespace = %s] message middleware: [broadcast demand]", self.namespace)
		self.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCCHECKPOINT)
	case executor.NOTIFY_UNICAST_INVALID:
		log.Debugf("[Namespace = %s] message middleware: [unicast invalid tx]", self.namespace)
		peerId := ev.Peers[0]
		if peerId == uint64(self.PeerManager.GetNodeId()) {
			self.executor.StoreInvalidTransaction(event.InvalidTxsEvent{
				Payload: ev.Payload,
			})
		} else {
			self.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_INVALIDRESP)
		}
	case executor.NOTIFY_BROADCAST_SINGLE:
		log.Debugf("[Namespace = %s] message middleware: [broadcast single]", self.namespace)
		self.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCSINGLE)
	case executor.NOTIFY_UNICAST_BLOCK:
		log.Debugf("[Namespace = %s] message middleware: [unicast block]", self.namespace)
		self.PeerManager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCBLOCK)
	case executor.NOTIFY_SYNC_REPLICA:
		log.Debugf("[Namespace = %s] message middleware: [sync replica]", self.namespace)
		chain := &types.Chain{}
		proto.Unmarshal(ev.Payload, chain)
		addr := self.PeerManager.GetLocalNode().GetNodeAddr()
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain:    chain,
			Ip:       []byte(addr.IP),
			Port:     int32(addr.Port),
			Namespace:[]byte(self.namespace),
		})
		peers := self.PeerManager.GetVPPeers()
		var peerIds = make([]uint64, len(peers))
		for idx, peer := range peers {
			peerIds[idx] = uint64(peer.PeerAddr.ID)
		}
		self.PeerManager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
		self.eventMux.Post(event.ReplicaInfoEvent{
			Payload: payload,
		})
	}
}
