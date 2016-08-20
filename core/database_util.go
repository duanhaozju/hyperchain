package core

import (
	"hyperchain-alpha/hyperdb"
	"hyperchain-alpha/core/types"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"hyperchain-alpha/core/node"
)

//-- 这里定义了各种结构在数据库或内存中存取key的前缀
var (
	TransactionHeaderKey = []byte("transactionHeader")
	BlockHeaderKey = []byte("blockHeader")
	BalanceHeaderKey = []byte("balanceHeader")
	ChainHeaderKey = []byte("chainHeaderKey")
	NodeHeaderKey = []byte("nodeHeaderKey")
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

//-- 初始化ldb 和 memdb
//-- 应该在程序开始初始化
//-- port为端口号
func InitDB(port int) {
	LDBPath = BaseLDBPath + strconv.Itoa(port)
	MemDB, _ = hyperdb.NewMemDatabase()
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

//-- --------------- about memdb -----------------------
//-- 这里定义了操作MemDB的
var (
	MemDB *hyperdb.MemDatabase;
)

//-- ------------------ memdb end ----------------------

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


//-- ------------------- Block ---------------------------------
func PutBlock(db hyperdb.Database, key []byte, t types.Block) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(BlockHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetBlock(db hyperdb.Database, key []byte) (types.Block, error){
	var block types.Block
	keyFact := append(BlockHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return block, err
	}
	err = decondeFromBytes(data, &block)
	return block, err
}

func DeleteBlock(db hyperdb.Database, key []byte) error {
	keyFact := append(BlockHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Block END ----------------------------------

//-- ------------------- Balance ---------------------------------
func PutBalance(db hyperdb.Database, key []byte, t types.Balance) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(BalanceHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetBalance(db hyperdb.Database, key []byte) (types.Balance, error){
	var balance types.Balance
	keyFact := append(BalanceHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return balance, err
	}
	err = decondeFromBytes(data, &balance)
	return balance, err
}

func DeleteBalance(db hyperdb.Database, key []byte) error {
	keyFact := append(BalanceHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Balance END ----------------------------------


//-- ------------------- Chain ---------------------------------
func PutChain(db hyperdb.Database, key []byte, t types.Chain) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(ChainHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetChain(db hyperdb.Database, key []byte) (types.Chain, error){
	var Chain types.Chain
	keyFact := append(ChainHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return Chain, err
	}
	err = decondeFromBytes(data, &Chain)
	return Chain, err
}

func DeleteChain(db hyperdb.Database, key []byte) error {
	keyFact := append(ChainHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Chain END ----------------------------------

//-- ------------------- Node ---------------------------------
func PutNode(db hyperdb.Database, key []byte, t node.Node) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(NodeHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetNode(db hyperdb.Database, key []byte) (node.Node, error){
	var Node node.Node
	keyFact := append(NodeHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return Node, err
	}
	err = decondeFromBytes(data, &Node)
	return Node, err
}

func DeleteNode(db hyperdb.Database, key []byte) error {
	keyFact := append(NodeHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Node END ----------------------------------