package executor

import (
	"hyperchain/event"
	"hyperchain/recovery"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"hyperchain/protos"
	"time"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/common"
)

// SendSyncRequest - send synchronization request to other nodes.
func (executor *Executor) sendSyncRequest(ev event.SendCheckpointSyncEvent) {
	err, stateUpdateMsg, target := executor.unmarshalStateUpdateMessage(ev)
	if err != nil {
		log.Errorf("[Namespace = %s] invalid state update message.", executor.namespace)
		return
	}
	log.Noticef("[Namespace = %s] send sync block request to fetch missing block, current height %d, target height %d", executor.namespace, edb.GetHeightOfChain(executor.namespace), target.Height)

	if executor.status.syncFlag.SyncRecoveryNum >= target.Height || edb.GetHeightOfChain(executor.namespace) > target.Height {
		log.Criticalf("[Namespace = %s] receive invalid state update request, just ignore it", executor.namespace)
		return
	}

	if edb.GetHeightOfChain(executor.namespace) == target.Height {
		log.Debugf("[Namespace = %s] recv target height same with current chain height", executor.namespace)
		if executor.isBlockHashEqual(target.CurrentBlockHash) == true {
			log.Infof("[Namespace = %s] current chain latest block hash equal with target hash, send state updated event", executor.namespace)
			executor.sendStateUpdatedEvent()
		} else {
			log.Warningf("[Namespace = %s] current chain latest block hash not equal with target hash, cut down local block %d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
			if err := executor.CutdownBlock(edb.GetHeightOfChain(executor.namespace)); err != nil {
				log.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, edb.GetHeightOfChain(executor.namespace))
				return
			}
		}
	}
	// send block request message to remot peer
	required := &recovery.CheckPointMessage{
		RequiredNumber: target.Height,
		CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
		PeerId:         stateUpdateMsg.Id,
	}

	executor.UpdateRequire(target.Height, target.CurrentBlockHash, target.Height)
	log.Noticef("[Namespace = %s] state update, current height %d, target height %d", executor.namespace, edb.GetHeightOfChain(executor.namespace), target.Height)
	// save context
	executor.status.syncFlag.SendReplicas = stateUpdateMsg.Replicas
	executor.status.syncFlag.LocalId = stateUpdateMsg.Id

	payload, err := proto.Marshal(required)
	if err != nil {
		log.Errorf("[Namespace = %s] SendSyncRequest marshal message failed", executor.namespace)
		return
	}
	executor.peerManager.SendMsgToPeers(payload, stateUpdateMsg.Replicas, recovery.Message_SYNCCHECKPOINT)
}

// ReceiveSyncRequest - receive synchronization request from some node, send back request blocks.
func (executor *Executor) receiveSyncRequest(ev event.StateUpdateEvent) {
	checkpointMsg := &recovery.CheckPointMessage{}
	proto.Unmarshal(ev.Payload, checkpointMsg)
	blocks := &types.Blocks{}
	for i := checkpointMsg.RequiredNumber; i > checkpointMsg.CurrentNumber; i -= 1 {
		block, err := edb.GetBlockByNumber(executor.namespace, i)
		if err != nil {
			log.Errorf("[Namespace = %s] no required block number: %d", executor.namespace, i)
			continue
		}

		if blocks.Batch == nil {
			blocks.Batch = append(blocks.Batch, block)
		} else {
			blocks.Batch[0] = block
		}

		payload, err := proto.Marshal(blocks)
		if err != nil {
			log.Errorf("[Namespace = %s] ReceiveSyncRequest marshal message failed", executor.namespace)
			continue
		}
		var peer []uint64
		peer = append(peer, checkpointMsg.PeerId)
		executor.peerManager.SendMsgToPeers(payload, peer, recovery.Message_SYNCBLOCK)
	}
}

// ReceiveSyncBlocks - receive request synchronization blocks from others.
func (executor *Executor) receiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	if executor.status.syncFlag.SyncRequireBlockNum != 0 {
		blocks := &types.Blocks{}
		proto.Unmarshal(ev.Payload, blocks)
		db, err := hyperdb.GetDBDatabase()
		if err != nil {
			log.Errorf("[Namespace = %s] no database handler found", executor.namespace)
			executor.reject()
			return
		}
		// store blocks into database only, not process them.
		for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {
			if blocks.Batch[i].Number <= executor.status.syncFlag.SyncRequireBlockNum {
				log.Debugf("[Namespace = %s] receive block #%d  hash %s", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())

				if blocks.Batch[i].Number == executor.status.syncFlag.SyncRequireBlockNum {
					// receive demand block
					acceptHash := blocks.Batch[i].HashBlock(executor.commonHash).Bytes()
					// compare received block's hash with target hash for guarantee block's integrity
					if common.Bytes2Hex(acceptHash) == common.Bytes2Hex(executor.status.syncFlag.SyncRequireBlockHash) {
						// put into db instead of memory for avoiding memory boom when sync huge number blocks
						edb.PersistBlock(db.NewBatch(), blocks.Batch[i], true, true)
						// updateRequire update require info
						if err := executor.updateRequire(blocks.Batch[i]); err != nil {
							log.Errorf("[Namespace = %s] UpdateRequired failed!", executor.namespace)
							executor.reject()
							return
						}
					}

				} else {
					// requested block with smaller number arrive earlier than expected
					// store in cache temporarily
					log.Debug("[Namespace = %s] Receive Block earily: ", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
					ret, existed := executor.syncBlockCache.Get(blocks.Batch[i].Number)
					if existed {
						blks := ret.(map[string]types.Block)
						if _, ok := blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()]; ok {
							log.Noticef("[Namespace = %s] Receive Duplicate Block: %d %s", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
							continue
						} else {
							log.Debugf("[Namespace = %s] Receive Sync Block with different hash: %d %s", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
							blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()] = *blocks.Batch[i]
							executor.syncBlockCache.Add(blocks.Batch[i].Number, blks)
						}
					} else {
						blks := make(map[string]types.Block)
						blks[common.BytesToHash(blocks.Batch[i].BlockHash).Hex()] = *blocks.Batch[i]
						executor.syncBlockCache.Add(blocks.Batch[i].Number, blks)
					}
				}
			}
		}
		executor.processSyncBlocks()
	}
}

func (executor *Executor) processSyncBlocks() {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Errorf("[Namespace = %s] no database handler found", executor.namespace)
		executor.reject()
		return
	}

	if executor.status.syncFlag.SyncRequireBlockNum <= edb.GetHeightOfChain(executor.namespace) {
		// get the first of SyncBlocks
		lastBlk, err := edb.GetBlockByNumber(db, executor.status.syncFlag.SyncRequireBlockNum +1)
		if err != nil {
			log.Errorf("[Namespace = %s] StateUpdate Failed!", executor.namespace)
			executor.reject()
			return
		}
		// check the latest block in local's correctness
		if common.Bytes2Hex(lastBlk.ParentHash) == common.Bytes2Hex(edb.GetLatestBlockHash(executor.namespace)) {

			executor.waitUtilSyncAvailable()
			defer executor.syncDone()
			// execute all received block at one time
			for i := executor.status.syncFlag.SyncRequireBlockNum + 1; i <= executor.status.syncFlag.SyncRecoveryNum; i += 1 {
				blk, err := edb.GetBlockByNumber(executor.namespace, i)
				if err != nil {
					log.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
						executor.namespace, executor.status.syncFlag.SyncRequireBlockNum +1, executor.status.syncFlag.SyncRecoveryNum, edb.GetHeightOfChain(executor.namespace))
					executor.reject()
					return
				} else {
					// set temporary block number as block number since block number is already here
					executor.setTempBlockNumber(blk.Number)
					executor.setDemandNumber(blk.Number)
					executor.setDemandSeqNo(blk.Number)
					err, result := executor.ApplyBlock(blk, blk.Number)
					if err != nil || executor.assertApplyResult(blk, result) == false {
						log.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
							executor.namespace, executor.status.syncFlag.SyncRequireBlockNum +1, executor.status.syncFlag.SyncRecoveryNum, edb.GetHeightOfChain(executor.namespace))
						executor.reject()
						return
					} else {
						// commit modified changes in this block and update chain.
						executor.accpet(blk.Number)
					}
				}
			}
			executor.setTempBlockNumber(executor.status.syncFlag.SyncRecoveryNum + 1)
			executor.setDemandSeqNo(executor.status.syncFlag.SyncRecoveryNum + 1)
			executor.setDemandNumber(executor.status.syncFlag.SyncRecoveryNum + 1)

			executor.UpdateRequire(uint64(0), []byte{}, uint64(0))
			executor.status.syncFlag.SendReplicas = nil
			executor.status.syncFlag.LocalId = 0

			executor.sendStateUpdatedEvent()
		} else {
			// the highest block in local is invalid, request the block
			if err := executor.CutdownBlock(lastBlk.Number - 1); err != nil {
				log.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, lastBlk.Number - 1)
				return
			}
			executor.broadcastDemandBlock(lastBlk.Number-1, lastBlk.ParentHash, executor.status.syncFlag.SendReplicas, executor.status.syncFlag.LocalId)
		}
	}
}

// broadcastDemandBlock - send block request message to others for demand block.
func (executor *Executor) broadcastDemandBlock(number uint64, hash []byte, replicas []uint64, peerId uint64) {
	required := &recovery.CheckPointMessage{
		RequiredNumber: number,
		CurrentNumber:  edb.GetHeightOfChain(executor.namespace),
		PeerId:         peerId,
	}
	payload, err := proto.Marshal(required)
	if err != nil {
		log.Errorf("[Namespace = %s] broadcastDemandBlock, marshal message failed", executor.namespace)
		return
	}
	executor.peerManager.SendMsgToPeers(payload, replicas, recovery.Message_SYNCSINGLE)
}

// updateRequire - update next required block number and block hash.
func (executor *Executor) updateRequire(block *types.Block) error {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		// TODO
		log.Errorf("[Namespace = %s] updateRequire get database failed", executor.namespace)
		return err
	}
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > edb.GetHeightOfChain(executor.namespace) {
		if executor.syncBlockCache.Contains(tmp) {
			ret, _ := executor.syncBlockCache.Get(tmp)
			blks := ret.(map[string]types.Block)
			for hash, blk := range blks {
				if hash == common.BytesToHash(tmpHash).Hex() {
					edb.PersistBlock(db.NewBatch(), &blk, true, true)
					executor.syncBlockCache.Remove(tmp)
					tmp = tmp - 1
					tmpHash = blk.ParentHash
					flag = true
					log.Debugf("[Namespace = %s] process sync block(block number = %d) stored in cache", executor.namespace, blk.Number)
					break
				} else {
					log.Debugf("[Namespace = %s] found Invalid sync block, discard block number %d, block hash %s\n", executor.namespace, blk.Number, common.BytesToHash(blk.BlockHash).Hex())
				}
			}
			if flag {
				flag = false
			} else {
				executor.syncBlockCache.Remove(tmp)
				break
			}
		} else {
			break
		}
	}
	executor.UpdateRequire(tmp, tmpHash, executor.status.syncFlag.SyncRecoveryNum)
	log.Debugf("[Namespace = %s] Next Required %d %s", executor.namespace, executor.status.syncFlag.SyncRequireBlockNum, common.BytesToHash(executor.status.syncFlag.SyncRequireBlockHash).Hex())
	return nil
}


func (executor *Executor) UpdateRequire(num uint64, hash []byte, recoveryNum uint64) error {
	executor.status.syncFlag.SyncRequireBlockNum = num
	executor.status.syncFlag.SyncRequireBlockHash = hash
	executor.status.syncFlag.SyncRecoveryNum = recoveryNum
	return nil
}

// sendStateUpdatedEvent - communicate with consensus, told him state update has finish.
func (executor *Executor) sendStateUpdatedEvent() {
	// state update success
	payload := &protos.StateUpdatedMessage{
		SeqNo: edb.GetHeightOfChain(executor.namespace),
	}
	msg, err := proto.Marshal(payload)
	if err != nil {
		log.Errorf("[Namespace = %s] StateUpdate marshal stateupdated message failed", executor.namespace)
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
	log.Debug("[Namespace = %s] send state updated message", executor.namespace)
	time.Sleep(2 * time.Second)
	// IMPORTANT clear block cache of blockpool
	executor.PurgeCache()
	executor.consenter.RecvMsg(msgPayload)
}

// reject - reject state update result.
func (executor *Executor) reject() {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error("get database handler failed.")
		return
	}
	batch := db.NewBatch()
	for i := edb.GetHeightOfChain(executor.namespace) + 1; i <= executor.status.syncFlag.SyncRecoveryNum; i += 1 {
		// delete persisted blocks number larger than chain height
		edb.DeleteBlockByNum(executor.namespace, batch, i, false, false)
	}
	batch.Write()
	executor.setDemandNumber(edb.GetHeightOfChain(executor.namespace) + 1)
	executor.setDemandSeqNo(edb.GetHeightOfChain(executor.namespace) + 1)
	executor.setTempBlockNumber(edb.GetHeightOfChain(executor.namespace) + 1)
	// reset state in block pool
	executor.ClearStateUnCommitted()
	// reset
	executor.UpdateRequire(uint64(0), []byte{}, uint64(0))
	executor.status.syncFlag.SendReplicas = nil
	executor.status.syncFlag.LocalId = 0
	executor.sendStateUpdatedEvent()
}