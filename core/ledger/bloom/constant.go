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
