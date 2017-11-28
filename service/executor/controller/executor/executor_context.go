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
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyperchain/hyperchain/common"
)

// Context a collection of all executor status.
type Context struct {
	skipValidation     int32 // validation type, execute normally or just skip.
	validateInProgress int32 // validation operation flag, validating or idle
	commitInProgress   int32 // commit operation flag, committing or idle
	validateQueueLen   int32 // validation buffer size
	commitQueueLen     int32 // commit buffer size

	demandNumber uint64 // demand number for commit
	demandSeqNo  uint64 // demand seqNo for validation

	demandOpLogIndex uint64 //demand index for OpLog

	exit chan struct{} // executor exit flag

	validationSuspend  chan bool // validation suspend notifier
	commitSuspend      chan bool // commit suspend notifier
	syncReplicaSuspend chan bool // replica sync suspend notifier

	stateUpdated chan struct{}
	closeW       sync.WaitGroup
	//syncCtx      *chainSyncContext // synchronization context
}

// newExecutorContext restores histrical status from db.
func newExecutorContext() *Context {

	context := &Context{
		exit:               make(chan struct{}),
		validationSuspend:  make(chan bool),
		commitSuspend:      make(chan bool),
		syncReplicaSuspend: make(chan bool),
		stateUpdated:       make(chan struct{}),
	}

	return context
}

// initDemand inits the demand of number and seqNo.
func (c *Context) initDemand(num uint64) {
	c.demandNumber = num
	c.demandSeqNo = num
}

func (c *Context) getDemandOpLogIndex() uint64 {
	return atomic.LoadUint64(&c.demandOpLogIndex)
}

func (c *Context) setDemandOpLogIndex(i uint64) {
	atomic.StoreUint64(&c.demandOpLogIndex, i)
}

// incDemand increases the demand by plussing one.
func (c *Context) incDemand(typ int) {
	if typ == DemandSeqNo {
		c.demandSeqNo += 1
	} else {
		c.demandNumber += 1
	}
}

// setDemand sets the demand with given num.
func (c *Context) setDemand(typ int, num uint64) {
	if typ == DemandSeqNo {
		c.demandSeqNo = num
	} else {
		c.demandNumber = num
	}
}

// getDemand gets the demand seqNo.
func (c *Context) getDemand(typ int) uint64 {
	if typ == DemandSeqNo {
		return c.demandSeqNo
	} else {
		return c.demandNumber
	}
}

// isDemand returns true if given seqNo is the demand one.
func (c *Context) isDemand(typ int, num uint64) bool {
	if typ == DemandSeqNo {
		return c.demandSeqNo == num
	} else {
		return c.demandNumber == num
	}
}

// turnOffValidationSwitch turns on validation switch, executor will process received event.
func (executor *Executor) turnOnValidationSwitch() {
	executor.logger.Debugf("turn on validation switch", executor.namespace)
	atomic.StoreInt32(&executor.context.skipValidation, VALIDATION_NORMAL)
}

// turnOffValidationSwitch turns off validation switch, executor will drop all received event when the switch turn off.
func (executor *Executor) turnOffValidationSwitch() {
	executor.logger.Debugf("turn off validation switch", executor.namespace)
	atomic.StoreInt32(&executor.context.skipValidation, VALIDATION_IGNORE)
}

// isReadyToValidation checks whether executor is ready to process validation event.
func (executor *Executor) isReadyToValidation() bool {
	if atomic.LoadInt32(&executor.context.skipValidation) == VALIDATION_NORMAL {
		return true
	}
	return false
}

// markValidationBusy marks executor is in validation.
func (executor *Executor) markValidationBusy() {
	executor.logger.Debugf("mark validation busy", executor.namespace)
	atomic.StoreInt32(&executor.context.validateInProgress, BUSY)
}

// markValidationBusy marks executor is idle.
func (executor *Executor) markValidationIdle() {
	executor.logger.Debugf("mark validation idle", executor.namespace)
	atomic.StoreInt32(&executor.context.validateInProgress, IDLE)
}

// markCommitBusy marks executor is in commit.
func (e *Executor) markCommitBusy() {
	e.logger.Debugf("mark commit busy", e.namespace)
	atomic.StoreInt32(&e.context.commitInProgress, BUSY)
}

// markCommitIdle marks executor is idle.
func (executor *Executor) markCommitIdle() {
	executor.logger.Debugf("mark commit idle")
	atomic.StoreInt32(&executor.context.commitInProgress, IDLE)
}

// waitUtilValidationIdle suspends thread util all validations event has been process done.
func (executor *Executor) waitUtilValidationIdle() {
	executor.logger.Debugf("wait validation idle", executor.namespace)
	defer executor.logger.Debugf("validation idle", executor.namespace)
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

// wailUtilCommitIdle suspends thread util all commit events has been process done.
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

// waitUtilRollbackAvailable waits validation processor and commit processor become idle.
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

// rollbackDone rollback callback function to notify rollback finish.
func (executor *Executor) rollbackDone() {
	executor.logger.Debugf("[Namespace = %s] roll back done", executor.namespace)
	executor.turnOnValidationSwitch()
}

/*
	CHAIN SYNCHRONIZATION
*/

// stateTransition resets statedb's seqNo and root to target.
func (executor *Executor) stateTransition(id uint64, root common.Hash) {
	executor.statedb.ResetToTarget(id, root)
}

// waitUtilSyncAvailable waits validation processor and commit processor become idle.
func (executor *Executor) waitUtilSyncAvailable() {
	executor.logger.Debugf("[Namespace = %s] wait util sync available", executor.namespace)
	defer executor.logger.Debugf("[Namespace = %s] sync available", executor.namespace)
	executor.turnOffValidationSwitch()
	executor.waitUtilValidationIdle()
	executor.wailUtilCommitIdle()

	// Clear all cached stuff
	executor.statedb.Purge()
}

// syncDone is the callback function to notify sync finish.
func (executor *Executor) syncDone() {
	executor.logger.Debugf("[Namespace = %s] sync done", executor.namespace)
	executor.turnOnValidationSwitch()
	executor.cache.syncCache.Purge()
}

// clearSyncFlag clears all sync flag fields.
func (executor *Executor) clearSyncFlag() {
	executor.context.stateUpdated <- struct{}{}
	executor.context.closeW.Wait()
	//executor.context.syncCtx = nil
}

// getSuspend the corresponding notifier with given identifier.
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

// setSuspend sets the corresponding suspend into true with given identifier.
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

// unsetSuspend sets the corresponding suspend into false with given identifier.
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
