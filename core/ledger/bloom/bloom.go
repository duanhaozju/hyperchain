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

package bloom

import (
	"sync"
	"time"

	"hyperchain/common"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/types"

	"github.com/op/go-logging"
	"github.com/willf/bloom"
)

// Bloomfilter implements interface TxBloomFilter. It's the muilti-namespace transaciton bloom filter
// and can help to do the transaction duplication checking.
type BloomFilterCache struct {
	config *common.Config  // Common config of bloom filter
	log    *logging.Logger // Log handler

	bloomMap map[string]*bloom.BloomFilter // Map of bloom filter for different namespace
	lock     sync.RWMutex                  // Lock for single cache read/write
	wg       sync.WaitGroup                // WaitGroup used in running loop

	closed     chan struct{} // Chanel for close signal
	registerCh chan request  // Channel for passing registration request
	updateCh   chan request  // Channel for passing updating request
	writeCh    chan message  // Channel for passing writing operation
	remove     chan string   // Channel for passing removing request
}

// request represents the registration/updating request for namespace-level bloom filter.
type request struct {
	namespace string
	filter    *bloom.BloomFilter
	resp      chan bool
}

// message represents request of writing new content to bloom filter.
type message struct {
	namespace string
	hashes    []common.Hash
	resp      chan bool
}

// NewBloomFilterCache constructs new BloomFilterCache.
func NewBloomFilterCache(config *common.Config) *BloomFilterCache {
	filter := &BloomFilterCache{
		config:     config,
		log:        common.GetLogger(common.DEFAULT_NAMESPACE, "bloom"),
		bloomMap:   make(map[string]*bloom.BloomFilter),
		closed:     make(chan struct{}),
		updateCh:   make(chan request),
		registerCh: make(chan request),
		writeCh:    make(chan message),
		remove:     make(chan string),
	}
	bloomFilterCache = filter

	return filter
}

// Start implements the corresponding method from interface bloom.TxBloomFilter.
// This function runs the two threads, loop for normal handler of requests, expire for
// rebuild new bloom cache if timer expired.
func (cache *BloomFilterCache) Start() {
	go cache.loop()
	go cache.expire()
}

// Register implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) Register(namespace string) error {
	size := cache.config.GetInt64(BloomBit)
	filter := cache.initBloomFilter(int(size), namespace, nil)
	if filter == nil {
		return ErrInitBloomFilterFailed
	}
	resp := make(chan bool)
	cache.registerCh <- request{namespace, filter, resp}
	if ret := <-resp; !ret {
		return ErrDuplicateBloomFilterRegister
	} else {
		cache.log.Noticef("register %s success", namespace)
		return nil
	}
}

// UnRegister implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) UnRegister(namespace string) error {
	cache.remove <- namespace
	return nil
}

// Get implements the corresponding method from interface bloom.TxBloomFilter.
// Only the copied will been return.
func (cache *BloomFilterCache) Get(namespace string) (*bloom.BloomFilter, error) {
	cache.lock.RLock()
	if filter, exist := cache.bloomMap[namespace]; !exist {
		cache.lock.RUnlock()
		return nil, ErrBloomFilterNotExist
	} else {
		ret := filter.Copy()
		cache.lock.RUnlock()
		return ret, nil
	}
}

// Put implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) Put(namespace string, filter *bloom.BloomFilter) error {
	resp := make(chan bool)
	cache.updateCh <- request{namespace, filter, resp}
	<-resp
	return nil
}

// Write implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) Write(namespace string, txs []*types.Transaction) (bool, error) {
	var hashes []common.Hash
	for _, tx := range txs {
		hashes = append(hashes, tx.GetHash())
	}
	resp := make(chan bool)
	cache.writeCh <- message{namespace, hashes, resp}
	if res := <-resp; res {
		return false, nil
	} else {
		return true, ErrBloomFilterNotExist
	}
}

// Look implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) Look(namespace string, hash common.Hash) (bool, error) {
	cache.lock.RLock()
	filter := cache.bloomMap[namespace]
	if filter == nil {
		cache.lock.RUnlock()
		return false, ErrBloomFilterNotExist
	} else {
		res := filter.Test(hash.Bytes())
		cache.lock.RUnlock()
		return res, nil
	}
}

// Close implements the corresponding method from interface bloom.TxBloomFilter.
func (cache *BloomFilterCache) Close() {
	close(cache.closed)
	cache.wg.Wait()
}

// loop listens a series of requests and handles them.
func (cache *BloomFilterCache) loop() {
	cache.wg.Add(1)
	for {
		select {
		// If Close called, close this loop
		case <-cache.closed:
			cache.wg.Done()
			return
		// Update the bloomMap of corresponding filter with given namespace
		case req := <-cache.updateCh:
			cache.lock.Lock()
			cache.bloomMap[req.namespace] = req.filter
			cache.lock.Unlock()
			close(req.resp)
		// Register in the bloomMap if the filter of given namespace hasn't been registered
		case req := <-cache.registerCh:
			if _, exist := cache.bloomMap[req.namespace]; exist {
				cache.log.Debug("the namespace bloom filter has been registered")
				req.resp <- false
			} else {
				cache.lock.Lock()
				cache.bloomMap[req.namespace] = req.filter
				cache.lock.Unlock()
				req.resp <- true
				cache.log.Debugf("register %s bloom filter successfully.", req.namespace)
			}
		// Remove the corresponding filter in the bloomMap with given namespace
		case ns := <-cache.remove:
			cache.lock.Lock()
			delete(cache.bloomMap, ns)
			cache.lock.Unlock()
			cache.log.Debugf("remove %s bloom filter successfully.", ns)
		// Write the transaction hashes into corresponding filter
		case msg := <-cache.writeCh:
			var filter *bloom.BloomFilter
			cache.lock.RLock()
			filter = cache.bloomMap[msg.namespace].Copy()
			cache.lock.RUnlock()
			// create a copy of filter, so that the computation procedure won't affect the online filter
			go func() {
				if filter == nil {
					msg.resp <- false
					cache.log.Debugf("write %s bloom filter failed.", msg.namespace)
				} else {
					for _, hash := range msg.hashes {
						filter.Add(hash.Bytes())
					}
					cache.lock.Lock()
					cache.bloomMap[msg.namespace] = filter
					cache.lock.Unlock()
					cache.log.Debugf("write %s bloom filter successfully.", msg.namespace)
					msg.resp <- true
				}
			}()
		}
	}
}

// expire listens if need to rebuild the filter with the set config
func (cache *BloomFilterCache) expire() {
	cache.wg.Add(1)

	var (
		rebuildTime int64 = cache.config.GetInt64(RebuildTime)
		interval    int64 = cache.config.GetInt64(RebuildInterval)
		timer       *time.Timer
		duration    time.Duration
	)

	duration = time.Duration(int64(24 + int(rebuildTime) - time.Now().Hour()))
	timer = time.NewTimer(duration * time.Hour)

	for {
		select {
		// Close this loop if Close is called
		case <-cache.closed:
			cache.wg.Done()
			return
		// Rebuild the filter if timer expired
		case <-timer.C:
			cache.rebuild(nil)
			timer.Reset(time.Duration(interval) * time.Hour)
		}
	}
}

// rebuild rebuilds bloom filter rebuild handler.
func (cache *BloomFilterCache) rebuild(hook func()) {
	var (
		pending map[string]struct{}           = make(map[string]struct{})
		finish  map[string]*bloom.BloomFilter = make(map[string]*bloom.BloomFilter)
		size    int64                         = cache.config.GetInt64(BloomBit)
	)

	// Clear all the filter in bloomMap
	cache.lock.RLock()
	for ns := range cache.bloomMap {
		pending[ns] = struct{}{}
	}
	cache.lock.RUnlock()

	// Rebuild serially to decrease the system IO pressure
	for ns := range pending {
		filter := cache.initBloomFilter(int(size), ns, hook)
		if filter != nil {
			finish[ns] = filter
		}
	}
	cache.lock.Lock()
	for ns, filter := range finish {
		cache.bloomMap[ns] = filter
	}
	cache.lock.Unlock()
}

// initBloomFilter init a bloom filter with given parameters.
func (cache *BloomFilterCache) initBloomFilter(bloomBit int, namespace string, sleepHook func()) *bloom.BloomFilter {
	var (
		start      time.Time = time.Now()
		now        int64     = time.Now().UnixNano()
		duration   time.Duration
		activeTime int64
		head       uint64
		genesis    uint64
		current    uint64
		counter    int
		repeat     int
		c          *types.Chain
		filter     *bloom.BloomFilter = bloom.New(uint(bloomBit), 3)
		err        error
	)

	c, err = chain.GetChain(namespace)
	if err != nil {
		cache.log.Error(err.Error())
		return nil
	}
	head = c.Height
	genesis = c.Genesis
	current = head

	duration, err = time.ParseDuration(cache.config.GetString(ActiveTime))
	if err != nil {
		cache.log.Error(err.Error())
		return nil
	}
	activeTime = duration.Nanoseconds()

	// Build bloom filter from the current head block
	// Break the loop if current block is old enough
	for {
		if current == 0 || current < genesis {
			break
		}
		blk, err := chain.GetBlockByNumber(namespace, current)
		if err != nil {
			cache.log.Warningf("missing block (#%d) when building tx bloom filter", current)
			current--
			continue
		}
		// Keep the last-24h txHashes in the bloom filter
		if blk.Timestamp+activeTime < now {
			break
		}
		for _, tx := range blk.Transactions {
			filter.Add(tx.GetTransactionHash())
		}
		counter++
		current--
	}

	if sleepHook != nil {
		sleepHook()
	}
	// Add new blocks while generated during the last stage.
	repeat = 0
	for current = head + 1; current <= chain.GetHeightOfChain(namespace); current++ {
		blk, err := chain.GetBlockByNumber(namespace, current)
		if err != nil {
			// Since block persistence is delayed compare with memory chain height updating,
			// use a tiny loop to capture the latest block content.
			current--
			// Prevent dead loop happen.
			if repeat > 100 {
				break
			} else {
				repeat++
			}
			continue
		}
		repeat = 0
		if blk.Timestamp+activeTime < now {
			continue
		}
		for _, tx := range blk.Transactions {
			filter.Add(tx.GetTransactionHash())
		}
		counter++
	}
	cache.log.Debugf("build bloom filter for namespace %s success. totally %d block been pushed into filter, elapsed %v", namespace, counter, time.Since(start))

	return filter
}
