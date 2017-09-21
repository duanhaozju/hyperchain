package db_utils

import (
	"errors"
	"github.com/op/go-logging"
	"github.com/willf/bloom"
	"hyperchain/common"
	"hyperchain/core/types"
	"sync"
	"time"
)

var (
	DuplicateBloomFilterRegisterErr = errors.New("duplicated bloom filter registeration")
	BloomFilterNotExistErr          = errors.New("bloom filter hasn't been registered")
	InitBloomFilterFailedErr        = errors.New("init bloom filter failed")
	BloomCacheFailureErr            = errors.New("bloom cache in failure")
	bloomFilterCache                *BloomFilterCache
)

const (
	RebuildTime     = "duplicate.remove.bloomfilter.rebuild_time"
	RebuildInterval = "duplicate.remove.bloomfilter.interval"
	BloomBit        = "duplicate.remove.bloomfilter.bloombit"
)

// Bloomfilter implement the muilti-namespace transaciton bloom filter.
// It can help to do the transaction duplication checking.
type BloomFilterCache struct {
	c          map[string]*bloom.BloomFilter
	lock       sync.RWMutex
	log        *logging.Logger
	closed     chan struct{}
	registerCh chan request // channel for passing registration request
	updateCh   chan request // channel for passing updating request
	writeCh    chan message // channel for passing writing operation
	remove     chan string  // channel for passing removing request
	config     *common.Config
	wg         sync.WaitGroup
}

// request - registration or updating request for namespace-level bloom filter.
type request struct {
	namespace string
	filter    *bloom.BloomFilter
	cont      chan bool
}

// message - write new content to bloom filter.
type message struct {
	namespace string
	hashes    []common.Hash
	cont      chan bool
}

func NewBloomCache(config *common.Config) *BloomFilterCache {
	filter := &BloomFilterCache{
		c:          make(map[string]*bloom.BloomFilter),
		log:        common.GetLogger(common.DEFAULT_NAMESPACE, "db_utils"),
		closed:     make(chan struct{}),
		updateCh:   make(chan request),
		registerCh: make(chan request),
		writeCh:    make(chan message),
		remove:     make(chan string),
		config:     config,
	}
	go filter.loop()
	go filter.expire()
	bloomFilterCache = filter
	return filter
}

func (cache *BloomFilterCache) loop() {
	cache.wg.Add(1)
	for {
		select {
		case <-cache.closed:
			cache.wg.Done()
			return
		case req := <-cache.updateCh:
			cache.lock.Lock()
			cache.c[req.namespace] = req.filter
			cache.lock.Unlock()
			close(req.cont)
		case req := <-cache.registerCh:
			if _, exist := cache.c[req.namespace]; exist {
				cache.log.Debug("the namespace bloom filter has been registered")
				req.cont <- false
			} else {
				cache.lock.Lock()
				cache.c[req.namespace] = req.filter
				cache.lock.Unlock()
				req.cont <- true
				cache.log.Debugf("register %s bloom filter successfully.", req.namespace)
			}
		case ns := <-cache.remove:
			cache.lock.Lock()
			delete(cache.c, ns)
			cache.lock.Unlock()
			cache.log.Debugf("remove %s bloom filter successfully.", ns)
		case msg := <-cache.writeCh:
			var filter *bloom.BloomFilter
			cache.lock.RLock()
			filter = cache.c[msg.namespace]
			cache.lock.RUnlock()
			// create a copy of filter, so that the computation procedure won't affect the online filter
			go func() {
				if filter == nil {
					msg.cont <- false
					cache.log.Debugf("write %s bloom filter failed.", msg.namespace)
				} else {
					filter = filter.Copy()
					for _, hash := range msg.hashes {
						filter.Add(hash.Bytes())
					}
					cache.lock.Lock()
					cache.c[msg.namespace] = filter
					cache.lock.Unlock()
					cache.log.Debugf("write %s bloom filter successfully.", msg.namespace)
					msg.cont <- true
				}
			}()
		}
	}
}

func (cache *BloomFilterCache) expire() {
	cache.wg.Add(1)
	var (
		rebuildTime int64 = cache.config.GetInt64(RebuildTime)
		interval    int64 = cache.config.GetInt64(RebuildInterval)
		timer       *time.Timer
	)

	var duration time.Duration = time.Duration(int64(24 + int(rebuildTime) - time.Now().Hour()))
	timer = time.NewTimer(duration * time.Hour)
	for {
		select {
		case <-cache.closed:
			cache.wg.Done()
			return
		case <-timer.C:
			cache.rebuild(nil)
			timer.Reset(time.Duration(interval) * time.Hour)
		}
	}
}

// rebuild bloom filter rebuild handler.
func (cache *BloomFilterCache) rebuild(hook func()) {
	var (
		pending map[string]struct{}           = make(map[string]struct{})
		finish  map[string]*bloom.BloomFilter = make(map[string]*bloom.BloomFilter)
		size    int64                         = cache.config.GetInt64(BloomBit)
	)
	cache.lock.RLock()
	for ns := range cache.c {
		pending[ns] = struct{}{}
	}
	cache.lock.RUnlock()
	// rebuild serially to decrease the system IO pressure
	for ns := range pending {
		filter := cache.InitBloomFilter(int(size), ns, hook)
		if filter != nil {
			finish[ns] = filter
		}
	}
	cache.lock.Lock()
	for ns, filter := range finish {
		cache.c[ns] = filter
	}
	cache.lock.Unlock()
}

// Register register a bloom filter with namespace as a parameter.
// Duplication error will been returned if the filter is already existed.
func (cache *BloomFilterCache) Register(namespace string) error {
	size := cache.config.GetInt64(BloomBit)
	filter := cache.InitBloomFilter(int(size), namespace, nil)
	if filter == nil {
		return InitBloomFilterFailedErr
	}
	wait := make(chan bool)
	cache.registerCh <- request{namespace, filter, wait}
	if ret := <-wait; !ret {
		return DuplicateBloomFilterRegisterErr
	} else {
		cache.log.Noticef("register %s success", namespace)
		return nil
	}
}

// Unregister remove the relative bloom filter.
func (cache *BloomFilterCache) UnRegister(namespace string) error {
	cache.remove <- namespace
	return nil
}

// Put force update the bloom filter in cache with the given one.
func (cache *BloomFilterCache) Put(namespace string, filter *bloom.BloomFilter) error {
	wait := make(chan bool)
	cache.updateCh <- request{namespace, filter, wait}
	<-wait
	return nil
}

// Get obtain bloom filter handler with given namespace.
// Only the copied will been return.
func (cache *BloomFilterCache) Get(namespace string) (error, *bloom.BloomFilter) {
	cache.lock.RLock()
	if filter, exist := cache.c[namespace]; !exist {
		cache.lock.RUnlock()
		return BloomFilterNotExistErr, nil
	} else {
		ret := filter.Copy()
		cache.lock.RUnlock()
		return nil, ret
	}
}

// Look input the given hash and make bloom filter checking.
func (cache *BloomFilterCache) Look(namespace string, hash common.Hash) (error, bool) {
	cache.lock.RLock()
	filter := cache.c[namespace]
	if filter == nil {
		cache.lock.RUnlock()
		return BloomFilterNotExistErr, false
	} else {
		res := filter.Test(hash.Bytes())
		cache.lock.RUnlock()
		return nil, res
	}
}

// Write write new content to relative bloom filter.
func (cache *BloomFilterCache) Write(namespace string, txs []*types.Transaction) (error, bool) {
	var hashes []common.Hash
	for _, tx := range txs {
		hashes = append(hashes, tx.GetHash())
	}
	cont := make(chan bool)
	cache.writeCh <- message{namespace, hashes, cont}
	if res := <-cont; res {
		return nil, false
	} else {
		return BloomFilterNotExistErr, true
	}
}

func (cache *BloomFilterCache) Close() {
	close(cache.closed)
	cache.wg.Wait()
}

func (cache *BloomFilterCache) InitBloomFilter(bloomBit int, namespace string, sleepHook func()) *bloom.BloomFilter {
	var (
		start   time.Time = time.Now()
		now     int64     = time.Now().UnixNano()
		head    uint64
		genesis uint64
		cur     uint64
		filter  *bloom.BloomFilter = bloom.New(uint(bloomBit), 3)
		counter int
		err     error
		chain   *types.Chain
	)

	chain, err = GetChain(namespace)
	if err != nil {
		return nil
	}
	head = chain.Height
	genesis = chain.Genesis
	cur = head

	// Build bloom filter from the current head block.
	// Break the loop if current block is old enough.
	for {
		if cur == 0 || cur < genesis {
			break
		}
		blk, err := GetBlockByNumber(namespace, cur)
		if err != nil {
			cache.log.Warningf("missing block (#%d) when building tx bloom filter", cur)
			cur -= 1
			continue
		}
		if blk.Timestamp+time.Duration(24*time.Hour).Nanoseconds() < now {
			break
		}
		for _, tx := range blk.Transactions {
			filter.Add(tx.GetTransactionHash())
		}
		counter += 1
		cur -= 1
	}

	if sleepHook != nil {
		sleepHook()
	}
	// Add new blocks while generated during the last stage.
	var repeat int = 0
	for cur = head + 1; cur <= GetHeightOfChain(namespace); cur += 1 {
		blk, err := GetBlockByNumber(namespace, cur)
		if err != nil {
			// Since block persistence is delayed compare with memory chain heigh updating.
			// So use a tiny loop to capture the latest block content.
			cur -= 1
			// Prevent dead loop happen.
			if repeat > 100 {
				break
			} else {
				repeat += 1
			}
			continue
		}
		repeat = 0
		if blk.Timestamp+time.Duration(24*time.Hour).Nanoseconds() < now {
			continue
		}
		for _, tx := range blk.Transactions {
			filter.Add(tx.GetTransactionHash())
		}
		counter += 1
	}
	cache.log.Debugf("build bloom filter for namespace %s success. totally %d block been pushed into filter, elapsed %v", namespace, counter, time.Since(start))
	return filter
}

func LookupTransaction(namespace string, txHash common.Hash) (error, bool) {
	if bloomFilterCache == nil {
		return BloomCacheFailureErr, false
	}
	return bloomFilterCache.Look(namespace, txHash)
}

func WriteTxBloomFilter(namespace string, txs []*types.Transaction) (error, bool) {
	if bloomFilterCache == nil {
		return BloomCacheFailureErr, false
	}
	return bloomFilterCache.Write(namespace, txs)
}
