//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core/executor"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"time"
	"hyperchain/admittance"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("eventhub")
}

type EventHub struct {
	namespace           string
	executor            *executor.Executor
	Peermanager         p2p.PeerManager

	nodeInfo            p2p.PeerInfos // node info ,store node status,ip,port
	consenter           consensus.Consenter

	AccountManager      *accounts.AccountManager
	commonHash          crypto.CommonHash

	eventMux            *event.TypeMux

	validateSub         event.Subscription
	commitSub           event.Subscription
	consensusSub        event.Subscription
	viewChangeSub       event.Subscription
	respSub             event.Subscription
	chainSyncSub        event.Subscription
	syncBlockSub        event.Subscription
	syncStatusSub       event.Subscription
	peerMaintainSub     event.Subscription

	executorSub         event.Subscription

	quitSync            chan struct{}
	syncBlockCache      *common.Cache
	replicaStatus       *common.Cache
	syncReplicaInterval time.Duration
	syncReplica         bool
	expired             chan bool
	expiredTime         time.Time
	initType            int
}
type NodeManager struct {
	peerManager p2p.PeerManager
}

var eventMuxAll *event.TypeMux

func NewEventHub(namespace string, executor *executor.Executor, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
	//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
	am *accounts.AccountManager, commonHash crypto.CommonHash, interval time.Duration, syncReplica bool, expired chan bool, expiredTime time.Time) *EventHub {
	synccache, _ := common.NewCache()
	replicacache, _ := common.NewCache()
	manager := &EventHub{
		namespace:          namespace,
		executor:           executor,
		eventMux:            eventMux,
		quitSync:            make(chan struct{}),
		consenter:           consenter,
		Peermanager:         peerManager,
		AccountManager:      am,
		commonHash:          commonHash,
		syncBlockCache:      synccache,
		replicaStatus:       replicacache,
		syncReplicaInterval: interval,
		syncReplica:         syncReplica,
		expired:             expired,
		expiredTime:         expiredTime,
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
	hub.validateSub = hub.eventMux.Subscribe(event.ExeTxsEvent{})
	hub.commitSub = hub.eventMux.Subscribe(event.CommitOrRollbackBlockEvent{})
	hub.chainSyncSub = hub.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{}, event.ReceiveSyncBlockEvent{})
	hub.respSub = hub.eventMux.Subscribe(event.RespInvalidTxsEvent{})
	hub.viewChangeSub = hub.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	go hub.validateLoop()
	go hub.commitLoop()
	hub.peerMaintainSub = hub.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.RecvNewPeerEvent{},
		event.DelPeerEvent{}, event.BroadcastDelPeerEvent{}, event.RecvDelPeerEvent{})
	hub.executorSub = hub.eventMux.Subscribe(event.ExecutorToConsensusEvent{}, event.ExecutorToP2PEvent{})
	go hub.ConsensusLoop()
	go hub.ListenSynchronizationEvent()
	go hub.listenExecutorEvent()
	go hub.respHandlerLoop()
	go hub.viewChangeLoop()
	go hub.peerMaintainLoop()
	go hub.checkExpired()
	if hub.syncReplica {
		hub.syncStatusSub = hub.eventMux.Subscribe(event.ReplicaStatusEvent{})
		go hub.syncReplicaStatusLoop()
		go hub.SyncReplicaStatus()
	}

	go hub.Peermanager.Start(c, hub.eventMux, cm)
	hub.initType = <- c
	if hub.initType == 0 {
		// start in normal mode
		hub.PassRouters()
		hub.NegotiateView()
	}
}

func (self *EventHub) ListenSynchronizationEvent() {
	for obj := range self.chainSyncSub.Chan() {
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
func (self *EventHub) validateLoop() {

	for obj := range self.validateSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.ExeTxsEvent:
			// start validation serially
			//self.executor.Validate(ev, self.commonHash, self.AccountManager.Encryption, self.Peermanager)
			self.executor.Validate(ev, self.Peermanager)
		}
	}
}

// listen commit msg
func (self *EventHub) commitLoop() {
	for obj := range self.commitSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.CommitOrRollbackBlockEvent:
			// start commit block serially
			self.executor.CommitBlock(ev, self.Peermanager)
		}
	}
}

func (self *EventHub) respHandlerLoop() {

	for obj := range self.respSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.RespInvalidTxsEvent:
			// receive invalid tx message, save to db
			self.executor.StoreInvalidTransaction(ev)
		}
	}
}

func (self *EventHub) viewChangeLoop() {

	for obj := range self.viewChangeSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			// receive invalid tx message, save to db
			self.executor.Rollback(ev)
		case event.InformPrimaryEvent:
			//log.Notice("InformPrimaryEvent")
			self.Peermanager.SetPrimary(ev.Primary)
		}
	}
}

func (self *EventHub) syncReplicaStatusLoop() {

	for obj := range self.syncStatusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ReplicaStatusEvent:
			// receive replicas status event
			self.RecordReplicaStatus(ev)
		}
	}
}

//listen consensus msg
func (self *EventHub) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.BroadcastConsensusEvent:
			log.Debug("######enter broadcast")
			go self.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go self.Peermanager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)

		case event.NewTxEvent:
			if ev.Simulate == true {
				tx := &types.Transaction{}
				proto.Unmarshal(ev.Payload, tx)
				self.executor.RunInSandBox(tx)
			} else {
				log.Debug("###### enter NewTxEvent")
				go self.sendMsg(ev.Payload)
			}

		case event.ConsensusEvent:
			//log.Error("enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)

		case event.NegoRoutersEvent:
			self.Peermanager.UpdateAllRoutingTable(ev.Payload)
		}
	}
}

func (self *EventHub) peerMaintainLoop() {

	for obj := range self.peerMaintainSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.NewPeerEvent:
			log.Debug("NewPeerEvent")
			// a new peer required to join the network and past the local CA validation
			// payload is the new peer's address information
			msg := &protos.AddNodeMessage{
				Payload: ev.Payload,
			}
			self.consenter.RecvLocal(msg)
		case event.BroadcastNewPeerEvent:
			log.Debug("BroadcastNewPeerEvent")
			// receive this event from consensus module
			// broadcast the local CA validition result to other replica
			peers := self.Peermanager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			self.Peermanager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_NEWPEER)
		case event.RecvNewPeerEvent:
			log.Debug("RecvNewPeerEvent")
			// receive from replica for a new peer CA validation
			// deliver it to consensus module
			self.consenter.RecvMsg(ev.Payload)
		case event.DelPeerEvent:
			// a peer submit a request to exit the alliance
			log.Debug("DelPeerEvent")
			payload := ev.Payload
			routerHash, id, del := self.Peermanager.GetRouterHashifDelete(string(payload))
			msg := &protos.DelNodeMessage{
				DelPayload: payload,
				RouterHash: routerHash,
				Id:         id,
				Del:		del,
			}
			self.consenter.RecvLocal(msg)
		case event.BroadcastDelPeerEvent:
			log.Debug("BroadcastDelPeerEvent")
			// receive this event from consensus module
			// broadcast to other replica
			peers := self.Peermanager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, uint64(peer.PeerAddr.ID))
			}
			self.Peermanager.SendMsgToPeers(ev.Payload, peerIds, recovery.Message_BROADCAST_DELPEER)
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
				self.Peermanager.UpdateRoutingTable(ev.Payload)
				self.PassRouters()
			} else {
				// remove a peer
				self.Peermanager.DeleteNode(string(ev.Payload))
				self.PassRouters()
			}
		case event.AlreadyInChainEvent:
			log.Debug("AlreadyInChainEvent")
			// send negotiate event
			if self.initType == 1 {
				self.Peermanager.SetOnline()
				payload :=self.Peermanager.GetLocalAddressPayload()
				msg := &protos.NewNodeMessage{
					Payload: payload,
				}
				self.consenter.RecvLocal(msg)
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
	self.Peermanager.BroadcastPeers(payload)

}

func (self *EventHub) GetNodeInfo() p2p.PeerInfos {
	self.nodeInfo = self.Peermanager.GetPeerInfo()
	log.Info("nodeInfo is ", self.nodeInfo)
	return self.nodeInfo
}

func (self *EventHub) PassRouters() {

	router := self.Peermanager.GetRouters()
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
		self.Peermanager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCCHECKPOINT)
	case executor.NOTIFY_UNICAST_INVALID:
		log.Debugf("[Namespace = %s] message middleware: [unicast invalid tx]", self.namespace)
		peerId := ev.Peers[0]
		if peerId == uint64(self.Peermanager.GetNodeId()) {
			self.executor.StoreInvalidTransaction(event.RespInvalidTxsEvent{
				Payload: ev.Payload,
			})
		} else {
			self.Peermanager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_INVALIDRESP)
		}
	case executor.NOTIFY_BROADCAST_SINGLE:
		log.Debugf("[Namespace = %s] message middleware: [broadcast single]", self.namespace)
		self.Peermanager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCSINGLE)
	case executor.NOTIFY_UNICAST_BLOCK:
		log.Debugf("[Namespace = %s] message middleware: [unicast block]", self.namespace)
		self.Peermanager.SendMsgToPeers(ev.Payload, ev.Peers, recovery.Message_SYNCBLOCK)
	}
}
