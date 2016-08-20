package core

import (
	"sync"
	"hyperchain-alpha/core/types"
)

type memTxPool struct {
	txPool types.TxPool
	lock sync.RWMutex
}

var memTxPoolMap *memTxPool;

func GetTxPool() types.TxPool {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	return memTxPoolMap.txPool
}


func AddTransactionToTxPool(tran types.Transaction) {
	memTxPoolMap.txPool.Transactions = append(memTxPoolMap.txPool.Transactions, tran)
}