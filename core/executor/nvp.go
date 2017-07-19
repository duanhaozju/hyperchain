package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"sync"
)

type NVP interface {
	ReceiveBlock(*types.Block)
}

type NVPContext interface {
	GetDemand() uint64
}

type NVPImpl struct {
	lock     sync.Mutex
	ctx      NVPContext
	executor *Executor
}

type NVPContextImpl struct {
	demand     uint64
	blockCache *common.Cache
}

func (nvp *NVPImpl) ReceiveBlock(block *types.Block) {
	nvp.lock.Lock()
	defer nvp.lock.Unlock()
	if err := nvp.PreProcess(block); err != nil {

	}
	if nvp.isDemand(block.Number) {
		if err := nvp.ApplyBlock(block); err != nil {

		}
	} else if nvp.isFuture(block.Number) && nvp.isFallBehind(block.Number) {
		nvp.FetchBlock()
	} else {
		// ignore
	}
}

func (nvp *NVPImpl) PreProcess(block *types.Block) error {
	// 1. check whether received block is needed
	// 2. if is demand(block number larger or equal to demand), check block integrity
	// 3. persist to database
	return nil
}

func (nvp *NVPImpl) ApplyBlock(block *types.Block) error {
	if err := nvp.process(block); err != nil {
		return err
	}
	nvp.updateDemand()
	return nil
}

func (nvp *NVPImpl) ApplyRemainBlock(number uint64) error {
	return nil
}

func (nvp *NVPImpl) process(block *types.Block) error {
	return nil
}

func (nvp *NVPImpl) FetchBlock() {

}

func (nvp *NVPImpl) updateDemand() {

}

func (nvp *NVPImpl) isDemand(id uint64) bool {
	return nvp.ctx.GetDemand() == id
}

func (nvp *NVPImpl) isFuture(id uint64) bool {
	return nvp.ctx.GetDemand() < id
}

func (nvp *NVPImpl) isFallBehind(id uint64) bool {
	return false
}
