//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/core"
	"hyperchain/core/blockpool"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/p2p"
	"hyperchain/p2p/peermessage"
	"hyperchain/protos"
	"hyperchain/recovery"
	"sync"
	"time"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("manager")
}

type ReplicaInfo struct {
	IP              string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Port            int64  `protobuf:"varint,2,opt,name=Port" json:"Port,omitempty"`
	Hash            string `protobuf:"bytes,3,opt,name=hash" json:"hash,omitempty"`
	ID              uint64 `protobuf:"varint,4,opt,name=ID" jsonson:"ID,omitempty"`

	LatestBlockHash []byte `protobuf:"bytes,1,opt,name=latestBlockHash,proto3" json:"latestBlockHash,omitempty"`
	ParentBlockHash []byte `proto3obuf:"bytes,2,opt,name=parentBlockHash,proto3" json:"parentBlockHash,omitempty"`
	Height          uint64 `protobuf:"varint,3,opt,name=height" json:"heightight,omitempty"`
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

	newBlockSub       event.Subscription
	consensusSub      event.Subscription
	viewChangeSub     event.Subscription
	respSub           event.Subscription
	syncCheckpointSub event.Subscription
	syncBlockSub      event.Subscription
	syncStatusSub     event.Subscription
	peerMaintainSub   event.Subscription
	quitSync          chan struct{}
	wg                sync.WaitGroup
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

func NewProtocolManager(blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, eventMux *event.TypeMux, consenter consensus.Consenter,
//encryption crypto.Encryption, commonHash crypto.CommonHash) (*ProtocolManager) {
am *accounts.AccountManager, commonHash crypto.CommonHash, interval time.Duration, syncReplica bool, initType int,  expired chan bool, expiredTime time.Time) *ProtocolManager {
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
		initType:            initType,
	}
	manager.nodeInfo = make(p2p.PeerInfos, 0, 1000)
	eventMuxAll = eventMux
	return manager
}
func (pm *ProtocolManager) GetEventObject() *event.TypeMux {
	return pm.eventMux
}
func GetEventObject() *event.TypeMux {
	return eventMuxAll
}

// start listen new block msg and consensus msg
func (pm *ProtocolManager) Start() {

	pm.wg.Add(1)
	pm.consensusSub = pm.eventMux.Subscribe(event.ConsensusEvent{}, event.TxUniqueCastEvent{}, event.BroadcastConsensusEvent{}, event.NewTxEvent{})
	pm.newBlockSub = pm.eventMux.Subscribe(event.CommitOrRollbackBlockEvent{}, event.ExeTxsEvent{})
	pm.syncCheckpointSub = pm.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{})
	pm.syncBlockSub = pm.eventMux.Subscribe(event.ReceiveSyncBlockEvent{})
	pm.respSub = pm.eventMux.Subscribe(event.RespInvalidTxsEvent{})
	pm.viewChangeSub = pm.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	pm.peerMaintainSub = pm.eventMux.Subscribe(event.NewPeerEvent{}, event.BroadcastNewPeerEvent{},
		event.UpdateRoutingTableEvent{}, event.AlreadyInChainEvent{}, event.RecvNewPeerEvent{},
		event.DelPeerEvent{}, event.BroadcastDelPeerEvent{}, event.RecvDelPeerEvent{})
	go pm.NewBlockLoop()
	go pm.ConsensusLoop()
	go pm.syncBlockLoop()
	go pm.syncCheckpointLoop()
	go pm.respHandlerLoop()
	go pm.viewChangeLoop()
	go pm.peerMaintainLoop()

	go pm.checkExpired()
	if pm.syncReplica {
		pm.syncStatusSub = pm.eventMux.Subscribe(event.ReplicaStatusEvent{})
		go pm.syncReplicaStatusLoop()
		go pm.SyncReplicaStatus()
	}
	if pm.initType == 0 {
		// start in normal mode
		pm.NegotiateView()
	}
	if pm.initType == 1 {
		// join the chain dynamically
		payload := pm.Peermanager.GetLocalAddressPayload()
		msg := &protos.NewNodeMessage{
			Payload: payload,
		}
		pm.consenter.RecvLocal(msg)
		pm.Peermanager.ConnectToOthers()
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

// listen block msg
func (self *ProtocolManager) NewBlockLoop() {

	for obj := range self.newBlockSub.Chan() {

		switch ev := obj.Data.(type) {

		case event.CommitOrRollbackBlockEvent:
			// start commit block serially
			self.blockPool.CommitBlock(ev, self.commonHash, self.Peermanager)

		case event.ExeTxsEvent:
			// start validation parallelly
			go self.blockPool.Validate(ev, self.commonHash, self.AccountManager.Encryption, self.Peermanager)
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
			log.Debug("###### enter ConsensusEvent")
			self.consenter.RecvMsg(ev.Payload)
		}
	}
}

func (self *ProtocolManager) peerMaintainLoop() {

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
				peerIds = append(peerIds, peer.ID)
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
			routerHash, id := self.Peermanager.GetRouterHashifDelete(string(payload))
			msg := &protos.DelNodeMessage{
				DelPayload: payload,
				RouterHash: routerHash,
				Id: id,
			}
			self.consenter.RecvLocal(msg)
		case event.BroadcastDelPeerEvent:
			log.Debug("BroadcastDelPeerEvent")
			// receive this event from consensus module
			// broadcast to other replica
			// TODO Don't send to the exit peer itself
			peers := self.Peermanager.GetAllPeers()
			var peerIds []uint64
			for _, peer := range peers {
				peerIds = append(peerIds, peer.ID)
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
			} else {
				// remove a peer
				self.Peermanager.DeleteNode(string(ev.Payload))
			}
		case event.AlreadyInChainEvent:
			log.Debug("AlreadyInChainEvent")
			// send negotiate event
			if self.initType == 1 {
				self.Peermanager.SetOnline()
				self.NegotiateView()
			}
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
	self.Peermanager.BroadcastPeers(payload)

}

func (self *ProtocolManager) GetNodeInfo() p2p.PeerInfos {
	self.nodeInfo = self.Peermanager.GetPeerInfo()
	log.Info("nodeInfo is ", self.nodeInfo)
	return self.nodeInfo
}

func (self *ProtocolManager) SendSyncRequest(ev event.SendCheckpointSyncEvent) {
	UpdateStateMessage := &protos.UpdateStateMessage{}
	proto.Unmarshal(ev.Payload, UpdateStateMessage)
	blockChainInfo := &protos.BlockchainInfo{}
	proto.Unmarshal(UpdateStateMessage.TargetId, blockChainInfo)

	if core.GetChainCopy().RecoveryNum >= blockChainInfo.Height {
		log.Notice("receive duplicate stateupdate request, just ignore it")
		return
	}
	required := &recovery.CheckPointMessage{
		RequiredNumber: blockChainInfo.Height,
		CurrentNumber:  core.GetChainCopy().Height,
		PeerId:         UpdateStateMessage.Id,
	}

	core.UpdateRequire(blockChainInfo.Height, blockChainInfo.CurrentBlockHash, blockChainInfo.Height)
	// save context
	core.SetReplicas(UpdateStateMessage.Replicas)
	core.SetId(UpdateStateMessage.Id)

	payload, err := proto.Marshal(required)
	if err != nil {
		log.Error("SendSyncRequest marshal message failed")
		return
	}
	self.Peermanager.SendMsgToPeers(payload, UpdateStateMessage.Replicas, recovery.Message_SYNCCHECKPOINT)
}

func (self *ProtocolManager) ReceiveSyncRequest(ev event.StateUpdateEvent) {
	checkpointMsg := &recovery.CheckPointMessage{}
	proto.Unmarshal(ev.Payload, checkpointMsg)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("No Database Found")
		return
	}
	blocks := &types.Blocks{}
	for i := checkpointMsg.RequiredNumber; i > checkpointMsg.CurrentNumber; i -= 1 {
		block, err := core.GetBlockByNumber(db, i)
		if err != nil {
			log.Error("no required block number:", i)
			continue
		}

		if blocks.Batch == nil {
			blocks.Batch = append(blocks.Batch, block)
		} else {
			blocks.Batch[0] = block
		}

		payload, err := proto.Marshal(blocks)
		if err != nil {
			log.Error("ReceiveSyncRequest marshal message failed")
			continue
		}
		var peers []uint64
		peers = append(peers, checkpointMsg.PeerId)
		self.Peermanager.SendMsgToPeers(payload, peers, recovery.Message_SYNCBLOCK)
	}
}

func (self *ProtocolManager) ReceiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	if core.GetChainCopy().RequiredBlockNum != 0 {
		blocks := &types.Blocks{}
		proto.Unmarshal(ev.Payload, blocks)
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Error("No Database Found")
			return
		}
		for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {
			if blocks.Batch[i].Number <= core.GetChainCopy().RequiredBlockNum {
				log.Debug("Receive Block: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
				if blocks.Batch[i].Number == core.GetChainCopy().RequiredBlockNum {
					acceptHash := blocks.Batch[i].HashBlock(self.commonHash).Bytes()
					if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(core.GetChainCopy().RequireBlockHash) {
						core.PersistBlock(db.NewBatch(),blocks.Batch[i], self.blockPool.GetConfig().BlockVersion, true, true)
						if err := self.updateRequire(blocks.Batch[i]); err != nil {
							log.Error("UpdateRequired failed!")
							return
						}

						// receive all block in chain
						if core.GetChainCopy().RequiredBlockNum <= core.GetChainCopy().Height {
							lastBlk, err := core.GetBlockByNumber(db, core.GetChainCopy().RequiredBlockNum + 1)
							if err != nil {
								log.Error("StateUpdate Failed!")
								return
							}
							if common.Bytes2Hex(lastBlk.ParentHash) == common.Bytes2Hex(core.GetChainCopy().LatestBlockHash) {
								// execute all received block at one time
								for i := core.GetChainCopy().RequiredBlockNum + 1; i <= core.GetChainCopy().RecoveryNum; i += 1 {
									blk, err := core.GetBlockByNumber(db, i)
									if err != nil {
										log.Error("StateUpdate Failed")
										return
									} else {
										self.blockPool.ProcessBlockInVm(blk.Transactions, nil, blk.Number)
										self.blockPool.SetDemandNumber(blk.Number + 1)
										self.blockPool.SetDemandSeqNo(blk.Number + 1)
									}
								}
								core.UpdateChainByBlcokNum(db, core.GetChainCopy().RecoveryNum)
								core.UpdateRequire(uint64(0), []byte{}, uint64(0))
								core.SetReplicas(nil)
								core.SetId(0)

								payload := &protos.StateUpdatedMessage{
									SeqNo: core.GetChainCopy().Height,
								}
								msg, err := proto.Marshal(payload)
								if err != nil {
									log.Error("StateUpdate marshal stateupdated message failed")
									return
								}
								msgSend := &protos.Message{
									Type:      protos.Message_STATE_UPDATED,
									Payload:   msg,
									Timestamp: time.Now().UnixNano(),
									Id:        1,
								}
								msgPayload, err := proto.Marshal(msgSend)
								if err != nil {
									log.Error(err)
									return
								}
								time.Sleep(2000 * time.Millisecond)
								self.consenter.RecvMsg(msgPayload)
								break
							} else {
								// the highest block in local is invalid, request the block
								// TODO clear global cache
								// TODO clear receipt, txmeta, tx
								self.blockPool.CutdownBlock(lastBlk.Number - 1)
								core.UpdateChainByBlcokNum(db, lastBlk.Number - 2)
								self.broadcastDemandBlock(lastBlk.Number - 1, lastBlk.ParentHash, core.GetReplicas(), core.GetId())
							}
						}
					}

				} else {
					// requested block with smaller number arrive earlier than expected
					// store in cache temporarily
					log.Debug("Receive Block earily: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
					ret, existed := self.syncBlockCache.Get(blocks.Batch[i].Number)
					if existed {
						blks := ret.(map[string]types.Block)
						if _, ok := blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()]; ok {
							log.Notice("Receive Duplicate Block: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
							continue
						} else {
							log.Debug("Receive Sync Block with different hash: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
							blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()] = *blocks.Batch[i]
							self.syncBlockCache.Add(blocks.Batch[i].Number, blks)
						}
					} else {
						blks := make(map[string]types.Block)
						blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()] = *blocks.Batch[i]
						self.syncBlockCache.Add(blocks.Batch[i].Number, blks)
					}
				}
			}
		}
	}

}

func (self *ProtocolManager) broadcastDemandBlock(number uint64, hash []byte, replicas []uint64, peerId uint64) {
	required := &recovery.CheckPointMessage{
		RequiredNumber: number,
		CurrentNumber:  core.GetChainCopy().Height,
		PeerId:         peerId,
	}
	payload, err := proto.Marshal(required)
	if err != nil {
		log.Error("broadcastDemandBlock, marshal message failed")
		return
	}
	self.Peermanager.SendMsgToPeers(payload, replicas, recovery.Message_SYNCSINGLE)
}

func (self *ProtocolManager) updateRequire(block *types.Block) error {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		// TODO
		log.Error("updateRequire get database failed")
		return err
	}
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > core.GetChainCopy().Height {
		if self.syncBlockCache.Contains(tmp) {
			ret, _ := self.syncBlockCache.Get(tmp)
			blks := ret.(map[string]types.Block)
			for hash, blk := range blks {
				if hash == common.BytesToHash(tmpHash).Hex() {
					core.PersistBlock(db.NewBatch(), &blk, self.blockPool.GetConfig().BlockVersion, true, true)
					self.syncBlockCache.Remove(tmp)
					tmp = tmp - 1
					tmpHash = blk.ParentHash
					flag = true
					log.Debug("process sync block stored in cache", blk.Number)
					break
				} else {
					log.Debug("found Invalid sync block, discard block number %d, block hash %s\n", blk.Number, common.BytesToHash(blk.BlockHash).Hex())
				}
			}
			if flag {
				flag = false
			} else {
				self.syncBlockCache.Remove(tmp)
				break
			}
		} else {
			break
		}
	}
	core.UpdateRequire(tmp, tmpHash, core.GetChainCopy().RecoveryNum)
	log.Debug("Next Required", core.GetChainCopy().RequiredBlockNum, common.BytesToHash(core.GetChainCopy().RequireBlockHash).Hex())
	return nil
}

func (self *ProtocolManager) SyncReplicaStatus() {
	ticker := time.NewTicker(self.syncReplicaInterval)
	for {
		select {
		case <-ticker.C:
			addr, chain := self.packReplicaStatus()
			if addr == nil || chain == nil {
				continue
			}
			status := &recovery.ReplicaStatus{
				Addr:  addr,
				Chain: chain,
			}
			payload, err := proto.Marshal(status)
			if err != nil {
				log.Error("marshal syncReplicaStatus message failed")
				continue
			}
			peers := self.Peermanager.GetAllPeers()
			var peerIds = make([]uint64, len(peers))
			for idx, peer := range peers {
				peerIds[idx] = peer.ID
			}
			self.Peermanager.SendMsgToPeers(payload, peerIds, recovery.Message_SYNCREPLICA)
			// post to self
			self.eventMux.Post(event.ReplicaStatusEvent{
				Payload: payload,
			})
		}
	}
}

func (self *ProtocolManager) RecordReplicaStatus(ev event.ReplicaStatusEvent) {
	status := &recovery.ReplicaStatus{}
	proto.Unmarshal(ev.Payload, status)
	addr := &peermessage.PeerAddress{}
	chain := &types.Chain{}
	proto.Unmarshal(status.Addr, addr)
	proto.Unmarshal(status.Chain, chain)
	replicaInfo := ReplicaInfo{
		IP:              addr.IP,
		Port:            addr.Port,
		Hash:            addr.Hash,
		ID:              addr.ID,
		LatestBlockHash: chain.LatestBlockHash,
		ParentBlockHash: chain.ParentBlockHash,
		Height:          chain.Height,
	}
	log.Critical("recv Replica Status", replicaInfo)
	self.replicaStatus.Add(addr.ID, replicaInfo)
}
func (self *ProtocolManager) packReplicaStatus() ([]byte, []byte) {
	peerAddress := self.Peermanager.GetLocalNode().GetNodeAddr()
	currentChain := core.GetChainCopy()
	// remove useless fields
	currentChain.RequiredBlockNum = 0
	currentChain.RequireBlockHash = nil
	currentChain.RecoveryNum = 0
	currentChain.CurrentTxSum = 0
	// marshal
	addrData, err := proto.Marshal(peerAddress)
	if err != nil {
		log.Error("packReplicaStatus failed!")
		return nil, nil
	}
	chainData, err := proto.Marshal(currentChain)
	if err != nil {
		log.Error("packReplicaStatus failed!")
		return nil, nil
	}
	return addrData, chainData
}

func (self *ProtocolManager) checkExpired() {
	expiredChecker := func(currentTime, expiredTime time.Time) bool {
		return currentTime.Before(expiredTime)
	}
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			if !expiredChecker(currentTime, self.expiredTime) {
				log.Error("License Expired")
				self.expired <- true
			}
		}
	}

}

func (self *ProtocolManager) NegotiateView() {
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
