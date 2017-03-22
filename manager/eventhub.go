//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/consensus"
	"hyperchain/consensus/pbft"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"time"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("eventhub")
}

const (
	SUB_VALIDATION = iota
	SUB_COMMIT
	SUB_CONSENSUS
	SUB_VIEWCHANGE
	SUB_SYNCCHAIN
	SUB_PEERMAINTAIN
	SUB_MISCELLANEOUS
	SUB_EXEC
)

type EventHub struct {
	namespace string
	// module services
	executor       *executor.Executor
	peerManager    p2p.PeerManager
	consenter      consensus.Consenter
	accountManager *accounts.AccountManager
	eventMux       *event.TypeMux
	// subscription
	subscriptions map[int]event.Subscription
	initType      int
}

func NewEventHub(namespace string, executor *executor.Executor, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
	am *accounts.AccountManager) *EventHub {
	manager := &EventHub{
		namespace:      namespace,
		executor:       executor,
		eventMux:       eventMux,
		consenter:      consenter,
		peerManager:    peerManager,
		accountManager: am,
		subscriptions:  make(map[int]event.Subscription),
	}
	return manager
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

// start listen new block msg and consensus msg
func (hub *EventHub) Start(c chan int, cm *admittance.CAManager) {
	hub.Subscribe()
	go hub.listenValidateEvent()
	go hub.listenCommitEvent()
	go hub.listenConsensusEvent()
	go hub.listenSynchronizationEvent()
	go hub.listenExecutorEvent()
	go hub.listenMiscellaneousEvent()
	go hub.listenViewChangeEvent()
	go hub.listenPeerMaintainEvent()

	go hub.peerManager.Start(c, hub.eventMux, cm)
	hub.initType = <-c
	if hub.initType == 0 {
		// start in normal mode
		hub.PassRouters()
		hub.NegotiateView()
	}
}

func (hub *EventHub) Subscribe() {
	hub.subscriptions[SUB_CONSENSUS] = hub.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{},
		event.NewTxEvent{}, event.NegoRoutersEvent{})
	hub.subscriptions[SUB_VALIDATION] = hub.eventMux.Subscribe(event.ValidationEvent{})
	hub.subscriptions[SUB_COMMIT] = hub.eventMux.Subscribe(event.CommitEvent{})
	hub.subscriptions[SUB_SYNCCHAIN] = hub.eventMux.Subscribe(event.SyncBlockReqEvent{}, event.ChainSyncReqEvent{}, event.SyncBlockReceiveEvent{})
	hub.subscriptions[SUB_VIEWCHANGE] = hub.eventMux.Subscribe(event.VCResetEvent{})
	hub.subscriptions[SUB_PEERMAINTAIN] = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.RecvNewPeerEvent{},
		event.DelPeerEvent{}, event.BroadcastDelPeerEvent{}, event.RecvDelPeerEvent{})
	hub.subscriptions[SUB_MISCELLANEOUS] = hub.eventMux.Subscribe(event.InvalidTxsEvent{}, event.ReplicaInfoEvent{}, event.InformPrimaryEvent{})
	hub.subscriptions[SUB_EXEC] = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})
}

func (hub *EventHub) GetSubscription(t int) event.Subscription {
	return hub.subscriptions[t]
}

func (hub *EventHub) listenSynchronizationEvent() {
	for obj := range hub.GetSubscription(SUB_SYNCCHAIN).Chan() {
		switch ev := obj.Data.(type) {
		case event.ChainSyncReqEvent:
			log.Debugf("[Namespace = %s] message middleware: [chain sync request]", hub.namespace)
			hub.executor.SendSyncRequest(ev)

		case event.SyncBlockReqEvent:
			log.Debugf("[Namespace = %s] message middleware: [sync block request]", hub.namespace)
			hub.executor.ReceiveSyncRequest(ev)

		case event.SyncBlockReceiveEvent:
			log.Debugf("[Namespace = %s] message middleware: [sync block receive]", hub.namespace)
			hub.executor.ReceiveSyncBlocks(ev)
		}
	}
}

// listen validate msg
func (hub *EventHub) listenValidateEvent() {
	for obj := range hub.GetSubscription(SUB_VALIDATION).Chan() {
		switch ev := obj.Data.(type) {
		case event.ValidationEvent:
			log.Debugf("[Namespace = %s] message middleware: [validation]", hub.namespace)
			hub.executor.Validate(ev)
		}
	}
}

// listen commit msg
func (hub *EventHub) listenCommitEvent() {
	for obj := range hub.GetSubscription(SUB_COMMIT).Chan() {
		switch ev := obj.Data.(type) {
		case event.CommitEvent:
			log.Debugf("[Namespace = %s] message middleware: [commit]", hub.namespace)
			hub.executor.CommitBlock(ev)
		}
	}
}

func (hub *EventHub) listenMiscellaneousEvent() {
	for obj := range hub.GetSubscription(SUB_MISCELLANEOUS).Chan() {
		switch ev := obj.Data.(type) {
		case event.InvalidTxsEvent:
			log.Debugf("[Namespace = %s] message middleware: [invalid tx]", hub.namespace)
			hub.executor.StoreInvalidTransaction(ev)
		case event.InformPrimaryEvent:
			log.Debugf("[Namespace = %s] message middleware: [inform primary]", hub.namespace)
			hub.peerManager.SetPrimary(ev.Primary)
		case event.ReplicaInfoEvent:
			log.Debugf("[Namespace = %s] message middleware: [sync replica receive]", hub.namespace)
			hub.executor.ReceiveReplicaInfo(ev)
		}
	}
}

func (hub *EventHub) listenViewChangeEvent() {
	for obj := range hub.GetSubscription(SUB_VIEWCHANGE).Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			log.Debugf("[Namespace = %s] message middleware: [vc reset]", hub.namespace)
			hub.executor.Rollback(ev)
		}
	}
}

func (hub *EventHub) listenConsensusEvent() {
	for obj := range hub.GetSubscription(SUB_CONSENSUS).Chan() {
		switch ev := obj.Data.(type) {
		case event.BroadcastConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [broadcast consensus]", hub.namespace)
			hub.peerManager.BroadcastPeers(ev.Payload)
		case event.TxUniqueCastEvent:
			log.Debugf("[Namespace = %s] message middleware: [tx unicast]", hub.namespace)
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go hub.peerManager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
		case event.NewTxEvent:
			log.Debugf("[Namespace = %s] message middleware: [new tx]", hub.namespace)
			if ev.Simulate == true {
				tx := &types.Transaction{}
				proto.Unmarshal(ev.Payload, tx)
				hub.executor.RunInSandBox(tx)
			} else {
				msg, err :=  proto.Marshal(&protos.Message{
					Type:    protos.Message_TRANSACTION,
					Payload: ev.Payload,
					Timestamp: time.Now().UnixNano(),
					Id:        0,
				})
				if err != nil {
					return
				}
				hub.consenter.RecvMsg(msg)
			}
		case event.ConsensusEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive consensus]", hub.namespace)
			hub.consenter.RecvMsg(ev.Payload)
		case event.NegoRoutersEvent:
			log.Debugf("[Namespace = %s] message middleware: [negotiate routers]", hub.namespace)
			hub.peerManager.UpdateAllRoutingTable(ev.Payload)
		}
	}
}

func (hub *EventHub) listenPeerMaintainEvent() {
	for obj := range hub.GetSubscription(SUB_PEERMAINTAIN).Chan() {
		switch ev := obj.Data.(type) {
		case event.NewPeerEvent:
			log.Debugf("[Namespace = %s] message middleware: [new peer]", hub.namespace)
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
			log.Debugf("[Namespace = %s] message middleware: [broadcast new peer]", hub.namespace)
			peers := hub.peerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			hub.peerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_NEWPEER)
		case event.RecvNewPeerEvent:
			log.Debugf("[Namespace = %s] message middleware: [recv new peer]", hub.namespace)
			hub.consenter.RecvMsg(ev.Payload)
		case event.DelPeerEvent:
			log.Debugf("[Namespace = %s] message middleware: [delete peer]", hub.namespace)
			payload := ev.Payload
			routerHash, id, del := hub.peerManager.GetRouterHashifDelete(string(payload))
			msg := &protos.DelNodeMessage{
				DelPayload: payload,
				RouterHash: routerHash,
				Id:         id,
				Del:        del,
			}
			e := &pbft.LocalEvent{
				Service:   pbft.NODE_MGR_SERVICE,
				EventType: pbft.NODE_MGR_DEL_NODE_EVENT,
				Event:     msg,
			}
			hub.consenter.RecvLocal(e)
		case event.BroadcastDelPeerEvent:
			log.Debugf("[Namespace = %s] message middleware: [broadcast delete peer]", hub.namespace)
			peers := hub.peerManager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			hub.peerManager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_DELPEER)
		case event.RecvDelPeerEvent:
			log.Debugf("[Namespace = %s] message middleware: [receive delete peer]", hub.namespace)
			hub.consenter.RecvMsg(ev.Payload)
		case event.UpdateRoutingTableEvent:
			log.Debugf("[Namespace = %s] message middleware: [update routing table]", hub.namespace)
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
			log.Debugf("[Namespace = %s] message middleware: [already in chain]", hub.namespace)
			if hub.initType == 1 {
				hub.peerManager.SetOnline()
				payload := hub.peerManager.GetLocalAddressPayload()
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
	for obj := range hub.GetSubscription(SUB_EXEC).Chan() {
		switch ev := obj.Data.(type) {
		case event.ExecutorToConsensusEvent:
			hub.dispatchExecutorToConsensus(ev)
		case event.ExecutorToP2PEvent:
			hub.dispatchExecutorToP2P(ev)
		}
	}
}

func (hub *EventHub) PassRouters() {
	router := hub.peerManager.GetRouters()
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
		log.Error("marshal nego view failed")
		return
	}
	hub.consenter.RecvMsg(msg)
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
