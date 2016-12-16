//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"strconv"
	"sync"
	"errors"
)

// the prefix of key, use to save to db
var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key")
	BlockNumPrefix           = []byte("blockNum-")
	TxMetaSuffix             = []byte{0x01}
	log                      *logging.Logger // package-level logger
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

/*
	Receipt
 */
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
	var receiptWrapper types.ReceiptWrapper
	err = proto.Unmarshal(data, &receiptWrapper)
	if err != nil {
		log.Errorf("GetReceipt err:", err)
		return nil
	}
	// TODO use different policy to unmarshal receipt depend on version
	var receipt types.Receipt
	err = proto.Unmarshal(receiptWrapper.Receipt, &receipt)
	if err != nil {
		log.Errorf("GetReceipt err:", err)
		return nil
	}
	return receipt.ToReceiptTrans()
}

// Persist receipt content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistReceipt(batch hyperdb.Batch, receipt *types.Receipt, version string, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if receipt == nil || batch == nil {
		return errors.New("empty pointer"), nil
	}
	// process
	data, err := proto.Marshal(receipt)
	if err != nil {
		log.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	wrapper := &types.ReceiptWrapper{
		ReceiptVersion: []byte(version),
		Receipt:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err := batch.Put(append(ReceiptsPrefix, receipt.TxHash...), data); err != nil {
		log.Error("Put receipt data into database failed! error msg, ", err.Error())
		return err, nil
	}
	// flush to disk immediately
	if flush  {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

func DeleteReceipt(db hyperdb.Database, key []byte) error {
	keyFact := append(ReceiptsPrefix, key...)
	return db.Delete(keyFact)
}

/*
	Transaction
 */
// Query Transaction
func GetTransaction(db hyperdb.Database, key []byte) (*types.Transaction, error) {
	if db == nil || key == nil {
		return nil, errors.New("empty pointer")
	}
	var wrapper     types.TransactionWrapper
	var transaction types.Transaction
	keyFact := append(TransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return &transaction, err
	}
	err = proto.Unmarshal(data, &wrapper)
	if err != nil {
		log.Errorf("GetTransaction err:", err)
		return &transaction, err
	}

	// TODO use different policy to unmarshal receipt depend on version
	err = proto.Unmarshal(wrapper.Transaction, &transaction)
	return &transaction, err
}
// Persist transaction content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistTransaction(batch hyperdb.Batch, transaction *types.Transaction, version string, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if transaction == nil || batch == nil {
		return errors.New("empty pointer"), nil
	}
	// process
	data, err := proto.Marshal(transaction)
	if err != nil {
		log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	wrapper := &types.TransactionWrapper{
		TransactionVersion: []byte(version),
		Transaction:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err := batch.Put(append(TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data); err != nil {
		log.Error("Put tx data into database failed! error msg, ", err.Error())
		return err, nil
	}
	// flush to disk immediately
	if flush  {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

// Persist transactions content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistTransactions(batch hyperdb.Batch, transactions []*types.Transaction, version string, flush bool, sync bool) error {
	// check pointer value
	if transactions == nil || batch == nil {
		return errors.New("empty pointer")
	}
	// process
	for _, transaction := range transactions {
		data, err := proto.Marshal(transaction)
		if err != nil {
			log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
			return err
		}
		wrapper := &types.TransactionWrapper{
			TransactionVersion: []byte(version),
			Transaction:        data,
		}
		data, err =  proto.Marshal(wrapper)
		if err := batch.Put(append(TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data); err != nil {
			log.Error("Put tx data into database failed! error msg, ", err.Error())
			return err
		}
	}
	// flush to disk immediately
	if flush  {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// Judge whether a transaction has been saved in database
func JudgeTransactionExist(db hyperdb.Database, key []byte) (bool, error) {
	var wrapper types.TransactionWrapper
	keyFact := append(TransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
			return false, err
	}
	err = proto.Unmarshal(data, &wrapper)
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
	iter := db.NewIteratorWithPrefix(TransactionPrefix)
	for iter.Next() {
		var wrapper types.TransactionWrapper
		var transaction types.Transaction
		value := iter.Value()
		proto.Unmarshal(value, &wrapper)
		proto.Unmarshal(wrapper.Transaction, &transaction)
		ts = append(ts, &transaction)
	}
	iter.Release()
	err := iter.Error()
	return ts, err
}

func GetAllDiscardTransaction(db *hyperdb.LDBDatabase) ([]*types.InvalidTransactionRecord, error) {
	var ts []*types.InvalidTransactionRecord = make([]*types.InvalidTransactionRecord, 0)
	iter := db.NewIteratorWithPrefix(InvalidTransactionPrefix)
	for iter.Next() {
		var t types.InvalidTransactionRecord
		value := iter.Value()
		proto.Unmarshal(value, &t)
		ts = append(ts, &t)
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

/*
	Transaction Meta
 */

// Persist tx meta content to a batch, KEEP IN MIND call batch.Write to flush all data to disk
func PersistTransactionMeta(batch hyperdb.Batch, transactionMeta *types.TransactionMeta, txHash common.Hash, flush bool, sync bool) error {
	if transactionMeta == nil || batch == nil {
		return errors.New("empty transactionmeta pointer")
	}
	data, err := proto.Marshal(transactionMeta)
	if err != nil {
		log.Error("Invalid txmeta struct to marshal! error msg, ", err.Error())
		return err
	}
	if err := batch.Put(append(txHash.Bytes(), TxMetaSuffix...), data); err != nil {
		log.Error("Put txmeta into database failed! error msg, ", err.Error())
		return err
	}
	// flush to disk immediately
	if flush  {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func DeleteTransactionMeta(db hyperdb.Database, key []byte) error {
	keyFact := append(key, TxMetaSuffix...)
	return db.Delete(keyFact)
}
/*
	Invalid Transaction
 */

func PersistInvalidTransactionRecord(batch hyperdb.Batch, invalidTx *types.InvalidTransactionRecord, flush bool, sync bool) (error, []byte) {
	// save to db
	if batch == nil || invalidTx == nil {
		return errors.New("empty  pointer"), nil
	}
	data, err := proto.Marshal(invalidTx)
	if err != nil {
		return err, nil
	}
	if err := batch.Put(append(InvalidTransactionPrefix, invalidTx.Tx.TransactionHash...), data); err != nil {
		log.Error("Put invalid tx into database failed! error msg, ", err.Error())
		return err, nil
	}
	// flush to disk immediately
	if flush  {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

/*
	Block
 */
func PersistBlock(batch hyperdb.Batch, block *types.Block, version string, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if block == nil || batch == nil {
		return errors.New("empty block pointer"), nil
	}
	// process
	data, err := proto.Marshal(block)
	if err != nil {
		log.Error("Invalid block struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	wrapper := &types.BlockWrapper{
		BlockVersion: []byte(version),
		Block:        data,
	}
	data, err =  proto.Marshal(wrapper)
	if err := batch.Put(append(BlockPrefix, block.BlockHash...), data); err != nil {
		log.Error("Put block data into database failed! error msg, ", err.Error())
		return err, nil
	}

	// save number <-> hash associate
	keyNum := strconv.FormatInt(int64(block.Number), 10)
	if err := batch.Put(append(BlockNumPrefix, keyNum...), block.BlockHash); err != nil {
		return err, nil
	}
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

func GetBlockHash(db hyperdb.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(BlockNumPrefix, keyNum...))
}

func GetBlock(db hyperdb.Database, key []byte) (*types.Block, error) {
	var wrapper types.BlockWrapper
	var block   types.Block
	key = append(BlockPrefix, key...)
	data, err := db.Get(key)
	if len(data) == 0 {
		return &block, err
	}
	err = proto.Unmarshal(data, &wrapper)
	if err != nil {
		return &block, err
	}
	data = wrapper.Block
	// TODO use different policy to unmarshal block data
	err = proto.Unmarshal(data, &block)
	return &block, err
}

func GetBlockByNumber(db hyperdb.Database, blockNumber uint64) (*types.Block, error) {
	if db == nil {
		return nil, errors.New("empty database pointer")
	}
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
