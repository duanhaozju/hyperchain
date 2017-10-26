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
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/hyperdb/mdb"
)

func genRandomBlock() *types.Block {
	var txs []*types.Transaction
	for i := 0; i <= 2; i++ {
		tx := genRandomTransaction()
		txs = append(txs, tx)
	}
	baseTime := time.Now().UnixNano()
	return &types.Block{
		Version:      genRandomByte(3),
		ParentHash:   genRandomByte(64),
		BlockHash:    genRandomByte(64),
		Transactions: txs,
		Timestamp:    baseTime,
		MerkleRoot:   genRandomByte(64),
		TxRoot:       genRandomByte(64),
		ReceiptRoot:  genRandomByte(64),
		Number:       uint64(rand.Intn(100)),
		WriteTime:    baseTime + int64(300*time.Millisecond),
		CommitTime:   baseTime + int64(100*time.Millisecond),
		EvmTime:      baseTime + int64(200*time.Millisecond),
		Bloom:        genRandomByte(32),
	}
}

func TestGetBlock(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	block := genRandomBlock()
	PersistBlock(db.NewBatch(), block, true, true)
	// Get block number
	dbBlock, _ := getBlockByNumberFunc(db, block.Number)
	if !reflect.DeepEqual(block, dbBlock) {
		t.Error("expect to be same")
	}
	// Get block hash
	dbBlock, _ = GetBlockFunc(db, block.BlockHash)
	if !reflect.DeepEqual(block, dbBlock) {
		t.Error("expect to be same")
	}
	// Get block hash by number
	blockHash, _ := getBlockHashFunc(db, block.Number)
	if bytes.Compare(block.BlockHash, blockHash) != 0 {
		t.Error("expect to be same")
	}
}

func TestPersistBlockWitoutFlush(t *testing.T) {
	var dbBlock *types.Block
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	block := genRandomBlock()
	// persist without flush
	PersistBlock(db.NewBatch(), block, false, false)
	dbBlock, _ = GetBlockFunc(db, block.BlockHash)
	if dbBlock != nil {
		t.Error("expect empty in memdb")
	}
	// persist blcok and flush immediately
	PersistBlock(db.NewBatch(), block, true, true)
	dbBlock, _ = GetBlockFunc(db, block.BlockHash)
	if dbBlock == nil {
		t.Error("expect block persisted in db")
	}
}

func TestPersistBlockWithVersion(t *testing.T) {
	var dbBlock *types.Block
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	block := genRandomBlock()
	// (1) persis block with default version tag
	PersistBlock(db.NewBatch(), block, true, true)
	dbBlock, _ = GetBlockFunc(db, block.BlockHash)
	if string(dbBlock.Version) != BlockVersion {
		t.Error("expect block version same with default")
	}
	for _, tx := range dbBlock.Transactions {
		if string(tx.Version) != TransactionVersion {
			t.Error("expect transaction version same with default value")
		}
	}
	// (2) persis block with specified version tag
	PersistBlock(db.NewBatch(), block, true, true, "1.4", "1.4")
	dbBlock, _ = GetBlockFunc(db, block.BlockHash)
	if string(dbBlock.Version) != "1.4" {
		t.Error("expect block version same with specified value")
	}
	for _, tx := range dbBlock.Transactions {
		if string(tx.Version) != "1.4" {
			t.Error("expect transaction version same with specified value")
		}
	}
}

func TestDeleteBlock(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	// delete block by hash
	block := genRandomBlock()
	PersistBlock(db.NewBatch(), block, true, true)
	deleteBlockFunc(db, db.NewBatch(), block.BlockHash, true, true)
	if blk, _ := GetBlockFunc(db, block.BlockHash); blk != nil {
		t.Error("expect deletion success")
	}
	// delete block by number
	PersistBlock(db.NewBatch(), block, true, true)
	deleteBlockByNumFunc(db, db.NewBatch(), block.Number, true, true)
	if blk, _ := getBlockByNumberFunc(db, block.Number); blk != nil {
		t.Error("expect deletion success")
	}
}
