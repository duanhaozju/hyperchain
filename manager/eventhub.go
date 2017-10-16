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

const (
	IdentificationVP int = iota
	IdentificationNVP
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

	filterSystem *flt.EventSystem

	initType int
	logger   *logging.Logger
	close    chan bool
	wg       sync.WaitGroup
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
		filterSystem:   flt.NewEventSystem(filterMux),
	}
	hub.logger = common.GetLogger(namespace, "eventhub")
	return hub
}

func (hub *EventHub) Start() {
	hub.close = make(chan bool)
	hub.Subscribe()
	go hub.listenValidateEvent()
	go hub.listenCommitEvent()
	go hub.listenConsensusEvent()
	go hub.listenExecutorEvent()
	go hub.listenMiscellaneousEvent()
	go hub.listenPeerMaintainEvent()
	go hub.listenSessionEvent()
	go hub.listenTransactionEvent()
	hub.logger.Notice("start eventhub successfully...")
}

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
	hub.subscriptions = make(map[int]event.Subscription)
	// Session stuff
	hub.subscriptions[SUB_SESSION] = hub.eventMux.Subscribe(event.SessionEvent{})
	// Internal stuff
	hub.subscriptions[SUB_CONSENSUS] = hub.eventMux.Subscribe(event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{})
	hub.subscriptions[SUB_VALIDATION] = hub.eventMux.Subscribe(event.ValidationEvent{})
	hub.subscriptions[SUB_COMMIT] = hub.eventMux.Subscribe(event.CommitEvent{})
	hub.subscriptions[SUB_PEERMAINTAIN] = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.DelVPEvent{}, event.DelNVPEvent{}, event.BroadcastDelPeerEvent{})
	hub.subscriptions[SUB_MISCELLANEOUS] = hub.eventMux.Subscribe(event.InformPrimaryEvent{}, event.VCResetEvent{}, event.ChainSyncReqEvent{},
		event.SnapshotEvent{}, event.DeleteSnapshotEvent{}, event.ArchiveEvent{})
	hub.subscriptions[SUB_EXEC] = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})
	hub.subscriptions[SUB_TRANSACTION] = hub.eventMux.Subscribe(event.NewTxEvent{}, event.NvpRelayTxEvent{})
}

func (hub *EventHub) Unsubscribe() {
	for typ := SUB_VALIDATION; typ <= SUB_TRANSACTION; typ += 1 {
		if hub.subscriptions[typ] != nil {
			hub.subscriptions[typ].Unsubscribe()
		}
	}
	hub.subscriptions = nil
}

func (hub *EventHub) GetSubscription(t int) event.Subscription {
	return hub.subscriptions[t]
}

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

// listen validate msg
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

// listen commit msg
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
				hub.invokePbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_ADD_NODE_EVENT, &protos.AddNodeMessage{ev.Payload})
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
				hub.invokePbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_DEL_NODE_EVENT, msg)
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
				hub.invokePbftLocal(rbft.NODE_MGR_SERVICE, rbft.NODE_MGR_NEW_NODE_EVENT, &protos.NewNodeMessage{payload})
				hub.PassRouters()
				hub.NegotiateView()
			}
		}
	}
}

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
	case executor.NOTIFY_VC_DONE:
		hub.logger.Debugf("message middleware: [vc done]")
		hub.invokePbftLocal(rbft.VIEW_CHANGE_SERVICE, rbft.VIEW_CHANGE_VC_RESET_DONE_EVENT, ev.Payload)
	case executor.NOTIFY_VALIDATION_RES:
		hub.logger.Debugf("message middleware: [validation result]")
		hub.invokePbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_VALIDATED_TXS_EVENT, ev.Payload)
	case executor.NOTIFY_SYNC_DONE:
		hub.logger.Debugf("message middleware: [sync done]")
		hub.invokePbftLocal(rbft.CORE_RBFT_SERVICE, rbft.CORE_STATE_UPDATE_EVENT, ev.Payload)
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
	case executor.NOTIFY_BROADCAST_SINGLE:
		hub.logger.Debugf("message middleware: [broadcast single]")
		hub.send(m.SessionMessage_BROADCAST_SINGLE_BLK, ev.Payload, ev.Peers)
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
		// TODO FIX this! @chenquan
		//addr := hub.peerManager.GetLocalNode().GetNodeAddr()
		payload, _ := proto.Marshal(&types.ReplicaInfo{
			Chain: chain,
			//Ip:        []byte(addr.IP),
			//Port:      int32(addr.Port),
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

func (hub *EventHub) NodeIdentification() int {
	if hub.peerManager.IsVP() {
		return IdentificationVP
	} else {
		return IdentificationNVP
	}
}

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
