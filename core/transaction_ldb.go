package core

import "hyperchain-alpha/hyperdb"

import (
	"hyperchain-alpha/core/types"
	"fmt"
)

//-- 将Transaction存入ldb数据库
func PutTransactionToLDB(key string, t types.Transaction) error{
	fmt.Println(t)
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return putTransaction(ldb, []byte(key), t)
}

//-- 在ldb中 根据Key获取的Transaction
func GetTransactionFromLDB(key string) (types.Transaction, error){

	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return types.Transaction{},err
	}
	return getTransaction(ldb, []byte(key))
}

//-- 从ldb中删除Transaction
func DeleteTransactionFromLDB(key string) error {
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return deleteTransaction(ldb, []byte(key))
}

//-- 从ldb中获取所有Transaction
func GetAllTransactionFromLDB() ([]types.Transaction, error) {
	var ts []types.Transaction
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return ts, err
	}

	iter := ldb.NewIterator()
	for iter.Next() {
		var t types.Transaction
		value := iter.Value()
		err = decondeFromBytes(value, &t)
		ts = append(ts, t)
	}
	iter.Release()
	err = iter.Error()
	return ts, err
}