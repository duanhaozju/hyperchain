package core

import "hyperchain-alpha/core/types"

//-- 将Chain存入内存
func PutChainToMEM(key string, t types.Chain) error{
	return PutChain(MemDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Chain
func GetChainFromMEM(key string) (types.Chain, error){
	return GetChain(MemDB, []byte(key))
}

//-- 从MEM中删除Chain
func DeleteChainFromMEM(key string) error {
	return DeleteChain(MemDB, []byte(key))
}

//-- 从MEM中获取所有Chain
func GetAllChainFromMEM() ([]types.Chain, error) {
	var ts []types.Chain

	Keys := MemDB.Keys()

	var err error
	for _, key := range Keys {
		var t types.Chain
		value , _ := MemDB.Get(key)
		err = decondeFromBytes(value, &t)
		ts = append(ts, t)
	}
	return ts, err
}