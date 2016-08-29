// Hash interface defined
// author: Lizhong kuang
// date: 2016-08-25
// last modified:2016-08-25
package core

import (
	"hyperchain/hyperdb"
	"os"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"sync"
	"strconv"
	"encoding/json"
)

var (
	transactionPrefix   = []byte("transaction-")
	blockPrefix    = []byte("block-")
	chainKey            = []byte("chain-key")
	balanceKey = []byte("balance-key")
	bodySuffix   = []byte("-body")
	txMetaSuffix        = []byte{0x01}
)
//-- 初始化ldb 和 memdb
//-- 应该在程序开始初始化
//-- port为端口号
func InitDB(port int) {
	lDBPath = baseLDBPath + strconv.Itoa(port)

	/*log.Println("删除现有本地数据库数据...")
	utils.RemoveContents(lDBPath)
	log.Println("初始化本地数据库...")*/
	memChainMap = newMemChain()
	memTxPoolMap = newMemTxPool()
}

//-- --------------- about ldb -----------------------

func getBaseDir() string {
	path := os.TempDir()
	return path
}

var (
	baseLDBPath = getBaseDir() + "/cache/"
	lDBPath string
)
func getLDBConnection()(*leveldb.DB,error){
	db, err := leveldb.OpenFile(lDBPath, nil )
	return db, err
}

//-- ------------------ ldb end ----------------------

//-- ------------------- Transaction ---------------------------------
func PutTransaction(db hyperdb.Database, key []byte, t types.Transaction) error {
	data, err := proto.Marshal(&t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(transactionPrefix, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetTransaction(db hyperdb.Database, key []byte) (types.Transaction, error){
	var transaction types.Transaction
	keyFact := append(transactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return transaction, err
	}
	err = proto.Unmarshal(data, &transaction)
	return transaction, err
}

func DeleteTransaction(db hyperdb.Database, key []byte) error {
	keyFact := append(transactionPrefix, key...)
	return db.Delete(keyFact)
}

func GetAllTransaction(db *hyperdb.LDBDatabase) ([]types.Transaction, error) {
	var ts []types.Transaction
	iter := db.NewIterator()
	for iter.Next() {
		key := iter.Key()
		if len(string(key)) >= len(transactionPrefix) && string(key[:len(transactionPrefix)]) == string(transactionPrefix) {
			var t types.Transaction
			value := iter.Value()
			//err = decondeFromBytes(value, &t)
			proto.Unmarshal(value, &t)
			ts = append(ts, t)
		}
	}
	iter.Release()
	err := iter.Error()
	return ts, err
}
//-- --------------------- Transaction END -----------------------------------


//-- ------------------- Block ---------------------------------
func PutBlock(db hyperdb.Database, key []byte, t types.Block) error {
	data, err := proto.Marshal(&t)
	if err != nil {
		return err
	}
	//-- 给key加上前缀,用于区分,实际存放的key
	keyFact := append(blockPrefix, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

func GetBlock(db hyperdb.Database, key []byte) (types.Block, error){
	var block types.Block
	keyFact := append(blockPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return block, err
	}
	err = proto.Unmarshal(data, &block)
	return block, err
}

func DeleteBlock(db hyperdb.Database, key []byte) error {
	keyFact := append(blockPrefix, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Block END ----------------------------------

//-- --------------------- BalanceMap ------------------------------------
func PutDBBalance(db hyperdb.Database, balance_db BalanceMap) error {
	data, err := json.Marshal(balance_db)
	if err != nil {
		return err
	}
	return db.Put(balanceKey, data)
}

func GetDBBalance(db hyperdb.Database) (BalanceMap, error) {
	var b BalanceMap
	data, err := db.Get(balanceKey)
	if err != nil {
		return b, err
	}
	if err = json.Unmarshal(data, &b); err != nil {
		return b, err
	}
	return b, nil
}
//-- --------------------- BalanceMap END---------------------------------

//-- ------------------- Chain ----------------------------------------

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
func GetLatestBlockHash() string {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return string(memChainMap.data.LatestBlockHash)
}

//-- 更新Chain，即更新最新的blockhash 并将height加1
//-- blockHash为最新区块的hash
func UpdateChain(blockHash []byte)  {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.LatestBlockHash = blockHash
	memChainMap.data.Height += 1
}

//-- 获取区块的高度
func GetHeightOfChain() int64 {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.Height
}

//-- 获取chain的拷贝
func GetChain() *types.Chain {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return &types.Chain{
		LatestBlockHash: memChainMap.data.LatestBlockHash,
		Height: memChainMap.data.Height,
	}
}

func putChain(db hyperdb.Database, t types.Chain) error {
	data, err := proto.Marshal(&t)
	if err != nil {
		return err
	}
	if err := db.Put(chainKey, data); err != nil {
		return err
	}
	return nil
}

func getChain(db hyperdb.Database) (types.Chain, error){
	var chain types.Chain
	data, err := db.Get(chainKey)
	if len(data) == 0 {
		return chain, err
	}
	err = proto.Unmarshal(data, &chain)
	return chain, err
}
//-- --------------------- Chain END ----------------------------------


//-- --------------------- TxPool -------------------------------------
type memTxPool struct {
	txPool TxPool
	lock sync.RWMutex
}

var memTxPoolMap *memTxPool;
var maxCapacity = 5

func newMemTxPool() *memTxPool {
	return &memTxPool{
		txPool:TxPool{
			MaxCapacity:maxCapacity,
			Transactions:make([]types.Transaction, 0, maxCapacity),
		},
	}
}

//-- 获取TxPool中所有的交易
func GetTransactionsFromTxPool() []types.Transaction {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return memTxPoolMap.txPool.Transactions
}

//-- 将交易加到Pool 但是并不能保证溢出
//-- 在调用这个方法之后应该先调用TxPoolIsFull()判断Pool是否为满
//-- 如果为满应该先清空
func AddTransactionToTxPool(tran types.Transaction){
	memTxPoolMap.lock.Lock()
	defer memTxPoolMap.lock.Unlock()
	memTxPoolMap.txPool.Transactions = append(memTxPoolMap.txPool.Transactions, tran)
}

//-- 判断TxPool是否为满状态
func TxPoolIsFull() bool {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return len(memTxPoolMap.txPool.Transactions) == memTxPoolMap.txPool.MaxCapacity
}

//-- 清空交易池
func ClearTxPool()  {
	memTxPoolMap.lock.Lock()
	defer memTxPoolMap.lock.Unlock()
	memTxPoolMap.txPool.Transactions = make([]types.Transaction, 0, maxCapacity)
}

//-- 获取交易池的容量
func GetTxPoolCapacity() int {
	memTxPoolMap.lock.RLock()
	defer memTxPoolMap.lock.RUnlock()
	return len(memTxPoolMap.txPool.Transactions)
}
//-- --------------------- TxPool END ---------------------------------