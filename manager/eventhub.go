//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"sync"
	"time"

	"github.com/hyperchain/hyperchain/admittance"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/consensus/rbft"
	"github.com/hyperchain/hyperchain/core/executor"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	flt "github.com/hyperchain/hyperchain/manager/filter"
	m "github.com/hyperchain/hyperchain/manager/message"
	"github.com/hyperchain/hyperchain/manager/protos"
	"github.com/hyperchain/hyperchain/p2p"

	"github.com/op/go-logging"
	"github.com/golang/protobuf/proto"

)

// This file defines a transfer station called Eventhub which is used to deliver
// messages between local modules.

// Each namespace has an independent Eventhub module to maintain the correctness
// and safety of local messages transferred in local modules including:
// 1. consensus module: which is used to ensure the order of execution
// 2. p2p module: which is mainly used to maintain network connection and deliver
//    messages between all hyperchain nodes.

// event types of subscription.
const (
	// session messages sent from other nodes through network, will be parsed and
	// dispatched to corresponding module.
	SUB_SESSION = iota

	// consensus messages sent from local consensus module, will be delivered to
	// p2p module.
	SUB_CONSENSUS

	// validate request sent from consensus module, will be delivered to executor module.
	SUB_VALIDATION

	// commit request sent from consensus module , will be delivered to executor module.
	SUB_COMMIT

	// p2p peer maintain messages used to maintain peer connections and help add/delete
	// node, sent from p2p module, will be parsed and dispatched to corresponding module.
	SUB_PEERMAINTAIN

	// executor messages sent from executor module , will be delivered to consensus or
	// p2p module.
	SUB_EXEC

	// transaction request sent from rpc module , will be delivered to consensus module or
	// executor module.
	SUB_TRANSACTION

	// other events...
	SUB_MISCELLANEOUS
)

// broadcast type, to all VP or all NVP or all peers.
const (
	BROADCAST_VP = iota
	BROADCAST_NVP
	BROADCAST_ALL
)

// node identification, VP(validating peer) or NVP(non-validating peer).
const (
	IdentificationVP int = iota
	IdentificationNVP
)

// Eventhub defines a transfer station which is used to deliver messages between local modules.
type EventHub struct {
	// Each namespace has one Eventhub instance to manage internal message delivery.
	namespace string

	// module services
	peerManager p2p.PeerManager
	consenter   consensus.Consenter

	eventMux  *event.TypeMux
	filterMux *event.TypeMux
	// subscriptions records all the internal event subscriptions registered to eventMux.
	subscriptions map[int]event.Subscription
	// filterSystem is used to manage external event subscriptions.
	filterSystem *flt.EventSystem

	initType int
	logger   *logging.Logger
	close    chan bool
	wg       sync.WaitGroup
}

// New creates and returns a new Eventhub instance with the given namespace.
func New(namespace string, eventMux *event.TypeMux, filterMux *event.TypeMux, peerManager p2p.PeerManager, consenter consensus.Consenter, cm *admittance.CAManager) *EventHub {
	eventHub := NewEventHub(namespace, peerManager, eventMux, filterMux, consenter)
	return eventHub
}

func NewEventHub(namespace string, peerManager p2p.PeerManager, eventMux *event.TypeMux, filterMux *event.TypeMux, consenter consensus.Consenter) *EventHub {
	hub := &EventHub{
		namespace:    namespace,
		eventMux:     eventMux,
		filterMux:    filterMux,
		consenter:    consenter,
		peerManager:  peerManager,
		filterSystem: flt.NewEventSystem(filterMux),
	}
	hub.logger = common.GetLogger(namespace, "eventhub")
	return hub
}

// Start initializes subscriptions and starts some go-routines to listen subscribed events.
func (hub *EventHub) Start() {
	hub.close = make(chan bool)
	hub.Subscribe()
	go hub.listenSessionEvent()
	go hub.listenConsensusEvent()
	go hub.listenPeerMaintainEvent()
	go hub.listenExecutorEvent()
	go hub.listenTransactionEvent()
	go hub.listenMiscellaneousEvent()
	hub.logger.Notice("start eventhub successfully...")
}

// Stop unsubscribes all events registered to system and close eventhub.
func (hub *EventHub) Stop() {
	if hub.close != nil {
		close(hub.close)
	}
	hub.wg.Wait()
	hub.Unsubscribe()
	hub.close = nil
	hub.logger.Notice("stop eventhub successfully...")
}

// Properties
func (hub *EventHub) GetEventObject() *event.TypeMux {
	return hub.eventMux
}

func (hub *EventHub) GetExternalEventObject() *event.TypeMux {
	return hub.filterMux
}

func (hub *EventHub) GetConsentor() consensus.Consenter {
	return hub.consenter
}

func (hub *EventHub) GetPeerManager() p2p.PeerManager {
	return hub.peerManager
}

func (hub *EventHub) GetFilterSystem() *flt.EventSystem {
	return hub.filterSystem
}

// Subscribe initializes subscriptions and dispatches different event to different subscription.
func (hub *EventHub) Subscribe() {
	hub.subscriptions = make(map[int]event.Subscription)
	// Session messages.
	hub.subscriptions[SUB_SESSION] = hub.eventMux.Subscribe(event.SessionEvent{})

	// Internal messages.
	// consensus unicast or broadcast event sent to p2p module.
	hub.subscriptions[SUB_CONSENSUS] = hub.eventMux.Subscribe(event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{})

	// p2p peer maintain messages used to maintain peer connections and help add/delete node.
	hub.subscriptions[SUB_PEERMAINTAIN] = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.DelVPEvent{}, event.DelNVPEvent{},
		event.BroadcastDelPeerEvent{})

	// executor messages sent to consensus or p2p module.
	hub.subscriptions[SUB_EXEC] = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})

	// transaction request sent from rpc module to internal module.
	hub.subscriptions[SUB_TRANSACTION] = hub.eventMux.Subscribe(event.NewTxEvent{})

	// other messages.
	hub.subscriptions[SUB_MISCELLANEOUS] = hub.eventMux.Subscribe(event.InformPrimaryEvent{},
		event.ChainSyncReqEvent{}, event.SnapshotEvent{}, event.DeleteSnapshotEvent{}, event.ArchiveEvent{}, event.ArchiveRestoreEvent{})
}

// Unsubscribe unsubscribes all events registered to system.
func (hub *EventHub) Unsubscribe() {
	for typ := SUB_SESSION; typ <= SUB_MISCELLANEOUS; typ += 1 {
		if hub.subscriptions[typ] != nil {
			hub.subscriptions[typ].Unsubscribe()
		}
	}
	hub.subscriptions = nil
}

func (hub *EventHub) GetSubscription(t int) event.Subscription {
	return hub.subscriptions[t]
}

// listenSessionEvent listens session messages and parses, dispatches to corresponding module.
func (hub *EventHub) listenSessionEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen session event exit")
			return
		case obj := <-hub.GetSubscription(SUB_SESSION).Chan():
			switch ev := obj.Data.(type) {
			case event.SessionEvent:
				hub.parseAndDispatch(ev)
			}

		}
	}
}

// listenConsensusEvent listens unicast and broadcast events sent from consensus module
// and forward to p2p module to send consensus messages to other nodes.
func (hub *EventHub) listenConsensusEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen consensus event exit")
			return
		case obj := <-hub.GetSubscription(SUB_CONSENSUS).Chan():
			switch ev := obj.Data.(type) {
			case event.BroadcastConsensusEvent:
				hub.logger.Debugf("message middleware: [broadcast consensus]")
				hub.broadcast(BROADCAST_VP, m.SessionMessage_CONSENSUS, ev.Payload)
			case event.TxUniqueCastEvent:
				hub.logger.Debugf("message middleware: [tx unicast]")
				hub.send(m.SessionMessage_FOWARD_TX, ev.Payload, []uint64{ev.PeerId})
			}
		}
	}
}

// listenPeerMaintainEvent listens add/delete node related messages and delivers to
// consensus module for VP related messages or delivers to p2p module for NVP related
// messages.
func (hub *EventHub) listenPeerMaintainEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen peer maintain event exit")
			return
		case obj := <-hub.GetSubscription(SUB_PEERMAINTAIN).Chan():
			switch ev := obj.Data.(type) {
			case event.NewPeerEvent:
				hub.logger.Debugf("message middleware: [new peer]")
				hub.invokeRbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_ADD_NODE_EVENT,
					&protos.AddNodeMessage{ev.Payload})
			case event.BroadcastNewPeerEvent:
				hub.logger.Debugf("message middleware: [broadcast new peer]")
				hub.broadcast(BROADCAST_VP, m.SessionMessage_ADD_PEER, ev.Payload)
			case event.DelVPEvent:
				hub.logger.Debugf("message middleware: [delete peer]")

				if hub.GetPeerManager().IsVP() {
					payload := ev.Payload
					routerHash, id, del := hub.peerManager.GetRouterHashifDelete(string(payload))
					msg := &protos.DelNodeMessage{
						DelPayload: payload,
						RouterHash: routerHash,
						Id:         id,
						Del:        del,
					}

					hub.invokeRbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_DEL_NODE_EVENT, msg)
				} else {
					hub.GetPeerManager().DeleteNode(string(ev.Payload))
				}

			case event.DelNVPEvent:
				hub.logger.Debugf("message middleware: [delete nvp peer]")
				hub.peerManager.DeleteNVPNode(string(ev.Payload))
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
				payload := hub.peerManager.GetLocalAddressPayload()
				hub.invokeRbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_NEW_NODE_EVENT,
					&protos.NewNodeMessage{payload})
				hub.PassRouters()
				hub.NegotiateView()
			}
		}
	}
}

// listenExecutorEvent listens executor events and delivers to consensus module or
// p2p module by its event type.
func (hub *EventHub) listenExecutorEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen executor event exit")
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

// listenTransactionEvent listens transaction event and delivers to different modules according
// to the following rules:
// 1. for simulate transaction, deliver to executor module to simulate execution.
// 2. for non-simulate transaction, if current node is a VP, then deliver to consensus module,
//    else, if current node is a NVP, then relay transaction to a random VP it has connected to.
// 3. for transaction relayed from NVP, deliver to consensus module.
func (hub *EventHub) listenTransactionEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen transaction event exit")
			return
		case obj := <-hub.GetSubscription(SUB_TRANSACTION).Chan():
			switch ev := obj.Data.(type) {
			case event.NewTxEvent:
				hub.logger.Debugf("message middleware: [new tx]")
				if ev.Simulate == true {
					//TODO handle simulate transactions
				} else {
					if hub.NodeIdentification() == IdentificationNVP {
						// TODO implement NVP relay in ordering module?
					} else {
						hub.consenter.RecvLocal(ev.Transaction)
					}
				}
			case event.NvpRelayTxEvent:
				transaction := &types.Transaction{}
				err := proto.Unmarshal(ev.Payload, transaction)
				if err != nil {
					hub.logger.Error("Relay tx, unmarshal payload failed")
				} else {
					transaction.Id = uint64(hub.GetPeerManager().GetNodeId())
					hub.consenter.RecvLocal(transaction)
				}
			}

		}
	}
}

// listenMiscellaneousEvent listens some miscellaneous events and delivers to different
// modules according to the following rules:
// 1. for InformPrimaryEvent, deliver to p2p module to set current primary ID.
// 2. for other events, deliver to executor module.
func (hub *EventHub) listenMiscellaneousEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen miscellaneous event exit")
			return
		case obj := <-hub.GetSubscription(SUB_MISCELLANEOUS).Chan():
			switch ev := obj.Data.(type) {
			case event.InformPrimaryEvent:
				hub.logger.Debugf("message middleware: [inform primary]")
				hub.peerManager.SetPrimary(ev.Primary)

			// TODO forward to consensus module
			case event.ChainSyncReqEvent:
				hub.logger.Debugf("message middleware: [chain sync request]")

			// TODO forward to consensus or directly forward to executor module?
			case event.SnapshotEvent:
				hub.logger.Debugf("message middleware: [snapshot request]")
			case event.DeleteSnapshotEvent:
				hub.logger.Debugf("message middleware: [delete snapshot request]")
			case event.ArchiveEvent:
				hub.logger.Debugf("message middleware: [archive request]")
			case event.ArchiveRestoreEvent:
				hub.logger.Debugf("message middleware: [archive restore]")
			}
		}
	}
}

// PassRouters is used to pass node routers from p2p module to consensus module.
func (hub *EventHub) PassRouters() {
	router := hub.peerManager.GetRouters()
	msg := protos.RoutersMessage{Routers: router}
	hub.consenter.RecvLocal(msg)
}

// NegotiateView starts negotiate view in consensus module.
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

// dispatchExecutorToConsensus dispatches executor event to consensus module by its type.
func (hub *EventHub) dispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_VC_DONE:
		hub.logger.Debugf("message middleware: [vc done]")
		hub.invokeRbftLocal(rbft.VIEW_CHANGE_SERVICE, rbft.VIEW_CHANGE_VC_RESET_DONE_EVENT, ev.Payload)
	case executor.NOTIFY_VALIDATION_RES:
		hub.logger.Debugf("message middleware: [validation result]")
		hub.invokeRbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_VALIDATED_TXS_EVENT, ev.Payload)
	case executor.NOTIFY_SYNC_DONE:
		hub.logger.Debugf("message middleware: [sync done]")
		hub.invokeRbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_STATE_UPDATE_EVENT, ev.Payload)
	}
}

// dispatchExecutorToP2P dispatches executor event to p2p module by its type.
func (hub *EventHub) dispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	defer func() {
		hub.logger.Debugf("send session event %d finish", ev.Type)
	}()
	hub.logger.Debugf("begin to send session event %d", ev.Type)
	switch ev.Type {
	case executor.NOTIFY_UNICAST_INVALID:
		hub.logger.Debugf("message middleware: [unicast invalid tx]")
		peerId := ev.Peers[0]
		peerHash := ev.PeersHash[0]
		if peerId == uint64(hub.peerManager.GetNodeId()) {
			if len(peerHash) == 0 {
				// TODO store invalid transactions?
			} else {
				hub.sendToNVP(m.SessionMessage_UNICAST_INVALID, ev.Payload, ev.PeersHash)
			}
		} else {
			hub.send(m.SessionMessage_UNICAST_INVALID, ev.Payload, ev.Peers)
		}
	}
}

// parseAndDispatch parses messages sent from p2p modules and dispatches to corresponding modules.
func (hub *EventHub) parseAndDispatch(ev event.SessionEvent) {
	var message m.SessionMessage
	err := proto.Unmarshal(ev.Message, &message)
	if err != nil {
		hub.logger.Error("unmarshal session message failed")
		return
	}
	defer func() {
		hub.logger.Debugf("dispatch session event %s finish", message.Type.String())
	}()

	hub.logger.Debugf("handle session event %s", message.Type.String())

	switch message.Type {
	case m.SessionMessage_CONSENSUS:
		fallthrough
	// TODO save NVP or not? If save it, should add interface to transfer tx from
	// TODO executor to consensus
	case m.SessionMessage_FOWARD_TX:
		fallthrough
	case m.SessionMessage_ADD_PEER:
		fallthrough
	case m.SessionMessage_DEL_PEER:
		hub.consenter.RecvMsg(message.Payload)
	case m.SessionMessage_UNICAST_BLK:
		if hub.NodeIdentification() == IdentificationVP {
			// TODO forward to consensus module (used in state update)
		} else {
			// TODO unitcast to NVP should be implemented in executor module?
		}

	default:
		hub.logger.Error("receive a undefined session event")
	}
}

// NodeIdentification returns the current node identification using the information
// stored in p2p module.
func (hub *EventHub) NodeIdentification() int {
	if hub.peerManager.IsVP() {
		return IdentificationVP
	} else {
		return IdentificationNVP
	}
}