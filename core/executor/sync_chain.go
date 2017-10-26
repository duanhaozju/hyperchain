// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package executor

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	cmd "os/exec"
	"path"
	"path/filepath"
	"sync/atomic"
	"time"

	"hyperchain/common"
	cm "hyperchain/core/common"
	"hyperchain/core/ledger/bloom"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/ledger/state"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

/*
	Sync chain initiator
*/

// SyncChain receive chain sync request from consensus module,
// trigger to sync.
func (executor *Executor) SyncChain(ev event.ChainSyncReqEvent) {
	executor.logger.Noticef("[Namespace = %s] send sync block request to fetch missing block, current height %d, target height %d",
		executor.namespace, edb.GetHeightOfChain(executor.namespace), ev.TargetHeight)
	// Ignore invalid state synchronization request.
	// Invalid request includes:
	// 1. last synchronization has not finished and current sync target is less than the in-flight one.
	// 2. sync target is less than chain height.
	if executor.context.syncCtx != nil && executor.context.syncCtx.target >= ev.TargetHeight || edb.GetHeightOfChain(executor.namespace) > ev.TargetHeight {
		executor.logger.Errorf("[Namespace = %s] receive invalid state update request, just ignore it", executor.namespace)
		executor.reject()
		return
	}

	// If sync target is equal with current chain height, compare sync target block hash with local block hash.
	if edb.GetHeightOfChain(executor.namespace) == ev.TargetHeight {
		executor.logger.Debugf("[Namespace = %s] recv target height same with current chain height", executor.namespace)
		if executor.isBlockHashEqual(ev.TargetBlockHash) == true {
			executor.logger.Infof("[Namespace = %s] current chain latest block hash equal with target hash, send state updated event", executor.namespace)
			executor.sendStateUpdatedEvent()
		} else {
			// Local latest block is invalid.
			executor.logger.Warningf("[Namespace = %s] current chain latest block hash not equal with target hash, cut down local block %d", executor.namespace, edb.GetHeightOfChain(executor.namespace))
			if err := executor.CutdownBlock(edb.GetHeightOfChain(executor.namespace)); err != nil {
				executor.logger.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, edb.GetHeightOfChain(executor.namespace))
				executor.reject()
				return
			}
		}
	}

	executor.setupContext(ev)

	// If synchronization has to depend on world state transition,
	// check whether world state transition is enable or not.
	if executor.context.syncCtx.updateGenesis && !executor.conf.IsSyncWsEable() {
		executor.logger.Noticef("World state transition is not supported, chain synchronization can not been arcieved since there has no required blocks over the network. system exit")
		os.Exit(1)
	}

	// If synchronization has to depend on world state transition,
	// check whether there exists available block data in network.
	if len(executor.context.syncCtx.getFullPeersId()) == 0 && len(executor.context.syncCtx.getPartPeersId()) == 0 {
		executor.logger.Noticef("There is no satisfied peers over the blockchain network to make chain synchronization, hold on some time to retry. system exit")
		os.Exit(1)
	}

	executor.SendSyncRequest(ev.TargetHeight, executor.calcuDownstream())

	executor.context.closeW.Add(1)
	go executor.syncChainResendBackend()
}

// syncChainResendBackend resend request if not meet requirement in time.
func (executor *Executor) syncChainResendBackend() {
	ticker := time.NewTicker(executor.conf.GetSyncResendInterval())
	up, down := executor.context.syncCtx.getRequest()
	for {
		select {
		case <-executor.context.stateUpdated:
			executor.context.closeW.Done()
			return
		case <-ticker.C:
			if executor.context.syncCtx.getResendMode() == ResendMode_Block {
				// resend block reqeust.
				curUp, curDown := executor.context.syncCtx.getRequest()
				if curUp == up && curDown == down {
					executor.logger.Noticef("resend sync request. want [%d] - [%d]", down, executor.context.syncCtx.demandBlockNum)
					executor.context.syncCtx.qosStat.feedBack(false)
					executor.context.syncCtx.setCurrentPeer(executor.context.syncCtx.qosStat.selectPeer())
					executor.SendSyncRequest(executor.context.syncCtx.demandBlockNum, down)
					executor.context.syncCtx.recordRequest(curUp, curDown)
				} else {
					up = curUp
					down = curDown
				}
			} else if executor.context.syncCtx.getResendMode() == ResendMode_WorldState_Hs {
				// resend world state transition handkshake request.
				executor.SendSyncRequestForWorldState(executor.context.syncCtx.demandBlockNum + 1)
			} else if executor.context.syncCtx.getResendMode() == ResendMode_WorldState_Piece {
				// resend world state transition piece request.
				ack := executor.constructWsAck(executor.context.syncCtx.hs.Ctx, executor.context.syncCtx.getWsId(), WsAck_OK, nil)
				if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
					executor.logger.Warning("send ws ack failed")
					return
				}
			}
		}
	}
}

// ReceiveSyncBlocks receives synchronization request blocks from others.
func (executor *Executor) ReceiveSyncBlocks(payload []byte) {

	// Check whether all blocks received or not.
	checkNeedMore := func() bool {
		var needNextFetch bool
		if !executor.context.syncCtx.updateGenesis {
			if executor.context.syncCtx.getDownstream() != edb.GetHeightOfChain(executor.namespace) {
				executor.logger.Notice("current downstream not equal to chain height")
				needNextFetch = true
			}
		} else {
			_, genesis := executor.context.syncCtx.getTargerGenesis()
			if executor.context.syncCtx.getDownstream() != genesis-1 {
				executor.logger.Notice("current downstream not equal to genesis")
				needNextFetch = true
			}
		}
		return needNextFetch
	}

	// Check whether the latest block received from network is equal with local related block.
	checker := func() bool {
		lastBlk, err := edb.GetBlockByNumber(executor.namespace, executor.context.syncCtx.demandBlockNum+1)
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

	// Send block request.
	reqNext := func(isbatch bool) {
		executor.context.syncCtx.qosStat.feedBack(true)
		executor.context.syncCtx.setCurrentPeer(executor.context.syncCtx.qosStat.selectPeer())
		if isbatch {
			executor.context.syncCtx.updateProgress(BlockReceiveProgress, int64(executor.conf.GetSyncMaxBatchSize()), 0)
		}
		prev := executor.context.syncCtx.getDownstream()
		next := executor.calcuDownstream()
		executor.SendSyncRequest(prev, next)
	}

	if executor.context.syncCtx != nil && executor.context.syncCtx.demandBlockNum != 0 {
		block := &types.Block{}
		if err := proto.Unmarshal(payload, block); err != nil {
			executor.logger.Warning("receive a block but unmarshal failed")
			return
		}
		// Check block intergrity via block hash
		if !VerifyBlockIntegrity(block) {
			executor.logger.Warningf("[Namespace = %s] receive a broken block %d, drop it", executor.namespace, block.Number)
			return
		}
		if block.Number <= executor.context.syncCtx.demandBlockNum {
			executor.logger.Debugf("[Namespace = %s] receive block #%d  hash %s", executor.namespace, block.Number, common.BytesToHash(block.BlockHash).Hex())
			// is demand
			if executor.isDemandSyncBlock(block) {
				// received block's struct definition may different from current
				// for backward compatibility, store with original version tag.
				// Persist receive block to dababase directly because store in memory
				// may cause memory exhausted.
				edb.PersistBlock(executor.db.NewBatch(), block, true, true, string(block.Version), getTxVersion(block))
				// update sync demand and scan any available blocks in cache and persist them to db.
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
			// a batch of request blocks all received
			executor.logger.Notice("receive a batch of blocks")
			needNextFetch := checkNeedMore()
			if needNextFetch {
				reqNext(true)
			} else {
				executor.logger.Noticef("receive all required blocks. from %d to %d", edb.GetHeightOfChain(executor.namespace), executor.context.syncCtx.target)
				executor.context.syncCtx.finishProgress(BlockReceiveProgress)
				if executor.context.syncCtx.updateGenesis {
					executor.logger.Notice("send request to fetch world state for status transition")
					executor.context.syncCtx.setResendMode(ResendMode_WorldState_Hs)
					executor.SendSyncRequestForWorldState(executor.context.syncCtx.demandBlockNum + 1)
				} else {
					// check
					if checker() {
						executor.context.syncCtx.setResendMode(ResendMode_Nope)
						executor.processSyncBlocks()
					} else {
						if err := executor.CutdownBlock(executor.context.syncCtx.demandBlockNum); err != nil {
							executor.logger.Errorf("[Namespace = %s] cut down block %d failed.", executor.namespace, executor.context.syncCtx.demandBlockNum)
							executor.reject()
							return
						}
						executor.logger.Noticef("cutdown block #%d", executor.context.syncCtx.demandBlockNum)
						reqNext(false)
					}
				}
			}
		}
	}
}

// ReceiveWsHandshake receive target peer's ws handshake packet,
// make some initialisation operations and send back ack.
func (executor *Executor) ReceiveWsHandshake(payload []byte) {
	// Short circuit if received duplicated handshake
	if executor.context.syncCtx.handshaked {
		return
	}
	var hs WsHandshake
	if err := proto.Unmarshal(payload, &hs); err != nil {
		executor.logger.Warning("unmarshal world state packet failed.")
		return
	}
	executor.logger.Noticef("receive ws handshake, content: [ total size (#%d), packet num (#%d), max packet size (#%d)KB ]",
		hs.Size, hs.PacketNum, hs.PacketSize/1024)

	executor.context.syncCtx.recordWsHandshake(&hs)

	// make `receive home`
	p := path.Join(cm.GetDatabaseHome(executor.namespace, executor.conf.conf), "ws", "ws_"+hs.Ctx.FilterId)
	err := os.MkdirAll(p, 0777)
	if err != nil {
		executor.logger.Warningf("make ws home for %s failed", hs.Ctx.FilterId)
		return
	}
	executor.context.syncCtx.setWsHome(p)
	// send back ack
	ack := executor.constructWsAck(hs.Ctx, executor.context.syncCtx.getWsId(), WsAck_OK, nil)
	if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
		executor.logger.Warning("send ws ack failed")
		return
	}
	executor.logger.Debugf("send ws ack (#%d) success", ack.PacketId)
}

func (executor *Executor) ReceiveWorldState(payload []byte) {
	var (
		process  func()
		finalize func() error
	)
	if executor.context.syncCtx != nil && executor.context.syncCtx.typ == VP {
		process = executor.processSyncBlocks
		finalize = nil
	} else if executor.context.syncCtx != nil && executor.context.syncCtx.typ == NVP {
		process = nil
		finalize = func() error {
			retry := 3
			for retry > 0 {
				if !atomic.CompareAndSwapInt32(&executor.context.syncCtx.waitTransition, 1, 0) {
					retry--
					continue
				}
				break
			}
			if retry == 0 {
				return nil
			} else {
				return errors.New("finalize world state transition failed")
			}
		}
	}
	executor.receiveWorldState(payload, process, finalize)
}

// ReceiveWsHandshake receive target peer's ws packet,
// store the slice packet and send back ack to fetch the next one.
// If all packets has been received, assemble them and trigger to apply.
func (executor *Executor) receiveWorldState(payload []byte, process func(), finailize func() error) {
	executor.logger.Debug("receive ws packet")
	var ws Ws
	if err := proto.Unmarshal(payload, &ws); err != nil {
		executor.logger.Warning("unmarshal world state packet failed.")
		return
	}

	store := func(payload []byte, packetId uint64, filterId string) error {
		// GRPC will prevent packet to be modified
		executor.logger.Debugf("receive ws (#%s) fragment (#%d), size (#%d)", filterId, packetId, len(payload))
		fname := fmt.Sprintf("ws_%d.tar.gz", packetId)
		if err := ioutil.WriteFile(path.Join(executor.context.syncCtx.getWsHome(), fname), payload, 0644); err != nil {
			return err
		}
		return nil
	}

	assemble := func() error {
		hs := executor.context.syncCtx.hs
		var i uint64 = 1
		fd, err := os.OpenFile(path.Join(executor.context.syncCtx.getWsHome(), "ws.tar.gz"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer fd.Close()
		for ; i <= hs.PacketNum; i += 1 {
			fname := fmt.Sprintf("ws_%d.tar.gz", i)
			buf, err := ioutil.ReadFile(path.Join(executor.context.syncCtx.getWsHome(), fname))
			if err != nil {
				return err
			}
			n, err := fd.Write(buf)
			if n != len(buf) || err != nil {
				return errors.New("assmble ws file failed")
			}
		}
		return nil
	}

	if ws.PacketId == executor.context.syncCtx.hs.PacketNum && !executor.context.syncCtx.receiveAll {
		// receive all, trigger to assemble and apply.
		store(ws.Payload, ws.PacketId, ws.Ctx.FilterId)
		ack := executor.constructWsAck(ws.Ctx, ws.PacketId, WsAck_OK, []byte("Done"))
		if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
			executor.logger.Warning("send ws ack failed")
			return
		}
		executor.context.syncCtx.setResendMode(ResendMode_Nope)
		executor.context.syncCtx.receiveAll = true

		executor.logger.Notice("receive all ws packets, begin to assemble")
		if err := assemble(); err != nil {
			executor.logger.Errorf("assemble failed, err detail %s", err.Error())
			return
		}
		// update genesis tag, regard the ws as the latest genesis status.
		var (
			height     uint64
			newGenesis *types.Block
			err        error
		)
		height = executor.context.syncCtx.hs.Height

		if newGenesis, err = edb.GetBlockByNumber(executor.namespace, height); err != nil {
			return
		}

		if err := executor.applyWorldState(path.Join(executor.context.syncCtx.getWsHome(), "ws.tar.gz"), ws.Ctx.FilterId, common.BytesToHash(newGenesis.MerkleRoot), newGenesis.Number); err != nil {
			executor.logger.Errorf("apply ws failed, err detail %s", err.Error())
			return
		}
		executor.context.syncCtx.setTransitioned()
		// Inserts received snapshot to local as genesis.
		go executor.InsertSnapshot(cm.Manifest{
			Height:     height,
			BlockHash:  common.Bytes2Hex(newGenesis.BlockHash),
			FilterId:   ws.Ctx.FilterId,
			MerkleRoot: common.Bytes2Hex(newGenesis.MerkleRoot),
			Date:       time.Unix(time.Now().Unix(), 0).Format("2006-01-02-15:04:05"),
			Namespace:  executor.namespace,
		})
		if process != nil {
			process()
		}
		if finailize != nil {
			finailize()
		}
	} else if ws.PacketId == executor.context.syncCtx.getWsId()+1 {
		// fetch the next slice packet.
		store(ws.Payload, ws.PacketId, ws.Ctx.FilterId)
		ack := executor.constructWsAck(ws.Ctx, ws.PacketId, WsAck_OK, nil)
		if err := executor.informP2P(NOTIFY_SEND_WS_ACK, ack); err != nil {
			executor.logger.Warning("send ws ack failed")
			return
		}
		executor.logger.Debugf("send ws (#%s) ack (#%d) success", ack.Ctx.FilterId, ack.PacketId)
		executor.context.syncCtx.setWsId(ws.PacketId)
	}
}

// SendSyncRequest sends synchronization request to other nodes.
func (executor *Executor) SendSyncRequest(upstream, downstream uint64) {
	peer := executor.context.syncCtx.getCurrentPeer()
	executor.logger.Debugf("send sync req to %d, require [%d] to [%d]", peer, downstream, upstream)
	if err := executor.informP2P(NOTIFY_BROADCAST_DEMAND, upstream, downstream, peer); err != nil {
		executor.logger.Errorf("[Namespace = %s] send sync req failed.", executor.namespace)
		executor.reject()
		return
	}
	executor.context.syncCtx.recordRequest(upstream, downstream)
}

func (executor *Executor) ApplyBlock(block *types.Block, seqNo uint64) (error, *ValidationResultRecord) {
	var filterLogs []*types.Log
	err, result := executor.applyTransactions(block.Transactions, nil, seqNo)
	if err != nil {
		return err, nil
	}
	batch := executor.statedb.FetchBatch(seqNo, state.BATCH_NORMAL)
	if err := executor.persistTransactions(batch, block.Transactions, seqNo); err != nil {
		return err, nil
	}
	if logs, err := executor.persistReceipts(batch, block.Transactions, result.Receipts, seqNo, common.BytesToHash(block.BlockHash)); err != nil {
		return err, nil
	} else {
		filterLogs = logs
	}
	executor.storeFilterData(result, block, filterLogs)
	return nil, result
}

// ClearStateUnCommitted removes all cached stuff
func (executor *Executor) clearStatedb() {
	executor.statedb.Purge()
}

// assertApplyResult checks apply result whether equal with other's.
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

// isBlockHashEqual compares block hash.
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

// processSyncBlocks executes all received block one by one.
func (executor *Executor) processSyncBlocks() {
	if executor.context.syncCtx == nil {
		return
	}
	if executor.context.syncCtx.demandBlockNum <= edb.GetHeightOfChain(executor.namespace) || executor.context.syncCtx.getTranstioned() {
		// get the first of SyncBlocks
		executor.waitUtilSyncAvailable()
		defer executor.syncDone()
		// execute all received block at one time
		var low uint64
		if executor.context.syncCtx.updateGenesis {
			_, low = executor.context.syncCtx.getTargerGenesis()
			low += 1
			executor.context.syncCtx.initProgres(BlockExecuteProgress, int64(executor.context.syncCtx.target-low+1), "block-process")
		} else {
			low = executor.context.syncCtx.demandBlockNum + 1
			executor.context.syncCtx.initProgres(BlockExecuteProgress, int64(executor.context.syncCtx.target-edb.GetHeightOfChain(executor.namespace)), "block-process")
		}

		for i := low; i <= executor.context.syncCtx.target; i += 1 {
			blk, err := edb.GetBlockByNumber(executor.namespace, i)
			if err != nil {
				executor.logger.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
					executor.namespace, executor.context.syncCtx.demandBlockNum+1, executor.context.syncCtx.target, edb.GetHeightOfChain(executor.namespace))
				executor.reject()
				return
			} else {
				// set temporary block number as block number since block number is already here
				executor.context.initDemand(blk.Number)
				executor.stateTransition(blk.Number+1, common.BytesToHash(blk.MerkleRoot))
				err, result := executor.ApplyBlock(blk, blk.Number)
				if err != nil || executor.assertApplyResult(blk, result) == false {
					executor.logger.Errorf("[Namespace = %s] state update from #%d to #%d failed. current chain height #%d",
						executor.namespace, executor.context.syncCtx.demandBlockNum+1, executor.context.syncCtx.target, edb.GetHeightOfChain(executor.namespace))
					if err != nil {
						executor.logger.Errorf("detail error %s", err.Error())
					}
					executor.reject()
					return
				} else {
					// commit modified changes in this block and update chain.
					if err := executor.accpet(blk.Number, blk, result); err != nil {
						executor.reject()
						return
					}
					executor.context.syncCtx.updateProgress(BlockExecuteProgress, 1, 10)
				}
			}
		}
		executor.context.syncCtx.finishProgress(BlockExecuteProgress)
		executor.context.initDemand(executor.context.syncCtx.target + 1)
		executor.clearSyncFlag()
		executor.sendStateUpdatedEvent()
	}
}

// SendSyncRequestForWorldState sends world state fetch request.
func (executor *Executor) SendSyncRequestForWorldState(number uint64) {
	executor.logger.Debugf("send req to fetch world state at height (#%d)", number)
	executor.informP2P(NOTIFY_REQUEST_WORLD_STATE, number, executor.context.syncCtx.getCurrentPeer(), false)
}

// updateSyncDemand updates next demand block number and block hash.
func (executor *Executor) updateSyncDemand(block *types.Block) error {
	var tmp = block.Number - 1
	var tmpHash = block.ParentHash
	flag := false
	for tmp > edb.GetHeightOfChain(executor.namespace) {
		if executor.cache.syncCache.Contains(tmp) {
			blks, _ := executor.fetchFromSyncCache(tmp)
			for hash, blk := range blks {
				if hash == common.Bytes2Hex(tmpHash) {
					edb.PersistBlock(executor.db.NewBatch(), &blk, true, true, string(block.Version), getTxVersion(block))
					executor.cache.syncCache.Remove(tmp)
					tmp = tmp - 1
					tmpHash = blk.ParentHash
					flag = true
					executor.logger.Debugf("[Namespace = %s] process sync block(block number = %d) stored in cache", executor.namespace, blk.Number)
					break
				} else {
					executor.logger.Debugf("[Namespace = %s] found invalid sync block, discard block number %d, block hash %s, required %s", executor.namespace, blk.Number, common.Bytes2Hex(blk.BlockHash), common.Bytes2Hex(tmpHash))
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
	executor.context.syncCtx.update(tmp, tmpHash)
	executor.logger.Debugf("[Namespace = %s] Next Demand %d %s", executor.namespace,
		executor.context.syncCtx.demandBlockNum, common.BytesToHash(executor.context.syncCtx.demandBlockHash).Hex())
	return nil
}

// sendStateUpdatedEvent communicates with consensus, told it state update has finished.
func (executor *Executor) sendStateUpdatedEvent() {
	// state update success
	executor.purgeCache()
	executor.informConsensus(NOTIFY_SYNC_DONE, protos.StateUpdatedMessage{edb.GetHeightOfChain(executor.namespace)})
}

// accpet accepts block synchronization result.
func (executor *Executor) accpet(seqNo uint64, block *types.Block, result *ValidationResultRecord) error {
	batch := executor.statedb.FetchBatch(seqNo, state.BATCH_NORMAL)
	if err := edb.UpdateChainByBlcokNum(executor.namespace, batch, seqNo, false, false); err != nil {
		executor.logger.Errorf("update chain to (#%d) failed, err: %s", err.Error())
		return err
	}
	// write bloom filter first
	bloom.WriteTxBloomFilter(executor.namespace, block.Transactions)

	if err := batch.Write(); err != nil {
		executor.logger.Errorf("commit (#%d) changes failed, err: %s", err.Error())
		return err
	}
	executor.statedb.MarkProcessFinish(seqNo)
	executor.TransitVerifiedBlock(block)
	executor.filterFeedback(result.Block, result.Logs)
	return nil
}

// reject clear all temporary cached stuff and send state updated event to consensus.
func (executor *Executor) reject() {
	executor.cache.syncCache.Purge()
	// clear all useless stuff
	batch := executor.db.NewBatch()
	for i := edb.GetHeightOfChain(executor.namespace) + 1; i <= executor.context.syncCtx.target; i += 1 {
		// delete persisted blocks number larger than chain height
		edb.DeleteBlockByNum(executor.namespace, batch, i, false, false)
	}
	batch.Write()
	executor.context.initDemand(edb.GetHeightOfChain(executor.namespace) + 1)
	executor.clearStatedb()
	executor.clearSyncFlag()
	executor.sendStateUpdatedEvent()
}

// isDemandSyncBlock checks whether the sync block is demand or not.
func (executor *Executor) isDemandSyncBlock(block *types.Block) bool {
	if block.Number == executor.context.syncCtx.demandBlockNum &&
		bytes.Compare(block.BlockHash, executor.context.syncCtx.demandBlockHash) == 0 {
		return true
	}
	return false
}

// calcuDownstream calculates a sync request downstream
// if a node required to sync too much blocks one time, the huge chain sync request will be split to several small one.
// a sync chain required block number can not more than `sync batch size` in config file.
func (executor *Executor) calcuDownstream() uint64 {
	if executor.context.syncCtx.updateGenesis {
		_, genesis := executor.context.syncCtx.getTargerGenesis()
		total := executor.context.syncCtx.getDownstream() - genesis + 1
		if total < executor.conf.GetSyncMaxBatchSize() {
			executor.context.syncCtx.setDownstream(genesis - 1)
		} else {
			executor.context.syncCtx.setDownstream(executor.context.syncCtx.getDownstream() - executor.conf.GetSyncMaxBatchSize())
		}
	} else {
		total := executor.context.syncCtx.getDownstream() - edb.GetHeightOfChain(executor.namespace)
		if total < executor.conf.GetSyncMaxBatchSize() {
			executor.context.syncCtx.setDownstream(edb.GetHeightOfChain(executor.namespace))
		} else {
			executor.context.syncCtx.setDownstream(executor.context.syncCtx.getDownstream() - executor.conf.GetSyncMaxBatchSize())
		}
	}

	executor.logger.Noticef("update temporarily downstream to %d", executor.context.syncCtx.getDownstream())
	return executor.context.syncCtx.getDownstream()
}

// receiveAllRequiredBlocks checks all required blocks has been received.
func (executor *Executor) receiveAllRequiredBlocks() bool {
	return executor.context.syncCtx.demandBlockNum == executor.context.syncCtx.getDownstream()
}

// storeFilterData stores filter data in record temporarily, avoid re-generated when using.
func (executor *Executor) storeFilterData(record *ValidationResultRecord, block *types.Block, logs []*types.Log) {
	record.Block = block
	record.Logs = logs
}

// syncinitialize initialize sync context and status.
func (executor *Executor) setupContext(ev event.ChainSyncReqEvent) {
	executor.context.syncCtx = newChainSyncContext(executor.namespace, ev, executor.conf.conf, executor.logger)
}

// applyWorldState apply the received world state, check the whole ledger hash before flush the changes to databse.
// if success, update the genesis tag with the world state tag(means genesis transition finished).
// Note, the whole operation is a atomic operation. If apply success, the genesis tag and current height will equal with
// given height; otherwise, no change will flush to db.
func (executor *Executor) applyWorldState(fPath string, filterId string, root common.Hash, genesis uint64) error {
	uncompressCmd := cmd.Command("tar", "-zxvf", fPath, "-C", filepath.Dir(fPath))
	if err := uncompressCmd.Run(); err != nil {
		return err
	}
	dbPath := path.Join("ws", "ws_"+filterId, "SNAPSHOT_"+filterId)
	wsDb, err := hyperdb.NewDatabase(executor.conf.conf, dbPath, hyperdb.GetDatabaseType(executor.conf.conf), executor.namespace)
	if err != nil {
		return err
	}
	defer wsDb.Close()

	writeBatch := executor.db.NewBatch()

	if err := executor.statedb.Merge(wsDb, writeBatch, root); err != nil {
		return err
	}

	if err := edb.UpdateGenesisTag(executor.namespace, genesis, writeBatch, false, false); err != nil {
		return err
	}

	// update current chain height equal with given world state height
	if err := edb.UpdateChainByBlcokNum(executor.namespace, writeBatch, genesis, false, false); err != nil {
		return err
	}

	if err := writeBatch.Write(); err != nil {
		return err
	}

	executor.logger.Noticef("apply ws pieces (%s) success", filterId)
	return nil
}

/*
	Sync chain Receiver
*/
// ReceiveSyncRequest receives` synchronization request from some nodes, and send back request blocks.
func (executor *Executor) ReceiveSyncRequest(payload []byte) {
	var request ChainSyncRequest
	if err := proto.Unmarshal(payload, &request); err != nil {
		executor.logger.Error("unmarshal sync request failed.")
		return
	}
	for i := request.RequiredNumber; i > request.CurrentNumber; i -= 1 {
		executor.informP2P(NOTIFY_UNICAST_BLOCK, i, request.PeerId, request.PeerHash)
	}
}

// ReceiveWorldStateSyncRequest receives ws request, send back handshake packet first time.
func (executor *Executor) ReceiveWorldStateSyncRequest(payload []byte) {
	var request WsRequest
	var fsize int64
	if err := proto.Unmarshal(payload, &request); err != nil {
		executor.logger.Warning("unmarshal world state sync request failed.")
		return
	}
	executor.logger.Noticef("receive world state sync req, required (#%d)", request.Target)
	err, manifest := executor.snapshotReg.rw.Search(request.Target)
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

	wsShardSize := executor.conf.GetStateFetchPacketSize()

	n := fsize / int64(wsShardSize)
	if fsize%int64(wsShardSize) > 0 {
		n += 1
	}
	hs := executor.constructWsHandshake(request, manifest.FilterId, uint64(fsize), uint64(n))
	if err := executor.informP2P(NOTIFY_SEND_WORLD_STATE_HANDSHAKE, hs); err != nil {
		executor.logger.Warningf("send world state (#%s) back to (%v) failed, err msg %s", manifest.FilterId, HashOrIdString(request.InitiatorIdOrHash), err.Error())
		return
	}
	executor.logger.Debugf("send world state (#%s) handshake back to (%v) success, total size %d, total packet num %d, max packet size %d",
		manifest.FilterId, HashOrIdString(request.InitiatorIdOrHash), hs.Size, hs.PacketNum, hs.PacketSize)
}

// ReceiveWsAck receive initiator's ack packet.
// send the next packet or resend the latest one depend on the ack status field.
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

	wsShardSize := int64(executor.conf.GetStateFetchPacketSize())

	sendWs := func(shardId uint64, filterId string, ws *WsAck) {
		fpath := executor.snapshotReg.CompressedSnapshotPath(filterId)
		err, reader := common.NewSectionReader(fpath, wsShardSize)
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
			executor.logger.Debugf("send ws(#%s) packet (#%d), packet size (#%d) to peer (#%v) success", ws.Ctx.FilterId, ws.PacketId, ws.PacketSize, HashOrIdString(ws.Ctx.ReceiverIdOrHash))
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
			sendWs(ack.PacketId+1, ack.Ctx.FilterId, &ack)
		}
	} else {
		// resend
		sendWs(ack.PacketId, ack.Ctx.FilterId, &ack)
	}

}

/*
	Net Packets
*/

// constructWsHandshake makes ws handshake packet.
func (executor *Executor) constructWsHandshake(req WsRequest, filterId string, size uint64, pn uint64) *WsHandshake {
	wsShardSize := int64(executor.conf.GetStateFetchPacketSize())
	return &WsHandshake{
		Ctx: &WsContext{
			FilterId:          filterId,
			InitiatorIdOrHash: common.CopyBytes(req.ReceiverIdOrHash),
			ReceiverIdOrHash:  common.CopyBytes(req.InitiatorIdOrHash),
		},
		Height:     req.Target,
		Size:       size,
		PacketSize: uint64(wsShardSize),
		PacketNum:  pn,
	}
}

// constructWsHandshake makes ws ack packet.
func (executor *Executor) constructWsAck(ctx *WsContext, packetId uint64, status WsAck_STATUS, message []byte) *WsAck {
	return &WsAck{
		Ctx: &WsContext{
			FilterId:          ctx.FilterId,
			InitiatorIdOrHash: common.CopyBytes(ctx.ReceiverIdOrHash),
			ReceiverIdOrHash:  common.CopyBytes(ctx.InitiatorIdOrHash),
		},
		PacketId: packetId,
		Status:   status,
		Message:  message,
	}
}

// constructWsHandshake makes ws packet.
func (executor *Executor) constrcutWs(ack *WsAck, packetId uint64, packetSize uint64, payload []byte) *Ws {
	executor.logger.Debugf("construct ws packet with %d size", len(payload))
	return &Ws{
		Ctx: &WsContext{
			FilterId:          ack.Ctx.FilterId,
			InitiatorIdOrHash: common.CopyBytes(ack.Ctx.ReceiverIdOrHash),
			ReceiverIdOrHash:  common.CopyBytes(ack.Ctx.InitiatorIdOrHash),
		},
		PacketId:   packetId,
		PacketSize: packetSize,
		Payload:    payload,
	}
}

func getTxVersion(block *types.Block) string {
	if block == nil || len(block.Transactions) == 0 {
		// short circuit if block is empty or no transaction embeded.
		return edb.TransactionVersion
	}
	return string(block.Transactions[0].Version)
}
