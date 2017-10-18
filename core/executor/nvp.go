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
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
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

type nvpContext struct {
	demandNumber uint64        // current demand number for commit
	inSync       bool          // indicates whether nvp is in status synchronization
	max          uint64        // max demand block number during sync
	upper        uint64        // max demand block number in a single request batch
	down         uint64        // min demand block number in a single request batch
	resendExit   chan struct{} // resend backend exit notifier
}

func (ctx *nvpContext) updateDemand() {
	ctx.demandNumber += 1
}

func (ctx *nvpContext) isDemand(id uint64) bool {
	return ctx.demandNumber == id
}

func (ctx *nvpContext) getUpper() uint64 {
	return atomic.LoadUint64(&ctx.upper)
}

func (ctx *nvpContext) setUpper(num uint64) {
	atomic.StoreUint64(&ctx.upper, num)
}

func (ctx *nvpContext) getDown() uint64 {
	return atomic.LoadUint64(&ctx.down)
}

func (ctx *nvpContext) setDown(num uint64) {
	atomic.StoreUint64(&ctx.down, num)
}

func (ctx *nvpContext) initSync(num, chainHeight uint64) {
	ctx.max, ctx.down = num, chainHeight
	ctx.inSync = true
	ctx.resendExit = make(chan struct{})
}

func (ctx *nvpContext) clear() {
	ctx.inSync = false
	ctx.max, ctx.down, ctx.upper = 0, 0, 0
	close(ctx.resendExit)
}

/*
	Non-Verifying Peer
*/

type nvp struct {
	lock     sync.Mutex // block processor preventor
	ctx      *nvpContext
	executor *Executor
	logger   *logging.Logger
}

func NewNVPImpl(executor *Executor) *nvp {
	return &nvp{
		ctx:      &nvpContext{demandNumber: chain.GetChainCopy(executor.namespace).Height + 1},
		executor: executor,
		logger:   executor.logger,
	}
}

// ReceiveBlock receives block from vp, process them in serial but may out of order.
func (nvp *nvp) ReceiveBlock(payload []byte) {
	// Hold the lock, prevent double access.
	nvp.lock.Lock()
	defer nvp.lock.Unlock()

	var err error

	block, err := nvp.preprocess(payload)
	if err != nil {
		nvp.logger.Error(err)
		return
	}
	if nvp.getCtx().isDemand(block.Number) {
		err = nvp.handleDemand(block)
	} else {
		err = nvp.handleUndemand(block)
	}
	if err != nil {
		nvp.logger.Errorf("handle block #%d failed. %s", block.Number, err.Error())
	}
}

// handleDemand receives demand block and process it.
// Note. Nvp maybe in state synchronization, if so, modify some control flags.
func (nvp *nvp) handleDemand(block *types.Block) error {
	if err := nvp.applyBlock(block); err != nil {
		nvp.logger.Errorf("apply block #%d failed. %s", block.Number, err.Error())
		return err
	}
	if err := nvp.applyRemainBlock(nvp.ctx.demandNumber); err != nil {
		nvp.logger.Errorf("apply remain block failed. %s", block.Number, err.Error())
		return err
	}
	if nvp.getCtx().inSync {
		if nvp.getCtx().max == chain.GetHeightOfChain(nvp.getExecutor().namespace)-1 {
			// All blocks received
			nvp.getCtx().clear()
		} else {
			// Still have some unreceived blocks, try to fetch another batch blocks.
			if nvp.getCtx().getDown() >= nvp.getCtx().getUpper() {
				nvp.calUpper()
				nvp.decUpper(block)
				nvp.sendSyncRequest(nvp.getCtx().getUpper(), nvp.getCtx().getDown())
				nvp.getExecutor().logger.Debugf("In sync phase! next batch: minNumInBatch is %v. maxNumInBatch is %v.", nvp.getCtx().getDown(), nvp.getCtx().getUpper())
			}
		}
	}
	return nil
}

// handleUndemand receives unexpect block, may trigger a synchronization procedure.
func (nvp *nvp) handleUndemand(block *types.Block) error {
	// The arrival block is not the demand one.
	nvp.logger.Debugf("receive unexpected block number: %d, hash: %s", block.Number, common.Bytes2Hex(block.BlockHash))
	if !nvp.getCtx().inSync {
		// TODO negotiate with vp to assure whether genesis transition is necessary.
		// Set up block synchronization if nvp is not in sync status.
		nvp.getCtx().initSync(block.Number-1, chain.GetHeightOfChain(nvp.getExecutor().namespace))
		nvp.logger.Debug("local blockchain is not continuous with the vp blockchain, start to synchronization")
		nvp.logger.Debugf("require to fetch (%d - %d] block", nvp.getCtx().getDown(), nvp.getCtx().max)
		nvp.sendSyncRequest(nvp.calUpper(), nvp.getCtx().getDown())
		go nvp.resendBackend()
	}
	if block.Number > nvp.getCtx().max {
		nvp.ctx.max = block.Number - 1
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
func (nvp *nvp) preprocess(payload []byte) (*types.Block, error) {
	nvp.getExecutor().logger.Debugf("pre-process block of NVP start!")
	block := &types.Block{}
	err := proto.Unmarshal(payload, block)
	if err != nil {
		nvp.getExecutor().logger.Errorf("receive invalid verified block, unmarshal failed.")
		return nil, err
	}
	if chain.GetHeightOfChain(nvp.getExecutor().namespace) < block.Number {
		blk, err := chain.GetBlockByNumber(nvp.getExecutor().namespace, block.Number)
		if err != nil {
			// The block with given number doesn't exist in database.
			if !VerifyBlockIntegrity(block) {
				errStr := fmt.Sprintf("verify block integrity fail! receive a broken block %d, drop it.", block.Number)
				return nil, errors.New(errStr)
			}
			if block.Version != nil {
				// Receive block with version tag
				err, _ = chain.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true, string(block.Version), getTxVersion(block))
			} else {
				err, _ = chain.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true)
			}
			if err != nil {
				return nil, errors.New("persisit block failed." + err.Error())
			}
			nvp.logger.Debugf("pre-process block #%d done!", block.Number)
			return block, nil
		} else {
			// Ignore duplicated block
			nvp.logger.Debugf("ignore duplicated block %d, hash %s, skip persistence", block.Number, common.Bytes2Hex(block.BlockHash))
			return blk, nil
		}
	} else {
		errStr := fmt.Sprintf("unexpected block #%d received", block.Number)
		return nil, errors.New(errStr)
	}
}

// applyBlock executes given block to make state transition.
func (nvp *nvp) applyBlock(block *types.Block) error {
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.getCtx().updateDemand()
	// increase down bound if nvp is in status synchronization
	if nvp.getCtx().inSync {
		nvp.getCtx().setDown(block.Number)
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
	nvp.getCtx().updateDemand()
	// increase down bound if nvp is in status synchronization
	if nvp.getCtx().inSync {
		nvp.getCtx().setDown(block.Number)
	}
	return nvp.applyRemainBlock(nvp.getCtx().demandNumber)
}

// process applys given block to make blockchain status transition.
func (nvp *nvp) process(block *types.Block) error {
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
			if !nvp.getCtx().inSync || i == nvp.getCtx().max+1 {
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
	nvp.logger.Debugf("sync request resend interval: ", nvp.getExecutor().conf.GetSyncResendInterval().String())
	up := nvp.getCtx().getUpper()
	down := nvp.getCtx().getDown()
	for {
		select {
		case <-nvp.getCtx().resendExit:
			nvp.logger.Notice("state update finish")
			return
		case <-ticker.C:
			// resend
			curUp := nvp.getCtx().getUpper()
			curDown := nvp.getCtx().getDown()
			if curUp == up && curDown == down {
				nvp.logger.Noticef("resend sync request of NVP. want [%d] - [%d]", down, up)
				nvp.sendSyncRequest(up, down)
			} else {
				up = curUp
				down = curDown
			}
		}
	}
}

func (nvp *nvp) getCtx() *nvpContext {
	return nvp.ctx
}

func (nvp *nvp) getExecutor() *Executor {
	return nvp.executor
}

// calUpper calculates max demand number in a sync request
// if a node required to sync too much blocks one time, the huge chain sync request will be split to several small one.
// a sync chain required block number can not more than `sync batch size` in config file.
func (nvp *nvp) calUpper() uint64 {
	total := nvp.getCtx().max - nvp.getCtx().getDown()
	if total < nvp.getExecutor().conf.GetSyncMaxBatchSize() {
		nvp.getCtx().setUpper(nvp.getCtx().max)
	} else {
		nvp.ctx.setUpper(nvp.getCtx().getDown() + nvp.getExecutor().conf.GetSyncMaxBatchSize())
	}
	nvp.getExecutor().logger.Debugf("update upper to %d", nvp.getCtx().getUpper())
	return nvp.getCtx().getUpper()
}

// sendSyncRequest sends synchronization request to other vp nodes.
func (nvp *nvp) sendSyncRequest(upstream, downstream uint64) {
	if err := nvp.getExecutor().informP2P(NOTIFY_NVP_SYNC, upstream, downstream); err != nil {
		nvp.getExecutor().logger.Errorf("[Namespace = %s] send sync req failed. %s", nvp.getExecutor().namespace, err.Error())
		return
	}
}

// decUpper decrease upper in a single batch request.
func (nvp *nvp) decUpper(block *types.Block) {
	if nvp.getCtx().getUpper() == block.Number {
		nvp.getCtx().setUpper(block.Number - 1)
	}
	for blk, err := chain.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getCtx().getUpper()); err == nil; {
		if nvp.getCtx().getUpper() <= nvp.getCtx().getDown() {
			break
		}
		nvp.logger.Debugf("db already has block number #%v. block hash %v.", blk.Number, common.Bytes2Hex(blk.BlockHash))
		nvp.getCtx().setUpper(nvp.getCtx().getUpper() - 1)
		blk, err = chain.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getCtx().getUpper())
	}
}
