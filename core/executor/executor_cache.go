package executor

import (
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"sync/atomic"
)

type Cache struct {
	validationEventC        chan event.ValidationEvent // validation event buffer
	commitEventC            chan event.CommitEvent     // commit event buffer
	validationResultCache   *common.Cache              // cache for validation result
	pendingValidationEventQ *common.Cache              // cache for storing validation event
	syncCache               *common.Cache              // cache for storing stuff in sync
	replicaInfoCache        *common.Cache              // cache for storing replica info
}

type Peer struct {
	Ip   string
	Port int32
}

func initializeExecutorCache(executor *Executor) error {
	validationResC, _ := common.NewCache()
	executor.cache.validationResultCache = validationResC
	validationEventQ, _ := common.NewCache()
	executor.cache.pendingValidationEventQ = validationEventQ
	syncCache, _ := common.NewCache()
	executor.cache.syncCache = syncCache
	replicaCache, _ := common.NewCache()
	executor.cache.replicaInfoCache = replicaCache
	executor.cache.validationEventC = make(chan event.ValidationEvent, VALIDATEQUEUESIZE)
	executor.cache.commitEventC = make(chan event.CommitEvent, COMMITQUEUESIZE)
	return nil
}

// PurgeCache - purge executor cache.
func (executor *Executor) PurgeCache() {
	executor.cache.validationResultCache.Purge()
	executor.clearPendingValidationEventQ()
	executor.logger.Debugf("[Namespace = %s] purge validation result cache and validation event cache success", executor.namespace)
}

// addPendingValidationEvent - push a validation event to pending queue.
func (executor *Executor) addPendingValidationEvent(validationEvent event.ValidationEvent) {
	executor.logger.Warningf("[Namespace = %s] receive validation event %d while %d is required, save into cache temporarily.", executor.namespace, validationEvent.SeqNo, executor.getDemandSeqNo())
	executor.cache.pendingValidationEventQ.Add(validationEvent.SeqNo, validationEvent)
}

// fetchPendingValidationEvent - fetch a validation event in pending queue via seqNo, return false if not exist.
func (executor *Executor) fetchPendingValidationEvent(seqNo uint64) (event.ValidationEvent, bool) {
	res, existed := executor.cache.pendingValidationEventQ.Get(seqNo)
	if existed == false {
		return event.ValidationEvent{}, false
	}
	ev := res.(event.ValidationEvent)
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
	atomic.AddInt32(&executor.context.validateQueueLen, -1*int32(length))
}

// addValidationResult - save a validation result to cache.
func (executor *Executor) addValidationResult(tag ValidationTag, res *ValidationResultRecord) {
	executor.cache.validationResultCache.Add(tag, res)
}

// fetchValidationResult - fetch a validation result via hash.
func (executor *Executor) fetchValidationResult(tag ValidationTag) (*ValidationResultRecord, bool) {
	v, existed := executor.cache.validationResultCache.Get(tag)
	if existed == false {
		return nil, false
	}
	return v.(*ValidationResultRecord), true
}

// addValidationEvent - push a validation event to channel buffer.
func (executor *Executor) addValidationEvent(ev event.ValidationEvent) {
	executor.cache.validationEventC <- ev
	atomic.AddInt32(&executor.context.validateQueueLen, 1)
	executor.logger.Debugf("[Namespace = %s] receive a validation event #%d", executor.namespace, ev.SeqNo)
}

// fetchValidationEvent - got a validation event from channel buffer.
func (executor *Executor) fetchValidationEvent() chan event.ValidationEvent {
	return executor.cache.validationEventC
}

// processValidationDone - validation finish callback.
func (executor *Executor) processValidationDone() {
	atomic.AddInt32(&executor.context.validateQueueLen, -1)
}

// addCommitEvent - push a commit event to channel buffer.
func (executor *Executor) addCommitEvent(ev event.CommitEvent) {
	executor.cache.commitEventC <- ev
	atomic.AddInt32(&executor.context.commitQueueLen, 1)
	executor.logger.Debugf("[Namespace = %s] receive a commit event #%d", executor.namespace, ev.SeqNo)
}

// fetchCommitEvent - got a commit event from channel buffer.
func (executor *Executor) fetchCommitEvent() chan event.CommitEvent {
	return executor.cache.commitEventC
}

// processCommitDone - commit process finish callback.
func (executor *Executor) processCommitDone() {
	atomic.AddInt32(&executor.context.commitQueueLen, -1)
}

// addToSyncCache - add a block to cache which arrives earlier than expect.
func (executor *Executor) addToSyncCache(block *types.Block) {
	blks, existed := executor.fetchFromSyncCache(block.Number)
	if existed {
		if _, ok := blks[common.Bytes2Hex(block.BlockHash)]; ok {
			executor.logger.Debugf("[Namespace = %s] receive duplicate block: %d %s", executor.namespace, block.Number, common.Bytes2Hex(block.BlockHash))
		} else {
			executor.logger.Debugf("[Namespace = %s] receive  block with different hash: %d %s", executor.namespace, block.Number, common.Bytes2Hex(block.BlockHash))
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

// addToReplicaCache - push a replica status to cache with ip and port as flag.
func (executor *Executor) addToReplicaCache(ip string, port int32, chain *types.Chain) {
	executor.cache.replicaInfoCache.Add(Peer{
		Ip:   ip,
		Port: port,
	}, chain)
}

// fetchFromReplicaCache - fetch a replica's status from cache with ip and port as key.
func (executor *Executor) fetchFromReplicaCache(ip string, port int32) (*types.Chain, bool) {
	v, existed := executor.cache.replicaInfoCache.Get(Peer{
		Ip:   ip,
		Port: port,
	})
	if existed == false {
		return nil, existed
	}
	return v.(*types.Chain), true
}
