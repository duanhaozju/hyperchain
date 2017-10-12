package bloom

import (
	"hyperchain/common"
	"hyperchain/core/types"
)

func LookupTransaction(namespace string, txHash common.Hash) (bool, error) {
	if bloomFilterCache == nil {
		return false, ErrBloomCacheFailure
	}
	return bloomFilterCache.Look(namespace, txHash)
}

func WriteTxBloomFilter(namespace string, txs []*types.Transaction) (bool, error) {
	if bloomFilterCache == nil {
		return false, ErrBloomCacheFailure
	}
	return bloomFilterCache.Write(namespace, txs)
}
