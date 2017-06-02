package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	edb "hyperchain/core/db_utils"
	"github.com/golang/protobuf/proto"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"
	"bytes"
	"time"
	"hyperchain/core/vm"
	"io/ioutil"
	"hyperchain/hyperdb"
	cmd "os/exec"
	"path"
	"path/filepath"
)

func (executor *Executor) SyncChain(ev event.ChainSyncReqEvent) {
	executor.logger.Noticef("[Namespace = %s] send sync block request to fetch missing block, current height %d, target height %d", executor.namespace, edb.GetHeightOfChain(executor.namespace), ev.TargetHeight)
	if executor.status.syncFlag.SyncTarget >= ev.TargetHeight || edb.GetHeightOfChain(executor.namespace) > ev.TargetHeight {
		executor.logger.Errorf("[Namespace = %s] receive invalid state update request, just ignore it", executor.namespace)
		executor.reject()
		return
	}

	if edb.GetHeightOfChain(executor.namespace) == ev.TargetHeight {
		executor.logger.Debugf("[Namespace = %s] recv target height same with current chain height", executor.namespace)
		if executor.isBlockHashEqual(ev.TargetBlockHash) == true {
			executor.logger.Infof("[Namespace = %s] current chain latest block hash equal with target hash, send state updated event", executor.namespace)
			executor.sendStateUpdatedEvent()
		} else {
			executor.logger.Warningf("[Namespace = %s] current chain latest block hash not equal with target hash, cut down local block %d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
			if err := executor.CutdownBlock(edb.GetHeightOfChain(executor.namespace)); err != nil {
				executor.logger.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, edb.GetHeightOfChain(executor.namespace))
				executor.reject()
				return
			}
		}
	}
	executor.syncInitialize(ev)
	executor.SendSyncRequest(ev.TargetHeight, executor.calcuDownstream())
	go executor.syncChainResendBackend()
}

func (executor *Executor) syncChainResendBackend() {
	ticker := time.NewTicker(executor.GetSyncResendInterval())
	up, down := executor.getSyncReqArgs()
	for {
		select {
		case <- executor.status.syncFlag.ResendExit:
			return
		case <-ticker.C:
		        // resend
			if executor.status.syncCtx.GetResendMode() == ResendMode_Block {
				curUp, curDown := executor.getSyncReqArgs()
				if curUp == up && curDown == down && !executor.isSyncInExecution() {
					executor.logger.Noticef("resend sync request. want [%d] - [%d]", down, executor.status.syncFlag.SyncDemandBlockNum)
					executor.status.syncFlag.Oracle.FeedBack(false)
					executor.status.syncCtx.SetCurrentPeer(executor.status.syncFlag.Oracle.SelectPeer())
					// TODO change peer may triggle context switch
					executor.SendSyncRequest(executor.status.syncFlag.SyncDemandBlockNum, down)
					executor.recordSyncReqArgs(curUp, curDown)
				} else {
					up = curUp
					down = curDown
				}
			} else if executor.status.syncCtx.GetResendMode() == ResendMode_WorldState {
				// TODO different resend strategy for world state
			}
		}
	}
}

// ReceiveSyncRequest - receive synchronization request from some nodes, and send back request blocks.
func (executor *Executor) ReceiveSyncRequest(payload []byte) {
	var request ChainSyncRequest
	if err := proto.Unmarshal(payload, &request); err != nil {
		executor.logger.Error("unmarshal sync request failed.")
		return
	}
	for i := request.RequiredNumber; i > request.CurrentNumber; i -= 1 {
		executor.informP2P(NOTIFY_UNICAST_BLOCK, i, request.PeerId)
	}
}

func (executor *Executor) ReceiveWorldStateSyncRequest(payload []byte) {
	var request WorldStateSyncRequest
	if err := proto.Unmarshal(payload, &request); err != nil {
		executor.logger.Warning("unmarshal world state sync request failed.")
		return
	}
	err, manifest := executor.snapshotReg.rwc.Search(request.Target)
	if err != nil {
		executor.logger.Warning("required snapshot doesn't exist")
		return
	}
	if err := executor.snapshotReg.CompressSnapshot(manifest.FilterId); err != nil {
		executor.logger.Warning("compress snapshot failed")
		return
	}

	if err := executor.informP2P(NOTIFY_SEND_WORLD_STATE, executor.snapshotReg.CompressedSnapshotPath(manifest.FilterId), request.PeerId); err != nil {
		executor.logger.Warningf("send world state (#%s) back to (%d) failed, err msg %s", manifest.FilterId, request.PeerId, err.Error())
		return
	}
	executor.logger.Noticef("send world state (#%s) back to (%d) success", manifest.FilterId, request.PeerId)
}

// ReceiveSyncBlocks - receive request synchronization blocks from others.
func (executor *Executor) ReceiveSyncBlocks(payload []byte) {
	if executor.status.syncFlag.SyncDemandBlockNum != 0 {
		block := &types.Block{}
		proto.Unmarshal(payload, block)
		// store blocks into database only, not process them.
		if !executor.verifyBlockIntegrity(block) {
			executor.logger.Warningf("[Namespace = %s] receive a broken block %d, drop it", executor.namespace, block.Number)
			return
		}
		if block.Number <= executor.status.syncFlag.SyncDemandBlockNum {
			executor.logger.Debugf("[Namespace = %s] receive block #%d  hash %s", executor.namespace, block.Number, common.BytesToHash(block.BlockHash).Hex())
			// is demand
			if executor.isDemandSyncBlock(block) {
				edb.PersistBlock(executor.db.NewBatch(), block, true, true)
				if err := executor.updateSyncDemand(block); err != nil {
					executor.logger.Errorf("[Namespace = %s] update sync demand failed.", executor.namespace)
					executor.reject()
					return
				}
			} else {
				// requested block with smaller number arrive earlier than expected
				// store in cache temporarily
				executor.logger.Debugf("[Namespace = %s] receive block #%d hash %s earily", executor.namespace, block.Number, common.BytesToHash(block.BlockHash).Hex())
				executor.addToSyncCache(block)
			}
		}
		if executor.receiveAllRequiredBlocks() {
			executor.logger.Debug("receive a batch of blocks")
			var needNextFetch bool
			if !executor.status.syncCtx.UpdateGenesis {
				if executor.getLatestSyncDownstream() != edb.GetHeightOfChain(executor.namespace) {
					executor.logger.Debug("current downstream not equal to chain height")
					needNextFetch = true
				}
			} else {
				_, genesis := executor.status.syncCtx.GetCurrentGenesis()
				if executor.getLatestSyncDownstream() != genesis - 1 {
					executor.logger.Debug("current downstream not equal to genesis")
					needNextFetch = true
				}
			}
			if needNextFetch {
				executor.logger.Debug("still have some blocks to fetch")
				executor.status.syncFlag.Oracle.FeedBack(true)
				executor.status.syncCtx.SetCurrentPeer(executor.status.syncFlag.Oracle.SelectPeer())
				prev := executor.getLatestSyncDownstream()
				next := executor.calcuDownstream()
				executor.SendSyncRequest(prev, next)
			} else {
				executor.logger.Debugf("receive all required blocks. from %d to %d", edb.GetHeightOfChain(executor.namespace), executor.status.syncFlag.SyncTarget)
				if executor.status.syncCtx.UpdateGenesis {
					// receive world state
					executor.logger.Debug("send request to fetch world state for status transition")
					executor.status.syncCtx.SetResendMode(ResendMode_WorldState)
					executor.SendSyncRequestForWorldState(executor.status.syncFlag.SyncDemandBlockNum + 1)
				} else {
					executor.processSyncBlocks()
				}
			}
		}
	}
}

func (executor *Executor) ReceiveWorldState(payload []byte) {
	executor.logger.Noticef("receive world state")
	var packet WorldStateContext
	if err := proto.Unmarshal(payload, &packet); err != nil {
		executor.logger.Warning("unmarshal world state packet failed.")
		return
	}
	// apply
	tmp, err := ioutil.TempDir(hyperdb.GetDatabaseHome(executor.conf), "WORLD_STATE")
	if err != nil {
		executor.logger.Warning("create temp dir for world state failed")
		return
	}

	//defer func() {
	//	os.RemoveAll(tmp)
	//}()

	fPath := path.Join(tmp, "world_state.tar.gz")
	if err := ioutil.WriteFile(fPath, packet.Payload, 0644); err != nil {
		executor.logger.Warning("write network packet to compress file failed")
		return
	}
	localCmd := cmd.Command("tar", "-C", filepath.Dir(fPath), "-zxvf", fPath)
	if err := localCmd.Run(); err != nil {
		executor.logger.Warning("uncompress world state failed")
		return
	}
}

// SendSyncRequest - send synchronization request to other nodes.
func (executor *Executor) SendSyncRequest(upstream, downstream uint64) {
	if executor.isSyncInExecution() == true {
		return
	}
	peer := executor.status.syncCtx.GetCurrentPeer()
	executor.logger.Noticef("send sync req to %d, require [%d] to [%d]", peer, downstream, upstream)
	if err := executor.informP2P(NOTIFY_BROADCAST_DEMAND, upstream, downstream, peer); err != nil {
		executor.logger.Errorf("[Namespace = %s] send sync req failed.", executor.namespace)
		executor.reject()
		return
	}
	executor.recordSyncReqArgs(upstream, downstream)
}

// ApplyBlock - apply all transactions in block into state during the `state update` process.
func (executor *Executor) ApplyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	if block.Transactions == nil {
		return EmptyPointerErr, nil
	}
	return executor.applyBlock(block, seqNo)
}

func (executor *Executor) applyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	var filterLogs []*vm.Log
	err, result := executor.applyTransactions(block.Transactions, nil, seqNo)
	if err != nil {
		return err, nil
	}
	batch := executor.statedb.FetchBatch(seqNo)
	if err := executor.persistTransactions(batch, block.Transactions, seqNo); err != nil {
		return err, nil
	}
	if err, logs := executor.persistReceipts(batch, result.Receipts, seqNo, common.BytesToHash(block.BlockHash)); err != nil {
		return err, nil
	} else {
		filterLogs = logs
	}
	executor.storeFilterData(result, block, filterLogs)
	return nil, result
}

// ClearStateUnCommitted - remove all cached stuff
func (executor *Executor) clearStatedb() {
	executor.statedb.Purge()
}

// assertApplyResult - check apply result whether equal with other's.
func (executor *Executor) assertApplyResult(block *types.Block, result *ValidationResultRecord) bool {
	if bytes.Compare(block.MerkleRoot, result.MerkleRoot) != 0 {
		executor.logger.Warningf("[Namespace = %s] mismatch in block merkle root  of #%d, demand %s, got %s",
			executor.namespace, block.Number, common.Bytes2Hex(block.MerkleRoot), common.Bytes2Hex(result.MerkleRoot))
		return false
	}
	if bytes.Compare(block.TxRoot, result.TxRoot) != 0 {
		executor.logger.Warningf("[Namespace = %s] mismatch in block transaction root  of #%d, demand %s, got %s",
			block.Number, common.Bytes2Hex(block.TxRoot), common.Bytes2Hex(result.TxRoot))
		return false

	}
	if bytes.Compare(block.ReceiptRoot, result.ReceiptRoot) != 0 {
		executor.logger.Warningf("[Namespace = %s] mismatch in block receipt root  of #%d, demand %s, got %s",
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
		executor.logger.Warningf("[Namespace = %s] missing match target blockhash and latest block's hash, target block hash %s, latest block hash %s",
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
			executor.logger.Errorf("[Namespace = %s] StateUpdate Failed!", executor.namespace)
			executor.reject()
			return
		}
		// check the latest block in local's correctness
		if bytes.Compare(lastBlk.ParentHash, edb.GetLatestBlockHash(executor.namespace)) == 0  {
			executor.waitUtilSyncAvailable()
			defer executor.syncDone()
			// execute all received block at one time
			for i := executor.status.syncFlag.SyncDemandBlockNum + 1; i <= executor.status.syncFlag.SyncTarget; i += 1 {
				executor.markSyncExecBegin()
				blk, err := edb.GetBlockByNumber(executor.namespace, i)
				if err != nil {
					executor.logger.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
						executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1, executor.status.syncFlag.SyncTarget, edb.GetHeightOfChain(executor.namespace))
					executor.reject()
					return
				} else {
					// set temporary block number as block number since block number is already here
					executor.initDemand(blk.Number)
					err, result := executor.ApplyBlock(blk, blk.Number)
					if err != nil || executor.assertApplyResult(blk, result) == false {
						executor.logger.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
							executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1, executor.status.syncFlag.SyncTarget, edb.GetHeightOfChain(executor.namespace))
						executor.reject()
						return
					} else {
						// commit modified changes in this block and update chain.
						if err := executor.accpet(blk.Number, result); err != nil {
							executor.reject()
							return
						}
					}
				}
			}
			executor.initDemand(executor.status.syncFlag.SyncTarget + 1)
			executor.clearSyncFlag()
			executor.sendStateUpdatedEvent()
		} else {
			// the highest block in local is invalid, request the block
			if err := executor.CutdownBlock(lastBlk.Number - 1); err != nil {
				executor.logger.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, lastBlk.Number - 1)
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

func (executor *Executor) SendSyncRequestForWorldState(number uint64) {
	executor.informP2P(NOTIFY_REQUEST_WORLD_STATE, number)
}

// updateSyncDemand - update next demand block number and block hash.
func (executor *Executor) updateSyncDemand(block *types.Block) error {
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > edb.GetHeightOfChain(executor.namespace) {
		if executor.cache.syncCache.Contains(tmp) {
			blks, _ := executor.fetchFromSyncCache(tmp)
			for hash, blk := range blks {
				if hash == common.BytesToHash(tmpHash).Hex() {
					edb.PersistBlock(executor.db.NewBatch(), &blk, true, true)
					executor.cache.syncCache.Remove(tmp)
					tmp = tmp - 1
					tmpHash = blk.ParentHash
					flag = true
					executor.logger.Debugf("[Namespace = %s] process sync block(block number = %d) stored in cache", executor.namespace, blk.Number)
					break
				} else {
					executor.logger.Debugf("[Namespace = %s] found invalid sync block, discard block number %d, block hash %s", executor.namespace, blk.Number, common.BytesToHash(blk.BlockHash).Hex())
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
	executor.logger.Debugf("[Namespace = %s] Next Demand %d %s", executor.namespace, executor.status.syncFlag.SyncDemandBlockNum, common.BytesToHash(executor.status.syncFlag.SyncDemandBlockHash).Hex())
	return nil
}

// sendStateUpdatedEvent - communicate with consensus, told it state update has finished.
func (executor *Executor) sendStateUpdatedEvent() {
	// state update success
	executor.PurgeCache()
	executor.informConsensus(NOTIFY_SYNC_DONE, protos.StateUpdatedMessage{edb.GetHeightOfChain(executor.namespace)})
}

// accpet - accept block synchronization result.
func (executor *Executor) accpet(seqNo uint64, result *ValidationResultRecord) error {
	batch := executor.statedb.FetchBatch(seqNo)
	if err := edb.UpdateChainByBlcokNum(executor.namespace, batch, seqNo, false, false); err != nil {
		executor.logger.Errorf("update chain to (#%d) failed, err: %s", err.Error())
		return err
	}
	if err := batch.Write(); err != nil {
		executor.logger.Errorf("commit (#%d) changes failed, err: %s", err.Error())
		return err
	}
	executor.statedb.MarkProcessFinish(seqNo)
	executor.filterFeedback(result.Block, result.Logs)
	return nil
}

// reject - reject state update result.
func (executor *Executor) reject() {
	executor.cache.syncCache.Purge()
	// clear all useless stuff
	batch := executor.db.NewBatch()
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
	if bytes.Compare(block.BlockHash, block.Hash().Bytes()) == 0 {
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

// calcuDownstream - calculate a sync request downstream
// if a node required to sync too much blocks one time, the huge chain sync request will be split to several small one.
// a sync chain required block number can not more than `sync batch size` in config file.
func (executor *Executor) calcuDownstream() uint64 {
	total := executor.getLatestSyncDownstream() - edb.GetHeightOfChain(executor.namespace)
	if total < executor.GetSyncMaxBatchSize() {
		_, genesis := executor.status.syncCtx.GetCurrentGenesis()
		if genesis > edb.GetHeightOfChain(executor.namespace) {
			// genesis block is also required
			executor.setLatestSyncDownstream(genesis - 1)
			executor.logger.Notice("update temporarily downstream with peer's genesis")
		} else {
			executor.setLatestSyncDownstream(edb.GetHeightOfChain(executor.namespace))
			executor.logger.Notice("update temporarily downstream with current chain height")
		}
	} else {
		_, genesis := executor.status.syncCtx.GetCurrentGenesis()
		if genesis > executor.getLatestSyncDownstream() - executor.GetSyncMaxBatchSize() {
			// genesis block is also required
			executor.setLatestSyncDownstream(genesis - 1)
			executor.logger.Notice("update temporarily downstream with peer's genesis")
		} else {
			executor.setLatestSyncDownstream(executor.getLatestSyncDownstream() - executor.GetSyncMaxBatchSize())
			executor.logger.Notice("update temporarily downstream with last temp downstream")
		}
	}
	executor.logger.Noticef("update temporarily downstream to %d", executor.getLatestSyncDownstream())
	return executor.getLatestSyncDownstream()
}

func (executor *Executor) receiveAllRequiredBlocks() bool {
	return executor.status.syncFlag.SyncDemandBlockNum == executor.getLatestSyncDownstream()
}

// storeFilterData - store filter data in record temporarily, avoid re-generated when using.
func (executor *Executor) storeFilterData(record *ValidationResultRecord, block *types.Block, logs []*vm.Log) {
	record.Block = block
	record.Logs = logs
}


func (executor *Executor) fetchRepliceIds(event event.ChainSyncReqEvent) []uint64 {
	var ret []uint64
	for _, r := range event.Replicas {
		ret = append(ret, r.Id)
	}
	return ret
}

// syncinitialize initialize sync context and status.
func (executor *Executor) syncInitialize(ev event.ChainSyncReqEvent) {
	ctx := NewChainSyncContext(executor.namespace, ev)
	executor.status.syncCtx = ctx

	executor.updateSyncFlag(ev.TargetHeight, ev.TargetBlockHash, ev.TargetHeight)
	executor.setLatestSyncDownstream(ev.TargetHeight)
	executor.recordSyncPeers(executor.fetchRepliceIds(ev), ev.Id)
	executor.status.syncFlag.Oracle = NewOracle(ctx, executor.conf, executor.logger)
	firstPeer := executor.status.syncFlag.Oracle.SelectPeer()
	ctx.SetCurrentPeer(firstPeer)
}


