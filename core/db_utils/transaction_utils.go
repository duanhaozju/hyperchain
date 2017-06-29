package db_utils

import (
	"github.com/golang/protobuf/proto"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"encoding/json"
)

// GetInvaildTxErrType - gets ErrType of invalid tx
// Return -1 if not existed in db.
func GetInvaildTxErrType(namespace string, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return -1, err
	}
	return GetInvalidTxErrTypeFunc(db, key)
}

func GetInvalidTxErrTypeFunc(db db.Database, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
	var invalidTx types.InvalidTransactionRecord
	keyFact := append(InvalidTransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return -1, err
	}
	err = proto.Unmarshal(data, &invalidTx)
	return invalidTx.ErrType, err
}

// GetTransaction - retrieve transaction by hash.
func GetTransaction(namespace string, key []byte) (*types.Transaction, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetTransactionFunc(db, key)
}

func GetTransactionFunc(db db.Database, key []byte) (*types.Transaction, error) {
	if key == nil {
		return nil, EmptyPointerErr
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
		return &transaction, err
	}
	err = proto.Unmarshal(wrapper.Transaction, &transaction)
	return &transaction, err
}

// Persist transaction content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistTransaction(batch db.Batch, transaction *types.Transaction, flush bool, sync bool, extra ...interface{}) (error, []byte) {
	// check pointer value
	if transaction == nil || batch == nil {
		return EmptyPointerErr, nil
	}
	err, data := encapsulateTransaction(transaction, extra...)
	if err != nil {
		return err, nil
	}
	if err := batch.Put(append(TransactionPrefix, transaction.GetHash().Bytes()...), data); err != nil {
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

// wrap transaction
func encapsulateTransaction(transaction *types.Transaction, extra ...interface{}) (error, []byte) {
	version := TransactionVersion
	if transaction == nil {
		return EmptyPointerErr, nil
	}
	if len(extra) > 0 {
		// parse version
		if tmp, ok := extra[0].(string); ok {
			version = tmp
		}
	}
	transaction.Version = []byte(version)
	data, err := proto.Marshal(transaction)
	if err != nil {
		return err, nil
	}
	wrapper := &types.TransactionWrapper{
		TransactionVersion: []byte(version),
		Transaction:        data,
	}
	data, err = proto.Marshal(wrapper)
	if err != nil {
		return err, nil
	}
	return nil, data
}

// Persist transactions content to a batch, KEEP IN MIND call batch.Write to flush all data to disk if `flush` is false
func PersistTransactions(batch db.Batch, transactions []*types.Transaction, version string, flush bool, sync bool, extra ...interface{}) error {
	// check pointer value
	if transactions == nil || batch == nil {
		return EmptyPointerErr
	}
	// process
	for _, transaction := range transactions {
		err, data := encapsulateTransaction(transaction, extra...)
		if err != nil {
			return err
		}
		if err := batch.Put(append(TransactionPrefix, transaction.GetHash().Bytes()...), data); err != nil {
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

func DeleteTransaction(batch db.Batch, key []byte, flush, sync bool) error {
	keyFact := append(TransactionPrefix, key...)
	err := batch.Delete(keyFact)
	if err != nil {
		return err
	}
	err = DeleteTransactionMeta(batch, key, false, false)
	if err != nil {
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

func GetAllTransaction(namespace string) ([]*types.Transaction, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetAllTransactionFunc(db)
}

func GetAllTransactionFunc(db db.Database) ([]*types.Transaction, error) {
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

// Judge whether a transaction has been saved in database
func JudgeTransactionExist(namespace string, key []byte) (bool, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return false, err
	}
	var wrapper types.TransactionWrapper
	keyFact := append(TransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return false, err
	}
	err = proto.Unmarshal(data, &wrapper)
	return true, err
}

/*
	txmeta
*/

// Persist tx meta content to a batch, KEEP IN MIND call batch.Write to flush all data to disk
func PersistTransactionMeta(batch db.Batch, transactionMeta *types.TransactionMeta, txHash common.Hash, flush bool, sync bool) error {
	if transactionMeta == nil || batch == nil {
		return EmptyPointerErr
	}
	data, err := proto.Marshal(transactionMeta)
	if err != nil {
		//logger.Error("Invalid txmeta struct to marshal! error msg, ", err.Error())
		return err
	}
	if err := batch.Put(append(txHash.Bytes(), TxMetaSuffix...), data); err != nil {
		//logger.Error("Put txmeta into database failed! error msg, ", err.Error())
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

func DeleteTransactionMeta(batch db.Batch, key []byte, flush, sync bool) error {
	keyFact := append(key, TxMetaSuffix...)
	err := batch.Delete(keyFact)
	if err != nil {
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

//get tx<-->block num,hash,index
func GetTxWithBlock(namespace string, key []byte) (uint64, int64) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return 0, 0
	}
	return GetTxWithBlockFunc(db, key)
}

func GetTxWithBlockFunc(db db.Database, key []byte) (uint64, int64) {
	dataMeta, _ := db.Get(append(key, TxMetaSuffix...))
	if len(dataMeta) == 0 {
		return 0, 0
	}
	meta := &types.TransactionMeta{}
	if err := proto.Unmarshal(dataMeta, meta); err != nil {
		return 0, 0
	}
	return meta.BlockIndex, meta.Index
}

/*
	Invalid Txs
*/
func PersistInvalidTransactionRecord(batch db.Batch, invalidTx *types.InvalidTransactionRecord, flush bool, sync bool) (error, []byte) {
	// save to db
	if batch == nil || invalidTx == nil {
		return EmptyPointerErr, nil
	}

	invalidTx.Tx.Version = []byte(TransactionVersion)
	data, err := proto.Marshal(invalidTx)
	if err != nil {
		return err, nil
	}
	if err := batch.Put(append(InvalidTransactionPrefix, invalidTx.Tx.TransactionHash...), data); err != nil {
		//logger.Error("Put invalid tx into database failed! error msg, ", err.Error())
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

func GetAllDiscardTransaction(namespace string) ([]*types.InvalidTransactionRecord, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetAllDiscardTransactionFunc(db)
}

func GetAllDiscardTransactionFunc(db db.Database) ([]*types.InvalidTransactionRecord, error) {
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

func DeleteAllDiscardTransaction(db db.Database, batch db.Batch, flush, sync bool) error {
	iter := db.NewIterator(InvalidTransactionPrefix)
	defer iter.Release()
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	err := iter.Error()
	// flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return err
}

func DumpDiscardTransactionInRange(db db.Database, batch db.Batch, dumpBatch db.Batch, start, end int64, flush, sync bool) (error, uint64) {
	// flush to disk immediately
	iter := db.NewIterator(InvalidTransactionPrefix)
	defer iter.Release()
	var cnt uint64
	for iter.Next() {
		var t types.InvalidTransactionRecord
		value := iter.Value()
		if err := proto.Unmarshal(value, &t); err != nil {
			continue
		}
		if t.Tx != nil && t.Tx.Timestamp < end && t.Tx.Timestamp >= start {
			cnt += 1
			batch.Delete(iter.Key())
			dumpBatch.Put(iter.Key(), iter.Value())
		}
	}
	err := iter.Error()
	if flush {
		if sync {
			batch.Write()
			dumpBatch.Write()
		} else {
			go batch.Write()
			go dumpBatch.Write()
		}
	}
	return err, cnt
}

func GetDiscardTransaction(namespace string, key []byte) (*types.InvalidTransactionRecord, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return GetDiscardTransactionFunc(db, key)
}

func GetDiscardTransactionFunc(db db.Database, key []byte) (*types.InvalidTransactionRecord, error) {
	var invalidTransaction types.InvalidTransactionRecord
	keyFact := append(InvalidTransactionPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return nil, err
	}
	err = proto.Unmarshal(data, &invalidTransaction)
	return &invalidTransaction, err
}

// EncodeReceipt generated bytes only used in consensus comparison
func EncodeTransaction(transaction *types.Transaction) ([]byte, error) {
	// There use transaction's signature directly
	// Since the same signature means the consensus field of both are equal.
	type ConsensusTransaction struct {
		From            []byte      `json:"from,omitempty"`
		To              []byte      `json:"to,omitempty"`
		Value           []byte      `json:"value,omitempty"`
		Timestamp       int64       `json:"timestamp,omitempty"`
		Nonce           int64       `json:"nonce,omitempty"`
	}

	return json.Marshal(&ConsensusTransaction{
		From:       transaction.From,
		To:         transaction.To,
		Value:      transaction.Value,
		Timestamp:  transaction.Timestamp,
		Nonce:      transaction.Nonce,
	})
}
