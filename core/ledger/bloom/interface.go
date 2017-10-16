package bloom

import (
	"hyperchain/common"
	"hyperchain/core/types"

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
