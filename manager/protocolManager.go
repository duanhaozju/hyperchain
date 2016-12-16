//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core/blockpool"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/protos"
	"hyperchain/recovery"
	"sync"
	"time"
)

var log *logging.Logger // package-level logger

func init() {
	log = logging.MustGetLogger("manager")
}


type ProtocolManager struct {
	serverPort        int
	blockPool         *blockpool.BlockPool
	Peermanager       p2p.PeerManager

	nodeInfo          p2p.PeerInfos // node info ,store node status,ip,port
	consenter         consensus.Consenter

	AccountManager    *accounts.AccountManager
	commonHash        crypto.CommonHash

	eventMux          *event.TypeMux

	validateSub       event.Subscription
	commitSub         event.Subscription
	consensusSub      event.Subscription
	viewChangeSub     event.Subscription
	respSub           event.Subscription
	syncCheckpointSub event.Subscription
	syncBlockSub      event.Subscription
	syncStatusSub     event.Subscription
	quitSync          chan struct{}
	wg                sync.WaitGroup
	syncBlockCache      *common.Cache
	replicaStatus       *common.Cache
	syncReplicaInterval time.Duration
	syncReplica         bool
	expired             chan bool
	expiredTime         time.Time
}
type NodeManager struct {
	peerManager p2p.PeerManager
}

var eventMuxAll *event.TypeMux

func NewProtocolManager(blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
am *accounts.AccountManager, commonHash crypto.CommonHash, interval time.Duration, syncReplica bool, expired chan bool, expiredTime time.Time) *ProtocolManager {
	synccache, _ := common.NewCache()
	replicacache, _ := common.NewCache()
	manager := &ProtocolManager{
		blockPool:           blockPool,
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

func GetEventObject() *event.TypeMux {
	return eventMuxAll
}

// start listen new block msg and consensus msg
func (pm *ProtocolManager) Start() {
	pm.wg.Add(1)
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{}, event.NewTxEvent{})
	pm.validateSub = pm.eventMux.Subscribe(event.ExeTxsEvent{})
	pm.commitSub = pm.eventMux.Subscribe(event.CommitOrRollbackBlockEvent{})
	pm.syncCheckpointSub = pm.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{})
	pm.syncBlockSub = pm.eventMux.Subscribe(event.ReceiveSyncBlockEvent{})
	pm.respSub = pm.eventMux.Subscribe(event.RespInvalidTxsEvent{})
	pm.viewChangeSub = pm.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	go pm.validateLoop()
	go pm.commitLoop()
	go pm.ConsensusLoop()
	go pm.syncBlockLoop()
	go pm.syncCheckpointLoop()
	go pm.respHandlerLoop()
	go pm.viewChangeLoop()
	go pm.checkExpired()
	if pm.syncReplica {
		pm.syncStatusSub = pm.eventMux.Subscribe(event.ReplicaStatusEvent{})
		go pm.syncReplicaStatusLoop()
		go pm.SyncReplicaStatus()
	}
	pm.wg.Wait()

}
func (self *ProtocolManager) syncCheckpointLoop() {
	self.wg.Add(-1)
	for obj := range self.syncCheckpointSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.SendCheckpointSyncEvent:
			// receive request from the consensus module, which containes required block
			// send this request to the peers
			self.SendSyncRequest(ev)

		case event.StateUpdateEvent:
			// receive synchronzation request from peers
			self.ReceiveSyncRequest(ev)
		}
	}
}

func (self *ProtocolManager) syncBlockLoop() {
	for obj := range self.syncBlockSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.ReceiveSyncBlockEvent:
			// receive block from outer peers
			self.ReceiveSyncBlocks(ev)
		}
	}
}

// listen validate msg
func (self *ProtocolManager) validateLoop() {

	for obj := range self.validateSub.Chan() {

		switch ev := obj.Data.(type) {
		case event.ExeTxsEvent:
			// start validation serially
			self.blockPool.Validate(ev, self.commonHash, self.AccountManager.Encryption, self.Peermanager)
		}
	}
}
// listen commit msg
func (self *ProtocolManager) commitLoop() {
	for obj := range self.validateSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.CommitOrRollbackBlockEvent:
			// start commit block serially
			self.blockPool.CommitBlock(ev, self.commonHash, self.Peermanager)
		}
	}
}


func (self *ProtocolManager) respHandlerLoop() {

	for obj := range self.respSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.RespInvalidTxsEvent:
			// receive invalid tx message, save to db
			self.blockPool.StoreInvalidResp(ev)
		}
	}
}
func (self *ProtocolManager) viewChangeLoop() {

	for obj := range self.viewChangeSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.VCResetEvent:
			// receive invalid tx message, save to db
			self.blockPool.ResetStatus(ev)
		case event.InformPrimaryEvent:
			//log.Notice("InformPrimaryEvent")
			self.Peermanager.SetPrimary(ev.Primary)
		}
	}
}
func (self *ProtocolManager) syncReplicaStatusLoop() {

	for obj := range self.syncStatusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.ReplicaStatusEvent:
			// receive replicas status event
			self.RecordReplicaStatus(ev)
		}
	}
}

//listen consensus msg
func (self *ProtocolManager) ConsensusLoop() {

	// automatically stops if unsubscribe
	for obj := range self.consensusSub.Chan() {
		switch ev := obj.Data.(type) {
		case event.BroadcastConsensusEvent:
			log.Info("######enter broadcast")
			go self.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go self.Peermanager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
		//go self.peerManager.SendMsgToPeers(ev.Payload,)
		case event.NewTxEvent:
			if ev.Simulate == true {
				tx := &types.Transaction{}
				proto.Unmarshal(ev.Payload, tx)
				self.blockPool.RunInSandBox(tx)
			} else {
				log.Debug("###### enter NewTxEvent")
				go self.sendMsg(ev.Payload)
			}

		case event.ConsensusEvent:
			//call consensus module
			log.Info("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)
		}
	}
}

func (self *ProtocolManager) sendMsg(payload []byte) {
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
func (self *ProtocolManager) BroadcastConsensus(payload []byte) {
	log.Debug("begin call broadcast")
	self.Peermanager.BroadcastPeers(payload)

}

func (self *ProtocolManager) GetNodeInfo() p2p.PeerInfos {
	self.nodeInfo = self.Peermanager.GetPeerInfo()
	log.Info("nodeInfo is ", self.nodeInfo)
	return self.nodeInfo
}


