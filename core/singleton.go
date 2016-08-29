package core

import (
	"sync"
	"hyperchain/core/types"
	"strconv"
	"fmt"
)

//-- --------------------- Balance ------------------------------------\
// Balance account balance
// The key is account public-key-hash
// The value is the account balance value
type Balance map[string]int64

type balance struct {
	data Balance
	lock sync.RWMutex
}

func newMemBalance() *balance {
	return &balance{
		data: make(map[string]types.Balance),
	}
}

var balanceInstance = &balance{
	data: make(map[string]types.Balance),
};

func initMem()  {
	if balanceInstance == nil {
		balanceInstance = &balance{
			data: make(map[string]types.Balance),
		}
	}
}

//-- 将Balance存入内存
func PutBalance(accountPublicKeyHash string, value int64){
	balanceInstance.lock.Lock()
	defer balanceInstance.lock.Unlock()
	balanceInstance.data[accountPublicKeyHash] = value
}

//-- 在MEM中 根据Key获取的Balance
func GetBalance(accountPublicKeyHash string) Balance{
	balanceInstance.lock.RLock()
	defer balanceInstance.lock.RUnlock()
	return balanceInstance.data[accountPublicKeyHash]
}

//-- 从MEM中删除Balance
func DeleteBalance(accountPublicKeyHash string) {
	balanceInstance.lock.Lock()
	defer balanceInstance.lock.Unlock()
	delete(balanceInstance.data, accountPublicKeyHash)
}

//-- 从MEM中获取所有Balance
func GetAllBalance() (Balance) {
	balanceInstance.lock.RLock()
	defer balanceInstance.lock.RUnlock()
	var bs Balance
	for key, value := range balanceInstance.data {
		bs[key] = value
	}
	return bs
}

//-- 更新balance表 需要一个新的区块
func UpdateBalance(block types.Block)  {
	balanceInstance.lock.Lock()
	defer balanceInstance.lock.Unlock()
	for _, trans := range block.Transactions {
		//-- 将交易里的From(账户的publickey)余额减去value
		//-- 如果余额表中没有这个From(实际上不可能，因为余额表中没有这个From，不可能会验证通过
		//-- 但是上帝用户例外，因为上帝用户可能会出现负数)，则新建一个
		//-- 如果余额表中有这个From，则覆盖publickey(覆盖的Publickey是一样的，实际上没改)

		value, err := strconv.Atoi(string(trans.Value))
		if err != nil {
			fmt.Println(err)
			break
		}
		b := balanceInstance.data[trans.From]
		b -= value
		balanceInstance.data[trans.From] = b
		//-- 将交易中的To(账户中的publickey)余额加上value
		//-- 如果余额表中没有这个To(就是所有publickey不含有To)
		//-- 新建一个balance，将交易的value直接赋给balance.value
		//-- 如果余额表中有这个To,则直接加上交易中的value
		if _, ok := balanceInstance.data[trans.To]; ok {
			b = balanceInstance.data[trans.To]
			b += value
			balanceInstance.data[trans.To] = b
		}else {
			b = value
			balanceInstance.data[trans.To] = b
		}
	}
}
//-- --------------------- Balance END --------------------------------
