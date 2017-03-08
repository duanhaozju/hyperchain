package executor

import (
	"hyperchain/event"
	"hyperchain/common"
	"sync/atomic"
)

type ExecutorCache struct {
	validationEventC        chan event.ExeTxsEvent                // validation event buffer
	commitEventC            chan event.CommitOrRollbackBlockEvent // commit event buffer
	validationResultCache   *common.Cache                         // cache for validation result
	pendingValidationEventQ *common.Cache                         // cache for storing validation event
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
	executor.cache.pendingValidationEventQ.Purge()
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

// addValidationResult - save a validation result to cache.
func (executor *Executor) addValidationResult(hash string, res ValidationResultRecord) {
	executor.cache.validationResultCache.Add(hash, res)
}

// fetchValidationResult - fetch a validation result via hash.
func (executor *Executor) fetchValidationResult(hash string) (ValidationResultRecord, bool) {
	v, existed := executor.cache.validationResultCache.Get(hash)
	if existed == false {
		return ValidationResultRecord{}, false
	}
	return v.(ValidationResultRecord), true
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

func (executor *Executor) processValidationDone() {
	atomic.AddInt32(&executor.status.validateQueueLen, -1)
}

// addCommitEvent - push a commit event to channel buffer.
func (executor *Executor) addCommitEvent(ev event.CommitOrRollbackBlockEvent) {
	executor.cache.commitEventC <- ev
	atomic.AddInt32(&executor.status.commitQueueLen, 1)
	log.Noticef("[Namespace = %s] receive a commit event #%d", executor.namespace, ev.SeqNo)
}

// fetchCommitEvent - got a commit event from channel buffer.
func (executor *Executor) fetchCommitEvent() event.CommitOrRollbackBlockEvent {
	ev := <- executor.cache.commitEventC
	log.Noticef("[Namespace = %s] fetch a commit event #%d", executor.namespace, ev.SeqNo)
	return ev
}

// processCommitDone - commit process finish callback.
func (executor *Executor) processCommitDone() {
	atomic.AddInt32(&executor.status.commitQueueLen, -1)
}


