// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chain

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/hyperdb/db"

	"github.com/golang/protobuf/proto"
)

// GetInvaildTxErrType gets the ErrType of invalid tx.
// Return -1 if not existed in db.
func GetInvaildTxErrType(namespace string, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return -1, err
	}
	return getInvalidTxErrTypeFunc(db, key)
}

// GetTransaction gets the transaction with given namespace and txHash.
func GetTransaction(namespace string, key []byte) (*types.Transaction, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	return getTransactionFunc(db, key)
}

// GetAllTransaction gets all the transactions with given namespace.
func GetAllTransaction(namespace string) ([]*types.Transaction, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	return getAllTransactionFunc(db)
}

// IsTransactionExist judges whether a transaction has been saved in database.
func IsTransactionExist(namespace string, key []byte) (bool, error) {
	tx, err := GetTransaction(namespace, key)
	if tx == nil || err != nil {
		return false, nil
	}
	return true, nil
}

/*
	Transaction Meta
*/

// PersistTransactionMeta persists the transaction meta into database.
// Write to flush all data to disk.
func PersistTransactionMeta(batch db.Batch, transactionMeta *types.TransactionMeta, txHash common.Hash, flush bool, sync bool) error {
	if transactionMeta == nil || batch == nil {
		return ErrEmptyPointer
	}
	data, err := proto.Marshal(transactionMeta)
	if err != nil {
		return err
	}
	if err := batch.Put(append(txHash.Bytes(), TxMetaSuffix...), data); err != nil {
		return err
	}
	// Flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// DeleteTransactionMeta deletes transaction meta with given parameters.
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

// GetTxWithBlock gets the transaction meta with given key,
// returns block num and tx index.
func GetTxWithBlock(namespace string, key []byte) (uint64, int64) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return 0, 0
	}
	return getTxWithBlockFunc(db, key)
}

/*
	Invalid Txs
*/
// PersistInvalidTransactionRecord persists the invalid transaction record into database.
func PersistInvalidTransactionRecord(batch db.Batch, invalidTx *types.InvalidTransactionRecord, flush bool, sync bool) (error, []byte) {
	// Save to db
	if batch == nil || invalidTx == nil {
		return ErrEmptyPointer, nil
	}

	// Covert the transaction version
	invalidTx.Tx.Version = []byte(TransactionVersion)
	data, err := proto.Marshal(invalidTx)
	if err != nil {
		return err, nil
	}
	if err := batch.Put(append(InvalidTxPrefix, invalidTx.Tx.TransactionHash...), data); err != nil {
		return err, nil
	}
	// Flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil, data
}

// GetAllDiscardTransaction gets all the discard transactions with given namespace.
func GetAllDiscardTransaction(namespace string) ([]*types.InvalidTransactionRecord, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	return getAllDiscardTransactionFunc(db)
}

// getAllDiscardTransactionFunc is the help function returns the all
// the invalid transaction records in database.
func getAllDiscardTransactionFunc(db db.Database) ([]*types.InvalidTransactionRecord, error) {
	var ts []*types.InvalidTransactionRecord = make([]*types.InvalidTransactionRecord, 0)
	iter := db.NewIterator(InvalidTxPrefix)
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

// DeleteAllDiscardTransaction deletes all the discard transactions in database.
func DeleteAllDiscardTransaction(db db.Database, batch db.Batch, flush, sync bool) error {
	iter := db.NewIterator(InvalidTxPrefix)
	defer iter.Release()
	for iter.Next() {
		batch.Delete(iter.Key())
	}
	err := iter.Error()
	// Flush to disk immediately
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return err
}

// DumpDiscardTransactionInRange dumps all the discard transactions in range.
func DumpDiscardTransactionInRange(db db.Database, batch db.Batch, dumpBatch db.Batch, start, end int64, flush, sync bool) (uint64, error) {

	iter := db.NewIterator(InvalidTxPrefix)
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

	// Flush to disk immediately
	if flush {
		if sync {
			batch.Write()
			dumpBatch.Write()
		} else {
			go batch.Write()
			go dumpBatch.Write()
		}
	}

	return cnt, iter.Error()
}

// GetDiscardTransaction gets the discard transaction with given key.
func GetDiscardTransaction(namespace string, key []byte) (*types.InvalidTransactionRecord, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace, hcom.DBNAME_BLOCKCHAIN)
	if err != nil {
		return nil, err
	}
	return getDiscardTransactionFunc(db, key)
}

// ==========================================================
// Private functions that invoked by inner service
// ==========================================================

// getInvalidTxErrTypeFunc is the help function that gets the ErrType
// of invalid tx from the database.
func getInvalidTxErrTypeFunc(db db.Database, key []byte) (types.InvalidTransactionRecord_ErrType, error) {
	var invalidTx types.InvalidTransactionRecord
	keyFact := append(InvalidTxPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return -1, err
	}
	err = proto.Unmarshal(data, &invalidTx)
	return invalidTx.ErrType, err
}

// getTransactionFunc is the helper function that gets the transaction from database
// with the given db handler and txHash.
func getTransactionFunc(db db.Database, key []byte) (*types.Transaction, error) {
	if key == nil {
		return nil, ErrEmptyPointer
	}

	bi, ti := getTxWithBlockFunc(db, key)
	if bi == 0 && ti == 0 {
		return nil, ErrNotFindTxMeta
	}

	block, err := getBlockByNumberFunc(db, bi)
	if err != nil {
		return nil, ErrNotFindBlock
	}

	if len(block.Transactions) < int(ti+1) {
		return nil, ErrOutOfSliceRange
	}

	return block.Transactions[int(ti)], nil
}

// getAllTransactionFunc is the help function that gets all the transactions
// with given db handler.
func getAllTransactionFunc(db db.Database) ([]*types.Transaction, error) {
	var (
		index uint64
		ts    []*types.Transaction = make([]*types.Transaction, 0)
	)

	chain, err := getChainFn(db)
	if err != nil {
		return nil, err
	}

	// Append transactions in [ 1, ChainHeight ]
	for index = 1; index <= chain.Height; index += 1 {
		block, err := getBlockByNumberFunc(db, index)
		if err != nil {
			return nil, ErrNotFindBlock
		}
		ts = append(ts, block.Transactions...)
	}

	return ts, nil
}

// isTransactionExistFunc is the help function returns
// the existence of a transaction in database.
func isTransactionExistFunc(db db.Database, key []byte) (bool, error) {
	tx, err := getTransactionFunc(db, key)
	if tx == nil || err != nil {
		return false, nil
	}
	return true, nil
}

// getTxWithBlockFunc is the help function returns the block num and tx index.
func getTxWithBlockFunc(db db.Database, key []byte) (uint64, int64) {
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

// getDiscardTransactionFunc is the help function returns the discard transaction with given key.
func getDiscardTransactionFunc(db db.Database, key []byte) (*types.InvalidTransactionRecord, error) {
	var invalidTransaction types.InvalidTransactionRecord
	keyFact := append(InvalidTxPrefix, key...)
	data, err := db.Get(keyFact)
	if len(data) == 0 {
		return nil, err
	}
	err = proto.Unmarshal(data, &invalidTransaction)
	return &invalidTransaction, err
}
