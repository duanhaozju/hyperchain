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
	"hyperchain/crypto"
	"strconv"
)

// the prefix of key, use to save to db
var (
	transactionPrefix   = []byte("transaction-")
	receiptsPrefix   = []byte("receipts-")
	blockPrefix         = []byte("block-")
	chainKey            = []byte("chain-key")
	balanceKey          = []byte("balance-key")
	blockNumPrefix      = []byte("blockNum-")
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
//---------------------- Receipt Start ---------------------------------
// WriteReceipts stores a batch of transaction receipts into the database.
func WriteReceipts(receipts types.Receipts) error {
	db,_ := hyperdb.GetLDBDatabase()
	batch := db.NewBatch()
	// Iterate over all the receipts and queue them for database injection
	for _, receipt := range receipts {
		data, err := proto.Marshal(receipt)
		if err != nil {
			return err
		}
		if err := batch.Put(append(receiptsPrefix, receipt.TxHash...), data); err != nil {
			return err
		}
	}
	// Write the scheduled data into the database
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}
// GetReceipt returns a receipt by hash
func GetReceipt(txHash common.Hash) *types.Receipt {
	db,_ = hyperdb.GetLDBDatabase()
	data, _ := db.Get(append(receiptsPrefix, txHash[:]...))
	if len(data) == 0 {
		return nil
	}
	var receipt types.Receipt
	err := proto.Unmarshal(data, &receipt)
	if err != nil {
		log.Info("GetReceipt err:", err)
	}
	return (*types.Receipt)(&receipt)
}
//---------------------- Receipts End ---------------------------------


//-- ------------------- Transaction ---------------------------------
func PutTransaction(db hyperdb.Database, key []byte, t *types.Transaction) error {
	data, err := proto.Marshal(t)
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

// PutTransactions put transaction into database using Batch
// Each Transaction's key is its hash
func PutTransactions(db hyperdb.Database, commonHash crypto.CommonHash ,ts []*types.Transaction) error {
	batch := db.NewBatch()
	for _, trans := range ts {
		key := trans.Hash(commonHash).Bytes()
		keyFact := append(transactionPrefix, key...)
		value, err := proto.Marshal(trans)
		if err != nil {
			return nil
		}
		batch.Put(keyFact, value)
	}
	return batch.Write()
}

func GetTransaction(db hyperdb.Database, key []byte) (*types.Transaction, error){
	var transaction types.Transaction
	keyFact := append(transactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &transaction, err
	}
	err = proto.Unmarshal(data, &transaction)
	return &transaction, err
}

func DeleteTransaction(db hyperdb.Database, key []byte) error {
	keyFact := append(transactionPrefix, key...)
	return db.Delete(keyFact)
}

func GetAllTransaction(db *hyperdb.LDBDatabase) ([]*types.Transaction, error) {
	var ts []*types.Transaction = make([]*types.Transaction, 0)
	iter := db.NewIterator()
	for ok := iter.Seek(transactionPrefix); ok; ok = iter.Next() {
		key := iter.Key()
		if len(string(key)) >= len(transactionPrefix) && string(key[:len(transactionPrefix)]) == string(transactionPrefix) {
			var t types.Transaction
			value := iter.Value()
			proto.Unmarshal(value, &t)
			ts = append(ts, &t)
		} else {
			break
		}
	}
	iter.Release()
	err := iter.Error()
	return ts, err
}
//-- --------------------- Transaction END -----------------------------------


//-- ------------------- Block ---------------------------------
func PutBlock(db hyperdb.Database, key []byte, t *types.Block) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	keyFact := append(blockPrefix, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	keyNum := strconv.FormatInt(int64(t.Number), 10)
	err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)
	return err
}

func GetBlockHash(db hyperdb.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(blockNumPrefix, keyNum...))
}

func GetBlock(db hyperdb.Database, key []byte) (*types.Block, error){
	var block types.Block
	keyFact := append(blockPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &block, err
	}
	err = proto.Unmarshal(data, &block)
	return &block, err
}

func GetBlockByNumber(db hyperdb.Database, blockNumber uint64) (*types.Block, error) {
	hash , err := GetBlockHash(db , blockNumber)
	if err != nil {
		return nil, err

	}
	return GetBlock(db, hash)
}


func DeleteBlock(db hyperdb.Database, key []byte) error {
	keyFact := append(blockPrefix, key...)
	return db.Delete(keyFact)
}

//-- --------------------- Block END ----------------------------------

//-- --------------------- BalanceMap ------------------------------------
type balanceMapJson map[string][]byte

// PutDBBalance put dbBalance into database
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

// GetDBBalance get dbBalance from database
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

// memChain manage safe chain
type memChain struct {
	data   types.Chain   // chain
	lock   sync.RWMutex  // the lock of chain
	cpChan chan types.Chain // when data.Height reach check point, will be writed
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
			cpChan: make(chan types.Chain),
		}
	}
	chain, err := getChain(db)
	if err == nil {
		return &memChain{
			data: *chain,
			cpChan: make(chan types.Chain),
		}
	}
	return &memChain{
		data: types.Chain{
			Height: 0,
		},
		cpChan:make(chan types.Chain),
	}
}
var memChainMap *memChain;

// GetLatestBlockHash get latest blockHash
func GetLatestBlockHash() []byte {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.LatestBlockHash
}

// GetParentBlockHash get the latest block's parentHash
func GetParentBlockHash() []byte {
	memChainMap.lock.RLock()
	defer memChainMap.lock.RUnlock()
	return memChainMap.data.ParentBlockHash
}

// UpdateChain update latest blockHash as given blockHash
// and the height of chain add 1
func UpdateChain(block *types.Block, genesis bool) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.LatestBlockHash = block.BlockHash
	memChainMap.data.ParentBlockHash = block.ParentHash
	if !genesis {
		memChainMap.data.Height += 1
	}
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db, &memChainMap.data)
}

//　根据blockNumber更新chain,chain的height直接赋值为block.Number
func UpdateChainByBlcokNum(db hyperdb.Database, blockNumber uint64) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	block, err :=GetBlockByNumber(db,blockNumber)
	if err != nil {
		log.Warning("no required block number")
	}
	memChainMap.data.LatestBlockHash = block.BlockHash
	memChainMap.data.ParentBlockHash = block.ParentHash
	memChainMap.data.Height = block.Number
	return putChain(db, &memChainMap.data)
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
		ParentBlockHash: memChainMap.data.ParentBlockHash,
		Height: memChainMap.data.Height,
		RequiredBlockNum:memChainMap.data.RequiredBlockNum,
		RequireBlockHash:memChainMap.data.RequireBlockHash,
		RecoveryNum:memChainMap.data.RecoveryNum,
	}
}

// WaitUtilHeightChan get chain from channel. if channel is empty, the func will be blocked
func GetChainUntil() *types.Chain {
	chain := <- memChainMap.cpChan
	return &chain
}

// WriteChain to channel, if the channel is not read, will be blocked
func WriteChainChan()  {
	memChainMap.cpChan <- memChainMap.data
}

// putChain put chain database
func putChain(db hyperdb.Database, t *types.Chain) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	if err := db.Put(chainKey, data); err != nil {
		return err
	}
	return nil
}

// UpdateRequire updates requireBlockNum and requireBlockHash
func UpdateRequire(num uint64, hash []byte,recoveryNum uint64) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.RequiredBlockNum = num
	memChainMap.data.RecoveryNum = recoveryNum
	memChainMap.data.RequireBlockHash = hash
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {return err}
	return putChain(db, &memChainMap.data)
}



// getChain get chain from database
func getChain(db hyperdb.Database) (*types.Chain, error){
	var chain types.Chain
	data, err := db.Get(chainKey)
	if len(data) == 0 {
		return &chain, err
	}
	err = proto.Unmarshal(data, &chain)
	return &chain, err
}
//-- --------------------- Chain END ----------------------------------

// GetCurrentAndParentBlockHash get current blockHash and parent blockHash
// for given blockNumber if there are any error, it will be return
func GetCurrentAndParentBlockHash(blockNumber uint64) (currentHash, parentHash []byte, err error) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil, nil, err
	}
	currentHash, err = GetBlockHash(db, blockNumber)
	if err != nil {
		return nil, nil, err
	}
	block, err := GetBlock(db, currentHash)
	if err != nil {
		return nil, nil, err
	}
	parentHash = block.ParentHash
	return
}