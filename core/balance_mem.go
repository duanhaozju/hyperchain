package core

import (
	"hyperchain-alpha/core/types"
	"sync"
)

type memBalance struct {
	data map[string]types.Balance
	lock sync.RWMutex
}

func newMemBalance() *memBalance {
	return &memBalance{
		data: make(map[string]types.Balance),
	}
}

var memBalanceMap *memBalance;

//-- 将Balance存入内存
func PutBalanceToMEM(t types.Balance){
	memBalanceMap.lock.Lock()
	defer memBalanceMap.lock.Unlock()
	key := t.AccountPublicKeyHash
	memBalanceMap.data[key] = t
}

//-- 在MEM中 根据Key获取的Balance
func GetBalanceFromMEM(accountPublicKeyHash string) types.Balance{
	memBalanceMap.lock.RLock()
	defer memBalanceMap.lock.RUnlock()
	return memBalanceMap.data[accountPublicKeyHash]
}

//-- 从MEM中删除Balance
func DeleteBalanceFromMEM(accountPublicKeyHash string) {
	memBalanceMap.lock.Lock()
	defer memBalanceMap.lock.Unlock()
	delete(memBalanceMap.data, accountPublicKeyHash)
}

//-- 从MEM中获取所有Balance
func GetAllBalanceFromMEM() ([]types.Balance) {
	memBalanceMap.lock.RLock()
	defer memBalanceMap.lock.RUnlock()
	var ts []types.Balance
	for _, m := range memBalanceMap.data {
		ts = append(ts, m)
	}
	return ts
}

func UpdateBalance(block types.Block)  {
	memBalanceMap.lock.Lock()
	defer memBalanceMap.lock.Unlock()
	for _, trans := range block.Transactions {
		memBalanceMap.data[trans.From] -= trans.Value
		if _, ok := memBalanceMap.data[trans.To]; ok {
			memBalanceMap.data[trans.To] += trans.Value
		}else {
			memBalanceMap.data[trans.To] = trans.Value
		}
	}
}

/*
//-- 将Balance存入内存
func PutBalanceToMEM(key string, t types.Balance) error{
	return putBalance(memDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Balance
func GetBalanceFromMEM(key string) (types.Balance, error){
	return getBalance(memDB, []byte(key))
}

//-- 从MEM中删除Balance
func DeleteBalanceFromMEM(key string) error {
	return deleteBalance(memDB, []byte(key))
}

//-- 从MEM中获取所有Balance
func GetAllBalanceFromMEM() ([]types.Balance, error) {
	var ts []types.Balance

	Keys := memDB.Keys()

	var err error
	for _, key := range Keys {
		if string(key[:len(balanceHeaderKey)]) == string(balanceHeaderKey) {
			var t types.Balance
			value , _ := memDB.Get(key)
			err = decondeFromBytes(value, &t)
			ts = append(ts, t)
		}
	}
	return ts, err
}*/