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
	"hyperchain/consensus/rbft"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	m "hyperchain/manager/message"
	"hyperchain/manager/protos"
	"hyperchain/p2p"
	"time"

	flt "hyperchain/manager/filter"
	"sync"
    "fmt"
)

// This file defines a transfer station called Eventhub which is used to deliver
// messages between local modules.

// Each namespace has an independent Eventhub module to maintain the correctness
// and safety of local messages transferred in local modules including:
// 1. consensus module: which is used to ensure the order of execution
// 2. p2p module: which is mainly used to maintain network connection and deliver
//    messages between all hyperchain nodes.
// 3. executor module: which is mainly used to validate and commit transactions.

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
	//executor       *executor.Executor
	executor       executor.IExecutor
	peerManager    p2p.PeerManager
	consenter      consensus.Consenter
	accountManager *accounts.AccountManager

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
func New(namespace string, eventMux *event.TypeMux, filterMux *event.TypeMux, executor executor.IExecutor, peerManager p2p.PeerManager, consenter consensus.Consenter, am *accounts.AccountManager, cm *admittance.CAManager) *EventHub {
	eventHub := NewEventHub(namespace, executor, peerManager, eventMux, filterMux, consenter, am)
	return eventHub
}

func NewEventHub(namespace string, executor executor.IExecutor, peerManager p2p.PeerManager, eventMux *event.TypeMux, filterMux *event.TypeMux, consenter consensus.Consenter,
	am *accounts.AccountManager) *EventHub {
	hub := &EventHub{
		namespace:      namespace,
		executor:       executor,
		eventMux:       eventMux,
		filterMux:      filterMux,
		consenter:      consenter,
		peerManager:    peerManager,
		accountManager: am,
		filterSystem:   flt.NewEventSystem(filterMux),
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
	go hub.listenValidateEvent()
	go hub.listenCommitEvent()
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

func (hub *EventHub) GetExecutor() executor.IExecutor {
	return hub.executor
}

func (hub *EventHub) GetAccountManager() *accounts.AccountManager {
	return hub.accountManager
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

	// consensus validate request sent to executor module.
	hub.subscriptions[SUB_VALIDATION] = hub.eventMux.Subscribe(event.ValidationEvent{})

	// consensus commit request sent to executor module.
	hub.subscriptions[SUB_COMMIT] = hub.eventMux.Subscribe(event.CommitEvent{})

	// p2p peer maintain messages used to maintain peer connections and help add/delete node.
	hub.subscriptions[SUB_PEERMAINTAIN] = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.DelVPEvent{}, event.DelNVPEvent{},
		event.BroadcastDelPeerEvent{})

	// executor messages sent to consensus or p2p module.
	hub.subscriptions[SUB_EXEC] = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})

	// transaction request sent from rpc module to internal module.
	hub.subscriptions[SUB_TRANSACTION] = hub.eventMux.Subscribe(event.NewTxEvent{}, event.NvpRelayTxEvent{})

	// other messages.
	hub.subscriptions[SUB_MISCELLANEOUS] = hub.eventMux.Subscribe(event.InformPrimaryEvent{}, event.VCResetEvent{},
		event.ChainSyncReqEvent{}, event.SnapshotEvent{}, event.DeleteSnapshotEvent{}, event.ArchiveEvent{})
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

// listenValidateEvent listens validate event sent from consensus module and forward to
// executor module to start validate batch.
func (hub *EventHub) listenValidateEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen validation event exit")
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

// listenCommitEvent listens commit event sent from consensus module and forward to
// executor module to start commit batch.
func (hub *EventHub) listenCommitEvent() {
	hub.wg.Add(1)
	for {
		select {
		case <-hub.close:
			hub.wg.Done()
			hub.logger.Debug("listen commit event exit")
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
				payload := ev.Payload
				routerHash, id, del := hub.peerManager.GetRouterHashifDelete(string(payload))
				msg := &protos.DelNodeMessage{
					DelPayload: payload,
					RouterHash: routerHash,
					Id:         id,
					Del:        del,
				}
				hub.invokeRbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_DEL_NODE_EVENT, msg)
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
				hub.DispatchExecutorToConsensus(ev)
			case event.ExecutorToP2PEvent:
				hub.DispatchExecutorToP2P(ev)
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
					hub.executor.RunInSandBox(ev.Transaction, ev.SnapshotId)
				} else {
					if hub.NodeIdentification() == IdentificationNVP {
						hub.RelayTx(ev.Transaction, ev.Ch)
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

// DispatchExecutorToConsensus dispatches executor event to consensus module by its type.
func (hub *EventHub) DispatchExecutorToConsensus(ev event.ExecutorToConsensusEvent) {
	switch ev.Type {
	case executor.NOTIFY_VC_DONE:
		hub.logger.Criticalf("message middleware: [vc done]")
        fmt.Println("message middleware: [vc done]")
		event := &protos.VcResetDone{}
		err := proto.Unmarshal(ev.Payload, event)
		if err != nil {
			hub.logger.Error(err)
			return
		}
		hub.invokeRbftLocal(rbft.VIEW_CHANGE_SERVICE, rbft.VIEW_CHANGE_VC_RESET_DONE_EVENT, *event)
	case executor.NOTIFY_VALIDATION_RES:
		hub.logger.Debugf("message middleware: [validation result]")

		event := &protos.ValidatedTxs{}
		err := proto.Unmarshal(ev.Payload, event)
		if err != nil {
			hub.logger.Error(err)
			return
		}
		hub.invokeRbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_VALIDATED_TXS_EVENT, *event)
	case executor.NOTIFY_SYNC_DONE:
		hub.logger.Debugf("message middleware: [sync done]")

		event := &protos.StateUpdatedMessage{}
		err := proto.Unmarshal(ev.Payload, event)
		if err != nil {
			hub.logger.Error(err)
			return
		}
		hub.invokeRbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_STATE_UPDATE_EVENT, *event)
	}
}

// DispatchExecutorToP2P dispatches executor event to p2p module by its type.
func (hub *EventHub) DispatchExecutorToP2P(ev event.ExecutorToP2PEvent) {
	switch ev.Type {
	case executor.NOTIFY_BROADCAST_DEMAND:
		hub.logger.Debugf("message middleware: [broadcast demand]")
		hub.send(m.SessionMessage_SYNC_REQ, ev.Payload, ev.Peers)
	case executor.NOTIFY_UNICAST_INVALID:
		hub.logger.Debugf("message middleware: [unicast invalid tx]")
		peerId := ev.Peers[0]
		peerHash := ev.PeersHash[0]
		if peerId == uint64(hub.peerManager.GetNodeId()) {
			if len(peerHash) == 0 {
				hub.executor.StoreInvalidTransaction(ev.Payload)
			} else {
				hub.sendToNVP(m.SessionMessage_UNICAST_INVALID, ev.Payload, ev.PeersHash)
			}
		} else {
			hub.send(m.SessionMessage_UNICAST_INVALID, ev.Payload, ev.Peers)
		}
	case executor.NOTIFY_UNICAST_BLOCK:
		hub.logger.Debugf("message middleware: [unicast block]")
		toNVP := func() bool {
			if len(ev.PeersHash) != 0 && len(ev.PeersHash[0]) != 0 {
				return true
			}
			return false
		}
		if toNVP() {
			hub.sendToNVP(m.SessionMessage_UNICAST_BLK, ev.Payload, ev.PeersHash)
		} else {
			hub.send(m.SessionMessage_UNICAST_BLK, ev.Payload, ev.Peers)
		}
	case executor.NOTIFY_SYNC_REPLICA:
		hub.logger.Debugf("message middleware: [sync replica]")
		chain := &types.Chain{}
		proto.Unmarshal(ev.Payload, chain)
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain:     chain,
			Namespace: []byte(hub.namespace),
		})
		hub.broadcast(BROADCAST_VP, m.SessionMessage_SYNC_REPLICA, payload)
		hub.executor.ReceiveReplicaInfo(payload)
	case executor.NOTIFY_TRANSIT_BLOCK:
		hub.logger.Debug("message middleware: [transit block]")
		hub.broadcast(BROADCAST_NVP, m.SessionMessage_TRANSIT_BLOCK, ev.Payload)
	case executor.NOTIFY_NVP_SYNC:
		hub.logger.Debug("message middleware: [NVP sync]")
		syncMsg := &executor.ChainSyncRequest{}
		proto.Unmarshal(ev.Payload, syncMsg)
		syncMsg.PeerHash = hub.peerManager.GetLocalNodeHash()
		payload, err := proto.Marshal(syncMsg)
		if err != nil {
			hub.logger.Error("message middleware: SendNVPSyncRequest marshal message failed")
			return
		}
		hub.sendToRandomVP(m.SessionMessage_SYNC_REQ, payload)
	case executor.NOTIFY_REQUEST_WORLD_STATE:
		hub.logger.Debugf("message middleware: [request world state]")
		hub.send(m.SessionMessage_SYNC_WORLD_STATE, ev.Payload, ev.Peers)
	case executor.NOTIFY_SEND_WORLD_STATE_HANDSHAKE:
		hub.logger.Debugf("message middleware: [send world state handshake packet]")
		hub.send(m.SessionMessage_SEND_WS_HS, ev.Payload, ev.Peers)
	case executor.NOTIFY_SEND_WORLD_STATE:
		hub.logger.Debugf("message middleware: [request world state]")
		hub.send(m.SessionMessage_SEND_WORLD_STATE, ev.Payload, ev.Peers)
	case executor.NOTIFY_SEND_WS_ACK:
		hub.logger.Debugf("message middleware: [send ws ack]")
		hub.send(m.SessionMessage_SEND_WS_ACK, ev.Payload, ev.Peers)
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
		if hub.NodeIdentification() == IdentificationVP {
			hub.executor.ReceiveSyncBlocks(message.Payload)
		} else {
			hub.executor.GetNVP().ReceiveBlock(message.Payload)
		}
	case m.SessionMessage_UNICAST_INVALID:
		sendToNVP, hash := hub.isSendToNVP(message.Payload)
		// for invalid transaction relayed from NVP, relay back, else, only store it.
		if sendToNVP {
			hub.sendToNVP(m.SessionMessage_UNICAST_INVALID, message.Payload, hash)
		} else {
			hub.executor.StoreInvalidTransaction(message.Payload)
		}
	case m.SessionMessage_SYNC_REPLICA:
		hub.executor.ReceiveReplicaInfo(message.Payload)
	case m.SessionMessage_BROADCAST_SINGLE_BLK:
		fallthrough
	case m.SessionMessage_SYNC_REQ:
		hub.executor.ReceiveSyncRequest(message.Payload)
	case m.SessionMessage_SYNC_WORLD_STATE:
		hub.executor.ReceiveWorldStateSyncRequest(message.Payload)
	case m.SessionMessage_SEND_WORLD_STATE:
		hub.executor.ReceiveWorldState(message.Payload)
	case m.SessionMessage_SEND_WS_HS:
		hub.executor.ReceiveWsHandshake(message.Payload)
	case m.SessionMessage_SEND_WS_ACK:
		hub.executor.ReceiveWsAck(message.Payload)
	case m.SessionMessage_TRANSIT_BLOCK:
		hub.executor.GetNVP().ReceiveBlock(message.Payload)
	case m.SessionMessage_NVP_RELAY:
		go hub.GetEventObject().Post(event.NvpRelayTxEvent{
			Payload: message.Payload,
		})
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

// RelayTx is used by NVP when it has received a non-simulate transaction to forward to
// a random VP it has connected to.
func (hub *EventHub) RelayTx(transaction *types.Transaction, ch chan bool) {
	payload, err := proto.Marshal(transaction)
	if err != nil {
		hub.logger.Error("Relay tx, marshal payload failed")
		ch <- false
	}
	err = hub.sendToRandomVP(m.SessionMessage_NVP_RELAY, payload)
	if err == nil {
		ch <- true
	} else {
		ch <- false
	}
}

// isSendToNVP is used when a VP detects an invalid transaction. If this transaction is
// relayed from an NVP, then return true and NVP's hash, else return false which means
// this transaction is relayed from a VP and we don't need to relay back this transaction
// to that VP.
func (hub *EventHub) isSendToNVP(payload []byte) (bool, []string) {
	invalidTx := &types.InvalidTransactionRecord{}
	err := proto.Unmarshal(payload, invalidTx)
	if err != nil {
		hub.logger.Error("unmarshal invalid transaction record payload failed")
	}
	hash := invalidTx.Tx.GetNVPHash()
	if err != nil {
		hub.logger.Error("get NVP hash failed. Err Msg: %v.", err.Error())
	}
	if len(hash) > 0 && hub.peerManager.IsVP() {
		return true, []string{hash}
	}
	return false, nil
}
