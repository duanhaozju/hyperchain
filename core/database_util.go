package core

import (
	"hyperchain-alpha/hyperdb"
	"hyperchain-alpha/core/types"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
)

var (
	TransactionHeaderKey = []byte("transactionHeader")
)

//TODO 改变成更高效的方式编码 解码
//-- 将一个对象编码成[]byte
//-- 这个版本先用json.Marshal序列化
func encodeToBytes(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

//-- 将一个[]byte解码成对象
//-- 这个版本先用jsonUnmarshal 反序列化
func decondeFromBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

//-- 初始化ldb 和memdb
//-- 应该在程序开始初始化
func InitDB(port int) {
	LDBPath = BaseLDBPath + strconv.Itoa(port)
}

//-- --------------- about ldb -----------------------

func getBasePath() string {
	path, _ := os.Getwd()
	return path
}

var (
	BaseLDBPath = getBasePath() + "/cache/"
	LDBPath string
)
func GetLDBConnection()(*leveldb.DB,error){
	db, err := leveldb.OpenFile(LDBPath, nil )
	return db, err
}
//-- ------------------ ldb end ----------------------


//-- ------------------- Transaction ---------------------------------
func PutTransaction(db hyperdb.Database, key []byte, t types.Transaction) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(TransactionHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetTransaction(db hyperdb.Database, key []byte) (types.Transaction, error){
	var transaction types.Transaction
	keyFact := append(TransactionHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return transaction, err
	}
	err = decondeFromBytes(data, &transaction)
	return transaction, err
}

func DeleteTransaction(db hyperdb.Database, key []byte) error {
	keyFact := append(TransactionHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Transaction END -----------------------------------