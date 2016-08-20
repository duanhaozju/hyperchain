package core

import (
	"hyperchain-alpha/core/types"
)

//-- 将Balance存入内存
func PutBalanceToMEM(key string, t types.Balance) error{
	return PutBalance(MemDB, []byte(key), t)
}

//-- 在MEM中 根据Key获取的Balance
func GetBalanceFromMEM(key string) (types.Balance, error){
	return GetBalance(MemDB, []byte(key))
}

//-- 从MEM中删除Balance
func DeleteBalanceFromMEM(key string) error {
	return DeleteBalance(MemDB, []byte(key))
}

//-- 从MEM中获取所有Balance
func GetAllBalanceFromMEM() ([]types.Balance, error) {
	var ts []types.Balance

	Keys := MemDB.Keys()

	var err error
	for _, key := range Keys {
		var t types.Balance
		value , _ := MemDB.Get(key)
		err = decondeFromBytes(value, &t)
		ts = append(ts, t)
	}
	return ts, err
}