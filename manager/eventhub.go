//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/pbft"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	m "hyperchain/manager/message"
	"hyperchain/manager/protos"
	"hyperchain/p2p"
	"time"

	flt "hyperchain/manager/filter"
)

const (
	SUB_VALIDATION = iota
	SUB_COMMIT
	SUB_CONSENSUS
	SUB_PEERMAINTAIN
	SUB_MISCELLANEOUS
	SUB_EXEC
	SUB_SESSION
	SUB_TRANSACTION
)

const (
	BROADCAST_VP = iota
	BROADCAST_NVP
	BROADCAST_ALL
)

type EventHub struct {
	namespace string
	// module services
	executor       *executor.Executor
	peerManager    p2p.PeerManager
	consenter      consensus.Consenter
	accountManager *accounts.AccountManager
	eventMux       *event.TypeMux
	filterMux      *event.TypeMux
	// subscription
	subscriptions map[int]event.Subscription

	filterSystem  *flt.EventSystem

	initType      int
	logger        *logging.Logger
	close         chan bool
}

func New(namespace string, eventMux *event.TypeMux, filterMux *event.TypeMux, executor *executor.Executor, peerManager p2p.PeerManager, consenter consensus.Consenter, am *accounts.AccountManager, cm *admittance.CAManager) *EventHub {
	eventHub := NewEventHub(namespace, executor, peerManager, eventMux, filterMux, consenter, am)
	return eventHub
}

func NewEventHub(namespace string, executor *executor.Executor, peerManager p2p.PeerManager, eventMux *event.TypeMux, filterMux *event.TypeMux, consenter consensus.Consenter,
	am *accounts.AccountManager) *EventHub {
	hub := &EventHub{
		namespace:      namespace,
		executor:       executor,
		eventMux:       eventMux,
		filterMux:      filterMux,
		consenter:      consenter,
		peerManager:    peerManager,
		accountManager: am,
		subscriptions:  make(map[int]event.Subscription),
		close:          make(chan bool),
		filterSystem:   flt.NewEventSystem(filterMux),
	}
	hub.logger = common.GetLogger(namespace, "eventhub")
	hub.Subscribe()
	return hub
}

func (hub *EventHub) Start() {
	go hub.listenValidateEvent()
	go hub.listenCommitEvent()
	go hub.listenConsensusEvent()
	go hub.listenExecutorEvent()
	go hub.listenMiscellaneousEvent()
	go hub.listenPeerMaintainEvent()
	go hub.listenSessionEvent()
	go hub.listenTransactionEvent()
}

func (hub *EventHub) Stop() {
	hub.close <- true
	hub.logger.Noticef("event hub stopped!")
}

// Properties
func (hub *EventHub) GetEventObject() *event.TypeMux {
	return hub.eventMux
}

func (hub *EventHub) GetConsentor() consensus.Consenter {
	return hub.consenter
}

func (hub *EventHub) GetPeerManager() p2p.PeerManager {
	return hub.peerManager
}

func (hub *EventHub) GetExecutor() *executor.Executor {
	return hub.executor
}

func (hub *EventHub) GetAccountManager() *accounts.AccountManager {
	return hub.accountManager
}

func (hub *EventHub) GetFilterSystem() *flt.EventSystem {
	return hub.filterSystem
}

func (hub *EventHub) Subscribe() {
	// Session stuff
	hub.subscriptions[SUB_SESSION] = hub.eventMux.Subscribe(event.SessionEvent{})
	// Internal stuff
	hub.subscriptions[SUB_CONSENSUS] = hub.eventMux.Subscribe(event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{}, event.NegoRoutersEvent{})
	hub.subscriptions[SUB_VALIDATION] = hub.eventMux.Subscribe(event.ValidationEvent{})
	hub.subscriptions[SUB_COMMIT] = hub.eventMux.Subscribe(event.CommitEvent{})
	hub.subscriptions[SUB_PEERMAINTAIN] = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.DelPeerEvent{}, event.BroadcastDelPeerEvent{})
	hub.subscriptions[SUB_MISCELLANEOUS] = hub.eventMux.Subscribe(event.InformPrimaryEvent{}, event.VCResetEvent{}, event.ChainSyncReqEvent{},
		event.SnapshotEvent{}, event.DeleteSnapshotEvent{}, event.ArchiveEvent{})
	hub.subscriptions[SUB_EXEC] = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})
	hub.subscriptions[SUB_TRANSACTION] = hub.eventMux.Subscribe(event.NewTxEvent{})
}

func (hub *EventHub) GetSubscription(t int) event.Subscription {
	return hub.subscriptions[t]
}

func (hub *EventHub) listenSessionEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_SESSION).Chan():
			switch ev := obj.Data.(type) {
			case event.SessionEvent:
				hub.parseAndDispatch(ev)
			}

		}
	}
}

func (hub *EventHub) listenTransactionEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_TRANSACTION).Chan():
			switch ev := obj.Data.(type) {
			case event.NewTxEvent:
				hub.logger.Debugf("message middleware: [new tx]")
				if ev.Simulate == true {
					hub.executor.RunInSandBox(ev.Transaction)
				} else {
					hub.consenter.RecvLocal(ev.Transaction)
				}
			}

		}
	}
}

// listen validate msg
func (hub *EventHub) listenValidateEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_VALIDATION).Chan():
			switch ev := obj.Data.(type) {
			case event.ValidationEvent:
				hub.logger.Debugf("message middleware: [validation]")
				hub.executor.Validate(ev)
			}

		}
	}

}

// listen commit msg
func (hub *EventHub) listenCommitEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_COMMIT).Chan():
			switch ev := obj.Data.(type) {
			case event.CommitEvent:
				hub.logger.Debugf("message middleware: [commit]")
				hub.executor.CommitBlock(ev)
			}

		}
	}

}

func (hub *EventHub) listenMiscellaneousEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_MISCELLANEOUS).Chan():
			switch ev := obj.Data.(type) {
			case event.InformPrimaryEvent:
				hub.logger.Debugf("message middleware: [inform primary]")
				hub.peerManager.SetPrimary(ev.Primary)
			case event.VCResetEvent:
				hub.logger.Debugf("message middleware: [vc reset]")
				hub.executor.Rollback(ev)
			case event.ChainSyncReqEvent:
				hub.logger.Debugf("message middleware: [chain sync request]")
				hub.executor.SyncChain(ev)
			case event.SnapshotEvent:
				hub.logger.Debugf("message middleware: [snapshot request]")
				hub.executor.Snapshot(ev)
			case event.DeleteSnapshotEvent:
				hub.logger.Debugf("message middleware: [delete snapshot request]")
				hub.executor.DeleteSnapshot(ev)
			case event.ArchiveEvent:
				hub.logger.Debugf("message middleware: [archive request]")
				hub.executor.Archive(ev)
			}
		}
	}
}

func (hub *EventHub) listenConsensusEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_CONSENSUS).Chan():
			switch ev := obj.Data.(type) {
			case event.BroadcastConsensusEvent:
				hub.logger.Debugf("message middleware: [broadcast consensus]")
				hub.broadcast(BROADCAST_VP, m.SessionMessage_CONSENSUS, ev.Payload)
			case event.TxUniqueCastEvent:
				hub.logger.Debugf("message middleware: [tx unicast]")
				hub.send(m.SessionMessage_FOWARD_TX, ev.Payload, []uint64{ev.PeerId})
			case event.NegoRoutersEvent:
				hub.logger.Debugf("message middleware: [negotiate routers]")
				hub.peerManager.UpdateAllRoutingTable(ev.Payload)
			}

		}
	}
}

func (hub *EventHub) listenPeerMaintainEvent() {

	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_PEERMAINTAIN).Chan():
			switch ev := obj.Data.(type) {
			case event.NewPeerEvent:
				hub.logger.Debugf("message middleware: [new peer]")
				hub.invokePbftLocal(pbft.NODE_MGR_SERVICE, pbft.NODE_MGR_ADD_NODE_EVENT, &protos.AddNodeMessage{ev.Payload})
			case event.BroadcastNewPeerEvent:
				hub.logger.Debugf("message middleware: [broadcast new peer]")
				hub.broadcast(BROADCAST_VP, m.SessionMessage_ADD_PEER, ev.Payload)
			case event.DelPeerEvent:
				hub.logger.Debugf("message middleware: [delete peer]")
				payload := ev.Payload
				routerHash, id, del := hub.peerManager.GetRouterHashifDelete(string(payload))
				msg := &protos.DelNodeMessage{
					DelPayload: payload,
					RouterHash: routerHash,
					Id:         id,
					Del:        del,
				}
				hub.invokePbftLocal(pbft.NODE_MGR_SERVICE, pbft.NODE_MGR_DEL_NODE_EVENT, msg)
			case event.BroadcastDelPeerEvent:
				hub.logger.Debugf("message middleware: [broadcast delete peer]")
				hub.broadcast(BROADCAST_VP, m.SessionMessage_DEL_PEER, ev.Payload)
			case event.UpdateRoutingTableEvent:
				hub.logger.Debugf("message middleware: [update routing table]")
				if ev.Type == true {
					// add a peer
					hub.peerManager.UpdateRoutingTable(ev.Payload)
					hub.PassRouters()
				} else {
					// remove a peer
					hub.peerManager.DeleteNode(string(ev.Payload))
					hub.PassRouters()
				}
			case event.AlreadyInChainEvent:
				hub.logger.Debugf("message middleware: [already in chain]")
				if hub.initType == 1 {
					hub.peerManager.SetOnline()
					payload := hub.peerManager.GetLocalAddressPayload()
					hub.invokePbftLocal(pbft.NODE_MGR_SERVICE, pbft.NODE_MGR_NEW_NODE_EVENT, &protos.NewNodeMessage{payload})
					hub.PassRouters()
					hub.NegotiateView()
				}
			}

		}
	}
}

func (hub *EventHub) listenExecutorEvent() {
	for {
		select {
		case <-hub.close:
			return
		case obj := <-hub.GetSubscription(SUB_EXEC).Chan():
			switch ev := obj.Data.(type) {
			case event.ExecutorToConsensusEvent:
				hub.dispatchExecutorToConsensus(ev)
			case event.ExecutorToP2PEvent:
				hub.dispatchExecutorToP2P(ev)
			}

		}
	}
}

func (hub *EventHub) PassRouters() {
	router := hub.peerManager.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	hub.consenter.RecvLocal(msg)

}

func (hub *EventHub) NegotiateView() {
	hub.logger.Debug("negotiate view")
	negoView := &protos.Message{
		Type:      protos.Message_NEGOTIATE_VIEW,
		Timestamp: time.Now().UnixNano(),
		Payload:   nil,
		Id:        0,
	}
	hub.consenter.RecvLocal(negoView)
}

func (hub *EventHub) dispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_REMOVE_CACHE:
		hub.logger.Debugf("message middleware: [remove cache]")
		hub.consenter.RecvLocal(ev.Payload)
	case executor.NOTIFY_VC_DONE:
		hub.logger.Debugf("message middleware: [vc done]")
		hub.invokePbftLocal(pbft.VIEW_CHANGE_SERVICE, pbft.VIEW_CHANGE_VC_RESET_DONE_EVENT, ev.Payload)
	case executor.NOTIFY_VALIDATION_RES:
		hub.logger.Debugf("message middleware: [validation result]")
		hub.invokePbftLocal(pbft.CORE_PBFT_SERVICE, pbft.CORE_VALIDATED_TXS_EVENT, ev.Payload)
	case executor.NOTIFY_SYNC_DONE:
		hub.logger.Debugf("message middleware: [sync done]")
		hub.invokePbftLocal(pbft.CORE_PBFT_SERVICE, pbft.CORE_STATE_UPDATE_EVENT, ev.Payload)
	}
}

func (hub *EventHub) dispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	switch ev.Type {
	case executor.NOTIFY_BROADCAST_DEMAND:
		hub.logger.Debugf("message middleware: [broadcast demand]")
		hub.send(m.SessionMessage_SYNC_REQ, ev.Payload, ev.Peers)
	case executor.NOTIFY_UNICAST_INVALID:
		hub.logger.Debugf("message middleware: [unicast invalid tx]")
		peerId := ev.Peers[0]
		if peerId == uint64(hub.peerManager.GetNodeId()) {
			hub.executor.StoreInvalidTransaction(ev.Payload)
		} else {
			hub.send(m.SessionMessage_UNICAST_INVALID, ev.Payload, ev.Peers)
		}
	case executor.NOTIFY_BROADCAST_SINGLE:
		hub.logger.Debugf("message middleware: [broadcast single]")
		hub.send(m.SessionMessage_BROADCAST_SINGLE_BLK, ev.Payload, ev.Peers)
	case executor.NOTIFY_UNICAST_BLOCK:
		hub.logger.Debugf("message middleware: [unicast block]")
		hub.send(m.SessionMessage_UNICAST_BLK, ev.Payload, ev.Peers)
	case executor.NOTIFY_SYNC_REPLICA:
		hub.logger.Debugf("message middleware: [sync replica]")
		chain := &types.Chain{}
		proto.Unmarshal(ev.Payload, chain)
		addr := hub.peerManager.GetLocalNode().GetNodeAddr()
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain:     chain,
			Ip:        []byte(addr.IP),
			Port:      int32(addr.Port),
			Namespace: []byte(hub.namespace),
		})
		hub.broadcast(BROADCAST_VP, m.SessionMessage_SYNC_REPLICA, payload)
		hub.executor.ReceiveReplicaInfo(payload)
	}
}

func (hub *EventHub) parseAndDispatch(ev event.SessionEvent) {
	var message m.SessionMessage
	err := proto.Unmarshal(ev.Message, &message)
	if err != nil {
		hub.logger.Error("unmarshal session message failed")
		return
	}

	switch message.Type {
	case m.SessionMessage_CONSENSUS:
		fallthrough
	case m.SessionMessage_FOWARD_TX:
		fallthrough
	case m.SessionMessage_ADD_PEER:
		fallthrough
	case m.SessionMessage_DEL_PEER:
		hub.consenter.RecvMsg(message.Payload)
	case m.SessionMessage_UNICAST_BLK:
		hub.executor.ReceiveSyncBlocks(message.Payload)
	case m.SessionMessage_UNICAST_INVALID:
		hub.executor.StoreInvalidTransaction(message.Payload)
	case m.SessionMessage_SYNC_REPLICA:
		hub.executor.ReceiveReplicaInfo(message.Payload)
	case m.SessionMessage_BROADCAST_SINGLE_BLK:
		fallthrough
	case m.SessionMessage_SYNC_REQ:
		hub.executor.ReceiveSyncRequest(message.Payload)
	default:
		hub.logger.Error("receive a undefined session event")
	}
}
