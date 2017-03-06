//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"os"
	"strconv"
	"sync"
	"time"
	"hyperchain/hyperdb/db"
)

// the prefix of key, use to save to db
var (
	TransactionPrefix        = []byte("transaction-")
	ReceiptsPrefix           = []byte("receipts-")
	InvalidTransactionPrefix = []byte("invalidtransaction-")
	BlockPrefix              = []byte("block-")
	ChainKey                 = []byte("chain-key-")
	BlockNumPrefix           = []byte("blockNum-")
	TxMetaSuffix             = []byte{0x01}
	log                      *logging.Logger // package-level logger
)

// using to count the number of rollback transactions
var RollbackDataSum int = 0

func init() {
	log = logging.MustGetLogger("")
}

// InitDB initialization db
// should be called while programming start-up

func InitDB(conf *common.Config, dbConfig string, port int) {
//func InitDB(conf *common.Config)
	hyperdb.SetDBConfig(dbConfig, strconv.Itoa(port))
	hyperdb.InitDatabase(conf, "Global")
	memChainMap = newMemChain()
	memChainStatusMap = newMemChainStatus()
}

/*
	Receipt
*/
// GetReceipt returns a receipt by hash
func GetReceipt(txHash common.Hash) *types.ReceiptTrans {
	db, err := hyperdb.GetDBDatabase()
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
func PersistReceipt(batch db.Batch, receipt *types.Receipt, version string, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if receipt == nil || batch == nil {
		return errors.New("empty pointer"), nil
	}
	// process
	err, data := WrapperReceipt(receipt, version)
	if err != nil {
		log.Error("wrapper receipt failed.")
		return err, nil
	}
	if err := batch.Put(append(ReceiptsPrefix, receipt.TxHash...), data); err != nil {
		log.Error("Put receipt data into database failed! error msg, ", err.Error())
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
func WrapperReceipt(receipt *types.Receipt, version string) (error, []byte) {
	if receipt == nil {
		return errors.New("empty pointer"), nil
	}
	receipt.Version = []byte(version)
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
	if err != nil {
		log.Error("Invalid receipt struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}

func DeleteReceipt(batch db.Batch, key []byte) error {
	keyFact := append(ReceiptsPrefix, key...)
	return batch.Delete(keyFact)
}

/*
	Transaction
*/
// Query Transaction
func GetTransaction(db db.Database, key []byte) (*types.Transaction, error) {
	if db == nil || key == nil {
		return nil, errors.New("empty pointer")
	}
	var wrapper types.TransactionWrapper
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
func PersistTransaction(batch db.Batch, transaction *types.Transaction, version string, flush bool, sync bool) (error, []byte) {
	// check pointer value
	if transaction == nil || batch == nil {
		return errors.New("empty pointer"), nil
	}
	err, data := WrapperTransaction(transaction, version)
	if err != nil {
		logger.Errorf("wrapper transaction failed.")
		return err, nil
	}
	if err := batch.Put(append(TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data); err != nil {
		log.Error("Put tx data into database failed! error msg, ", err.Error())
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

func WrapperTransaction(transaction *types.Transaction, version string) (error, []byte) {
	if transaction == nil {
		return errors.New("empty pointer"), nil
	}
	// process
	transaction.Version = []byte(version)
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
	if err != nil {
		log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}

// Persist transactions content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistTransactions(batch db.Batch, transactions []*types.Transaction, version string, flush bool, sync bool) error {
	// check pointer value
	if transactions == nil || batch == nil {
		return errors.New("empty pointer")
	}
	// process
	for _, transaction := range transactions {
		err, data := WrapperTransaction(transaction, version)
		if err != nil {
			logger.Errorf("wrapper transaction failed.")
			return err
		}
		if err := batch.Put(append(TransactionPrefix, transaction.GetTransactionHash().Bytes()...), data); err != nil {
			log.Error("Put tx data into database failed! error msg, ", err.Error())
			return err
		}
	}
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// Judge whether a transaction has been saved in database
func JudgeTransactionExist(db db.Database, key []byte) (bool, error) {
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
func GetTxWithBlock(db db.Database, key []byte) (uint64, int64) {
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

func DeleteTransaction(batch db.Batch, key []byte) error {
	keyFact := append(TransactionPrefix, key...)
	return batch.Delete(keyFact)
}

func GetAllTransaction(db db.Database) ([]*types.Transaction, error) {
	var ts []*types.Transaction = make([]*types.Transaction, 0)
	iter := db.NewIterator(TransactionPrefix)
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

func GetAllDiscardTransaction(db db.Database) ([]*types.InvalidTransactionRecord, error) {
	var ts []*types.InvalidTransactionRecord = make([]*types.InvalidTransactionRecord, 0)
	iter := db.NewIterator(InvalidTransactionPrefix)
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

func GetDiscardTransaction(db db.Database, key []byte) (*types.InvalidTransactionRecord, error) {
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
func PersistTransactionMeta(batch db.Batch, transactionMeta *types.TransactionMeta, txHash common.Hash, flush bool, sync bool) error {
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
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

func DeleteTransactionMeta(batch db.Batch, key []byte) error {
	keyFact := append(key, TxMetaSuffix...)
	return batch.Delete(keyFact)
}

/*
	Invalid Transaction
*/

func PersistInvalidTransactionRecord(batch db.Batch, invalidTx *types.InvalidTransactionRecord, flush bool, sync bool) (error, []byte) {
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
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

func blockTime(block *types.Block) {
	time1 := block.WriteTime - block.Timestamp
	f, err1 := os.OpenFile(hyperdb.GetLogPath(), os.O_WRONLY|os.O_CREATE, 0644)
	if err1 != nil {
		fmt.Println("db.log file create failed. err: " + err1.Error())
	} else {
		n, _ := f.Seek(0, os.SEEK_END)
		currentTime := time.Now().Local()
		newFormat := currentTime.Format("2006-01-02 15:04:05.000")
		str := strconv.FormatUint(block.Number, 10) + "#" + newFormat + "#" + strconv.FormatInt(time1, 10) + "\n"
		_, err1 = f.WriteAt([]byte(str), n)
		f.Close()
	}
}

/*
	Block
*/
func PersistBlock(batch db.Batch, block *types.Block, version string, flush bool, sync bool) (error, []byte) {
	if hyperdb.IfLogStatus() {
		go blockTime(block)
	}

	// check pointer value
	if block == nil || batch == nil {
		return errors.New("empty block pointer"), nil
	}
	err, data := WrapperBlock(block, version)
	if err != nil {
		logger.Errorf("wrapper transaction failed.")
		return err, nil
	}
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

func WrapperBlock(block *types.Block, version string) (error, []byte) {
	if block == nil {
		return errors.New("empty pointer"), nil
	}
	// process
	block.Version = []byte(version)
	data, err := proto.Marshal(block)
	if err != nil {
		log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	wrapper := &types.BlockWrapper{
		BlockVersion: []byte(version),
		Block:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		log.Error("Invalid Transaction struct to marshal! error msg, ", err.Error())
		return err, nil
	}
	return nil, data
}

func GetBlockHash(db db.Database, blockNumber uint64) ([]byte, error) {
	keyNum := strconv.FormatInt(int64(blockNumber), 10)
	return db.Get(append(BlockNumPrefix, keyNum...))
}

func GetBlock(db db.Database, key []byte) (*types.Block, error) {
	var wrapper types.BlockWrapper
	var block types.Block
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

func GetBlockByNumber(db db.Database, blockNumber uint64) (*types.Block, error) {
	if db == nil {
		return nil, errors.New("empty database pointer")
	}
	hash, err := GetBlockHash(db, blockNumber)
	if err != nil {
		return nil, err
	}
	return GetBlock(db, hash)
}

func DeleteBlock(db db.Database, key []byte) error {
	keyFact := append(BlockPrefix, key...)
	// todo remove block.num<--->block.hash assosiation
	return db.Delete(keyFact)
}

//delete block data and block.num<--->block.hash
func DeleteBlockByNum(batch db.Batch, blockNum uint64) error {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err
	}
	hash, err := GetBlockHash(db, blockNum)
	if err != nil {
		return err
	}
	keyFact := append(BlockPrefix, hash...)
	if err := batch.Delete(keyFact); err != nil {
		return err
	}
	keyNum := strconv.FormatInt(int64(blockNum), 10)
	return batch.Delete(append(BlockNumPrefix, keyNum...))
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
	db, err := hyperdb.GetDBDatabase()
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
func UpdateChain(batch db.Batch, block *types.Block, genesis bool, flush bool, sync bool) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.LatestBlockHash = block.BlockHash
	memChainMap.data.ParentBlockHash = block.ParentHash
	if genesis {
		memChainMap.data.Height = 0
		memChainMap.data.CurrentTxSum = 0
	} else {
		if memChainMap.data.Height <= block.Number {
			memChainMap.data.Height = block.Number
			memChainMap.data.CurrentTxSum += uint64(len(block.Transactions))
		} else {
			memChainMap.data.Height = block.Number
			memChainMap.data.CurrentTxSum -= uint64(RollbackDataSum)
			RollbackDataSum = 0
		}
	}
	return putChain(batch, &memChainMap.data, flush, sync)
}

// update chain according block number, set chain current height to the block number
// return error if correspondent block missing
func UpdateChainByBlcokNum(batch db.Batch, blockNumber uint64, flush bool, sync bool) error {

	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err
	}
	block, err := GetBlockByNumber(db, blockNumber)
	if err != nil {
		log.Warning("no required block number")
		return err
	}
	return UpdateChain(batch, block, block.Number == 0, flush, sync)
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
func putChain(batch db.Batch, t *types.Chain, flush bool, sync bool) error {
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	if err := batch.Put(ChainKey, data); err != nil {
		return err
	}
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// IsGenesisFinish - check whether genesis block has been mined into blockchain
func IsGenesisFinish() bool {
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		logger.Error("get database handler failed.")
		return false
	}
	_, err = GetBlockByNumber(db, 0)
	if err != nil {
		logger.Warning("missing genesis block")
		return false
	} else {
		return true
	}
}

// UpdateRequire updates requireBlockNum and requireBlockHash
func UpdateRequire(num uint64, hash []byte, recoveryNum uint64) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.RequiredBlockNum = num
	memChainMap.data.RecoveryNum = recoveryNum
	memChainMap.data.RequireBlockHash = hash
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db.NewBatch(), &memChainMap.data, true, true)
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

// Deprecated
func UpdateChainByViewChange(height uint64, latestHash []byte) error {
	memChainMap.lock.Lock()
	defer memChainMap.lock.Unlock()
	memChainMap.data.Height = height
	//memChainMap.data.ParentBlockHash = parentHash
	memChainMap.data.LatestBlockHash = latestHash
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		return err
	}
	return putChain(db.NewBatch(), &memChainMap.data, true, true)
}

//GetInvaildTxErrType gets ErrType of invalid tx
func GetInvaildTxErrType(db db.Database, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
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
func getChain(db db.Database) (*types.Chain, error) {
	var chain types.Chain
	data, err := db.Get(ChainKey)
	if len(data) == 0 {
		return &chain, err
	}
	err = proto.Unmarshal(data, &chain)
	return &chain, err
}

//-- --------------------- Chain END ----------------------------------
