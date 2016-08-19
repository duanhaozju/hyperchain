package core

import "hyperchain-alpha/hyperdb"

import (
	"hyperchain-alpha/core/types"
	"encoding/json"
)


func PutTransactionToLDB(key string, t types.Transaction) error{
	ldb, err := hyperdb.NewLDBDataBase(LDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return PutTransaction(ldb, []byte(key), t)
}

func GetTransactionFromLDB(key string) (types.Transaction, error){

	ldb, err := hyperdb.NewLDBDataBase(LDBPath)
	defer ldb.Close()
	if err != nil {
		return types.Transaction{},err
	}
	return GetTransaction(ldb, []byte(key))
}

func DeleteTransactionFromLDB(key string) error {
	ldb, err := hyperdb.NewLDBDataBase(LDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return DeleteTransaction(ldb, []byte(key))
}

func GetAllTransactionFromLDB() ([]types.Transaction, error) {
	var ts []types.Transaction
	ldb, err := hyperdb.NewLDBDataBase(LDBPath)
	defer ldb.Close()
	if err != nil {
		return ts, err
	}

	iter := ldb.NewIterator()
	for iter.Next() {
		var t types.Transaction
		value := iter.Value()
		err = json.Unmarshal(value, &t)
		ts = append(ts, t)
	}
	iter.Release()
	err = iter.Error()
	return ts, err
}