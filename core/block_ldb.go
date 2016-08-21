package core

import (
	"hyperchain-alpha/hyperdb"
	"hyperchain-alpha/core/types"
)


//-- 将Block存入ldb数据库
func PutBlockToLDB(key string, t types.Block) error{
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return putBlock(ldb, []byte(key), t)
}

//-- 在ldb中 根据Key获取的Block
func GetBlockFromLDB(key string) (types.Block, error){

	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return types.Block{},err
	}
	return getBlock(ldb, []byte(key))
}

//-- 从ldb中删除Block
func DeleteBlockFromLDB(key string) error {
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return err
	}
	return deleteBlock(ldb, []byte(key))
}

//-- 从ldb中获取所有Block
func GetAllBlockFromLDB() ([]types.Block, error) {
	var ts []types.Block
	ldb, err := hyperdb.NewLDBDataBase(lDBPath)
	defer ldb.Close()
	if err != nil {
		return ts, err
	}

	iter := ldb.NewIterator()
	for iter.Next() {
		key := iter.Key()
		if len(string(key)) >= len(blockHeaderKey) && string(key[:len(blockHeaderKey)]) == string(blockHeaderKey) {
			var t types.Block
			value := iter.Value()
			err = decondeFromBytes(value, &t)
			ts = append(ts, t)
		}
	}
	iter.Release()
	err = iter.Error()
	return ts, err
}