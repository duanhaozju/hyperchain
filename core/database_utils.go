//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"strconv"
	"sync"
)

// the prefix of key, use to save to db
var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key")
	BlockNumPrefix           = []byte("blockNum-")
	//bodySuffix               = []byte("-body")
	TxMetaSuffix = []byte{0x01}
	log          *logging.Logger // package-level logger
)

func init() {
	log = logging.MustGetLogger("")
}

// InitDB initialization ldb and memdb
// should be called while programming start-up
// port: the server port
func InitDB(dbPath string, port int) {
	hyperdb.SetLDBPath(dbPath, port)
	memChainMap = newMemChain()
	memChainStatusMap = newMemChainStatus()
}

//---------------------- Receipt Start ---------------------------------
// GetReceipt returns a receipt by hash
func GetReceipt(txHash common.Hash) *types.ReceiptTrans {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return nil
	}
	data, _ := db.Get(append(ReceiptsPrefix, txHash[:]...))
	if len(data) == 0 {
		return nil
	}
	var receipt types.Receipt
	err = proto.Unmarshal(data, &receipt)
	if err != nil {
		log.Errorf("GetReceipt err:", err)
	}
	//a:= receipt.ToReceiptTrans()
	//fmt.Println("address",common.ToHex(a.ContractAddress))
	//fmt.Println("TxHash",a.TxHash)
	return receipt.ToReceiptTrans()
}

//---------------------- Receipts End ---------------------------------

//-- ------------------- Transaction ---------------------------------
func PutTransaction(db hyperdb.Database, key []byte, t *types.Transaction) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	// add key by prefix to identification for a key-value database
	keyFact := append(TransactionPrefix, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	return nil
}

// PutTransactions put transaction into database using Batch
// Each Transaction's key is its hash
func PutTransactions(db hyperdb.Database, commonHash crypto.CommonHash, ts []*types.Transaction) error {
	batch := db.NewBatch()
	for _, trans := range ts {
		key := trans.Hash(commonHash).Bytes()
		keyFact := append(TransactionPrefix, key...)
		value, err := proto.Marshal(trans)
		if err != nil {
			return nil
		}
		batch.Put(keyFact, value)
	}
	return batch.Write()
}

func GetTransaction(db hyperdb.Database, key []byte) (*types.Transaction, error) {
	var transaction types.Transaction
	keyFact := append(TransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &transaction, err
	}
	err = proto.Unmarshal(data, &transaction)
	return &transaction, err
}
func JudgeTransactionExist(db hyperdb.Database, key []byte) (bool, error) {
	var transaction types.Transaction
	keyFact := append(TransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return false, err
	}
	err = proto.Unmarshal(data, &transaction)
	return true, err
}



//get tx<-->block num,hash,index
func GetTxWithBlock(db hyperdb.Database, key []byte) (uint64, int64) {
	dataMeta, _ := db.Get(append(key, TxMetaSuffix...))
	if len(dataMeta) == 0 {
		return 0, 0
	}
	meta := &types.TransactionMeta{}
	if err := proto.Unmarshal(dataMeta, meta); err != nil {
		log.Error(err)
		return 0, 0
	}
	return meta.BlockIndex, meta.Index
}

func DeleteTransaction(db hyperdb.Database, key []byte) error {
	keyFact := append(TransactionPrefix, key...)
	return db.Delete(keyFact)
}

func GetAllTransaction(db *hyperdb.LDBDatabase) ([]*types.Transaction, error) {
	var ts []*types.Transaction = make([]*types.Transaction, 0)
	iter := db.NewIterator()
	for ok := iter.Seek(TransactionPrefix); ok; ok = iter.Next() {
		key := iter.Key()
		if len(string(key)) >= len(TransactionPrefix) && string(key[:len(TransactionPrefix)]) == string(TransactionPrefix) {
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

func GetAllDiscardTransaction(db *hyperdb.LDBDatabase) ([]*types.InvalidTransactionRecord, error) {
	var ts []*types.InvalidTransactionRecord = make([]*types.InvalidTransactionRecord, 0)
	iter := db.NewIterator()
	for ok := iter.Seek(InvalidTransactionPrefix); ok; ok = iter.Next() {
		key := iter.Key()
		if len(string(key)) >= len(InvalidTransactionPrefix) && string(key[:len(InvalidTransactionPrefix)]) == string(InvalidTransactionPrefix) {
			var t types.InvalidTransactionRecord
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

func GetDiscardTransaction(db hyperdb.Database, key []byte) (*types.InvalidTransactionRecord, error) {
	var invalidTransaction types.InvalidTransactionRecord
	keyFact := append(InvalidTransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &invalidTransaction, err
	}
	err = proto.Unmarshal(data, &invalidTransaction)
	return &invalidTransaction, err
}

//-- --------------------- Transaction END -----------------------------------

//-- ------------------- Block ---------------------------------
func PutBlock(db hyperdb.Database, key []byte, t *types.Block) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	keyFact := append(BlockPrefix, key...)
	if err := db.Put(keyFact, data); err != nil {
		return err
	}
	keyNum := strconv.FormatInt(int64(t.Number), 10)
	err = db.Put(append(BlockNumPrefix, keyNum...), t.BlockHash)
	return err
}
func PutBlockTx(db hyperdb.Database, commonHash crypto.CommonHash, key []byte, t *types.Block) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	keyFact := append(BlockPrefix, key...)
	batch := db.NewBatch()
	if err := batch.Put(keyFact, data); err != nil {
		return err
	}
	keyNum := strconv.FormatInt(int64(t.Number), 10)
	//err = db.Put(append(blockNumPrefix, keyNum...), t.BlockHash)

	//err = batch.Put(append(blockNumPrefix, keyNum...), t.BlockHash)
	if err := batch.Put(append(BlockNumPrefix, keyNum...), t.BlockHash); err != nil {
		return err
	}

	return batch.Write()
}

func GetBlockHash(db hyperdb.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(BlockNumPrefix, keyNum...))
}

func GetBlock(db hyperdb.Database, key []byte) (*types.Block, error) {
	var block types.Block
	keyFact := append(BlockPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &block, err
	}
	err = proto.Unmarshal(data, &block)
	return &block, err
}

func GetBlockByNumber(db hyperdb.Database, blockNumber uint64) (*types.Block, error) {
	hash, err := GetBlockHash(db, blockNumber)
	if err != nil {
		return nil, err

	}
	return GetBlock(db, hash)
}

func DeleteBlock(db hyperdb.Database, key []byte) error {
	keyFact := append(BlockPrefix, key...)
	return db.Delete(keyFact)
}

//delete block data and block.num<--->block.hash
func DeleteBlockByNum(db hyperdb.Database, blockNum uint64) error {
	hash, err := GetBlockHash(db, blockNum)
	if err != nil {
		return err
	}
	keyFact := append(BlockPrefix, hash...)
	if err := db.Delete(keyFact); err != nil {
		return err
	}
	keyNum := strconv.FormatInt(int64(blockNum), 10)
	return db.Delete(append(BlockNumPrefix, keyNum...))
}

//-- --------------------- Block END ----------------------------------

//-- --------------------- BalanceMap ------------------------------------
type balanceMapJson map[string][]byte

//-- --------------------- BalanceMap END---------------------------------

//-- ------------------- Chain ----------------------------------------

// memChain manage safe chain
type memChain struct {
	data   types.Chain      // chain
	lock   sync.RWMutex     // the lock of chain
	cpChan chan types.Chain // when data.Height reach check point, will be writed
}

type memChainStatus struct {
	data types.ChainStatus
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
			cpChan: make(chan types.Chain),
		}
	}
	chain, err := getChain(db)
	if err == nil {
		return &memChain{
			data:   *chain,
			cpChan: make(chan types.Chain),
		}
	}
	return &memChain{
		data: types.Chain{
			Height: 0,
		},
		cpChan: make(chan types.Chain),
	}
}
func newMemChainStatus() *memChainStatus {
	return &memChainStatus{
		data: types.ChainStatus{},
	}
}

var memChainMap *memChain
var memChainStatusMap *memChainStatus

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
		memChainMap.data.Height = block.Number
		memChainMap.data.CurrentTxSum += uint64(len(block.Transactions))
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
	block, err := GetBlockByNumber(db, blockNumber)
	if err != nil {
		log.Warning("no required block number")
		return err
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
		LatestBlockHash:  memChainMap.data.LatestBlockHash,
		ParentBlockHash:  memChainMap.data.ParentBlockHash,
		Height:           memChainMap.data.Height,
		RequiredBlockNum: memChainMap.data.RequiredBlockNum,
		RequireBlockHash: memChainMap.data.RequireBlockHash,
		RecoveryNum:      memChainMap.data.RecoveryNum,
		CurrentTxSum:     memChainMap.data.CurrentTxSum,
	}
}

// WaitUtilHeightChan get chain from channel. if channel is empty, the func will be blocked
func GetChainUntil() *types.Chain {
	chain := <-memChainMap.cpChan
	return &chain
}

// WriteChain to channel, if the channel is not read, will be blocked
func WriteChainChan() {
	memChainMap.cpChan <- memChainMap.data
}

// putChain put chain database
func putChain(db hyperdb.Database, t *types.Chain) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	if err := db.Put(ChainKey, data); err != nil {
		return err
	}
	return nil
}

// UpdateRequire updates requireBlockNum and requireBlockHash
func UpdateRequire(num uint64, hash []byte, recoveryNum uint64) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.RequiredBlockNum = num
	memChainMap.data.RecoveryNum = recoveryNum
	memChainMap.data.RequireBlockHash = hash
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db, &memChainMap.data)
}

func SetReplicas(replicas []uint64) {
	memChainStatusMap.lock.Lock()
	defer memChainStatusMap.lock.Unlock()
	memChainStatusMap.data.Replicas = replicas
}

func GetReplicas() []uint64 {
	memChainStatusMap.lock.Lock()
	defer memChainStatusMap.lock.Unlock()
	return memChainStatusMap.data.Replicas
}

func SetId(id uint64) {
	memChainStatusMap.lock.Lock()
	defer memChainStatusMap.lock.Unlock()
	memChainStatusMap.data.Id = id
}

func GetId() uint64 {
	memChainStatusMap.lock.Lock()
	defer memChainStatusMap.lock.Unlock()
	return memChainStatusMap.data.Id
}

func UpdateChainByViewChange(height uint64, latestHash []byte) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.Height = height
	//memChainMap.data.ParentBlockHash = parentHash
	memChainMap.data.LatestBlockHash = latestHash
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db, &memChainMap.data)
}

//GetInvaildTxErrType gets ErrType of invalid tx
func GetInvaildTxErrType(db hyperdb.Database, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
	var invalidTx types.InvalidTransactionRecord
	keyFact := append(InvalidTransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return -1, err
	}
	err = proto.Unmarshal(data, &invalidTx)
	return invalidTx.ErrType, err
}

// getChain get chain from database
func getChain(db hyperdb.Database) (*types.Chain, error) {
	var chain types.Chain
	data, err := db.Get(ChainKey)
	if len(data) == 0 {
		return &chain, err
	}
	err = proto.Unmarshal(data, &chain)
	return &chain, err
}

//-- --------------------- Chain END ----------------------------------
