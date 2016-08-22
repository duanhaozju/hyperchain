package core

import (
	"hyperchain-alpha/core/types"
	"sync"
)

type memChain struct {
	data types.Chain
	lock sync.RWMutex
}

func newMemChain() *memChain {
	return &memChain{
		data: types.Chain{
			Height: 0,
		},
	}
}
var memChainMap *memChain;

//-- 获取最新的blockhash
func GetLashestBlockHash() string {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.LastestBlockHash
}

//-- 更新Chain，即更新最新的blockhash 并将height加1
//-- blockHash为最新区块的hash
func UpdateChain(blockHash string)  {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.LastestBlockHash = blockHash
	memChainMap.data.Height += 1
}

//-- 获取区块的高度
func GetHeightOfChain() int {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.Height
}

//-- 获取chain的拷贝
func GetChain() *types.Chain {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return &types.Chain{
		LastestBlockHash: memChainMap.data.LastestBlockHash,
		Height: memChainMap.data.Height,
	}
}