// Hash interface defined
// author: Lizhong kuang
// date: 2016-08-25
// last modified:2016-08-25
package core

import (
	"hyperchain/hyperdb"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"sync"
	"encoding/json"
	"hyperchain/common"
)

// the prefix of key, use to save to db
var (
	transactionPrefix   = []byte("transaction-")
	blockPrefix         = []byte("block-")
	chainKey            = []byte("chain-key")
	balanceKey          = []byte("balance-key")
	bodySuffix          = []byte("-body")
	txMetaSuffix        = []byte{0x01}
)

// InitDB initialization ldb and memdb
// should be called while programming start-up
// port: the server port
func InitDB(port int) {
	hyperdb.SetLDBPath(port)
	memChainMap = newMemChain()
}


//-- ------------------- Transaction ---------------------------------
func PutTransaction(db hyperdb.Database, key []byte, t types.Transaction) error {
	data, err := proto.Marshal(&t)
	if err != nil {
		return err
	}
	// add key by prefix to identification for a key-value database
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
type balanceMapJson map[string][]byte

func PutDBBalance(db hyperdb.Database, balance_db BalanceMap) error {

	var bJson = make(balanceMapJson)
	for key, value := range balance_db{
		bJson[key.Str()] = value
	}

	data, err := json.Marshal(bJson)
	if err != nil {
		return err
	}
	return db.Put(balanceKey, data)
}

func GetDBBalance(db hyperdb.Database) (BalanceMap, error) {
	var bJson balanceMapJson
	var b = make(BalanceMap)
	data, err := db.Get(balanceKey)
	if err != nil {
		return b, err
	}
	if err = json.Unmarshal(data, &bJson); err != nil {
		return b, err
	}
	for key, value := range bJson {
		b[common.StringToAddress(key)] = value
	}
	return b, nil
}
//-- --------------------- BalanceMap END---------------------------------

//-- ------------------- Chain ----------------------------------------

type memChain struct {
	data types.Chain
	lock sync.RWMutex
}

// newMenChain new a memChain instance
// it read from db firstly, if not exist, create a empty chain
func newMemChain() *memChain {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return &memChain{
			data: types.Chain{
				Height: 0,
			},
		}
	}
	chain, err := getChain(db)
	if err == nil {
		return &memChain{
			data: chain,
		}
	}
	return &memChain{
		data: types.Chain{
			Height: 0,
		},
	}
}
var memChainMap *memChain;

// GetLatestBlockHash get latest blockHash
func GetLatestBlockHash() []byte {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.LatestBlockHash
}

// UpdateChain update latest blockHash as given blockHash
// and the height of chain add 1
func UpdateChain(blockHash []byte) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.LatestBlockHash = blockHash
	if!genesis{
		memChainMap.data.Height += 1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db, memChainMap.data)
}

// GetHeightOfChain get height of chain
func GetHeightOfChain() uint64 {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.Height
}

// GetChainCopy get copy of chain
func GetChainCopy() *types.Chain {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return &types.Chain{
		LatestBlockHash: memChainMap.data.LatestBlockHash,
		Height: memChainMap.data.Height,
	}
}

// putChain put chain database
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

// getChain get chain from database
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