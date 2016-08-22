package core

import (
	"hyperchain-alpha/core/types"
	"sync"
	"log"
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

//-- 更新balance表 需要一个新的区块
func UpdateBalance(block types.Block)  {
	log.Println("更新余额表...")
	memBalanceMap.lock.Lock()
	defer memBalanceMap.lock.Unlock()
	for _, trans := range block.Transactions {
		//-- 将交易里的From(账户的publickey)余额减去value
		//-- 如果余额表中没有这个From(实际上不可能，因为余额表中没有这个From，不可能会验证通过
		//-- 但是上帝用户例外，因为上帝用户可能会出现负数)，则新建一个
		//-- 如果余额表中有这个From，则覆盖publickey(覆盖的Publickey是一样的，实际上没改)
		b := memBalanceMap.data[trans.From]
		b.Value -= trans.Value
		b.AccountPublicKeyHash = trans.From
		memBalanceMap.data[trans.From] = b
		//-- 将交易中的To(账户中的publickey)余额加上value
		//-- 如果余额表中没有这个To(就是所有publickey不含有To)
		//-- 新建一个balance，将交易的value直接赋给balance.value
		//-- 如果余额表中有这个To,则直接加上交易中的value
		if _, ok := memBalanceMap.data[trans.To]; ok {
			b = memBalanceMap.data[trans.To]
			b.Value += trans.Value
			memBalanceMap.data[trans.To] = b
		}else {
			b = types.Balance{
				AccountPublicKeyHash: trans.To,
				Value: trans.Value,
			}
			memBalanceMap.data[trans.To] = b
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
