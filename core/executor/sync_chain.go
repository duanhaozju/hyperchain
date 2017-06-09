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
	"os"
	"path"
	"hyperchain/hyperdb"
	"io/ioutil"
	"fmt"
	"errors"
	cmd "os/exec"
	"path/filepath"
	cm "hyperchain/core/common"
)


/*
	Sync chain initiator
 */

// SyncChain receive chain sync request from consensus module,
// trigger to sync.
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

// ReceiveSyncBlocks - receive request synchronization blocks from others.
func (executor *Executor) ReceiveSyncBlocks(payload []byte) {

	checkNeedMore := func() bool {
		var needNextFetch bool
		if !executor.status.syncCtx.UpdateGenesis {
			if executor.getLatestSyncDownstream() != edb.GetHeightOfChain(executor.namespace) {
				executor.logger.Notice("current downstream not equal to chain height")
				needNextFetch = true
			}
		} else {
			_, genesis := executor.status.syncCtx.GetCurrentGenesis()
			if executor.getLatestSyncDownstream() != genesis - 1 {
				executor.logger.Notice("current downstream not equal to genesis")
				needNextFetch = true
			}
		}
		return needNextFetch
	}

	checker := func() bool {
		lastBlk, err := edb.GetBlockByNumber(executor.namespace, executor.status.syncFlag.SyncDemandBlockNum +1)
		if err != nil {
			return false
		}
		latestBlk, err := edb.GetLatestBlock(executor.namespace)
		if err != nil {
			return false
		}
		if bytes.Compare(lastBlk.ParentHash, latestBlk.BlockHash) != 0 {
			return false
		}
		return true
	}

	reqNext := func() {
		executor.logger.Notice("still have some blocks to fetch")
		executor.status.syncFlag.Oracle.FeedBack(true)
		executor.status.syncCtx.SetCurrentPeer(executor.status.syncFlag.Oracle.SelectPeer())
		prev := executor.getLatestSyncDownstream()
		next := executor.calcuDownstream()
		executor.SendSyncRequest(prev, next)
	}

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
			executor.logger.Notice("receive a batch of blocks")
			needNextFetch := checkNeedMore()
			if needNextFetch {
				reqNext()
			} else {
				executor.logger.Noticef("receive all required blocks. from %d to %d", edb.GetHeightOfChain(executor.namespace), executor.status.syncFlag.SyncTarget)
				if executor.status.syncCtx.UpdateGenesis {
					// receive world state
					executor.logger.Notice("send request to fetch world state for status transition")
					executor.status.syncCtx.SetResendMode(ResendMode_WorldState)
					executor.SendSyncRequestForWorldState(executor.status.syncFlag.SyncDemandBlockNum + 1)
				} else {
					// check
					if checker() {
						executor.status.syncCtx.SetResendMode(ResendMode_Nope)
						executor.processSyncBlocks()
					} else {
						if err := executor.CutdownBlock(executor.status.syncFlag.SyncDemandBlockNum); err != nil {
							executor.logger.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, executor.status.syncFlag.SyncDemandBlockNum)
							executor.reject()
							return
						}
						executor.logger.Noticef("cutdown block #%d", executor.status.syncFlag.SyncDemandBlockNum)
						reqNext()
					}
				}
			}
		}
	}
}

func (executor *Executor) ReceiveWsHandshake(payload []byte) {
	var hs WsHandshake
	if err := proto.Unmarshal(payload, &hs); err != nil {
		executor.logger.Warning("unmarshal world state packet failed.")
		return
	}
	executor.logger.Noticef("receive ws handshake, content: [ total size (#%d), packet num (#%d), max packet size (#%d)",
		hs.Size, hs.PacketNum, hs.PacketSize)
	executor.status.syncCtx.hs = hs

	// make `receive home`
	p := path.Join(hyperdb.GetDatabaseHome(executor.conf), "ws", "ws_" + hs.Ctx.FilterId)
	err := os.MkdirAll(p, 0777)
	if err != nil {
		executor.logger.Warningf("make ws home for %s failed", hs.Ctx.FilterId)
		return
	}
	executor.status.syncCtx.SetWsHome(p)
	// send back ack
	ack := executor.constructWsAck(hs.Ctx, 0, WsAck_OK, nil)
	if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
		executor.logger.Warning("send ws ack failed")
		return
	}
	executor.logger.Noticef("send ws ack (#%d) success", ack.PacketId)
}

func (executor *Executor) ReceiveWorldState(payload []byte) {
	executor.logger.Noticef("receive ws packet")
	var ws Ws
	if err := proto.Unmarshal(payload, &ws); err != nil {
		executor.logger.Warning("unmarshal world state packet failed.")
		return
	}

	store := func(payload []byte, packetId uint64, filterId string) error {
		// TODO add checksum
		// GRPC will prevent packet to be modified
		executor.logger.Noticef("receive ws (#%s) fragment (#%d), size (#%d)", filterId, packetId, len(payload))
		fname := fmt.Sprintf("ws_%d.tar.gz", packetId)
		if err := ioutil.WriteFile(path.Join(executor.status.syncCtx.GetWsHome(), fname), payload, 0644); err != nil {
			return err
		}
		return nil
	}

	assemble := func() error {
		hs := executor.status.syncCtx.hs
		var i uint64 = 1;
		fd, err := os.OpenFile(path.Join(executor.status.syncCtx.GetWsHome(), "ws.tar.gz"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		for ; i <= hs.PacketNum; i += 1 {
			fname := fmt.Sprintf("ws_%d.tar.gz", i)
			buf, err := ioutil.ReadFile(path.Join(executor.status.syncCtx.GetWsHome(), fname))
			if err != nil {
				return err
			}
			n, err := fd.Write(buf)
			if n != len(buf) || err != nil {
				return errors.New("assmble ws file failed")
			}
		}
		fd.Close()
		return nil
	}

	if ws.PacketId == executor.status.syncCtx.hs.PacketNum {
		// the last packet
		store(ws.Payload, ws.PacketId, ws.Ctx.FilterId)
		ack := executor.constructWsAck(ws.Ctx, ws.PacketId, WsAck_OK, []byte("Done"))
		if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
			executor.logger.Warning("send ws ack failed")
			return
		}
		executor.logger.Notice("receive all ws packet, begin to assemble")
		if err := assemble(); err != nil {
			executor.logger.Errorf("assemble failed, err detail %s", err.Error())
			return
		}

		_, nGenesis := executor.status.syncCtx.GetCurrentGenesis()
		newGenesis, err := edb.GetBlockByNumber(executor.namespace, nGenesis)
		if err != nil {
			return
		}
		if err := executor.applyWorldState(path.Join(executor.status.syncCtx.GetWsHome(), "ws.tar.gz"), ws.Ctx.FilterId, common.BytesToHash(newGenesis.MerkleRoot), newGenesis.Number); err != nil {
			executor.logger.Errorf("apply ws failed, err detail %s", err.Error())
			return
		}
		executor.status.syncCtx.SetTransitioned()
		executor.processSyncBlocks()
	} else {
		store(ws.Payload, ws.PacketId, ws.Ctx.FilterId)
		ack := executor.constructWsAck(ws.Ctx, ws.PacketId, WsAck_OK, nil)
		if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
			executor.logger.Warning("send ws ack failed")
			return
		}
		executor.logger.Noticef("send ws (#%s) ack (#%d) success", ack.Ctx.FilterId, ack.PacketId)
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
	if executor.status.syncFlag.SyncDemandBlockNum <= edb.GetHeightOfChain(executor.namespace) || executor.status.syncCtx.GenesisTranstioned {
		// get the first of SyncBlocks
		executor.waitUtilSyncAvailable()
		defer executor.syncDone()
		// execute all received block at one time
		var low uint64
		if executor.status.syncCtx.UpdateGenesis {
			_, low = executor.status.syncCtx.GetCurrentGenesis()
			low += 1
		} else {
			low = executor.status.syncFlag.SyncDemandBlockNum + 1;
		}

		for i := low; i <= executor.status.syncFlag.SyncTarget; i += 1 {
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
				executor.stateTranstion(blk.Number + 1, common.BytesToHash(blk.MerkleRoot))
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
	}
}

// broadcastDemandBlock - send block request message to others for demand block.
func (executor *Executor) SendSyncRequestForSingle(number uint64) {
	executor.informP2P(NOTIFY_BROADCAST_SINGLE, number)
}

func (executor *Executor) SendSyncRequestForWorldState(number uint64) {
	executor.logger.Noticef("send req to fetch world state at height (#%d)", number)
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
	if executor.status.syncCtx.UpdateGenesis {
		_, genesis := executor.status.syncCtx.GetCurrentGenesis()
		total := executor.getLatestSyncDownstream() - genesis + 1
		if total < executor.GetSyncMaxBatchSize() {
			executor.setLatestSyncDownstream(genesis - 1)
		} else {
			executor.setLatestSyncDownstream(executor.getLatestSyncDownstream() - executor.GetSyncMaxBatchSize())
		}
	} else {
		total := executor.getLatestSyncDownstream() - edb.GetHeightOfChain(executor.namespace)
		if total < executor.GetSyncMaxBatchSize() {
			executor.setLatestSyncDownstream(edb.GetHeightOfChain(executor.namespace))
		} else {
			executor.setLatestSyncDownstream(executor.getLatestSyncDownstream() - executor.GetSyncMaxBatchSize())
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

func (executor *Executor) applyWorldState(fPath string, filterId string, root common.Hash, genesis uint64) error {
	uncompressCmd := cmd.Command("tar", "-zxvf", fPath, "-C", filepath.Dir(fPath))
	if err := uncompressCmd.Run(); err != nil {
		return err
	}
	dbPath := path.Join("ws", "ws_" + filterId, "SNAPSHOT_" + filterId)
	wsDb, err := hyperdb.NewDatabase(executor.conf, dbPath, hyperdb.GetDatabaseType(executor.conf), executor.namespace)
	if err != nil {
		return err
	}
	defer wsDb.Close()

	// TODO just for test
	// TODO ATOMIC ASSURANCE
	writeBatch := executor.db.NewBatch()
	entries := cm.RetrieveSnapshotFileds()
	for _, entry := range entries {
		iter := wsDb.NewIterator([]byte(entry))
		for iter.Next() {
			writeBatch.Put(iter.Key(), iter.Value())
		}
		iter.Release()
	}

	writeBatch.Write()

	hash, err := executor.statedb.RecomputeCryptoHash()
	if err != nil || hash != root {
		return ApplyWsErr
	}

	edb.UpdateGenesisTag(executor.namespace, genesis, writeBatch, true, true)

	executor.logger.Noticef("apply ws pieces (%s) success", filterId)
	return nil
}

/*
	Sync chain Receiver
 */
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

// ReceiveWorldStateSyncRequest - receive ws request, send back handshake packet first time.
func (executor *Executor) ReceiveWorldStateSyncRequest(payload []byte) {
	var request WsRequest
	var fsize int64
	if err := proto.Unmarshal(payload, &request); err != nil {
		executor.logger.Warning("unmarshal world state sync request failed.")
		return
	}
	executor.logger.Noticef("receive world state sync req, required (#%d)", request.Target)
	err, manifest := executor.snapshotReg.rwc.Search(request.Target)
	if err != nil {
		executor.logger.Warning("required snapshot doesn't exist")
		return
	}
	if err, size := executor.snapshotReg.CompressSnapshot(manifest.FilterId); err != nil {
		executor.logger.Warning("compress snapshot failed")
		return
	} else {
		fsize = size
	}

	n := fsize / int64(WsShardLen)
	if fsize % int64(WsShardLen) > 0 {
		n += 1
	}
	hs := executor.constructWsHandshake(request, manifest.FilterId, uint64(fsize), uint64(n))
	if err := executor.informP2P(NOTIFY_SEND_WORLD_STATE_HANDSHAKE, hs); err != nil {
		executor.logger.Warningf("send world state (#%s) back to (%d) failed, err msg %s", manifest.FilterId, request.InitiatorId, err.Error())
		return
	}
	executor.logger.Noticef("send world state (#%s) handshake back to (%d) success, total size %d, total packet num %d, max packet size %d",
		manifest.FilterId, request.InitiatorId, hs.Size, hs.PacketNum, hs.PacketSize)
}

func (executor *Executor) ReceiveWsAck(payload []byte) {
	var ack WsAck
	if err := proto.Unmarshal(payload, &ack); err != nil {
		executor.logger.Warning("unmarshal ws ack failed.")
		return
	}
	remove := func(filterId string) {
		fpath := executor.snapshotReg.CompressedSnapshotPath(ack.Ctx.FilterId)
		os.Remove(fpath)
	}

	sendWs := func(shardId uint64, filterId string, ws *WsAck) {
		fpath := executor.snapshotReg.CompressedSnapshotPath(filterId)
		err, reader := common.NewSectionReader(fpath, WsShardLen)
		defer reader.Close()
		if err != nil {
			return
		}
		n, ctx, err := reader.ReadAt(int64(shardId))
		if n > 0 {
			ws := executor.constrcutWs(&ack, shardId, uint64(n), ctx[:n])
			if err := executor.informP2P(NOTIFY_SEND_WORLD_STATE, ws); err != nil {
				return
			}
			executor.logger.Noticef("send ws(#%s) packet (#%d), packet size (#%d) to peer (#%d) success", ws.Ctx.FilterId, ws.PacketId, ws.PacketSize, ws.Ctx.ReceiverId)
		} else if n == 0 && err != nil {
			// TODO handler invalid ws req
			return
		}
	}

	if ack.Status == WsAck_OK {
		if string(ack.Message) == "Done" {
			// remove compressed file
			remove(ack.Ctx.FilterId)
		} else {
			// send next one
			sendWs(ack.PacketId + 1, ack.Ctx.FilterId, &ack)
		}
	} else {
		// resend
		sendWs(ack.PacketId, ack.Ctx.FilterId, &ack)
	}

}

/*
	Net Packets
 */
func (executor *Executor) constructWsHandshake(req WsRequest, filterId string, size uint64, pn uint64) *WsHandshake {
	return &WsHandshake{
		Ctx:          &WsContext{
			FilterId:     filterId,
			InitiatorId:  req.ReceiverId,
			ReceiverId:   req.InitiatorId,
		},
		Height:       req.Target,
		Size:         size,
		PacketSize:   uint64(WsShardLen),
		PacketNum:    pn,
	}
}

func (executor *Executor) constructWsAck(ctx *WsContext, packetId uint64, status WsAck_STATUS, message []byte) *WsAck {
	return &WsAck{
		Ctx:       &WsContext{
			FilterId:      ctx.FilterId,
			InitiatorId:   ctx.ReceiverId,
			ReceiverId:    ctx.InitiatorId,
		},
		PacketId:  packetId,
		Status:    status,
		Message:   message,
	}
}

func (executor *Executor) constrcutWs(ack *WsAck, packetId uint64, packetSize uint64, payload []byte) *Ws {
	executor.logger.Noticef("construct ws packet with %d size", len(payload))
	return &Ws{
		Ctx:       &WsContext{
			FilterId:      ack.Ctx.FilterId,
			InitiatorId:   ack.Ctx.ReceiverId,
			ReceiverId:    ack.Ctx.InitiatorId,
		},
		PacketId:    packetId,
		PacketSize:  packetSize,
		Payload:     payload,
	}
}

