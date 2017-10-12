package executor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core/ledger/db_utils"
	"hyperchain/core/types"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// TransitVerifiedBlock - transit a verified block to non-verified peers.
func (executor *Executor) TransitVerifiedBlock(block *types.Block) {
	data, err := proto.Marshal(block)
	if err != nil {
		executor.logger.Errorf("marshal verified block for #%d failed.", block.Number)
		return
	}
	executor.informP2P(NOTIFY_TRANSIT_BLOCK, data)
}

/*-------------------------NVPContext------------------------*/

type NVPContext interface {
	getDemand() uint64
	updateDemand()
	isDemand(id uint64) bool

	getMax() uint64
	setMax(num uint64)
	getIsInSync() bool
	setIsInSync(flag bool)
	getUpper() uint64
	setUpper(num uint64)
	getDown() uint64
	setDown(num uint64)
	getResendExit() chan bool

	initSync(num, chainHeight uint64)
	initResendExit()
	clear()
}

type NVPContextImpl struct {
	demandNumber uint64 // current demand number for commit

	isInSync   bool      // for NVP use, to indicate whether send demand block repeatedly.
	max        uint64    // for NVP use, get max demand block number during sync
	upper      uint64    // for NVP use, get max demand block number in batch during sync
	down       uint64    // for NVP use, get min demand block number in batch during sync
	resendExit chan bool // for NVP use, resend backend process notifier
}

func NewNVPContextImpl(executor *Executor) *NVPContextImpl {
	currentChain := db_utils.GetChainCopy(executor.namespace)
	return &NVPContextImpl{
		demandNumber: currentChain.Height + 1,
	}
}

func (ctx *NVPContextImpl) getDemand() uint64 {
	return ctx.demandNumber
}

func (ctx *NVPContextImpl) updateDemand() {
	ctx.demandNumber += 1
}

func (ctx *NVPContextImpl) isDemand(id uint64) bool {
	return ctx.getDemand() == id
}

func (ctx *NVPContextImpl) getMax() uint64 {
	return ctx.max
}

func (ctx *NVPContextImpl) setMax(num uint64) {
	ctx.max = num
}

func (ctx *NVPContextImpl) getIsInSync() bool {
	return ctx.isInSync
}

func (ctx *NVPContextImpl) setIsInSync(flag bool) {
	ctx.isInSync = flag
}

func (ctx *NVPContextImpl) getUpper() uint64 {
	return atomic.LoadUint64(&ctx.upper)
}

func (ctx *NVPContextImpl) setUpper(num uint64) {
	atomic.StoreUint64(&ctx.upper, num)
}

func (ctx *NVPContextImpl) getDown() uint64 {
	return atomic.LoadUint64(&ctx.down)
}

func (ctx *NVPContextImpl) setDown(num uint64) {
	atomic.StoreUint64(&ctx.down, num)
}

func (ctx *NVPContextImpl) getResendExit() chan bool {
	return ctx.resendExit
}

func (ctx *NVPContextImpl) initSync(num, chainHeight uint64) {
	ctx.setMax(num)
	ctx.setDown(chainHeight)
	ctx.setIsInSync(true)
}

func (ctx *NVPContextImpl) initResendExit() {
	ctx.resendExit = make(chan bool)
}

func (ctx *NVPContextImpl) clear() {
	ctx.setIsInSync(false)
	ctx.setMax(0)
	ctx.setDown(0)
	ctx.setUpper(0)
	ctx.getResendExit() <- true
}

/*-------------------------NVP------------------------*/

type NVP interface {
	ReceiveBlock(payload []byte)
}

type NVPImpl struct {
	lock     sync.Mutex
	ctx      NVPContext
	executor *Executor
}

func NewNVPImpl(executor *Executor) *NVPImpl {
	return &NVPImpl{
		ctx:      NewNVPContextImpl(executor),
		executor: executor,
	}
}

// ReceiveBlock - receive block from vp.
// process in serial but may out of order.
func (nvp *NVPImpl) ReceiveBlock(payload []byte) {

	nvp.getExecutor().logger.Debug("receive block")
	nvp.lock.Lock()
	defer nvp.lock.Unlock()

	block, err := nvp.preProcess(payload)
	if err != nil {
		nvp.executor.logger.Error(err)
		return
	}
	if nvp.getCtx().isDemand(block.Number) {
		if err := nvp.applyBlock(block); err != nil {
			nvp.executor.logger.Errorf("apply block #%d failed. %s", block.Number, err.Error())
			return
		}
		if err := nvp.applyRemainBlock(nvp.ctx.getDemand()); err != nil {
			nvp.executor.logger.Errorf("apply remain block failed. %s", block.Number, err.Error())
			return
		}
		if nvp.isInSync() {
			nvp.getExecutor().logger.Debugf("In sync phase! chain height is %v. minNumInBatch is %v. maxNum is %v", db_utils.GetHeightOfChain(nvp.getExecutor().namespace), nvp.getCtx().getDown(), nvp.getCtx().getMax())
			isSyncDone := func() bool {
				return nvp.getCtx().getMax() == db_utils.GetHeightOfChain(nvp.getExecutor().namespace)-1
			}
			if isSyncDone() {
				nvp.getCtx().clear()
			} else {
				nvp.getExecutor().logger.Debugf("In sync phase! now batch: minNumInBatch is %v. maxNumInBatch is %v.", nvp.getCtx().getDown(), nvp.getCtx().getUpper())
				getNextBatch := func() {
					if nvp.getCtx().getDown() >= nvp.getCtx().getUpper() {
						nvp.calUpper()
						nvp.decUpper(block)
						nvp.sendSyncRequest(nvp.getCtx().getUpper(), nvp.getCtx().getDown())
						nvp.getExecutor().logger.Debugf("In sync phase! next batch: minNumInBatch is %v. maxNumInBatch is %v.", nvp.getCtx().getDown(), nvp.getCtx().getUpper())
					}
				}
				getNextBatch()
			}
		}
	} else {
		if !nvp.isInSync() {
			nvp.getCtx().initSync(block.Number-1, db_utils.GetHeightOfChain(nvp.getExecutor().namespace))
			nvp.getExecutor().logger.Debugf("sync init result: maxNum: %v, minNumInBatch: %v, isInSync: %v", nvp.getCtx().getMax(), nvp.getCtx().getDown(), nvp.isInSync())
			nvp.sendSyncRequest(nvp.calUpper(), nvp.getCtx().getDown())
			nvp.getCtx().initResendExit()
			go nvp.resendBackend()
		}
		if block.Number > nvp.getCtx().getMax() {
			nvp.getCtx().setMax(block.Number - 1)
		}
		nvp.getExecutor().logger.Debugf("receive block number: %v, MaxNum: %v.", block.Number, nvp.getCtx().getMax())
		nvp.decUpper(block)
		nvp.getExecutor().logger.Debugf("receive block number: %v, MaxNumInBatch: %v.", block.Number, nvp.getCtx().getUpper())
	}
}

func (nvp *NVPImpl) preProcess(payload []byte) (*types.Block, error) {
	nvp.getExecutor().logger.Debugf("pre-process block of NVP start!")
	block := &types.Block{}
	err := proto.Unmarshal(payload, block)
	if err != nil {
		nvp.getExecutor().logger.Errorf("receive invalid verified block, unmarshal failed.")
		return nil, err
	}
	if db_utils.GetHeightOfChain(nvp.getExecutor().namespace) < block.Number {
		blk, err := db_utils.GetBlockByNumber(nvp.getExecutor().namespace, block.Number)
		if err != nil {
			if !VerifyBlockIntegrity(block) {
				errStr := fmt.Sprintf("verify block integrity fail! receive a broken block %d, drop it.", block.Number)
				return nil, errors.New(errStr)
			}
			if block.Version != nil {
				// Receive block with version tag
				err, _ = db_utils.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true, string(block.Version), getTxVersion(block))
			} else {
				err, _ = db_utils.PersistBlock(nvp.getExecutor().db.NewBatch(), block, true, true)
			}
			if err != nil {
				return nil, errors.New("put block into DB failed." + err.Error())
			}
			nvp.getExecutor().logger.Debugf("pre-process block #%d done!", block.Number)
			return block, nil
		} else {
			nvp.getExecutor().logger.Debugf("db has persisted block #%d. the block in db num is #%d, hash is %v. Just ignore it.", block.Number, blk.Number, common.Bytes2Hex(blk.BlockHash))
			return blk, nil
		}
	} else {
		errStr := fmt.Sprintf("core height is %v, ignore block#%v.", db_utils.GetHeightOfChain(nvp.getExecutor().namespace), block.Number)
		return nil, errors.New(errStr)
	}
}

func (nvp *NVPImpl) applyBlock(block *types.Block) error {
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.getCtx().updateDemand()
	return nil
}

func (nvp *NVPImpl) applyRemainBlock(number uint64) error {
	block, err := db_utils.GetBlockByNumber(nvp.getExecutor().namespace, number)
	if err != nil {
		return nil
	}
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.getCtx().updateDemand()
	return nvp.applyRemainBlock(nvp.getCtx().getDemand())
}

func (nvp *NVPImpl) process(block *types.Block) error {
	err, result := nvp.getExecutor().ApplyBlock(block, block.Number)
	if err != nil {
		return err
	}
	if nvp.getExecutor().assertApplyResult(block, result) == false {
		if nvp.getExecutor().GetExitFlag() {
			batch := nvp.getExecutor().db.NewBatch()
			for i := block.Number; ; i += 1 {
				// delete persisted blocks number larger than chain height
				err := db_utils.DeleteBlockByNum(nvp.getExecutor().namespace, batch, i, false, false)
				if err != nil {
					nvp.getExecutor().logger.Errorf("delete block number #%v in batch failed! ErrMsg: %v.", i, err.Error())
				} else {
					nvp.getExecutor().logger.Noticef("delete block number #%v in batch success!", i)
				}
				if !nvp.isInSync() || i == nvp.getCtx().getMax()+1 {
					break
				}
			}
			err := batch.Write()
			if err != nil {
				nvp.getExecutor().logger.Error("delete blocks in db failed! ErrMsg: %v.", err.Error())
			}
			nvp.getExecutor().clearStatedb()
			nvp.getExecutor().logger.Error("assert failed! exit hyperchain.")
			syscall.Exit(0)
		}
	}
	nvp.getExecutor().accpet(block.Number, block, result)
	if nvp.isInSync() {
		nvp.getCtx().setDown(db_utils.GetHeightOfChain(nvp.getExecutor().namespace))
	}
	nvp.getExecutor().logger.Notice("Block number", block.Number)
	nvp.getExecutor().logger.Notice("Block hash", hex.EncodeToString(block.BlockHash))
	return nil
}

func (nvp *NVPImpl) resendBackend() {
	ticker := time.NewTicker(nvp.getExecutor().GetSyncResendInterval())
	nvp.getExecutor().logger.Debugf("sync request resend interval: ", nvp.getExecutor().GetSyncResendInterval().String())
	up := nvp.getCtx().getUpper()
	down := nvp.getCtx().getDown()
	for {
		select {
		case <-nvp.getCtx().getResendExit():
			nvp.getExecutor().logger.Notice("resend mechanism in sync finish!")
			return
		case <-ticker.C:
			// resend
			curUp := nvp.getCtx().getUpper()
			curDown := nvp.getCtx().getDown()
			if curUp == up && curDown == down {
				nvp.getExecutor().logger.Noticef("resend sync request of NVP. want [%d] - [%d]", down, up)
				nvp.sendSyncRequest(up, down)
			} else {
				up = curUp
				down = curDown
			}
		}
	}
}

func (nvp *NVPImpl) getCtx() NVPContext {
	return nvp.ctx
}

func (nvp *NVPImpl) getExecutor() *Executor {
	return nvp.executor
}

// calUpper - calculate max demand number in a sync request
// if a node required to sync too much blocks one time, the huge chain sync request will be split to several small one.
// a sync chain required block number can not more than `sync batch size` in config file.
func (nvp *NVPImpl) calUpper() uint64 {
	total := nvp.getCtx().getMax() - nvp.getCtx().getDown()
	if total < nvp.getExecutor().GetSyncMaxBatchSize() {
		nvp.getCtx().setUpper(nvp.getCtx().getMax())
	} else {
		nvp.ctx.setUpper(nvp.getCtx().getDown() + nvp.getExecutor().GetSyncMaxBatchSize())
	}
	nvp.getExecutor().logger.Debugf("update max demand number in batch to %d of NVP", nvp.getCtx().getUpper())
	return nvp.getCtx().getUpper()
}

// sendSyncRequest - send synchronization request to other vp nodes.
func (nvp *NVPImpl) sendSyncRequest(upstream, downstream uint64) {
	nvp.getCtx().setUpper(upstream)
	nvp.getCtx().setDown(downstream)
	if err := nvp.getExecutor().informP2P(NOTIFY_NVP_SYNC, upstream, downstream); err != nil {
		nvp.getExecutor().logger.Errorf("[Namespace = %s] send sync req failed. %s", nvp.getExecutor().namespace, err.Error())
		return
	}
}

func (nvp *NVPImpl) decUpper(block *types.Block) {
	if nvp.getCtx().getUpper() == block.Number {
		nvp.getCtx().setUpper(block.Number - 1)
	}
	for blk, err := db_utils.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getCtx().getUpper()); err == nil; {
		if nvp.getCtx().getUpper() <= nvp.getCtx().getDown() {
			break
		}
		nvp.getExecutor().logger.Debugf("db already has block number #%v. block hash %v.", blk.Number, common.Bytes2Hex(blk.BlockHash))
		nvp.getCtx().setUpper(nvp.getCtx().getUpper() - 1)
		blk, err = db_utils.GetBlockByNumber(nvp.getExecutor().namespace, nvp.getCtx().getUpper())
	}
}

func (nvp *NVPImpl) isInSync() bool {
	return nvp.getCtx().getIsInSync()
}
