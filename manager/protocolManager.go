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
	"hyperchain/p2p/peer"
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
	serverPort          int
	blockPool           *blockpool.BlockPool
	Peermanager         p2p.PeerManager

	nodeInfo            client.PeerInfos // node info ,store node status,ip,port
	consenter           consensus.Consenter

	AccountManager      *accounts.AccountManager
	commonHash          crypto.CommonHash

	eventMux            *event.TypeMux

	newBlockSub         event.Subscription
	consensusSub        event.Subscription
	viewChangeSub       event.Subscription
	respSub             event.Subscription
	syncCheckpointSub   event.Subscription
	syncBlockSub        event.Subscription
	syncStatusSub       event.Subscription
	quitSync            chan struct{}
	wg                  sync.WaitGroup
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
	manager.nodeInfo = make(client.PeerInfos, 0, 1000)
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
	pm.newBlockSub = pm.eventMux.Subscribe(event.CommitOrRollbackBlockEvent{}, event.ExeTxsEvent{})
	pm.syncCheckpointSub = pm.eventMux.Subscribe(event.StateUpdateEvent{}, event.SendCheckpointSyncEvent{})
	pm.syncBlockSub = pm.eventMux.Subscribe(event.ReceiveSyncBlockEvent{})
	pm.respSub = pm.eventMux.Subscribe(event.RespInvalidTxsEvent{})
	pm.viewChangeSub = pm.eventMux.Subscribe(event.VCResetEvent{}, event.InformPrimaryEvent{})
	go pm.NewBlockLoop()
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
			log.Debug("######enter broadcast")
			go self.BroadcastConsensus(ev.Payload)
		case event.TxUniqueCastEvent:
			var peers []uint64
			peers = append(peers, ev.PeerId)
			go self.Peermanager.SendMsgToPeers(ev.Payload, peers, recovery.Message_RELAYTX)
		//go self.peerManager.SendMsgToPeers(ev.Payload,)
		case event.NewTxEvent:
			log.Debug("###### enter NewTxEvent")
			go self.sendMsg(ev.Payload)

		case event.ConsensusEvent:
			//call consensus module
			log.Debug("###### enter ConsensusEvent")
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
	msgSend, _ := proto.Marshal(msg)
	self.consenter.RecvMsg(msgSend)

}

// Broadcast consensus msg to a batch of peers not knowing about it
func (self *ProtocolManager) BroadcastConsensus(payload []byte) {
	log.Debug("begin call broadcast")
	self.Peermanager.BroadcastPeers(payload)

}

func (self *ProtocolManager) GetNodeInfo() client.PeerInfos {
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
		log.Info("receive duplicate stateupdate request, just ignore it")
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

	payload, _ := proto.Marshal(required)
	self.Peermanager.SendMsgToPeers(payload, UpdateStateMessage.Replicas, recovery.Message_SYNCCHECKPOINT)
}

func (self *ProtocolManager) ReceiveSyncRequest(ev event.StateUpdateEvent) {
	checkpointMsg := &recovery.CheckPointMessage{}
	proto.Unmarshal(ev.Payload, checkpointMsg)
	db, _ := hyperdb.GetLDBDatabase()
	blocks := &types.Blocks{}
	for i := checkpointMsg.RequiredNumber; i > checkpointMsg.CurrentNumber; i -= 1 {
		block, err := core.GetBlockByNumber(db, i)
		if err != nil {
			log.Warning("no required block number")
		}

		if blocks.Batch == nil {
			blocks.Batch = append(blocks.Batch, block)
		} else {
			blocks.Batch[0] = block

		}

		payload, _ := proto.Marshal(blocks)
		var peers []uint64
		peers = append(peers, checkpointMsg.PeerId)
		self.Peermanager.SendMsgToPeers(payload, peers, recovery.Message_SYNCBLOCK)
	}
}

func (self *ProtocolManager) ReceiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	if core.GetChainCopy().RequiredBlockNum != 0 {
		blocks := &types.Blocks{}
		proto.Unmarshal(ev.Payload, blocks)
		db, _ := hyperdb.GetLDBDatabase()
		for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {
			if blocks.Batch[i].Number <= core.GetChainCopy().RequiredBlockNum {
				log.Debug("Receive Block: ", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
				if blocks.Batch[i].Number == core.GetChainCopy().RequiredBlockNum {
					acceptHash := blocks.Batch[i].HashBlock(self.commonHash).Bytes()
					if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(core.GetChainCopy().RequireBlockHash) {
						core.PutBlockTx(db, self.commonHash, blocks.Batch[i].BlockHash, blocks.Batch[i])
						self.updateRequire(blocks.Batch[i])

						// receive all block in chain
						if core.GetChainCopy().RequiredBlockNum <= core.GetChainCopy().Height {
							lastBlk, _ := core.GetBlockByNumber(db, core.GetChainCopy().RequiredBlockNum + 1)
							if common.Bytes2Hex(lastBlk.ParentHash) == common.Bytes2Hex(core.GetChainCopy().LatestBlockHash) {
								// execute all received block at one time
								for i := core.GetChainCopy().RequiredBlockNum + 1; i <= core.GetChainCopy().RecoveryNum; i += 1 {
									blk, err := core.GetBlockByNumber(db, i)
									if err != nil {
										continue
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
								msg, _ := proto.Marshal(payload)
								msgSend := &protos.Message{
									Type:      protos.Message_STATE_UPDATED,
									Payload:   msg,
									Timestamp: time.Now().UnixNano(),
									Id:        1,
								}
								msgPayload, err := proto.Marshal(msgSend)
								if err != nil {
									log.Error(err)
								}
								time.Sleep(2000 * time.Millisecond)
								self.consenter.RecvMsg(msgPayload)
								break
							} else {
								// the highest block in local is invalid, request the block
								core.DeleteBlockByNum(db, lastBlk.Number - 1)
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
	payload, _ := proto.Marshal(required)
	self.Peermanager.SendMsgToPeers(payload, replicas, recovery.Message_SYNCSINGLE)
}

func (self *ProtocolManager) updateRequire(block *types.Block) {
	db, _ := hyperdb.GetLDBDatabase()
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > core.GetChainCopy().Height {
		if self.syncBlockCache.Contains(tmp) {
			ret, _ := self.syncBlockCache.Get(tmp)
			blks := ret.(map[string]types.Block)
			for hash, blk := range blks {
				if hash == common.BytesToHash(tmpHash).Hex() {
					core.PutBlockTx(db, self.commonHash, blk.BlockHash, &blk)
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
}

func (self *ProtocolManager) SyncReplicaStatus() {
	ticker := time.NewTicker(self.syncReplicaInterval)
	for {
		select {
		case <-ticker.C:
			addr, chain := self.packReplicaStatus()
			status := &recovery.ReplicaStatus{
				Addr:  addr,
				Chain: chain,
			}
			payload, _ := proto.Marshal(status)
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
	addrData, _ := proto.Marshal(peerAddress)
	chainData, _ := proto.Marshal(currentChain)
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
