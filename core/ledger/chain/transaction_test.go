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
	"math/rand"
	"reflect"
	"testing"
	"time"

	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb/mdb"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

func genRandomByte(length int) []byte {
	var seed *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	ret := make([]byte, length)
	for i := range ret {
		ret[i] = charset[seed.Intn(len(charset))]
	}

	return ret
}

func genRandomTransaction() *types.Transaction {
	return &types.Transaction{
		Version:         genRandomByte(3),
		From:            genRandomByte(40),
		To:              genRandomByte(40),
		Value:           genRandomByte(5),
		Timestamp:       time.Now().UnixNano(),
		Signature:       genRandomByte(128),
		Id:              rand.Uint64(),
		TransactionHash: genRandomByte(64),
		Nonce:           rand.Int63(),
		Other:           &types.NonHash{genRandomByte(32)},
	}
}

func TestGetTransaction(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	block := genRandomBlock()
	PersistBlock(db.NewBatch(), block, true, true)

	for idx, tx := range block.Transactions {
		meta := &types.TransactionMeta{
			BlockIndex: block.Number,
			Index:      int64(idx),
		}
		PersistTransactionMeta(db.NewBatch(), meta, common.BytesToHash(tx.TransactionHash), true, true)
	}

	for _, tx := range block.Transactions {
		dbTx, err := getTransactionFunc(db, common.BytesToHash(tx.TransactionHash).Bytes())
		if err != nil {
			t.Error(err.Error())
		}
		if !CompareTransaction(tx, dbTx) {
			t.Error("expect to be same")
		}
	}
}

func TestJudgeTransactionExist(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	block := genRandomBlock()
	PersistBlock(db.NewBatch(), block, true, true)
	for idx, tx := range block.Transactions {
		meta := &types.TransactionMeta{
			BlockIndex: block.Number,
			Index:      int64(idx),
		}
		PersistTransactionMeta(db.NewBatch(), meta, common.BytesToHash(tx.TransactionHash), true, true)
	}

	for _, tx := range block.Transactions {
		if exist, _ := isTransactionExistFunc(db, common.BytesToHash(tx.TransactionHash).Bytes()); !exist {
			t.Error("expect to be not exist")
		}
	}
}

// TestGetInvaildTx tests for GetDiscardTransaction
func TestGetInvaildTx(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	tx := genRandomTransaction()
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	PersistInvalidTransactionRecord(db.NewBatch(), record, true, true)

	dbRecord, err := getDiscardTransactionFunc(db, tx.TransactionHash)
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(record, dbRecord) {
		t.Error("expect to be same")
	}
}

func CompareTransaction(tx1 *types.Transaction, tx2 *types.Transaction) bool {
	if reflect.DeepEqual(tx1, tx2) {
		return true
	} else {
		return false
	}
}
