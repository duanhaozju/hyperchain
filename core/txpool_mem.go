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
var maxCapacity = 5

func newMemTxPool() *memTxPool {
	return &memTxPool{
		txPool:types.TxPool{
			MaxCapacity:maxCapacity,
			Transactions:make([]types.Transaction, 0, maxCapacity),
		},
	}
}

//-- 获取TxPool中所有的交易
func GetTransactionsFromTxPool() []types.Transaction {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return memTxPoolMap.txPool.Transactions
}

//-- 将交易加到Pool 但是并不能保证溢出
//-- 在调用这个方法之后应该先调用TxPoolIsFull()判断Pool是否为满
//-- 如果为满应该先清空
func AddTransactionToTxPool(tran types.Transaction){
	memTxPoolMap.lock.Lock()
	defer memTxPoolMap.lock.Unlock()
	memTxPoolMap.txPool.Transactions = append(memTxPoolMap.txPool.Transactions, tran)
}

//-- 判断TxPool是否为满状态
func TxPoolIsFull() bool {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return len(memTxPoolMap.txPool.Transactions) == memTxPoolMap.txPool.MaxCapacity
}

//-- 清空交易池
func ClearTxPool()  {
	memTxPoolMap.lock.Lock()
	defer memTxPoolMap.lock.Unlock()
	memTxPoolMap.txPool.Transactions = make([]types.Transaction, 0, maxCapacity)
}

//-- 获取交易池的容量
func GetTxPoolCapacity() int {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return len(memTxPoolMap.txPool.Transactions)
}
