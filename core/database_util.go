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
	transactionHeaderKey = []byte("transactionHeader")
	blockHeaderKey = []byte("blockHeader")
	balanceHeaderKey = []byte("balanceHeader")
	chainHeaderKey = []byte("chainHeaderKey") 
	nodeHeaderKey = []byte("nodeHeaderKey")
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
	lDBPath = baseLDBPath + strconv.Itoa(port)
	memDB, _ = hyperdb.NewMemDatabase()  //-- 暂时没用
	memBalanceMap = newMemBalance()
	memChainMap = newMemChain()
	memNodeMap = newMemNode()
}

//-- --------------- about ldb -----------------------

func getBasePath() string {
	//path, _ := os.Getwd()
	path := os.TempDir()
	return path
}

var (
	baseLDBPath = getBasePath() + "/cache/"
	lDBPath string
)
func getLDBConnection()(*leveldb.DB,error){
	db, err := leveldb.OpenFile(lDBPath, nil )
	return db, err
}
//-- ------------------ ldb end ----------------------

//-- --------------- about memdb -----------------------
//-- 这里定义了操作MemDB的
var (
	memDB *hyperdb.MemDatabase;
)

//-- ------------------ memdb end ----------------------

//-- ------------------- Transaction ---------------------------------
func putTransaction(db hyperdb.Database, key []byte, t types.Transaction) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(transactionHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func getTransaction(db hyperdb.Database, key []byte) (types.Transaction, error){
	var transaction types.Transaction
	keyFact := append(transactionHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return transaction, err
	}
	err = decondeFromBytes(data, &transaction)
	return transaction, err
}

func deleteTransaction(db hyperdb.Database, key []byte) error {
	keyFact := append(transactionHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Transaction END -----------------------------------


//-- ------------------- Block ---------------------------------
func putBlock(db hyperdb.Database, key []byte, t types.Block) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(blockHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func getBlock(db hyperdb.Database, key []byte) (types.Block, error){
	var block types.Block
	keyFact := append(blockHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return block, err
	}
	err = decondeFromBytes(data, &block)
	return block, err
}

func deleteBlock(db hyperdb.Database, key []byte) error {
	keyFact := append(blockHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Block END ----------------------------------

//-- ------------------- Balance ---------------------------------
func putBalance(db hyperdb.Database, key []byte, t types.Balance) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(balanceHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func getBalance(db hyperdb.Database, key []byte) (types.Balance, error){
	var balance types.Balance
	keyFact := append(balanceHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return balance, err
	}
	err = decondeFromBytes(data, &balance)
	return balance, err
}

func deleteBalance(db hyperdb.Database, key []byte) error {
	keyFact := append(balanceHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Balance END ----------------------------------


//-- ------------------- Chain ---------------------------------
func putChain(db hyperdb.Database, key []byte, t types.Chain) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(chainHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func getChain(db hyperdb.Database, key []byte) (types.Chain, error){
	var Chain types.Chain
	keyFact := append(chainHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return Chain, err
	}
	err = decondeFromBytes(data, &Chain)
	return Chain, err
}

func deleteChain(db hyperdb.Database, key []byte) error {
	keyFact := append(chainHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Chain END ----------------------------------

//-- ------------------- Node ---------------------------------
func putNode(db hyperdb.Database, key []byte, t node.Node) error {
	data, err := encodeToBytes(t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(nodeHeaderKey, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func getNode(db hyperdb.Database, key []byte) (node.Node, error){
	var Node node.Node
	keyFact := append(nodeHeaderKey, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return Node, err
	}
	err = decondeFromBytes(data, &Node)
	return Node, err
}

func deleteNode(db hyperdb.Database, key []byte) error {
	keyFact := append(nodeHeaderKey, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Node END ----------------------------------