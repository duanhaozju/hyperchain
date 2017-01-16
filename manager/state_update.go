package manager

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/blockpool"
	"hyperchain/core/types"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/protos"
	"hyperchain/recovery"
	"time"
)

// SendSyncRequest - send synchronization request to other nodes.
func (self *ProtocolManager) SendSyncRequest(ev event.SendCheckpointSyncEvent) {
	err, stateUpdateMsg, target := self.unmarshalStateUpdateMessage(ev)
	if err != nil {
		log.Errorf("invalid state update message.")
		return
	}
	log.Noticef("send sync block request to fetch missing block, current height %d, target height %d", core.GetChainCopy().Height, target.Height)
	if core.GetChainCopy().RecoveryNum >= target.Height || core.GetChainCopy().Height > target.Height {
		log.Criticalf("receive invalid state update request, just ignore it")
		return
	}

	if core.GetChainCopy().Height == target.Height {
		log.Debug("recv target height same with current chain height")
		if self.isBlockHashEqual(target.CurrentBlockHash) == true {
			log.Info("current chain latest block hash equal with target hash, send state updated event")
			self.sendStateUpdatedEvent()
		} else {
			log.Warningf("current chain latest block hash not equal with target hash, cut down local block %d", core.GetChainCopy().Height)
			self.blockPool.CutdownBlock(core.GetChainCopy().Height)
		}
	}
	// send block request message to remot peer
	required := &recovery.CheckPointMessage{
		RequiredNumber: target.Height,
		CurrentNumber:  core.GetChainCopy().Height,
		PeerId:         stateUpdateMsg.Id,
	}

	core.UpdateRequire(target.Height, target.CurrentBlockHash, target.Height)
	log.Noticef("state update, current height %d, target height %d", core.GetChainCopy().Height, target.Height)
	// save context
	core.SetReplicas(stateUpdateMsg.Replicas)
	core.SetId(stateUpdateMsg.Id)

	payload, err := proto.Marshal(required)
	if err != nil {
		log.Error("SendSyncRequest marshal message failed")
		return
	}
	self.Peermanager.SendMsgToPeers(payload, stateUpdateMsg.Replicas, recovery.Message_SYNCCHECKPOINT)
}

// ReceiveSyncRequest - receive synchronization request from some node, send back request blocks.
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

// ReceiveSyncBlocks - receive request blocks from others.
func (self *ProtocolManager) ReceiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	if core.GetChainCopy().RequiredBlockNum != 0 {
		blocks := &types.Blocks{}
		proto.Unmarshal(ev.Payload, blocks)
		db, err := hyperdb.GetLDBDatabase()
		if err != nil {
			log.Error("no database handler found")
			self.reject()
			return
		}
		for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {
			if blocks.Batch[i].Number <= core.GetChainCopy().RequiredBlockNum {

				log.Debugf("receive block #%d  hash %s", blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())

				if blocks.Batch[i].Number == core.GetChainCopy().RequiredBlockNum {
					// receive demand block
					acceptHash := blocks.Batch[i].HashBlock(self.commonHash).Bytes()
					// compare received block's hash with target hash for guarantee block's integrity
					if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(core.GetChainCopy().RequireBlockHash) {
						// put into db instead of memory for avoiding memory boom when sync huge number blocks
						core.PersistBlock(db.NewBatch(), blocks.Batch[i], self.blockPool.GetBlockVersion(), true, true)
						if err := self.updateRequire(blocks.Batch[i]); err != nil {
							log.Error("UpdateRequired failed!")
							self.reject()
							return
						}

						// receive all block in chain
						if core.GetChainCopy().RequiredBlockNum <= core.GetChainCopy().Height {
							lastBlk, err := core.GetBlockByNumber(db, core.GetChainCopy().RequiredBlockNum+1)
							if err != nil {
								log.Error("StateUpdate Failed!")
								self.reject()
								return
							}
							// check the latest block in local's correctness
							if common.Bytes2Hex(lastBlk.ParentHash) == common.Bytes2Hex(core.GetChainCopy().LatestBlockHash) {
								// execute all received block at one time
								for i := core.GetChainCopy().RequiredBlockNum + 1; i <= core.GetChainCopy().RecoveryNum; i += 1 {
									blk, err := core.GetBlockByNumber(db, i)
									if err != nil {
										log.Errorf("state update from #%d to #%d failed. current chain height #%d",
											core.GetChainCopy().RequiredBlockNum+1, core.GetChainCopy().RecoveryNum, core.GetChainCopy().Height)
										self.reject()
										return
									} else {
										// set temporary block number as block number since block number is already here
										self.blockPool.SetTempBlockNumber(blk.Number)
										self.blockPool.SetDemandNumber(blk.Number)
										self.blockPool.SetDemandSeqNo(blk.Number)
										err, result := self.blockPool.ApplyBlock(blk, blk.Number)
										if err != nil || self.assertApplyResult(blk, result) == false {
											log.Errorf("state update from #%d to #%d failed. current chain height #%d",
												core.GetChainCopy().RequiredBlockNum+1, core.GetChainCopy().RecoveryNum, core.GetChainCopy().Height)
											self.reject()
											return
										} else {
											// commit modified changes in this block and update chain.
											self.accpet(blk.Number)
										}
									}
								}
								self.blockPool.SetTempBlockNumber(core.GetChainCopy().RecoveryNum + 1)
								self.blockPool.SetDemandSeqNo(core.GetChainCopy().RecoveryNum + 1)
								self.blockPool.SetDemandNumber(core.GetChainCopy().RecoveryNum + 1)

								core.UpdateRequire(uint64(0), []byte{}, uint64(0))
								core.SetReplicas(nil)
								core.SetId(0)

								self.sendStateUpdatedEvent()
								break
							} else {
								// the highest block in local is invalid, request the block
								self.blockPool.CutdownBlock(lastBlk.Number - 1)
								self.broadcastDemandBlock(lastBlk.Number-1, lastBlk.ParentHash, core.GetReplicas(), core.GetId())
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

// broadcastDemandBlock - send block request message to others for demand block.
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

// updateRequire - update next required block number and block hash.
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
					core.PersistBlock(db.NewBatch(), &blk, self.blockPool.GetBlockVersion(), true, true)
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

// sendStateUpdatedEvent - communicate with consensus, told him state update has finish.
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
	self.blockPool.PurgeBlockCache()
	self.consenter.RecvMsg(msgPayload)
}

// isBlockHashEqual - compare block hash.
func (self *ProtocolManager) isBlockHashEqual(targetHash []byte) bool {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database handler failed in state update")
		return false
	}
	// compare current latest block and peer's block hash
	latestBlock, err := core.GetBlockByNumber(db, core.GetChainCopy().Height)
	if err != nil || latestBlock == nil || bytes.Compare(targetHash, latestBlock.BlockHash) != 0 {
		log.Warningf("missing match target blockhash and latest block's hash, target block hash %s, latest block hash %s",
			common.Bytes2Hex(targetHash), common.Bytes2Hex(latestBlock.BlockHash))
		return false
	}
	return true
}

// assertApplyResult - check apply result whether equal with other's.
func (self *ProtocolManager) assertApplyResult(block *types.Block, result *blockpool.BlockRecord) bool {
	if bytes.Compare(block.MerkleRoot, result.MerkleRoot) != 0 {
		log.Warningf("mismatch in block merkle root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.MerkleRoot), common.Bytes2Hex(result.MerkleRoot))
		return false
	}
	if bytes.Compare(block.TxRoot, result.TxRoot) != 0 {
		log.Warningf("mismatch in block transaction root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.TxRoot), common.Bytes2Hex(result.TxRoot))
		return false

	}
	if bytes.Compare(block.ReceiptRoot, result.ReceiptRoot) != 0 {
		log.Warningf("mismatch in block receipt root  of #%d, required %s, got %s",
			block.Number, common.Bytes2Hex(block.ReceiptRoot), common.Bytes2Hex(result.ReceiptRoot))
		return false
	}
	return true
}

// reject - reject state update result.
func (self *ProtocolManager) reject() {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Error("get database handler failed.")
		return
	}
	batch := db.NewBatch()
	for i := core.GetChainCopy().Height + 1; i <= core.GetChainCopy().RecoveryNum; i += 1 {
		// delete persisted blocks number larger than chain height
		core.DeleteBlockByNum(batch, i)
	}
	batch.Write()
	self.blockPool.SetDemandNumber(core.GetChainCopy().Height + 1)
	self.blockPool.SetDemandSeqNo(core.GetChainCopy().Height + 1)
	self.blockPool.SetTempBlockNumber(core.GetChainCopy().Height + 1)
	// reset state in block pool
	self.blockPool.ClearStateUnCommitted()
	// reset
	core.UpdateRequire(uint64(0), []byte{}, uint64(0))
	core.SetReplicas(nil)
	core.SetId(0)
	self.sendStateUpdatedEvent()
}

// accpet - accept state update result.
func (self *ProtocolManager) accpet(seqNo uint64) {
	self.blockPool.SubmitForStateUpdate(seqNo)
}

// unmarshalStateUpdateMessage - unmarshal state update message sent from consensus module and return a state update target.
func (self *ProtocolManager) unmarshalStateUpdateMessage(ev event.SendCheckpointSyncEvent) (error, *protos.UpdateStateMessage, *protos.BlockchainInfo) {
	updateStateMessage := &protos.UpdateStateMessage{}
	err := proto.Unmarshal(ev.Payload, updateStateMessage)
	if err != nil {
		log.Errorf("unmarshal state update message failed. %s", err)
		return err, nil, nil
	}
	blockChainInfo := &protos.BlockchainInfo{}
	err = proto.Unmarshal(updateStateMessage.TargetId, blockChainInfo)
	if err != nil {
		log.Errorf("unmarshal block chain info failed. %s", err)
		return err, nil, nil
	}
	return nil, updateStateMessage, blockChainInfo
}