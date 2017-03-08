package executor

import (
	"sync/atomic"
	edb "hyperchain/core/db_utils"
	"hyperchain/common"
	"time"
)
type ExecutorStatus struct {
	validateBehaveFlag    int32                  // validation type, normal or just ignore.
	validateInProgress    int32                  // validation operation flag, validating or idle
	commitInProgress      int32                  // commit operation flag, committing or idle
	validateQueueLen      int32                  // validation buffer size
	commitQueueLen        int32                  // commit buffer size

	demandNumber          uint64              // current demand number for commit
	demandSeqNo           uint64              // current demand seqNo for validation
	tempBlockNumber       uint64              // temporarily block number
	lastValidationState   atomic.Value        // latest state root hash
}

func initializeExecutorStatus(executor *Executor) error {
	currentChain := edb.GetChainCopy(executor.namespace)
	executor.status.demandNumber = currentChain.Height + 1
	executor.status.demandSeqNo = currentChain.Height + 1
	executor.status.tempBlockNumber = currentChain.Height + 1

	blk, err := edb.GetBlockByNumber(executor.namespace, currentChain.Height)
	if err != nil {
		log.Errorf("[Namespace = %s] get block #%d failed.", executor.namespace, currentChain.Height)
		executor.status.lastValidationState.Store(common.Hash{})
		return err
	} else {
		executor.status.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
		log.Noticef("[Namespace = %s] initialize executor status success. demand block number %d, demand seqNo %d",
			executor.namespace, executor.status.demandNumber, executor.status.demandSeqNo)
		return nil
	}
}

// Demand block number
func (executor *Executor) incDemandNumber() {
	executor.status.demandNumber += 1
	log.Noticef("[Namespace = %s] increase demand number to %d", executor.namespace, executor.status.demandNumber)
}

func (executor *Executor) setDemandNumber(num uint64) {
	executor.status.demandNumber = num
	log.Noticef("[Namespace = %s] set demand number to %d", executor.namespace, executor.status.demandNumber)
}

func (executor *Executor) getDemandNumber() uint64 {
	return executor.status.demandNumber
}

func (executor *Executor) isDemandNumber(num uint64) bool {
	return executor.status.demandNumber == num
}

// Demand seqNo
func (executor *Executor) incDemandSeqNo() {
	executor.status.demandSeqNo += 1
	log.Noticef("[Namespace = %s] increase demand seqNo to %d", executor.namespace, executor.status.demandSeqNo)
}

func (executor *Executor) setDemandSeqNo(num uint64) {
	executor.status.demandSeqNo = num
	log.Noticef("[Namespace = %s] set demand seqNo to %d", executor.namespace, executor.status.demandSeqNo)
}

func (executor *Executor) getDemandSeqNo() uint64 {
	return executor.status.demandSeqNo
}

func (executor *Executor) isDemandSeqNo(num uint64) bool {
	return executor.status.demandSeqNo == num
}

// Temp block number
func (executor *Executor) incTempBlockNumber() {
	executor.status.tempBlockNumber += 1
	log.Noticef("[Namespace = %s] increase temp block number to %d", executor.namespace, executor.status.tempBlockNumber)
}

func (executor *Executor) setTempBlockNumber(num uint64) {
	executor.status.tempBlockNumber = num
	log.Noticef("[Namespace = %s] set temp block number to %d", executor.namespace, executor.status.tempBlockNumber)
}

func (executor *Executor) getTempBlockNumber() uint64 {
	return executor.status.tempBlockNumber
}

// recordStateHash - set latest statedb's hash.
func (executor *Executor) recordStateHash(hash common.Hash) {
	executor.status.lastValidationState.Store(hash)
}

// retrieveStateHash - get latest statedb's hash.
func (executor *Executor) retrieveStateHash() common.Hash {
	v := executor.status.lastValidationState.Load()
	return v.(common.Hash)
}

// turnOffValidationSwitch - turn on validation switch, executor will process received event.
func (executor *Executor) turnOnValidationSwitch() {
	log.Debugf("[Namespace = %s] turn on validation switch", executor.namespace)
	atomic.StoreInt32(&executor.status.validateBehaveFlag, VALIDATION_NORMAL)
}

// turnOffValidationSwitch - turn off validation switch, executor will drop all received event when the switch turn off.
func (executor *Executor) turnOffValidationSwitch() {
	log.Debugf("[Namespace = %s] turn off validation switch", executor.namespace)
	atomic.StoreInt32(&executor.status.validateBehaveFlag, VALIDATION_IGNORE)
}

// isReadyToValidation - check whether executor is ready to process validation event.
func (executor *Executor) isReadyToValidation() bool {
	if atomic.LoadInt32(&executor.status.validateBehaveFlag) == VALIDATION_NORMAL {
		return true
	}
	return false
}

// markValidationBusy - mark executor is in validation.
func (executor *Executor) markValidationBusy() {
	log.Debugf("[Namespace = %s] mark validation busy", executor.namespace)
	atomic.StoreInt32(&executor.status.validateInProgress, BUSY)
}

// markValidationBusy - mark executor is idle.
func (executor *Executor) markValidationIdle() {
	log.Debugf("[Namespace = %s] mark validation idle", executor.namespace)
	atomic.StoreInt32(&executor.status.validateInProgress, IDLE)
}

// markCommitBusy - mark executor is in commit.
func (executor *Executor) markCommitBusy() {
	log.Debugf("[Namespace = %s] mark commit busy", executor.namespace)
	atomic.StoreInt32(&executor.status.commitInProgress, BUSY)
}

// markCommitIdle - mark executor is idle.
func (executor *Executor) markCommitIdle() {
	log.Debugf("[Namespace = %s] mark commit idle", executor.namespace)
	atomic.StoreInt32(&executor.status.commitInProgress, IDLE)
}

// waitUtilValidationIdle - suspend thread util all validations event has been process done.
func (executor *Executor) waitUtilValidationIdle() {
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&executor.status.validateQueueLen) == 0 && atomic.LoadInt32(&executor.status.validateInProgress) == IDLE {
				return
			} else {
				continue
			}
		}
	}
}

// wailUtilCommitIdle - suspend thread util all commit events has been process done.
func (executor *Executor) wailUtilCommitIdle() {
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&executor.status.commitQueueLen) == 0 && atomic.LoadInt32(&executor.status.commitInProgress) == IDLE {
				return
			} else {
				continue
			}
		}
	}
}

// waitUtilRollbackAvailable - wait validation processor and commit processor become idle.
func (executor *Executor) waitUtilRollbackAvailable() {
	executor.turnOffValidationSwitch()
	executor.waitUtilValidationIdle()
	executor.wailUtilCommitIdle()
}

// rollbackDone - rollback callback function to notify rollback finish.
func (executor *Executor) rollbackDone() {
	executor.turnOnValidationSwitch()
}


