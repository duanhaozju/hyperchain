package core

import (
	"hyperchain-alpha/core/types"
	"sync"
)

type memChain struct {
	data map[string]types.Chain
	lock sync.RWMutex
}

func newMemChain() *memChain {
	return &memChain{
		data: make(map[string]types.Chain),
	}
}

var memChainMap *memChain;

//-- 将Chain存入内存
func PutChainToMEM(key string, t types.Chain){
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data[key] = t
}

/*GetLastestBlockHash()
UpdateChain(hash)*/
//-- 在MEM中 根据Key获取的Chain
func GetChainFromMEM(key string) types.Chain{
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data[key]
}

//-- 从MEM中删除Chain
func DeleteChainFromMEM(key string) {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	delete(memChainMap.data, key)
}

//-- 从MEM中获取所有Chain
func GetAllChainFromMEM() ([]types.Chain) {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	var ts []types.Chain
	for _, m := range memChainMap.data {
		ts = append(ts, m)
	}
	return ts
}

/*
//-- 将Chain存入内存
func PutChainToMEM(key string, t types.Chain) error{
	return putChain(memDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Chain
func GetChainFromMEM(key string) (types.Chain, error){
	return getChain(memDB, []byte(key))
}

//-- 从MEM中删除Chain
func DeleteChainFromMEM(key string) error {
	return deleteChain(memDB, []byte(key))
}

//-- 从MEM中获取所有Chain
func GetAllChainFromMEM() ([]types.Chain, error) {
	var ts []types.Chain

	Keys := memDB.Keys()
	var err error
	for _, key := range Keys {
		if string(key[:len(chainHeaderKey)]) == string(chainHeaderKey) {
			var t types.Chain
			value, _ := memDB.Get(key)
			err = decondeFromBytes(value, &t)
			ts = append(ts, t)
		}
	}
	return ts, err
}*/
