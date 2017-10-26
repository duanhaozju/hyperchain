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
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"

	"github.com/willf/bloom"
)

// TxBloomFilter provider functions related to transaction-bloom-filter to be invoked by outer service
type TxBloomFilter interface {
	// Start starts the BloomFilterCache
	Start()

	// Register registers a bloom filter with given namespace
	Register(namespace string) error

	// UnRegister removes the relative bloom filter with given namespace
	UnRegister(namespace string) error

	// Get obtains the handler of bloom filter with given namespace
	Get(namespace string) (*bloom.BloomFilter, error)

	// Put force updates the bloom filter in cache with the given one
	Put(namespace string, filter *bloom.BloomFilter) error

	// Write writes new content to relative bloom filter
	Write(namespace string, txs []*types.Transaction) (bool, error)

	// Look inputs the given hash and make bloom filter checking
	Look(namespace string, hash common.Hash) (bool, error)

	// Close closes the BloomFilter
	Close()
}
