package db_utils

import (
	"bytes"
	"hyperchain/common"
	"hyperchain/core/test_util"
	"hyperchain/core/types"
	"hyperchain/hyperdb/mdb"
	"reflect"
	"testing"
)

func TestGetBlock(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	// Get block number
	dbBlock, _ := getBlockByNumberFunc(db, test_util.BlockCases.Number)
	if !reflect.DeepEqual(&test_util.BlockCases, dbBlock) {
		t.Error("expect to be same")
	}
	// Get block hash
	dbBlock, _ = GetBlockFunc(db, test_util.BlockCases.BlockHash)
	if !reflect.DeepEqual(&test_util.BlockCases, dbBlock) {
		t.Error("expect to be same")
	}
	// Get block hash by number
	blockHash, _ := getBlockHashFunc(db, test_util.BlockCases.Number)
	if bytes.Compare(test_util.BlockCases.BlockHash, blockHash) != 0 {
		t.Error("expect to be same")
	}
}

func TestPersistBlockWitoutFlush(t *testing.T) {
	var (
		dbBlock *types.Block
	)
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	// persist without flush
	PersistBlock(db.NewBatch(), &test_util.BlockCases, false, false)
	dbBlock, _ = GetBlockFunc(db, test_util.BlockCases.BlockHash)
	if dbBlock != nil {
		t.Error("expect empty in memdb")
	}
	// persist blcok and flush immediately
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	dbBlock, _ = GetBlockFunc(db, test_util.BlockCases.BlockHash)
	if dbBlock == nil {
		t.Error("expect block persisted in db")
	}
}

func TestPersistBlockWithVersion(t *testing.T) {
	var (
		dbBlock *types.Block
	)
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	// (1) persis block with default version tag
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	dbBlock, _ = GetBlockFunc(db, test_util.BlockCases.BlockHash)
	if string(dbBlock.Version) != BlockVersion {
		t.Error("expect block version same with default")
	}
	for _, tx := range dbBlock.Transactions {
		if string(tx.Version) != TransactionVersion {
			t.Error("expect transaction version same with default value")
		}
	}
	// (2) persis block with specified version tag
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true, "1.4", "1.4")
	dbBlock, _ = GetBlockFunc(db, test_util.BlockCases.BlockHash)
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
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	deleteBlockFunc(db, db.NewBatch(), test_util.BlockCases.BlockHash, true, true)
	if blk, _ := GetBlockFunc(db, test_util.BlockCases.BlockHash); blk != nil {
		t.Error("expect deletion success")
	}
	// delete block by number
	PersistBlock(db.NewBatch(), &test_util.BlockCases, true, true)
	deleteBlockByNumFunc(db, db.NewBatch(), test_util.BlockCases.Number, true, true)
	if blk, _ := getBlockByNumberFunc(db, test_util.BlockCases.Number); blk != nil {
		t.Error("expect deletion success")
	}
}
