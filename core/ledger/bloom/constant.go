package bloom

import (
	"errors"
)

var (
	// DuplicateBloomFilterRegisterErr is thrown by (BloomFilterCache)Register call if a bloom filter with the given namespace is already exists
	ErrDuplicateBloomFilterRegister = errors.New("duplicated bloom filter registeration")
	// BloomFilterNotExistErr is thrown by (BloomFilterCache)Get call if a bloom filter with the given namespace does not exist
	ErrBloomFilterNotExist = errors.New("bloom filter hasn't been registered")
	// InitBloomFilterFailedErr is thrown by (BloomFilterCache)Register call if a bloom filter init fails
	ErrInitBloomFilterFailed = errors.New("init bloom filter failed")
	// BloomCacheFailureErr is thrown by LookupTransaction or WriteTxBloomFilter call if bloomFilterCache is nil
	ErrBloomCacheFailure = errors.New("bloom cache in failure")
	// The singleton of BloomFilterCache for transactions of all namespaces
	bloomFilterCache *BloomFilterCache
)

// Constants of BloomFilterCache that have been set in config file
const (
	RebuildTime     = "duplicate.remove.bloomfilter.rebuild_time"
	ActiveTime      = "duplicate.remove.bloomfilter.active_time"
	RebuildInterval = "duplicate.remove.bloomfilter.interval"
	BloomBit        = "duplicate.remove.bloomfilter.bloombit"
)
