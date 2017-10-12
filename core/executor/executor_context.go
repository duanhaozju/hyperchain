package executor

import (
	"hyperchain/common"
	edb "hyperchain/core/ledger/db_utils"
	"sync/atomic"
	"time"
)

type ExecutorContext struct {
	validateBehaveFlag int32 // validation type, normal or just ignore.
	validateInProgress int32 // validation operation flag, validating or idle
	commitInProgress   int32 // commit operation flag, committing or idle
	validateQueueLen   int32 // validation buffer size
	commitQueueLen     int32 // commit buffer size

	demandNumber        uint64       // current demand number for commit
	demandSeqNo         uint64       // current demand seqNo for validation
	lastValidationState atomic.Value // latest state root hash
	validationExit      chan bool    // validation exit notifier
	commitExit          chan bool    // commit exit notifier
	replicaSyncExit     chan bool    // replica sync exit notifier

	validationSuspend  chan bool // validation suspend notifier
	commitSuspend      chan bool // commit suspend notifier
	syncReplicaSuspend chan bool // replica sync suspend notifier

	syncFlag SyncFlag          // store temp variables during chain sync
	syncCtx  *ChainSyncContext // synchronization context
}

type SyncFlag struct {
	SyncDemandBlockNum  uint64    // the block num in sync process
	SyncDemandBlockHash []byte    // the block hash in sync process
	SyncTarget          uint64    // the target block number of this synchronization
	SyncPeers           []uint64  // peers to fetch sync blocks
	LocalId             uint64    // local node id
	TempDownstream      uint64    // sync request low height, from calculation
	LatestUpstream      uint64    // latest sync request high height
	LatestDownstream    uint64    // latest sync request low height. always equal to `TempDownstream`
	InExecution         uint32    // flag mark in execution
	ResendMode          uint32    // sync request resend context
	ResendExit          chan bool // resend backend process notifier
	qosStat             *QosStat  // peer selector before send sync request, adhere `BEST PEER` algorithm
}

func initializeExecutorContext(executor *Executor) error {
	executor.context.validationExit = make(chan bool)
	executor.context.commitExit = make(chan bool)
	executor.context.replicaSyncExit = make(chan bool)

	executor.context.validationSuspend = make(chan bool)
	executor.context.commitSuspend = make(chan bool)
	executor.context.syncReplicaSuspend = make(chan bool)

	currentChain := edb.GetChainCopy(executor.namespace)
	executor.initDemand(currentChain.Height + 1)
	blk, err := edb.GetBlockByNumber(executor.namespace, currentChain.Height)
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] get block #%d failed.", executor.namespace, currentChain.Height)
		executor.context.lastValidationState.Store(common.Hash{})
		return err
	} else {
		executor.context.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
		executor.logger.Noticef("[Namespace = %s] initialize executor status success. demand block number %d, demand seqNo %d, latest state hash %s",
			executor.namespace, executor.context.demandNumber, executor.context.demandSeqNo, common.Bytes2Hex(blk.MerkleRoot))
		return nil
	}
}

func (executor *Executor) initDemand(num uint64) {
	executor.context.demandNumber = num
	executor.context.demandSeqNo = num
}

func (executor *Executor) stateTransition(id uint64, root common.Hash) {
	executor.statedb.ResetToTarget(id, root)
}

// Demand block number
func (executor *Executor) incDemandNumber() {
	executor.context.demandNumber += 1
	executor.logger.Debugf("[Namespace = %s] increase demand number to %d", executor.namespace, executor.context.demandNumber)
}

func (executor *Executor) setDemandNumber(num uint64) {
	executor.context.demandNumber = num
	executor.logger.Debugf("[Namespace = %s] set demand number to %d", executor.namespace, executor.context.demandNumber)
}

func (executor *Executor) getDemandNumber() uint64 {
	return executor.context.demandNumber
}

func (executor *Executor) isDemandNumber(num uint64) bool {
	return executor.context.demandNumber == num
}

// Demand seqNo
func (executor *Executor) incDemandSeqNo() {
	executor.context.demandSeqNo += 1
	executor.logger.Debugf("[Namespace = %s] increase demand seqNo to %d", executor.namespace, executor.context.demandSeqNo)
}

func (executor *Executor) setDemandSeqNo(num uint64) {
	executor.context.demandSeqNo = num
	executor.logger.Debugf("[Namespace = %s] set demand seqNo to %d", executor.namespace, executor.context.demandSeqNo)
}

func (executor *Executor) getDemandSeqNo() uint64 {
	return executor.context.demandSeqNo
}

func (executor *Executor) isDemandSeqNo(num uint64) bool {
	return executor.context.demandSeqNo == num
}

// recordStateHash - set latest statedb's hash.
func (executor *Executor) recordStateHash(hash common.Hash) {
	executor.context.lastValidationState.Store(hash)
}

// retrieveStateHash - get latest statedb's hash.
func (executor *Executor) retrieveStateHash() common.Hash {
	v := executor.context.lastValidationState.Load()
	return v.(common.Hash)
}

// turnOffValidationSwitch - turn on validation switch, executor will process received event.
func (executor *Executor) turnOnValidationSwitch() {
	executor.logger.Debugf("[Namespace = %s] turn on validation switch", executor.namespace)
	atomic.StoreInt32(&executor.context.validateBehaveFlag, VALIDATION_NORMAL)
}

// turnOffValidationSwitch - turn off validation switch, executor will drop all received event when the switch turn off.
func (executor *Executor) turnOffValidationSwitch() {
	executor.logger.Debugf("[Namespace = %s] turn off validation switch", executor.namespace)
	atomic.StoreInt32(&executor.context.validateBehaveFlag, VALIDATION_IGNORE)
}

// isReadyToValidation - check whether executor is ready to process validation event.
func (executor *Executor) isReadyToValidation() bool {
	if atomic.LoadInt32(&executor.context.validateBehaveFlag) == VALIDATION_NORMAL {
		return true
	}
	return false
}

// markValidationBusy - mark executor is in validation.
func (executor *Executor) markValidationBusy() {
	executor.logger.Debugf("[Namespace = %s] mark validation busy", executor.namespace)
	atomic.StoreInt32(&executor.context.validateInProgress, BUSY)
}

// markValidationBusy - mark executor is idle.
func (executor *Executor) markValidationIdle() {
	executor.logger.Debugf("[Namespace = %s] mark validation idle", executor.namespace)
	atomic.StoreInt32(&executor.context.validateInProgress, IDLE)
}

// markCommitBusy - mark executor is in commit.
func (executor *Executor) markCommitBusy() {
	executor.logger.Debugf("[Namespace = %s] mark commit busy", executor.namespace)
	atomic.StoreInt32(&executor.context.commitInProgress, BUSY)
}

// markCommitIdle - mark executor is idle.
func (executor *Executor) markCommitIdle() {
	executor.logger.Debugf("[Namespace = %s] mark commit idle", executor.namespace)
	atomic.StoreInt32(&executor.context.commitInProgress, IDLE)
}

// waitUtilValidationIdle - suspend thread util all validations event has been process done.
func (executor *Executor) waitUtilValidationIdle() {
	executor.logger.Debugf("[Namespace = %s] wait validation idle", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] validation idle", executor.namespace)
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&executor.context.validateQueueLen) == 0 && atomic.LoadInt32(&executor.context.validateInProgress) == IDLE {
				return
			} else {
				continue
			}
		}
	}
}

// wailUtilCommitIdle - suspend thread util all commit events has been process done.
func (executor *Executor) wailUtilCommitIdle() {
	executor.logger.Debugf("[Namespace = %s] wait commit idle", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] commit idle", executor.namespace)
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if atomic.LoadInt32(&executor.context.commitQueueLen) == 0 && atomic.LoadInt32(&executor.context.commitInProgress) == IDLE {
				return
			} else {
				continue
			}
		}
	}
}

// waitUtilRollbackAvailable - wait validation processor and commit processor become idle.
func (executor *Executor) waitUtilRollbackAvailable() {
	executor.logger.Debugf("[Namespace = %s] wait util rollback available", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] rollback available", executor.namespace)
	executor.clearPendingValidationEventQ()
	executor.turnOffValidationSwitch()
	executor.waitUtilValidationIdle()
	executor.wailUtilCommitIdle()

	// clear all cached stuff
	executor.statedb.Purge()
}

// rollbackDone - rollback callback function to notify rollback finish.
func (executor *Executor) rollbackDone() {
	executor.logger.Debugf("[Namespace = %s] roll back done", executor.namespace)
	executor.turnOnValidationSwitch()
}

/*
	CHAIN SYNCHRONIZATION
*/
// waitUtilSyncAvailable - wait validation processor and commit processor become idle.
func (executor *Executor) waitUtilSyncAvailable() {
	executor.logger.Debugf("[Namespace = %s] wait util sync available", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] sync available", executor.namespace)
	executor.turnOffValidationSwitch()
	executor.waitUtilValidationIdle()
	executor.wailUtilCommitIdle()

	// clear all cached stuff
	executor.statedb.Purge()
}

// syncDone - sync callback function to notify sync finish.
func (executor *Executor) syncDone() {
	executor.logger.Debugf("[Namespace = %s] sync done", executor.namespace)
	executor.turnOnValidationSwitch()
	executor.cache.syncCache.Purge()
}

// updateSyncFlag - update demand block number, related hash and target during the sync.
func (executor *Executor) updateSyncFlag(num uint64, hash []byte, target uint64) error {
	executor.context.syncFlag.SyncDemandBlockNum = num
	executor.context.syncFlag.SyncDemandBlockHash = hash
	executor.context.syncFlag.SyncTarget = target
	return nil
}

// clearSyncFlag - clear all sync flag fields.
func (executor *Executor) clearSyncFlag() {
	executor.setSyncChainExit()
	executor.context.syncFlag.SyncDemandBlockNum = 0
	executor.context.syncFlag.SyncDemandBlockHash = nil
	executor.context.syncFlag.SyncTarget = 0
	executor.context.syncFlag.LocalId = 0
	executor.context.syncFlag.SyncPeers = nil
	executor.context.syncFlag.TempDownstream = 0
	executor.context.syncFlag.LatestUpstream = 0
	executor.context.syncFlag.LatestDownstream = 0
	executor.context.syncCtx = nil
	executor.context.syncFlag.qosStat = nil
}

// recordSyncPeers - record peers id and self during the sync.
func (executor *Executor) recordSyncPeers(peers []uint64, localId uint64) {
	executor.context.syncFlag.SyncPeers = peers
	executor.context.syncFlag.LocalId = localId
}

// recordSyncReqArgs - record current sync request's high height and low height.
func (executor *Executor) recordSyncReqArgs(upstream, downstream uint64) {
	atomic.StoreUint64(&executor.context.syncFlag.LatestUpstream, upstream)
	atomic.StoreUint64(&executor.context.syncFlag.LatestDownstream, downstream)
}

// getSyncReqArgs - get latest sync request's high height and low height.
func (executor *Executor) getSyncReqArgs() (uint64, uint64) {
	return atomic.LoadUint64(&executor.context.syncFlag.LatestUpstream), atomic.LoadUint64(&executor.context.syncFlag.LatestDownstream)
}

func (executor *Executor) setSyncChainExit() {
	executor.context.syncFlag.ResendExit <- true
}

// setLatestSyncDownstream - save latest sync request down stream.
// return 0 if hasn't been set.
func (executor *Executor) setLatestSyncDownstream(num uint64) {
	executor.context.syncFlag.TempDownstream = num
}

// getLatestSyncDownstream - get latest sync request down stream
func (executor *Executor) getLatestSyncDownstream() uint64 {
	return executor.context.syncFlag.TempDownstream
}

// markSyncExecBegin - set execution(sync) flag true
func (executor *Executor) markSyncExecBegin() {
	atomic.StoreUint32(&executor.context.syncFlag.InExecution, 1)
}

// markSyncExecFinish - set execution(sync) flag false
func (executor *Executor) markSyncExecFinish() {
	atomic.StoreUint32(&executor.context.syncFlag.InExecution, 0)
}

// isSyncInExecution - query execution(sync) flag
func (executor *Executor) isSyncInExecution() bool {
	return atomic.LoadUint32(&executor.context.syncFlag.InExecution) == 1
}

// setValidationExit - notify validation backend process to exit.
func (executor *Executor) setValidationExit() {
	executor.context.validationExit <- true
}

// getValidationExit - get exit flag.
func (executor *Executor) getValidationExit() chan bool {
	return executor.context.validationExit
}

// setCommitExit - notify commit backend process to exit.
func (executor *Executor) setCommitExit() {
	executor.context.commitExit <- true
}

// getCommitExit - get exit flag.
func (executor *Executor) getCommitExit() chan bool {
	return executor.context.commitExit
}

// setReplicaSyncExit - notify replica sync backend process to exit.
func (executor *Executor) setReplicaSyncExit() {
	executor.context.replicaSyncExit <- true
}

// getReplicaSyncExit - get exit flag.
func (executor *Executor) getReplicaSyncExit() chan bool {
	return executor.context.replicaSyncExit
}

// getExit - get exit status.
func (executor *Executor) getExit(identifier int) chan bool {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		return executor.getValidationExit()
	case IDENTIFIER_COMMIT:
		return executor.getCommitExit()
	case IDENTIFIER_REPLICA_SYNC:
		return executor.getReplicaSyncExit()
	}
	return nil
}

func (executor *Executor) getSuspend(identifier int) chan bool {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		return executor.context.validationSuspend
	case IDENTIFIER_COMMIT:
		return executor.context.commitSuspend
	case IDENTIFIER_REPLICA_SYNC:
		return executor.context.syncReplicaSuspend
	}
	return nil
}

func (executor *Executor) setSuspend(identifier int) {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		executor.context.validationSuspend <- true
	case IDENTIFIER_COMMIT:
		executor.context.commitSuspend <- true
	case IDENTIFIER_REPLICA_SYNC:
		executor.context.syncReplicaSuspend <- true
	}
}

func (executor *Executor) unsetSuspend(identifier int) {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		executor.context.validationSuspend <- false
	case IDENTIFIER_COMMIT:
		executor.context.commitSuspend <- false
	case IDENTIFIER_REPLICA_SYNC:
		executor.context.syncReplicaSuspend <- false
	}
}

// getSyncTarget - get SyncTarget.
func (executor *Executor) getSyncTarget() uint64 {
	return executor.context.syncFlag.SyncTarget
}
