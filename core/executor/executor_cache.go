package executor

import (
	"hyperchain/event"
	"hyperchain/common"
	"sync/atomic"
	"hyperchain/core/types"
)

type ExecutorCache struct {
	validationEventC        chan event.ExeTxsEvent                // validation event buffer
	commitEventC            chan event.CommitOrRollbackBlockEvent // commit event buffer
	validationResultCache   *common.Cache                         // cache for validation result
	pendingValidationEventQ *common.Cache                         // cache for storing validation event
	syncCache               *common.Cache                         // cache for storing stuff in sync
}

func initializeExecutorCache(executor *Executor) error {
	validationResC, err := common.NewCache()
	if err != nil {
		return err
	}
	executor.cache.validationResultCache = validationResC
	validationEventQ, err := common.NewCache()
	if err != nil {
		return err
	}
	executor.cache.pendingValidationEventQ = validationEventQ
	executor.cache.validationEventC = make(chan event.ExeTxsEvent, VALIDATEQUEUESIZE)
	executor.cache.commitEventC = make(chan event.CommitOrRollbackBlockEvent, COMMITQUEUESIZE)
	return nil
}

// PurgeCache - purge executor cache.
func (executor *Executor) PurgeCache() {
	executor.cache.validationResultCache.Purge()
	executor.clearPendingValidationEventQ()
	log.Noticef("[Namespace = %s] purge validation result cache and validation event cache success", executor.namespace)
}

// addPendingValidationEvent - push a validation event to pending queue.
func (executor *Executor) addPendingValidationEvent(validationEvent event.ExeTxsEvent) {
	log.Errorf("[Namespace = %s] receive validation event %d while %d is required, save into cache temporarily.", executor.namespace, validationEvent.SeqNo, executor.getDemandSeqNo())
	executor.cache.pendingValidationEventQ.Add(validationEvent.SeqNo, validationEvent)
}

// fetchPendingValidationEvent - fetch a validation event in pending queue via seqNo, return false if not exist.
func (executor *Executor) fetchPendingValidationEvent(seqNo uint64) (event.ExeTxsEvent, bool) {
	res, existed := executor.cache.pendingValidationEventQ.Get(seqNo)
	if existed == false {
		return event.ExeTxsEvent{}, false
	}
	ev := res.(event.ExeTxsEvent)
	return ev, true
}

// pendingValidationEventQLen - retrieve pending validation event queue length.
func (executor *Executor) pendingValidationEventQLen() int {
	return executor.cache.pendingValidationEventQ.Len()
}

// clearPendingValidationEventQ - purge validation event q.
func (executor *Executor) clearPendingValidationEventQ() {
	length := executor.pendingValidationEventQLen()
	executor.cache.pendingValidationEventQ.Purge()
	atomic.AddInt32(&executor.status.validateQueueLen, -1 * int32(length))
}

// addValidationResult - save a validation result to cache.
func (executor *Executor) addValidationResult(hash string, res *ValidationResultRecord) {
	executor.cache.validationResultCache.Add(hash, res)
}

// fetchValidationResult - fetch a validation result via hash.
func (executor *Executor) fetchValidationResult(hash string) (*ValidationResultRecord, bool) {
	v, existed := executor.cache.validationResultCache.Get(hash)
	if existed == false {
		return nil, false
	}
	return v.(*ValidationResultRecord), true
}

// addValidationEvent - push a validation event to channel buffer.
func (executor *Executor) addValidationEvent(ev event.ExeTxsEvent) {
	executor.cache.validationEventC <- ev
	atomic.AddInt32(&executor.status.validateQueueLen, 1)
	log.Noticef("[Namespace = %s] receive a validation event #%d", executor.namespace, ev.SeqNo)
}

// fetchValidationEvent - got a validation event from channel buffer.
func (executor *Executor) fetchValidationEvent() event.ExeTxsEvent {
	ev := <- executor.cache.validationEventC
	log.Noticef("[Namespace = %s] fetch a validation event #%d", executor.namespace, ev.SeqNo)
	return ev
}

// processValidationDone - validation finish callback.
func (executor *Executor) processValidationDone() {
	atomic.AddInt32(&executor.status.validateQueueLen, -1)
}

// addCommitEvent - push a commit event to channel buffer.
func (executor *Executor) addCommitEvent(ev event.CommitOrRollbackBlockEvent) {
	executor.cache.commitEventC <- ev
	atomic.AddInt32(&executor.status.commitQueueLen, 1)
	log.Debugf("[Namespace = %s] receive a commit event #%d", executor.namespace, ev.SeqNo)
}

// fetchCommitEvent - got a commit event from channel buffer.
func (executor *Executor) fetchCommitEvent() event.CommitOrRollbackBlockEvent {
	ev := <- executor.cache.commitEventC
	log.Debugf("[Namespace = %s] fetch a commit event #%d", executor.namespace, ev.SeqNo)
	return ev
}

// processCommitDone - commit process finish callback.
func (executor *Executor) processCommitDone() {
	atomic.AddInt32(&executor.status.commitQueueLen, -1)
}

// addToSyncCache - add a block to cache which arrives earlier than expect.
func (executor *Executor) addToSyncCache(block *types.Block) {
	blks, existed := executor.fetchFromSyncCache(block.Number)
	if existed {
		if _, ok := blks[common.Bytes2Hex(block.BlockHash)]; ok {
			log.Debugf("[Namespace = %s] receive duplicate block: %d %s", executor.namespace, block.Number, common.Bytes2Hex(block.BlockHash))
		} else {
			log.Debugf("[Namespace = %s] receive  block with different hash: %d %s", executor.namespace, block.Number, common.Bytes2Hex(block.BlockHash))
			blks[common.Bytes2Hex(block.BlockHash)] = *block
			executor.cache.syncCache.Add(block.Number, blks)
		}
	} else {
		blks := make(map[string]types.Block)
		blks[common.Bytes2Hex(block.BlockHash)] = *block
		executor.cache.syncCache.Add(block.Number, blks)
	}
}

// fetchFromSyncCache - fetch blocks from sync cache.
func (executor *Executor) fetchFromSyncCache(number uint64) (map[string]types.Block, bool) {
	ret, existed := executor.cache.syncCache.Get(number)
	if !existed {
		return nil, false
	}
	blks := ret.(map[string]types.Block)
	return blks, true
}


