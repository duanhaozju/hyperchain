package manager

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/recovery"
	"hyperchain/hyperdb"
	"time"
	"hyperchain/event"
	"hyperchain/protos"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/common"
	"bytes"
)


// Entry of state update
func (self *ProtocolManager) SendSyncRequest(ev event.SendCheckpointSyncEvent) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("Get database handle failed")
		return
	}
	UpdateStateMessage := &protos.UpdateStateMessage{}
	proto.Unmarshal(ev.Payload, UpdateStateMessage)
	blockChainInfo := &protos.BlockchainInfo{}
	proto.Unmarshal(UpdateStateMessage.TargetId, blockChainInfo)
	if core.GetChainCopy().RecoveryNum >= blockChainInfo.Height || core.GetChainCopy().Height > blockChainInfo.Height {
		log.Info("receive invalid state update request, just ignore it")
		return
	}
	if core.GetChainCopy().Height == blockChainInfo.Height {
		// compare current latest block and peer's block hash
		latestBlock, err := core.GetBlockByNumber(db, core.GetChainCopy().Height)
		if err != nil || latestBlock == nil || bytes.Compare(blockChainInfo.CurrentBlockHash, latestBlock.BlockHash) != 0 {
			log.Infof("missing match target blockhash and latest block's hash, target block hash %s, latest block hash %s",
				common.Bytes2Hex(blockChainInfo.CurrentBlockHash), common.Bytes2Hex(latestBlock.BlockHash))
			// cut down block to latest stable checkpoint
			self.blockPool.CutdownBlock(blockChainInfo.Height)
			// update demand number and demand seq no
			self.blockPool.SetDemandNumber(blockChainInfo.Height)
			self.blockPool.SetDemandSeqNo(blockChainInfo.Height)
			// update chain
			core.UpdateChainByBlcokNum(db, blockChainInfo.Height-1)
		} else {
			log.Info("match target blockhash and latest block's hash")
			self.sendStateUpdatedEvent()
			return
		}
	}
	// send block request message to remote peer
	required := &recovery.CheckPointMessage{
		RequiredNumber: blockChainInfo.Height,
		CurrentNumber:  core.GetChainCopy().Height,
		PeerId:         UpdateStateMessage.Id,
	}

	core.UpdateRequire(blockChainInfo.Height, blockChainInfo.CurrentBlockHash, blockChainInfo.Height)
	log.Noticef("state update, current height %d, target height %d", core.GetChainCopy().Height, blockChainInfo.Height)
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
								self.sendStateUpdatedEvent()
								break
							} else {
								// the highest block in local is invalid, request the block
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

func (self *ProtocolManager) sendStateUpdatedEvent() {
	// state update success
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
	log.Debug("send state updated message")
	time.Sleep(2 * time.Second)
	// IMPORTANT clear block cache of blockpool
	self.blockPool.PurgeValidateQueue()
	self.consenter.RecvMsg(msgPayload)
}
