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
	"sync/atomic"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
)

// Cache represents the caches that used in executor.
type Cache struct {
	validationEventC        chan event.TransactionBlock // validation event buffer
	commitEventC            chan event.CommitEvent      // commit event buffer
	validationResultCache   *common.Cache               // cache for validation result
	pendingValidationEventQ *common.Cache               // cache for storing validation event
	syncCache               *common.Cache               // cache for storing stuff in sync
	replicaInfoCache        *common.Cache               // cache for storing replica info
}

// Peer stores the ip and port of a peer that used in replicaInCache.
type Peer struct {
	Ip   string
	Port int32
}

// newExecutorCache creates the cache for executor.
func newExecutorCache() *Cache {
	cache := &Cache{
		validationEventC: make(chan event.TransactionBlock, VALIDATEQUEUESIZE),
		commitEventC:     make(chan event.CommitEvent, COMMITQUEUESIZE),
	}
	validationResC, _ := common.NewCache()
	cache.validationResultCache = validationResC
	validationEventQ, _ := common.NewCache()
	cache.pendingValidationEventQ = validationEventQ
	syncCache, _ := common.NewCache()
	cache.syncCache = syncCache
	replicaCache, _ := common.NewCache()
	cache.replicaInfoCache = replicaCache

	return cache
}

// purgeCache purges executor cache.
func (executor *Executor) purgeCache() {
	executor.cache.validationResultCache.Purge()
	executor.clearPendingValidationEventQ()
	executor.logger.Debugf("[Namespace = %s] purge validation result cache and validation event cache success", executor.namespace)
}

// addPendingValidationEvent pushes a validation event to pending queue.
func (executor *Executor) addPendingValidationEvent(validationEvent event.TransactionBlock) {
	executor.logger.Warningf("[Namespace = %s] receive validation event %d while %d is required, save into cache temporarily.", executor.namespace, validationEvent.SeqNo, executor.context.getDemand(DemandSeqNo))
	executor.cache.pendingValidationEventQ.Add(validationEvent.SeqNo, validationEvent)
}

// fetchPendingValidationEvent fetches a validation event in pending queue via seqNo, return false if not exist.
func (executor *Executor) fetchPendingValidationEvent(seqNo uint64) (event.TransactionBlock, bool) {
	res, existed := executor.cache.pendingValidationEventQ.Get(seqNo)
	if existed == false {
		return event.TransactionBlock{}, false
	}
	ev := res.(event.TransactionBlock)
	return ev, true
}

// pendingValidationEventQLen retrieves the length of pending validation event queue.
func (executor *Executor) pendingValidationEventQLen() int {
	return executor.cache.pendingValidationEventQ.Len()
}

// clearPendingValidationEventQ purges validation event queue.
func (executor *Executor) clearPendingValidationEventQ() {
	length := executor.pendingValidationEventQLen()
	executor.cache.pendingValidationEventQ.Purge()
	atomic.AddInt32(&executor.context.validateQueueLen, -1*int32(length))
}

// addValidationResult saves a validation result to cache.
func (executor *Executor) addValidationResult(tag ValidationTag, res *ValidationResultRecord) {
	executor.cache.validationResultCache.Add(tag, res)
}

// fetchValidationResult fetches the validation result with given hash.
func (executor *Executor) fetchValidationResult(tag ValidationTag) (*ValidationResultRecord, bool) {
	v, existed := executor.cache.validationResultCache.Get(tag)
	if existed == false {
		return nil, false
	}
	return v.(*ValidationResultRecord), true
}

// addValidationEvent pushes a validation event to channel buffer.
func (executor *Executor) addValidationEvent(ev event.TransactionBlock) {
	executor.cache.validationEventC <- ev
	atomic.AddInt32(&executor.context.validateQueueLen, 1)
	executor.logger.Debugf("[Namespace = %s] receive a validation event #%d", executor.namespace, ev.SeqNo)
}

// fetchValidationEvent fetches a validation event from channel buffer.
func (executor *Executor) fetchValidationEvent() chan event.TransactionBlock {
	return executor.cache.validationEventC
}

// processValidationDone is the callback func after validation process finished.
func (executor *Executor) processValidationDone() {
	atomic.AddInt32(&executor.context.validateQueueLen, -1)
}

// addCommitEvent pushes a commit event to channel buffer.
func (executor *Executor) addCommitEvent(ev event.CommitEvent) {
	executor.cache.commitEventC <- ev
	atomic.AddInt32(&executor.context.commitQueueLen, 1)
	executor.logger.Debugf("[Namespace = %s] receive a commit event #%d", executor.namespace, ev.SeqNo)
}

// fetchCommitEvent fetches a commit event from channel buffer.
func (executor *Executor) fetchCommitEvent() chan event.CommitEvent {
	return executor.cache.commitEventC
}

// processCommitDone is the callback func after commit process finished.
func (executor *Executor) processCommitDone() {
	atomic.AddInt32(&executor.context.commitQueueLen, -1)
}

// addToSyncCache adds a block to cache which arrives earlier than expect.
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

// fetchFromSyncCache fetches blocks from sync cache.
func (executor *Executor) fetchFromSyncCache(number uint64) (map[string]types.Block, bool) {
	ret, existed := executor.cache.syncCache.Get(number)
	if !existed {
		return nil, false
	}
	blks := ret.(map[string]types.Block)
	return blks, true
}

// addToReplicaCache pushes a replica status to cache with ip and port as flag.
func (executor *Executor) addToReplicaCache(ip string, port int32, chain *types.Chain) {
	executor.cache.replicaInfoCache.Add(Peer{
		Ip:   ip,
		Port: port,
	}, chain)
}

// fetchFromReplicaCache fetches a replica's status from cache with ip and port as key.
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
