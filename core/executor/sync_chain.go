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
	er "hyperchain/core/errors"
	"github.com/cheggaaa/pb"
)

var (
	receivePb *pb.ProgressBar
	processPb *pb.ProgressBar
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

	executor.updateSyncFlag(ev.TargetHeight, ev.TargetBlockHash, ev.TargetHeight)
	executor.status.syncFlag.ResendExit = make(chan bool)
	executor.setLatestSyncDownstream(ev.TargetHeight)
	executor.recordSyncPeers(ev.Replicas, ev.Id)
	executor.status.syncFlag.Oracle = NewOracle(ev.Replicas, executor.conf, executor.logger)
	executor.SendSyncRequest(ev.TargetHeight, executor.calcuDownstream())
	receivePb = common.InitPb(int64(ev.TargetHeight - edb.GetHeightOfChain(executor.namespace)), "receive block")
	receivePb.Start()
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
			curUp, curDown := executor.getSyncReqArgs()
			if curUp == up && curDown == down && !executor.isSyncInExecution() {
				executor.logger.Noticef("resend sync request. want [%d] - [%d]", down, executor.status.syncFlag.SyncDemandBlockNum)
				executor.status.syncFlag.Oracle.FeedBack(false)
				executor.SendSyncRequest(executor.status.syncFlag.SyncDemandBlockNum, down)
				executor.recordSyncReqArgs(curUp, curDown)
			} else {
				up = curUp
				down = curDown
			}
		}
	}
}

// ReceiveSyncRequest - receive synchronization request from some nodes, and send back request blocks.
func (executor *Executor) ReceiveSyncRequest(payload []byte) {
	var request ChainSyncRequest
	proto.Unmarshal(payload, &request)
	for i := request.RequiredNumber; i > request.CurrentNumber; i -= 1 {
		executor.informP2P(NOTIFY_UNICAST_BLOCK, i, request.PeerId, request.PeerHash)
	}
}

// ReceiveSyncBlocks - receive request synchronization blocks from others.
func (executor *Executor) ReceiveSyncBlocks(payload []byte) {
	if executor.status.syncFlag.SyncDemandBlockNum != 0 {
		block := &types.Block{}
		if err := proto.Unmarshal(payload, block); err != nil {
			executor.logger.Warning("receive a block but unmarshal failed")
			return
		}
		// store blocks into database only, not process them.
		if !VerifyBlockIntegrity(block) {
			executor.logger.Warningf("[Namespace = %s] receive a broken block %d, drop it", executor.namespace, block.Number)
			return
		}
		if block.Number <= executor.status.syncFlag.SyncDemandBlockNum {
			executor.logger.Debugf("[Namespace = %s] receive block #%d  hash %s", executor.namespace, block.Number, common.BytesToHash(block.BlockHash).Hex())
			// is demand
			if executor.isDemandSyncBlock(block) {
				// received block's struct definition may different from current
				// for backward compatibility, store with original version tag.
				edb.PersistBlock(executor.db.NewBatch(), block, true, true, string(block.Version))
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
			if executor.getLatestSyncDownstream() != edb.GetHeightOfChain(executor.namespace) {
				common.AddPb(receivePb, int64(executor.GetSyncMaxBatchSize()))
				common.PrintPb(receivePb, 0, executor.logger)
				prev := executor.getLatestSyncDownstream()
				next := executor.calcuDownstream()
				executor.status.syncFlag.Oracle.FeedBack(true)
				executor.SendSyncRequest(prev, next)
			} else {
				executor.logger.Debugf("receive all required blocks. from %d to %d", edb.GetHeightOfChain(executor.namespace), executor.status.syncFlag.SyncTarget)
				common.SetPb(receivePb, receivePb.Total)
				common.PrintPb(receivePb, 0, executor.logger)
				receivePb.Finish()
			}
		}
		executor.processSyncBlocks()
	}
}

// SendSyncRequest - send synchronization request to other nodes.
func (executor *Executor) SendSyncRequest(upstream, downstream uint64) {
	if executor.isSyncInExecution() == true {
		return
	}
	peer := executor.status.syncFlag.Oracle.SelectPeer()
	// peer := executor.status.syncFlag.SyncPeers[rand.Intn(len(executor.status.syncFlag.SyncPeers))]
	executor.logger.Debugf("send sync req to %d, require [%d] to [%d]", peer, downstream, upstream)
	if err := executor.informP2P(NOTIFY_BROADCAST_DEMAND, upstream, downstream, peer); err != nil {
		executor.logger.Errorf("[Namespace = %s] send sync req failed.", executor.namespace)
		executor.reject()
		return
	}
	executor.recordSyncReqArgs(upstream, downstream)
}

// ApplyBlock - apply all transactions in block into state during the `state update` process.
func (executor *Executor) ApplyBlock(block *types.Block, seqNo, tempBlockNumber uint64) (error, *ValidationResultRecord) {
	if block.Transactions == nil {
		return er.EmptyPointerErr, nil
	}
	return executor.applyBlock(block, seqNo, tempBlockNumber)
}

func (executor *Executor) applyBlock(block *types.Block, seqNo, tempBlockNumber uint64) (error, *ValidationResultRecord) {
	err, result := executor.applyTransactions(block.Transactions, nil, seqNo, tempBlockNumber)
	if err != nil {
		return err, nil
	}
	batch := executor.statedb.FetchBatch(seqNo)
	if err := executor.persistTransactions(batch, block.Transactions, seqNo); err != nil {
		return err, nil
	}
	if err := executor.persistReceipts(batch, block.Transactions, result.Receipts, seqNo, common.BytesToHash(block.BlockHash)); err != nil {
		return err, nil
	}
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
		latestBlk, _ := edb.GetBlockByNumber(executor.namespace, edb.GetHeightOfChain(executor.namespace))

		executor.logger.Debugf("compare latest block %d hash, correct %s, current %s",
			latestBlk.Number, common.Bytes2Hex(lastBlk.ParentHash), common.Bytes2Hex(latestBlk.BlockHash))

		if bytes.Compare(lastBlk.ParentHash, latestBlk.BlockHash) == 0  {
			executor.waitUtilSyncAvailable()
			defer executor.syncDone()
			// execute all received block at one time
			processPb = common.InitPb(int64(executor.getSyncTarget() - edb.GetHeightOfChain(executor.namespace)), "process block")
			processPb.Start()
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
					err, result := executor.ApplyBlock(blk, blk.Number, executor.getTempBlockNumber())
					if err != nil || executor.assertApplyResult(blk, result) == false {
						executor.logger.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
							executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1, executor.status.syncFlag.SyncTarget, edb.GetHeightOfChain(executor.namespace))
						executor.reject()
						return
					} else {
						// commit modified changes in this block and update chain.
						if err := executor.accpet(blk.Number); err != nil {
							executor.reject()
							return
						}
						common.AddPb(processPb, 1)
						common.PrintPb(processPb, 10, executor.logger)
					}
				}
			}
			if !common.IsPrintPb(processPb, 10) {
				common.PrintPb(processPb, 0, executor.logger)
			}
			processPb.Finish()
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
			executor.logger.Debugf("cutdown block #%d success", latestBlk.Number)

			prev := executor.getLatestSyncDownstream()
			next := executor.calcuDownstream()
			executor.status.syncFlag.Oracle.FeedBack(true)
			executor.SendSyncRequest(prev, next)
		}
	}
}

// broadcastDemandBlock - send block request message to others for demand block.
func (executor *Executor) SendSyncRequestForSingle(number uint64) {
	executor.informP2P(NOTIFY_BROADCAST_SINGLE, number)
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
func (executor *Executor) accpet(seqNo uint64) error {
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
	return nil
}

// reject - reject state update result.
func (executor *Executor) reject() {
	executor.cache.syncCache.Purge()
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
		executor.setLatestSyncDownstream(edb.GetHeightOfChain(executor.namespace))
	} else {
		executor.setLatestSyncDownstream(executor.getLatestSyncDownstream() - executor.GetSyncMaxBatchSize())
	}
	executor.logger.Debugf("update temporarily downstream to %d", executor.getLatestSyncDownstream())
	return executor.getLatestSyncDownstream()
}

func (executor *Executor) receiveAllRequiredBlocks() bool {
	return executor.status.syncFlag.SyncDemandBlockNum == executor.getLatestSyncDownstream()
}
