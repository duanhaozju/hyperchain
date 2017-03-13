package executor

import (
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	edb "hyperchain/core/db_utils"
	"github.com/golang/protobuf/proto"
	"hyperchain/event"
	"hyperchain/protos"
	"bytes"
	"hyperchain/recovery"
	"hyperchain/hyperdb"
)

// SendSyncRequest - send synchronization request to other nodes.
func (executor *Executor) SendSyncRequest(ev event.SendCheckpointSyncEvent) {
	err, stateUpdateMsg, target := executor.unmarshalStateUpdateMessage(ev)
	if err != nil {
		log.Errorf("[Namespace = %s] invalid state update message.", executor.namespace)
		executor.reject()
		return
	}
	log.Noticef("[Namespace = %s] send sync block request to fetch missing block, current height %d, target height %d", executor.namespace, edb.GetHeightOfChain(executor.namespace), target.Height)

	if executor.status.syncFlag.SyncTarget >= target.Height || edb.GetHeightOfChain(executor.namespace) > target.Height {
		log.Errorf("[Namespace = %s] receive invalid state update request, just ignore it", executor.namespace)
		executor.reject()
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
				executor.reject()
				return
			}
		}
	}

	executor.updateSyncFlag(target.Height, target.CurrentBlockHash, target.Height)
	executor.recordSyncPeers(stateUpdateMsg.Replicas, stateUpdateMsg.Id)
	if err := executor.informP2P(NOTIFY_BROADCAST_DEMAND, nil); err != nil {
		log.Errorf("[Namespace = %s] send sync req failed.", executor.namespace)
		executor.reject()
		return
	}
}

// ReceiveSyncRequest - receive synchronization request from some nodes, and send back request blocks.
func (executor *Executor) ReceiveSyncRequest(ev event.StateUpdateEvent) {
	syncReqMsg := &recovery.CheckPointMessage{}
	proto.Unmarshal(ev.Payload, syncReqMsg)
	for i := syncReqMsg.RequiredNumber; i > syncReqMsg.CurrentNumber; i -= 1 {
		executor.informP2P(NOTIFY_UNICAST_BLOCK, i, syncReqMsg.PeerId)
	}
}

// ReceiveSyncBlocks - receive request synchronization blocks from others.
func (executor *Executor) ReceiveSyncBlocks(ev event.ReceiveSyncBlockEvent) {
	if executor.status.syncFlag.SyncDemandBlockNum != 0 {
		blocks := &types.Blocks{}
		proto.Unmarshal(ev.Payload, blocks)
		db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
		if err != nil {
			log.Errorf("[Namespace = %s] no database handler found", executor.namespace)
			executor.reject()
			return
		}
		// store blocks into database only, not process them.
		for i := len(blocks.Batch) - 1; i >= 0; i -= 1 {
			if !executor.verifyBlockIntegrity(blocks.Batch[i]) {
				log.Warningf("[Namespace = %s] receive a broken block %d, drop it", executor.namespace, blocks.Batch[i].Number)
				continue
			}
			if blocks.Batch[i].Number <= executor.status.syncFlag.SyncDemandBlockNum {
				log.Debugf("[Namespace = %s] receive block #%d  hash %s", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
				// is demand
				if executor.isDemandSyncBlock(blocks.Batch[i]) {
					edb.PersistBlock(db.NewBatch(), blocks.Batch[i], true, true)
					if err := executor.updateSyncDemand(blocks.Batch[i]); err != nil {
						log.Errorf("[Namespace = %s] update sync demand failed.", executor.namespace)
						executor.reject()
						return
					}
				} else {
					// requested block with smaller number arrive earlier than expected
					// store in cache temporarily
					log.Debugf("[Namespace = %s] receive block #%d hash %s earily", executor.namespace, blocks.Batch[i].Number, common.BytesToHash(blocks.Batch[i].BlockHash).Hex())
					executor.addToSyncCache(blocks.Batch[i])
				}
			}
		}
		executor.processSyncBlocks()
	}
}

// ApplyBlock - apply all transactions in block into state during the `state update` process.
func (executor *Executor) ApplyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	if block.Transactions == nil {
		return EmptyPointerErr, nil
	}
	return executor.applyBlock(block, seqNo)
}

func (executor *Executor) applyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	executor.initTransactionHashCalculator()
	executor.initReceiptHashCalculator()

	batch := executor.statedb.FetchBatch(seqNo)
	executor.statedb.MarkProcessStart(executor.getTempBlockNumber())
	// initialize execution environment rule set
	env := initEnvironment(executor.statedb, executor.getTempBlockNumber())
	// execute transaction one by one
	for i, tx := range block.Transactions {
		executor.statedb.StartRecord(tx.GetHash(), common.Hash{}, i)
		receipt, _, _, err := core.ExecTransaction(tx, env)
		// just ignore invalid transactions
		if err != nil {
			log.Warningf("invalid transaction found during the state update process in #%d", seqNo)
			continue
		}
		executor.calculateTransactionsFingerprint(tx, false)
		executor.calculateReceiptFingerprint(receipt, false)

		// different with normal process, because during the state update, block number and seqNo are always same
		// persist transaction here
		if err, _ := edb.PersistTransaction(batch, tx, false, false); err != nil {
			log.Errorf("persist transaction for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist transaction meta data
		meta := &types.TransactionMeta{
			BlockIndex: seqNo,
			Index:      int64(i),
		}
		if err := edb.PersistTransactionMeta(batch, meta, tx.GetHash(), false, false); err != nil {
			log.Errorf("persist transaction meta for index %d in #%d failed.", i, seqNo)
			continue
		}
		// persist receipt
		if err, _ := edb.PersistReceipt(batch, receipt, false, false); err != nil {
			log.Errorf("persist receipt for index %d in #%d failed.", i, seqNo)
			continue
		}
	}
	// submit validation result
	err, merkleRoot, txRoot, receiptRoot := executor.submitValidationResult(batch)
	if err != nil {
		log.Error("submit validation result failed.", err.Error())
		return err, nil
	}
	log.Debugf("validate result temp block number #%d, vid #%d, merkle root [%s],  transaction root [%s],  receipt root [%s]",
		executor.getTempBlockNumber(), seqNo, common.Bytes2Hex(merkleRoot), common.Bytes2Hex(txRoot), common.Bytes2Hex(receiptRoot))
	return nil, &ValidationResultRecord{
		TxRoot:      txRoot,
		ReceiptRoot: receiptRoot,
		MerkleRoot:  merkleRoot,
	}
}

// ClearStateUnCommitted - remove all cached stuff
func (executor *Executor) clearStatedb() {
	executor.statedb.Purge()
}

// unmarshalStateUpdateMessage - unmarshal block synchronization message sent from consensus module and return a block synchronization target.
func (executor *Executor) unmarshalStateUpdateMessage(ev event.SendCheckpointSyncEvent) (error, *protos.UpdateStateMessage, *protos.BlockchainInfo) {
	updateStateMessage := &protos.UpdateStateMessage{}
	err := proto.Unmarshal(ev.Payload, updateStateMessage)
	if err != nil {
		log.Errorf("[Namespace = %s] unmarshal state update message failed. %s", executor.namespace, err)
		return err, nil, nil
	}
	blockChainInfo := &protos.BlockchainInfo{}
	err = proto.Unmarshal(updateStateMessage.TargetId, blockChainInfo)
	if err != nil {
		log.Errorf("[Namespace = %s] unmarshal block chain info failed. %s", executor.namespace, err)
		return err, nil, nil
	}
	return nil, updateStateMessage, blockChainInfo
}

// assertApplyResult - check apply result whether equal with other's.
func (executor *Executor) assertApplyResult(block *types.Block, result *ValidationResultRecord) bool {
	if bytes.Compare(block.MerkleRoot, result.MerkleRoot) != 0 {
		log.Warningf("[Namespace = %s] mismatch in block merkle root  of #%d, demand %s, got %s",
			executor.namespace, block.Number, common.Bytes2Hex(block.MerkleRoot), common.Bytes2Hex(result.MerkleRoot))
		return false
	}
	if bytes.Compare(block.TxRoot, result.TxRoot) != 0 {
		log.Warningf("[Namespace = %s] mismatch in block transaction root  of #%d, demand %s, got %s",
			block.Number, common.Bytes2Hex(block.TxRoot), common.Bytes2Hex(result.TxRoot))
		return false

	}
	if bytes.Compare(block.ReceiptRoot, result.ReceiptRoot) != 0 {
		log.Warningf("[Namespace = %s] mismatch in block receipt root  of #%d, demand %s, got %s",
			executor.namespace, block.Number, common.Bytes2Hex(block.ReceiptRoot), common.Bytes2Hex(result.ReceiptRoot))
		return false
	}
	return true
}

// isBlockHashEqual - compare block hash.
func (executor *Executor) isBlockHashEqual(targetHash []byte) bool {
	// compare current latest block and peer's block hash
	latestBlock, err := edb.GetBlockByNumber(executor.namespace, edb.GetHeightOfChain(executor.namespace))
	if err != nil || latestBlock == nil || bytes.Compare(targetHash, latestBlock.BlockHash) != 0 {
		log.Warningf("[Namespace = %s] missing match target blockhash and latest block's hash, target block hash %s, latest block hash %s",
			executor.namespace, common.Bytes2Hex(targetHash), common.Bytes2Hex(latestBlock.BlockHash))
		return false
	}
	return true
}

// processSyncBlocks - execute all received block one by one.
func (executor *Executor) processSyncBlocks() {
	if executor.status.syncFlag.SyncDemandBlockNum <= edb.GetHeightOfChain(executor.namespace) {
		// get the first of SyncBlocks
		lastBlk, err := edb.GetBlockByNumber(executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1)
		if err != nil {
			log.Errorf("[Namespace = %s] StateUpdate Failed!", executor.namespace)
			executor.reject()
			return
		}
		// check the latest block in local's correctness
		if bytes.Compare(lastBlk.ParentHash, edb.GetLatestBlockHash(executor.namespace)) == 0  {
			executor.waitUtilSyncAvailable()
			defer executor.syncDone()
			// execute all received block at one time
			for i := executor.status.syncFlag.SyncDemandBlockNum + 1; i <= executor.status.syncFlag.SyncTarget; i += 1 {
				blk, err := edb.GetBlockByNumber(executor.namespace, i)
				if err != nil {
					log.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
						executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1, executor.status.syncFlag.SyncTarget, edb.GetHeightOfChain(executor.namespace))
					executor.reject()
					return
				} else {
					// set temporary block number as block number since block number is already here
					executor.initDemand(blk.Number)
					err, result := executor.ApplyBlock(blk, blk.Number)
					if err != nil || executor.assertApplyResult(blk, result) == false {
						log.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
							executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1, executor.status.syncFlag.SyncTarget, edb.GetHeightOfChain(executor.namespace))
						executor.reject()
						return
					} else {
						// commit modified changes in this block and update chain.
						executor.accpet(blk.Number)
					}
				}
			}
			executor.initDemand(executor.status.syncFlag.SyncTarget + 1)
			executor.clearSyncFlag()
			executor.sendStateUpdatedEvent()
		} else {
			// the highest block in local is invalid, request the block
			if err := executor.CutdownBlock(lastBlk.Number - 1); err != nil {
				log.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, lastBlk.Number - 1)
				executor.reject()
				return
			}
			executor.SendSyncRequestForSingle(lastBlk.Number - 1)
		}
	}
}

// broadcastDemandBlock - send block request message to others for demand block.
func (executor *Executor) SendSyncRequestForSingle(number uint64) {
	executor.informP2P(NOTIFY_BROADCAST_SINGLE, number)
}

// updateSyncDemand - update next demand block number and block hash.
func (executor *Executor) updateSyncDemand(block *types.Block) error {
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Errorf("[Namespace = %s] get database failed", executor.namespace)
		return err
	}
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > edb.GetHeightOfChain(executor.namespace) {
		if executor.cache.syncCache.Contains(tmp) {
			blks, _ := executor.fetchFromSyncCache(tmp)
			for hash, blk := range blks {
				if hash == common.BytesToHash(tmpHash).Hex() {
					edb.PersistBlock(db.NewBatch(), &blk, true, true)
					executor.cache.syncCache.Remove(tmp)
					tmp = tmp - 1
					tmpHash = blk.ParentHash
					flag = true
					log.Debugf("[Namespace = %s] process sync block(block number = %d) stored in cache", executor.namespace, blk.Number)
					break
				} else {
					log.Debugf("[Namespace = %s] found invalid sync block, discard block number %d, block hash %s", executor.namespace, blk.Number, common.BytesToHash(blk.BlockHash).Hex())
				}
			}
			if flag {
				flag = false
			} else {
				executor.cache.syncCache.Remove(tmp)
				break
			}
		} else {
			break
		}
	}
	executor.updateSyncFlag(tmp, tmpHash, executor.status.syncFlag.SyncTarget)
	log.Debugf("[Namespace = %s] Next Demand %d %s", executor.namespace, executor.status.syncFlag.SyncDemandBlockNum, common.BytesToHash(executor.status.syncFlag.SyncDemandBlockHash).Hex())
	return nil
}

// sendStateUpdatedEvent - communicate with consensus, told it state update has finished.
func (executor *Executor) sendStateUpdatedEvent() {
	// state update success
	executor.PurgeCache()
	executor.informConsensus(NOTIFY_SYNC_DONE, nil)
}

// accpet - accept block synchronization result.
func (executor *Executor) accpet(seqNo uint64) {
	batch := executor.statedb.FetchBatch(seqNo)
	edb.UpdateChainByBlcokNum(executor.namespace, batch, seqNo, false, false)
	batch.Write()
	executor.statedb.MarkProcessFinish(seqNo)
}

// reject - reject state update result.
func (executor *Executor) reject() {
	executor.cache.syncCache.Purge()
	db, err := hyperdb.GetDBDatabaseByNamespace(executor.namespace)
	if err != nil {
		log.Error("get database handler failed.")
		return
	}
	batch := db.NewBatch()
	for i := edb.GetHeightOfChain(executor.namespace) + 1; i <= executor.status.syncFlag.SyncTarget; i += 1 {
		// delete persisted blocks number larger than chain height
		edb.DeleteBlockByNum(executor.namespace, batch, i, false, false)
	}
	batch.Write()
	executor.initDemand(edb.GetHeightOfChain(executor.namespace) + 1)
	executor.clearStatedb()
	executor.clearSyncFlag()
	executor.sendStateUpdatedEvent()
}

// verifyBlockIntegrity - make sure block content doesn't change.
func (executor *Executor) verifyBlockIntegrity(block *types.Block) bool {
	if bytes.Compare(block.BlockHash, block.Hash(executor.commonHash).Bytes()) == 0 {
		return true
	}
	return false
}

// isDemandSyncBlock - check whether is the demand sync block.
func (executor *Executor) isDemandSyncBlock(block *types.Block) bool {
	if block.Number == executor.status.syncFlag.SyncDemandBlockNum &&
		bytes.Compare(block.BlockHash, executor.status.syncFlag.SyncDemandBlockHash) == 0 {
		return true
	}
	return false
}
