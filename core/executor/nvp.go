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
	"encoding/hex"
	"math"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"hyperchain/common"
	er "hyperchain/core/errors"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/types"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
)

// TransitVerifiedBlock transits a verified block to non-verified peers.
func (executor *Executor) TransitVerifiedBlock(block *types.Block) {
	data, err := proto.Marshal(block)
	if err != nil {
		executor.logger.Errorf("marshal verified block for #%d failed.", block.Number)
		return
	}
	executor.informP2P(NOTIFY_TRANSIT_BLOCK, data)
}

/*
	Non-Verifying Peer
*/

// nvp implements the NonVerifyingPeer interface.
type nvp struct {
	lock         sync.Mutex
	executor     *Executor
	logger       *logging.Logger
	localhash    string
	demandNumber uint64
	resendExit   chan struct{}
}

// NewNVPImpl creates the nvp object.
func NewNVPImpl(executor *Executor, localhash string) *nvp {
	return &nvp{
		demandNumber: chain.GetChainCopy(executor.namespace).Height + 1,
		localhash:    localhash,
		executor:     executor,
		logger:       executor.logger,
		resendExit:   make(chan struct{}),
	}
}

// ReceiveBlock receives block from vp, process them in serial but may out of order.
func (nvp *nvp) ReceiveBlock(payload []byte) {
	// Hold the lock, prevent double access.
	nvp.lock.Lock()
	defer nvp.lock.Unlock()

	block, err, skip := nvp.preprocess(payload)
	if skip {
		return
	}
	if err != nil {
		nvp.logger.Error(err)
		return
	}

	if nvp.isDemand(block.Number) {
		err = nvp.handleDemand(block)
	} else {
		err = nvp.handleUndemand(block)
	}
	if err != nil {
		nvp.logger.Errorf("handle block #%d failed. %s", block.Number, err.Error())
	}
}

// ReceiveConsult
func (nvp *nvp) ReceiveConsult(payload []byte, n int) {
	nvp.logger.Debugf("receive consult reply, max reply number %d", n)
	if nvp.getSyncContext() == nil {
		// ignore consult reply if state update finish.
		return
	}
	if wait := atomic.LoadInt32(&nvp.getSyncContext().waitConsultReply); wait == 0 {
		// ignore consult reply if not in wait stage
		return
	}
	reply := &ConsultReply{}
	if err := proto.Unmarshal(payload, reply); err != nil {
		nvp.logger.Error("unmarshal consult reply failed", err.Error())
		return
	}
	nvp.getSyncContext().replys = append(nvp.getSyncContext().replys, reply)
	if !reply.Transition || len(nvp.getSyncContext().replys) == n {
		retry := 3
		for retry > 0 {
			if atomic.CompareAndSwapInt32(&nvp.getSyncContext().waitConsultReply, 1, 0) {
				nvp.logger.Notice("receive enough consult reply, stop waiting")
				break
			}
			retry--
		}
		if retry == 0 {
			nvp.logger.Error("receive enough consult replys, but shut down wait failed")
		}
	}
}

// GetLocalHash returns nvp's identification.
func (nvp *nvp) GetLocalHash() string {
	return nvp.localhash
}

// handleDemand receives demand block and process it.
// Note. Nvp maybe in state synchronization, if so, modify some control flags.
func (nvp *nvp) handleDemand(block *types.Block) error {
	if err := nvp.applyBlock(block); err != nil {
		nvp.logger.Errorf("apply block #%d failed. %s", block.Number, err.Error())
		return err
	}
	if err := nvp.applyRemainBlock(nvp.demandNumber); err != nil {
		nvp.logger.Errorf("apply remain block failed. %s", block.Number, err.Error())
		return err
	}
	if nvp.getSyncContext() != nil {
		if nvp.getSyncContext().max == chain.GetHeightOfChain(nvp.getExecutor().namespace)-1 {
			// All blocks received
			nvp.clearSyncContext()
		} else {
			// Still have some unreceived blocks, try to fetch another batch blocks.
			if nvp.getSyncContext().getDown() >= nvp.getSyncContext().getUpper() {
				nvp.calUpper()
				nvp.decUpper(block)
				nvp.sendSyncRequest(nvp.getSyncContext().getUpper(), nvp.getSyncContext().getDown(),
					nvp.getSyncContext().remote)
				nvp.logger.Debugf("try to fetch next batch: (%d - %d]",
					nvp.getSyncContext().getDown(), nvp.getSyncContext().getUpper())
			}
		}
	}
	return nil
}

// handleUndemand receives unexpect block, may trigger a synchronization procedure.
func (nvp *nvp) handleUndemand(block *types.Block) error {
	// The arrival block is not the demand one.
	nvp.logger.Debugf("receive unexpected block number: %d, hash: %s", block.Number, common.Bytes2Hex(block.BlockHash))
	if nvp.getSyncContext() == nil {
		nvp.setupSyncContext(block.Number-1, chain.GetHeightOfChain(nvp.getExecutor().namespace))
		nvp.logger.Debug("local blockchain is not continuous with the vp blockchain, start to synchronization")
		nvp.logger.Debugf("require to fetch (%d - %d] block", nvp.getSyncContext().getDown(), nvp.getSyncContext().max)

		var reply *ConsultReply = nvp.consult()
		if reply == nil {
			nvp.logger.Error("doesn't receive enough consult reply in 5 minutes. Maybe this peer is isolated with others")
			nvp.clearSyncContext()
			return er.NotEnoughReplyErr
		}
		if reply.Transition {
			// Make world state transition.
			// Same with consult, this procedure will also block util it finish.

			// Persist new genesis block here.
			chain.PersistBlock(nvp.executor.db.NewBatch(), reply.GenesisBlock, true, true, getTxVersion(reply.GenesisBlock))
			nvp.logger.Debugf("world state transition is required. target genesis %d, target peer %d",
				reply.GenesisBlock.Number, reply.PeerId)

			nvp.getSyncContext().remote = reply.PeerId
			nvp.getSyncContext().down = reply.GenesisBlock.Number
			nvp.demandNumber = reply.GenesisBlock.Number + 1

			if err := nvp.fetchWorldState(reply); err != nil {
				nvp.clearSyncContext()
				return err
			}
		}
		nvp.sendSyncRequest(nvp.calUpper(), nvp.getSyncContext().getDown(), nvp.getSyncContext().remote)
		go nvp.resendBackend()
	}
	if block.Number > nvp.getSyncContext().max {
		nvp.getSyncContext().max = block.Number - 1
	}
	nvp.decUpper(block)
	return nil
}

// preProcess do all preparatory work before execute a received block.
// Includes:
// 1. Unserialize block from serialized payload;
// 2. Ignore block which number is not larger than chain height;
// 3. Verify block intergrity via block hash;
// 4. Persist to database if meet demand;
func (nvp *nvp) preprocess(payload []byte) (*types.Block, error, bool) {
	nvp.logger.Debugf("preprocess received block")
	block := &types.Block{}
	err := proto.Unmarshal(payload, block)
	if err != nil {
		nvp.logger.Errorf("receive invalid verified block, unmarshal failed.")
		return nil, err, false
	}
	if chain.GetHeightOfChain(nvp.getExecutor().namespace) < block.Number {
		blk, err := chain.GetBlockByNumber(nvp.getExecutor().namespace, block.Number)
		if err != nil {
			// The block with given number doesn't exist in database.
			if !VerifyBlockIntegrity(block) {
				return nil, er.BlockIntegrityErr, false
			}
			if block.Version != nil {
				// Receive block with version tag
				_, err = chain.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true, string(block.Version), getTxVersion(block))
			} else {
				_, err = chain.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true)
			}
			if err != nil {
				return nil, err, false
			}
			nvp.logger.Debugf("pre-process block #%d done!", block.Number)
			return block, nil, false
		} else {
			// Ignore duplicated block
			nvp.logger.Debugf("ignore duplicated block %d, hash %s, skip persistence", block.Number, common.Bytes2Hex(block.BlockHash))
			return blk, nil, false
		}
	} else {
		return nil, nil, true
	}
}

// applyBlock executes given block to make state transition.
func (nvp *nvp) applyBlock(block *types.Block) error {
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.updateDemand()
	// increase down bound if nvp is in status synchronization
	if nvp.getSyncContext() != nil {
		nvp.getSyncContext().setDown(block.Number)
	}
	return nil
}

// applyRemainBlock checks whether there exists earlier arrival continuous blocks and execute them.
func (nvp *nvp) applyRemainBlock(number uint64) error {
	block, err := chain.GetBlockByNumber(nvp.getExecutor().namespace, number)
	if err != nil {
		// No continuous block found.
		return nil
	}
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.updateDemand()
	// increase down bound if nvp is in status synchronization
	if nvp.getSyncContext() != nil {
		nvp.getSyncContext().setDown(block.Number)
	}
	return nvp.applyRemainBlock(nvp.demandNumber)
}

// process applys given block to make blockchain status transition.
func (nvp *nvp) process(block *types.Block) error {
	nvp.executor.stateTransition(block.Number+1, common.BytesToHash(block.MerkleRoot))
	err, result := nvp.getExecutor().ApplyBlock(block, block.Number)
	if err != nil {
		return err
	}
	if nvp.getExecutor().assertApplyResult(block, result) == false && nvp.getExecutor().conf.GetExitFlag() {
		batch := nvp.getExecutor().db.NewBatch()
		for i := block.Number; ; i += 1 {
			// delete persisted blocks number larger than chain height
			err := chain.DeleteBlockByNum(nvp.getExecutor().namespace, batch, i, false, false)
			if err != nil {
				nvp.logger.Debugf("delete block number #%v in batch failed! ErrMsg: %v.", i, err.Error())
			} else {
				nvp.logger.Debugf("delete block number #%v in batch success!", i)
			}
			if nvp.getSyncContext() == nil || i == nvp.getSyncContext().max+1 {
				break
			}
		}
		err := batch.Write()
		if err != nil {
			nvp.logger.Error("delete blocks in db failed! ErrMsg: %v.", err.Error())
		}
		nvp.getExecutor().clearStatedb()
		nvp.logger.Error("assert failed, exit hyperchain.")
		syscall.Exit(0)
	}
	nvp.getExecutor().accpet(block.Number, block, result)
	nvp.logger.Notice("Block number", block.Number)
	nvp.logger.Notice("Block hash", hex.EncodeToString(block.BlockHash))
	return nil
}

func (nvp *nvp) resendBackend() {
	ticker := time.NewTicker(nvp.getExecutor().conf.GetSyncResendInterval())
	nvp.logger.Debug("sync request resend interval: ", nvp.getExecutor().conf.GetSyncResendInterval().String())
	up := nvp.getSyncContext().getUpper()
	down := nvp.getSyncContext().getDown()
	for {
		select {
		case <-nvp.resendExit:
			return
		case <-ticker.C:
			// resend
			if nvp.getSyncContext() == nil {
				continue
			}
			curUp := nvp.getSyncContext().getUpper()
			curDown := nvp.getSyncContext().getDown()
			if curUp == up && curDown == down {
				nvp.logger.Noticef("resend sync request, want [%d] - [%d]", down, up)
				nvp.sendSyncRequest(up, down, nvp.getSyncContext().remote)
			} else {
				up = curUp
				down = curDown
			}
		}
	}
}

func (nvp *nvp) getExecutor() *Executor {
	return nvp.executor
}

// calUpper calculates max demand number in a sync request
// if a node required to sync too much blocks one time, the huge chain sync request will be split to several small one.
// a sync chain required block number can not more than `sync batch size` in config file.
func (nvp *nvp) calUpper() uint64 {
	total := nvp.getSyncContext().max - nvp.getSyncContext().getDown()
	if total < nvp.getExecutor().conf.GetSyncMaxBatchSize() {
		nvp.getSyncContext().setUpper(nvp.getSyncContext().max)
	} else {
		nvp.getSyncContext().setUpper(nvp.getSyncContext().getDown() + nvp.getExecutor().conf.GetSyncMaxBatchSize())
	}
	nvp.logger.Debugf("update upper to %d", nvp.getSyncContext().getUpper())
	return nvp.getSyncContext().getUpper()
}

// sendSyncRequest sends synchronization request to other vp nodes.
func (nvp *nvp) sendSyncRequest(upstream, downstream, peerId uint64) {
	if err := nvp.getExecutor().informP2P(NOTIFY_NVP_SYNC, upstream, downstream, peerId); err != nil {
		nvp.getExecutor().logger.Errorf("[Namespace = %s] send sync req failed. %s", nvp.getExecutor().namespace, err.Error())
		return
	}
}

// decUpper decrease upper in a single batch request.
func (nvp *nvp) decUpper(block *types.Block) {
	if nvp.getSyncContext().getUpper() == block.Number {
		nvp.getSyncContext().setUpper(block.Number - 1)
	}
	for blk, err := chain.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getSyncContext().getUpper()); err == nil; {
		// TODO error can not be NOTFOUND
		if nvp.getSyncContext().getUpper() <= nvp.getSyncContext().getDown() {
			break
		}
		nvp.logger.Debugf("db already has block number #%v. block hash %v.", blk.Number, common.Bytes2Hex(blk.BlockHash))
		nvp.getSyncContext().setUpper(nvp.getSyncContext().getUpper() - 1)
		blk, err = chain.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getSyncContext().getUpper())
	}
}

// negotiate nvp sends a consult message to connected vps to assure whether genesis status transition is necesary.
// Presume sync context is not nil.
func (nvp *nvp) consult() *ConsultReply {
	// hold the lock
	retry := 3
	for retry > 0 {
		if atomic.CompareAndSwapInt32(&nvp.getSyncContext().waitConsultReply, 0, 1) {
			break
		}
		retry--
	}
	if retry == 0 {
		return nil
	}

	nvp.executor.informP2P(NOTIFY_NVP_CONSULT)
	// wait consult result. return false if doesn't get result in time.
	ticker := time.NewTicker(time.Second)
	timeout := time.NewTimer(5 * time.Minute)

	analysis := func() *ConsultReply {
		replys := nvp.getSyncContext().replys
		// clear immediately
		nvp.getSyncContext().replys = nil
		var (
			reply *ConsultReply
			min   uint64 = math.MaxUint64
		)
		for _, r := range replys {
			// full node always has the highest priority
			if !r.Transition {
				return r
			}
			// lower genesis node has higher priority
			if min > r.GenesisBlock.Number {
				min = r.GenesisBlock.Number
				reply = r
			}
		}
		nvp.getSyncContext().reply = reply
		return reply
	}
	for {
		select {
		case <-ticker.C:
			if wait := atomic.LoadInt32(&nvp.getSyncContext().waitConsultReply); wait == 0 {
				return analysis()
			} else {
				nvp.logger.Debug("haven't receive enough reply")
			}
		case <-timeout.C:
			nvp.logger.Warning("wait more than 5 minutes to receive enough replys. return any available right now")
			atomic.StoreInt32(&nvp.getSyncContext().waitConsultReply, 0)
			return analysis()
		}
	}
}

func (nvp *nvp) fetchWorldState(reply *ConsultReply) error {
	// mark world transition begin
	retry := 3
	for retry > 0 {
		if atomic.CompareAndSwapInt32(&nvp.getSyncContext().waitTransition, 0, 1) {
			break
		}
		retry--
	}
	if retry == 0 {
		return er.SeizeLockFailedErr
	}
	nvp.executor.informP2P(NOTIFY_REQUEST_WORLD_STATE, reply.GenesisBlock.Number, reply.PeerId, true)
	ticker := time.NewTicker(time.Second)
	timeout := time.NewTimer(24 * time.Hour)
	for {
		select {
		case <-ticker.C:
			if wait := atomic.LoadInt32(&nvp.getSyncContext().waitTransition); wait == 0 {
				return nil
			} else {
				nvp.logger.Debug("wait world state transition finish")
			}
		case <-timeout.C:
			nvp.logger.Error("exceed 24 hours waiting for world state transition")
			return er.TimeoutErr
		}
	}
}

func (nvp *nvp) getSyncContext() *chainSyncContext {
	return nvp.executor.context.syncCtx
}

func (nvp *nvp) setupSyncContext(num, chainHeight uint64) {
	nvp.executor.context.syncCtx = newNVPSyncContext(num, chainHeight)
}

func (nvp *nvp) clearSyncContext() {
	nvp.executor.context.syncCtx = nil
	nvp.resendExit <- struct{}{}
	nvp.logger.Notice("state update finish")
}

func (nvp *nvp) updateDemand() {
	nvp.demandNumber += 1
}

func (nvp *nvp) isDemand(id uint64) bool {
	return nvp.demandNumber == id
}
