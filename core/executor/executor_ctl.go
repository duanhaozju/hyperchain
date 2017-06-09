package executor

import (
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/tree/bucket"
	"sync/atomic"
	"time"
)

type ExecutorStatus struct {
	validateBehaveFlag int32 // validation type, normal or just ignore.
	validateInProgress int32 // validation operation flag, validating or idle
	commitInProgress   int32 // commit operation flag, committing or idle
	validateQueueLen   int32 // validation buffer size
	commitQueueLen     int32 // commit buffer size

	demandNumber        uint64       // current demand number for commit
	demandSeqNo         uint64       // current demand seqNo for validation
	tempBlockNumber     uint64       // temporarily block number
	lastValidationState atomic.Value // latest state root hash
	validationExit      chan bool    // validation exit notifier
	commitExit          chan bool    // commit exit notifier
	replicaSyncExit     chan bool    // replica sync exit notifier

	validationSuspend   chan bool     // validation suspend notifier
	commitSuspend       chan bool     // commit suspend notifier
	syncReplicaSuspend  chan bool     // replica sync suspend notifier

	syncFlag            SyncFlag     // store temp variables during chain sync
	syncCtx             *ChainSyncContext // synchronization context
}

type SyncFlag struct {
	SyncDemandBlockNum  uint64     // the block num in sync process
	SyncDemandBlockHash []byte     // the block hash in sync process
	SyncTarget          uint64     // the target block number of this synchronization
	SyncPeers           []uint64   // peers to fetch sync blocks
	LocalId             uint64     // local node id
	TempDownstream      uint64     // sync request low height, from calculation
	LatestUpstream      uint64     // latest sync request high height
	LatestDownstream    uint64     // latest sync request low height. always equal to `TempDownstream`
	InExecution         uint32     // flag mark in execution
	ResendMode          uint32     // sync request resend context
	ResendExit          chan bool  // resend backend process notifier
	Oracle              *Oracle    // peer selector before send sync request, adhere `BEST PEER` algorithm
}

func initializeExecutorStatus(executor *Executor) error {
	executor.status.validationExit = make(chan bool)
	executor.status.commitExit = make(chan bool)
	executor.status.replicaSyncExit = make(chan bool)

	executor.status.validationSuspend = make(chan bool)
	executor.status.commitSuspend = make(chan bool)
	executor.status.syncReplicaSuspend = make(chan bool)

	currentChain := edb.GetChainCopy(executor.namespace)
	executor.initDemand(currentChain.Height + 1)
	blk, err := edb.GetBlockByNumber(executor.namespace, currentChain.Height)
	if err != nil {
		executor.logger.Errorf("[Namespace = %s] get block #%d failed.", executor.namespace, currentChain.Height)
		executor.status.lastValidationState.Store(common.Hash{})
		return err
	} else {
		executor.status.lastValidationState.Store(common.BytesToHash(blk.MerkleRoot))
		executor.logger.Noticef("[Namespace = %s] initialize executor status success. demand block number %d, demand seqNo %d, latest state hash %s",
			executor.namespace, executor.status.demandNumber, executor.status.demandSeqNo, common.Bytes2Hex(blk.MerkleRoot))
		return nil
	}
}

func (executor *Executor) initDemand(num uint64) {
	executor.status.demandNumber = num
	executor.status.demandSeqNo = num
	executor.status.tempBlockNumber = num
}

func (executor *Executor) stateTransition(id uint64, root common.Hash) {
	executor.statedb.ResetToTarget(id, root)
}

// Demand block number
func (executor *Executor) incDemandNumber() {
	executor.status.demandNumber += 1
	executor.logger.Debugf("[Namespace = %s] increase demand number to %d", executor.namespace, executor.status.demandNumber)
}

func (executor *Executor) setDemandNumber(num uint64) {
	executor.status.demandNumber = num
	executor.logger.Debugf("[Namespace = %s] set demand number to %d", executor.namespace, executor.status.demandNumber)
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
	executor.logger.Debugf("[Namespace = %s] increase demand seqNo to %d", executor.namespace, executor.status.demandSeqNo)
}

func (executor *Executor) setDemandSeqNo(num uint64) {
	executor.status.demandSeqNo = num
	executor.logger.Debugf("[Namespace = %s] set demand seqNo to %d", executor.namespace, executor.status.demandSeqNo)
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
	executor.logger.Debugf("[Namespace = %s] increase temp block number to %d", executor.namespace, executor.status.tempBlockNumber)
}

func (executor *Executor) setTempBlockNumber(num uint64) {
	executor.status.tempBlockNumber = num
	executor.logger.Debugf("[Namespace = %s] set temp block number to %d", executor.namespace, executor.status.tempBlockNumber)
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
	executor.logger.Debugf("[Namespace = %s] turn on validation switch", executor.namespace)
	atomic.StoreInt32(&executor.status.validateBehaveFlag, VALIDATION_NORMAL)
}

// turnOffValidationSwitch - turn off validation switch, executor will drop all received event when the switch turn off.
func (executor *Executor) turnOffValidationSwitch() {
	executor.logger.Debugf("[Namespace = %s] turn off validation switch", executor.namespace)
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
	executor.logger.Debugf("[Namespace = %s] mark validation busy", executor.namespace)
	atomic.StoreInt32(&executor.status.validateInProgress, BUSY)
}

// markValidationBusy - mark executor is idle.
func (executor *Executor) markValidationIdle() {
	executor.logger.Debugf("[Namespace = %s] mark validation idle", executor.namespace)
	atomic.StoreInt32(&executor.status.validateInProgress, IDLE)
}

// markCommitBusy - mark executor is in commit.
func (executor *Executor) markCommitBusy() {
	executor.logger.Debugf("[Namespace = %s] mark commit busy", executor.namespace)
	atomic.StoreInt32(&executor.status.commitInProgress, BUSY)
}

// markCommitIdle - mark executor is idle.
func (executor *Executor) markCommitIdle() {
	executor.logger.Debugf("[Namespace = %s] mark commit idle", executor.namespace)
	atomic.StoreInt32(&executor.status.commitInProgress, IDLE)
}

// waitUtilValidationIdle - suspend thread util all validations event has been process done.
func (executor *Executor) waitUtilValidationIdle() {
	executor.logger.Debugf("[Namespace = %s] wait validation idle", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] validation idle", executor.namespace)
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
	executor.logger.Debugf("[Namespace = %s] wait commit idle", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] commit idle", executor.namespace)
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
	executor.logger.Debugf("[Namespace = %s] wait util rollback available", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] rollback available", executor.namespace)
	executor.clearPendingValidationEventQ()
	executor.turnOffValidationSwitch()
	executor.waitUtilValidationIdle()
	executor.wailUtilCommitIdle()

	// clear all cached stuff
	executor.statedb.Purge()
	tree := executor.statedb.GetTree()
	bucketTree := tree.(*bucket.BucketTree)
	bucketTree.ClearAllCache()
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
	tree := executor.statedb.GetTree()
	bucketTree := tree.(*bucket.BucketTree)
	bucketTree.ClearAllCache()
}

// syncDone - sync callback function to notify sync finish.
func (executor *Executor) syncDone() {
	executor.logger.Debugf("[Namespace = %s] sync done", executor.namespace)
	executor.turnOnValidationSwitch()
	executor.cache.syncCache.Purge()
}

// updateSyncFlag - update demand block number, related hash and target during the sync.
func (executor *Executor) updateSyncFlag(num uint64, hash []byte, target uint64) error {
	executor.status.syncFlag.SyncDemandBlockNum = num
	executor.status.syncFlag.SyncDemandBlockHash = hash
	executor.status.syncFlag.SyncTarget = target
	executor.status.syncFlag.ResendExit = make(chan bool)
	return nil
}

// clearSyncFlag - clear all sync flag fields.
func (executor *Executor) clearSyncFlag() {
	executor.setSyncChainExit()
	executor.status.syncFlag.SyncDemandBlockNum = 0
	executor.status.syncFlag.SyncDemandBlockHash = nil
	executor.status.syncFlag.SyncTarget = 0
	executor.status.syncFlag.LocalId = 0
	executor.status.syncFlag.SyncPeers = nil
	executor.status.syncFlag.TempDownstream = 0
	executor.status.syncFlag.LatestUpstream = 0
	executor.status.syncFlag.LatestDownstream = 0
}

// recordSyncPeers - record peers id and self during the sync.
func (executor *Executor) recordSyncPeers(peers []uint64, localId uint64) {
	executor.status.syncFlag.SyncPeers = peers
	executor.status.syncFlag.LocalId = localId
}

// recordSyncReqArgs - record current sync request's high height and low height.
func (executor *Executor) recordSyncReqArgs(upstream, downstream uint64) {
	atomic.StoreUint64(&executor.status.syncFlag.LatestUpstream, upstream)
	atomic.StoreUint64(&executor.status.syncFlag.LatestDownstream, downstream)
}

// getSyncReqArgs - get latest sync request's high height and low height.
func (executor *Executor) getSyncReqArgs() (uint64, uint64) {
	return atomic.LoadUint64(&executor.status.syncFlag.LatestUpstream), atomic.LoadUint64(&executor.status.syncFlag.LatestDownstream)
}

func (executor *Executor) setSyncChainExit() {
	executor.status.syncFlag.ResendExit <- true
}

// setLatestSyncDownstream - save latest sync request down stream.
// return 0 if hasn't been set.
func (executor *Executor) setLatestSyncDownstream(num uint64) {
	executor.status.syncFlag.TempDownstream = num
}

// getLatestSyncDownstream - get latest sync request down stream
func (executor *Executor) getLatestSyncDownstream() uint64 {
	return executor.status.syncFlag.TempDownstream
}

// markSyncExecBegin - set execution(sync) flag true
func (executor *Executor) markSyncExecBegin() {
	atomic.StoreUint32(&executor.status.syncFlag.InExecution, 1)
}

// markSyncExecFinish - set execution(sync) flag false
func (executor *Executor) markSyncExecFinish() {
	atomic.StoreUint32(&executor.status.syncFlag.InExecution, 0)
}

// isSyncInExecution - query execution(sync) flag
func (executor *Executor) isSyncInExecution() bool {
	return atomic.LoadUint32(&executor.status.syncFlag.InExecution) == 1
}

// setValidationExit - notify validation backend process to exit.
func (executor *Executor) setValidationExit() {
	executor.status.validationExit <- true
}

// getValidationExit - get exit flag.
func (executor *Executor) getValidationExit() chan bool {
	return executor.status.validationExit
}

// setCommitExit - notify commit backend process to exit.
func (executor *Executor) setCommitExit() {
	executor.status.commitExit <- true
}

// getCommitExit - get exit flag.
func (executor *Executor) getCommitExit() chan bool {
	return executor.status.commitExit
}

// setReplicaSyncExit - notify replica sync backend process to exit.
func (executor *Executor) setReplicaSyncExit() {
	executor.status.replicaSyncExit <- true
}

// getReplicaSyncExit - get exit flag.
func (executor *Executor) getReplicaSyncExit() chan bool {
	return executor.status.replicaSyncExit
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
		return executor.status.validationSuspend
	case IDENTIFIER_COMMIT:
		return executor.status.commitSuspend
	case IDENTIFIER_REPLICA_SYNC:
		return executor.status.syncReplicaSuspend
	}
	return nil
}

func (executor *Executor) setSuspend(identifier int) {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		executor.status.validationSuspend <- true
	case IDENTIFIER_COMMIT:
		executor.status.commitSuspend <- true
	case IDENTIFIER_REPLICA_SYNC:
		executor.status.syncReplicaSuspend <- true
	}
}

func (executor *Executor) unsetSuspend(identifier int) {
	switch identifier {
	case IDENTIFIER_VALIDATION:
		executor.status.validationSuspend <- false
	case IDENTIFIER_COMMIT:
		executor.status.commitSuspend <- false
	case IDENTIFIER_REPLICA_SYNC:
		executor.status.syncReplicaSuspend <- false
	}
}
