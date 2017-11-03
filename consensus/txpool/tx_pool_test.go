//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/stretchr/testify/assert"
)

func TestNewTxPoolImpl(t *testing.T) {
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 400, eventMux, 40)
	assert.Equal(t, nil, err, "new txPool fail")

	if txPool.batchSize != 40 || txPool.poolSize != 400 {
		t.Error("new txPool wrong")
	}
}

func TestPrimaryAddNewTx(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	batchSub := eventMux.Subscribe(TxHashBatch{})
	txPool, err := newTxPoolImpl(namespace, 400, eventMux, 40)
	ast.Equal(nil, err, "new txPool fail")
	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test1"),
		Id:              uint64(1),
		TransactionHash: []byte("hash1"),
	}
	isGenerated, err := txPool.primaryAddNewTx(tx1, false)
	ast.Equal(nil, err, "primary add transaction fail")
	ast.Equal(false, isGenerated, "should not generate a batch")

	if len(txPool.txPoolHash) != 1 || len(txPool.txPool) != 1 {
		t.Error("Only add one transaction to txPool")
	}
	txPool.batchSize = 2
	isGenerated, err = txPool.primaryAddNewTx(tx1, false)
	ast.NotEqual(nil, err, "should return a duplicate transaction error")
	ast.Equal(false, isGenerated, "should not generate a batch")

	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	go func() {
		isGenerated, err = txPool.primaryAddNewTx(tx2, false)
		assert.Equal(t, nil, err, "primary add transaction fail")
	}()
	select {
	case e := <-batchSub.Chan():
		batch, ok := e.Data.(TxHashBatch)
		ast.Equal(true, ok, "Should get a TxHashBatch")
		ast.Equal(2, len(batch.TxList), "This batch should contain two transactions")
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for message to fire")
	}

	if len(txPool.txPoolHash) != 0 || len(txPool.txPool) != 0 || len(txPool.batchStore) != 1 {
		t.Error("should generate a batch, but fail")
	}
}

func TestReplicaAddNewTx(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 1, eventMux, 40)
	ast.Equal(nil, err, "new txPool fail")
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	_, err = txPool.replicaAddNewTx(tx, true)
	ast.Nil(err, "replica add transaction fail")
	if len(txPool.txPoolHash) != 1 || len(txPool.txPool) != 1 {
		t.Error("Only add one transaction to txPool")
	}
	_, err = txPool.replicaAddNewTx(tx, true)
	ast.Equal(ErrPoolFull, err, "pool should be full and return a full error")
	txPool.poolSize = 2
	_, err = txPool.replicaAddNewTx(tx, true)
	ast.Equal(ErrDuplicateTx, err, "pool get a existed transaction")
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	_, err = txPool.replicaAddNewTx(tx2, true)
	ast.Equal(nil, err, "should add a new transaction, but fail now")
	if len(txPool.txPool) != 2 || len(txPool.txPoolHash) != 2 {
		t.Error("should store 2 transaction, expect 2")
	}
}

func TestAddNewTx(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 10, eventMux, 40)
	ast.Equal(nil, err, "new txPool fail")

	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	_, err = txPool.AddNewTx(tx, false, true)
	ast.Equal(nil, err, "add transaction false")
	ast.Equal(1, len(txPool.txPool), "should store 1 transaction")
	ast.Equal(1, len(txPool.txPoolHash), "should store 1 transaction")

	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	_, err = txPool.AddNewTx(tx2, true, false)
	ast.Equal(nil, err, "add transaction false")
	ast.Equal(2, len(txPool.txPool), "should store 2 transaction")
	ast.Equal(2, len(txPool.txPoolHash), "should store 2 transaction")
}

func TestAddTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 10, eventMux, 40)
	ast.Equal(nil, err, "new txPool fail")

	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	txs := []*types.Transaction{tx, tx}
	err = txPool.addTxs(txs)
	ast.Equal(nil, err, "add transaction false")
	ast.Equal(1, len(txPool.txPool), "add dupilcate transactions, expect 1")
	ast.Equal(1, len(txPool.txPoolHash), "add dupilcate transactions, expect 1")

	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	txs = []*types.Transaction{tx, tx2}
	err = txPool.addTxs(txs)
	ast.Equal(nil, err, "add transaction false")
	ast.Equal(2, len(txPool.txPool), "add dupilcate transactions, expect 2")
	ast.Equal(2, len(txPool.txPoolHash), "add dupilcate transactions, expect 2")
}

func TestReturnFetchTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}})
	missingHashList := make(map[uint64]string)
	missingHashList[uint64(1)] = "1"
	_, err = txPool.ReturnFetchTxs("1", missingHashList)
	ast.NotEqual(nil, err, fmt.Sprintf("should not find this batch, err is %v", err))
	txBatch := &TxHashBatch{
		TxHashList: []string{"1", "2", "3", "4"},
		TxList: []*types.Transaction{
			{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test1"),
				Id:              uint64(1),
				TransactionHash: []byte("hash1"),
			},
			{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test2"),
				Id:              uint64(1),
				TransactionHash: []byte("hash2"),
			},
			{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test3"),
				Id:              uint64(1),
				TransactionHash: []byte("hash3"),
			},
			{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test4"),
				Id:              uint64(1),
				TransactionHash: []byte("hash4"),
			},
		},
	}
	txBatch.BatchHash = hash(txBatch)
	txPool.batchStore = append(txPool.batchStore, txBatch)
	missingHashList = make(map[uint64]string)
	missingHashList[uint64(1)] = "5"
	txs, err := txPool.ReturnFetchTxs(txBatch.BatchHash, missingHashList)
	ast.NotEqual(nil, err, "should not find this transanction")

	missingHashList = make(map[uint64]string)
	missingHashList[uint64(1)] = "2"
	missingHashList[uint64(3)] = "4"
	txs, err = txPool.ReturnFetchTxs(txBatch.BatchHash, missingHashList)
	ast.Equal(nil, err, "should find these transanctions")
	ast.Equal(2, len(txs), "should get two transanctions")
	_, ok1 := txs[uint64(1)]
	ast.Equal(true, ok1, "should return this transaction")
	_, ok3 := txs[uint64(3)]
	ast.Equal(true, ok3, "should return this transaction")
	if ok1 {
		ast.Equal("hash2", string(txs[uint64(1)].TransactionHash), "should find transanctions whose hash values are 'hash2' and 'hash4'")
	}
	if ok3 {
		ast.Equal("hash4", string(txs[uint64(3)].TransactionHash), "should find transanctions whose hash values are 'hash2' and 'hash4'")
	}
}

func TestGotMissingTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txs := make(map[uint64]*types.Transaction)
	txs[uint64(2)] = &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	txs[uint64(4)] = &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test4"),
		Id:              uint64(1),
		TransactionHash: []byte("hash4"),
	}
	err = txPool.GotMissingTxs("2", txs)
	ast.Equal(ErrNoBatch, err, "should not find this missingTx, expect ErrNoBatch")

	txPool.missingTxs["2"] = make(map[uint64]string)
	txPool.missingTxs["2"][uint64(2)] = txs[uint64(2)].GetHash().Hex()
	err = txPool.GotMissingTxs("2", txs)
	ast.Equal(ErrMismatch, err, "not match, expect ErrMismatch")

	txPool.missingTxs["2"][uint64(4)] = "notexist"
	err = txPool.GotMissingTxs("2", txs)
	ast.Equal(ErrMismatch, err, "not match, expect ErrMismatch")

	txPool.missingTxs["2"][uint64(4)] = txs[uint64(4)].GetHash().Hex()
	err = txPool.GotMissingTxs("2", txs)
	ast.Equal(nil, err, err)
}

func TestGetTxsByHashList(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txlist := []*types.Transaction{
		{
			From:            []byte{1},
			To:              []byte{2},
			Value:           []byte{1},
			Timestamp:       time.Now().UnixNano(),
			Signature:       []byte("test2"),
			Id:              uint64(1),
			TransactionHash: []byte("hash2"),
		},
		{
			From:            []byte{1},
			To:              []byte{2},
			Value:           []byte{1},
			Timestamp:       time.Now().UnixNano(),
			Signature:       []byte("test4"),
			Id:              uint64(1),
			TransactionHash: []byte("hash4"),
		},
	}
	hashBatch1 := &TxHashBatch{TxHashList: []string{"1", "2"}, TxList: txlist}
	id := hash(hashBatch1)
	hashBatch1.BatchHash = id
	txPool.batchStore = append(txPool.batchStore, hashBatch1)
	txs, _, err := txPool.GetTxsByHashList(id, []string{})
	ast.Equal(txlist, txs, "This batch already exists, should return itself")

	missingBatchId := "miss"
	missingBatchHash := make(map[uint64]string)
	missingBatchHash[uint64(0)] = "hello"
	missingBatchHash[uint64(1)] = "world"
	txPool.missingTxs[missingBatchId] = missingBatchHash
	_, miss, err := txPool.GetTxsByHashList(missingBatchId, nil)
	ast.Equal(missingBatchHash, miss, "Should return before missingTxsHash")

	txPool.txPool["a"] = &types.Transaction{}
	txPool.txPool["b"] = &types.Transaction{}
	txPool.txPool["c"] = &types.Transaction{}
	txPool.txPool["d"] = &types.Transaction{}
	txPool.txPoolHash = []string{"a", "b", "c", "d"}
	txPool.batchedTxs["b"] = true

	_, _, err = txPool.GetTxsByHashList("2", []string{"b", "e", "f"})
	ast.Equal(ErrDuplicateTx, err, "Should return ErrDuplicateTx")

	delete(txPool.batchedTxs, "b")
	txs, missingHash, err := txPool.GetTxsByHashList("2", []string{"b", "e", "f"})
	if err != nil || txs != nil || len(missingHash) != 2 {
		t.Error("should return missing error and miss two transactions in total")
	}
	txs, missingHash2, err := txPool.GetTxsByHashList("2", []string{"b", "c", "d"})
	ast.Equal(missingHash, missingHash2, "should return the before missingHash")

	txs, missingHash, err = txPool.GetTxsByHashList("3", []string{"b", "c", "d"})
	if len(txs) != 3 || missingHash != nil || err != nil {
		t.Error("should generate a batch")
	}
	if txPool.batchStore[1].BatchHash != "3" || len(txPool.batchStore[1].TxList) != 3 {
		t.Error("The new batch hash should be '3' and it's length is 3")
	}
	if txPool.batchedTxs["b"] != true || txPool.batchedTxs["c"] != true || txPool.batchedTxs["d"] != true {
		t.Error("should store batched transactions in batchedTxs")
	}
}

func TestRemoveBatchedTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 1, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "2"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "3"})
	hashList := []string{"1", "2"}
	err = txPool.RemoveBatches(hashList)
	ast.Equal(nil, err, "remove batched transactions failed")
	ast.Equal(1, len(txPool.batchStore), "remove two batch, expect 1")

	hashList = []string{"2", "3"}
	err = txPool.RemoveBatches(hashList)
	ast.Equal(nil, err, "remove batched transactions failed")
	ast.Equal(0, len(txPool.batchStore), "remove one batch, expect 0")
}

func TestRemoveOneBatchedTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "2"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "3"})
	_, _, err = txPool.getBatchById("4")
	ast.Equal(ErrNoBatch, err, "should not find the batch whose hash is '4'")

	txPool.removeOneBatchedTxs(1)
	if txPool.batchStore[1].BatchHash == "2" {
		t.Error("the second batch in batchStore should be removed")
	}
}

func TestHasTxInPool(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	ast.Equal(false, txPool.HasTxInPool(), "contains no tx in txPool, expect false")

	txPool.txPool["1"] = nil
	ast.Equal(true, txPool.HasTxInPool(), "contains one tx in txPool, expect false")
}

func TestGetBatchesBack(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test1"),
		Id:              uint64(1),
		TransactionHash: []byte("hash1"),
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	batch1 := &TxHashBatch{
		TxList:     []*types.Transaction{tx1},
		TxHashList: []string{tx1.GetHash().Hex()},
	}
	batch2 := &TxHashBatch{
		TxList:     []*types.Transaction{tx2},
		TxHashList: []string{tx2.GetHash().Hex()},
	}
	batch1.BatchHash = hash(batch1)
	batch2.BatchHash = hash(batch2)

	txPool.batchStore = append(txPool.batchStore, batch1)
	txPool.batchedTxs[tx1.GetHash().Hex()] = true
	txPool.batchStore = append(txPool.batchStore, batch2)
	txPool.batchedTxs[tx2.GetHash().Hex()] = true
	err = txPool.GetBatchesBackExcept([]string{batch1.BatchHash, "1"})
	ast.Equal(1, len(txPool.batchStore), "all batches should be moved because there is no batch whose hash is '1'")

	err = txPool.GetBatchesBackExcept([]string{batch1.BatchHash, batch2.BatchHash})
	if err != nil || txPool.batchStore == nil {
		t.Error("batch1 should remain")
	}
	if txPool.txPoolHash[0] != tx2.GetHash().Hex() {
		t.Error("The first transaction in batch should be the first transaction in txpool")
	}
}

func TestGetOneTxsBack(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test1"),
		Id:              uint64(1),
		TransactionHash: []byte("hash1"),
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	batch1 := &TxHashBatch{
		TxList:     []*types.Transaction{tx1},
		TxHashList: []string{tx1.GetHash().Hex()},
	}
	batch2 := &TxHashBatch{
		TxList:     []*types.Transaction{tx2},
		TxHashList: []string{tx2.GetHash().Hex()},
	}
	batch1.BatchHash = hash(batch1)
	batch2.BatchHash = hash(batch2)

	txPool.batchStore = append(txPool.batchStore, batch1)
	txPool.batchedTxs[tx1.GetHash().Hex()] = true
	txPool.batchStore = append(txPool.batchStore, batch2)
	txPool.batchedTxs[tx2.GetHash().Hex()] = true

	err = txPool.GetOneBatchBack("1")
	ast.Equal(ErrNoBatch, err, "should not move because there is no batch whose hash is '1'")

	err = txPool.GetOneBatchBack(batch1.BatchHash)
	ast.Nil(err, "should be moved back successfully")
	ast.Equal(1, len(txPool.batchStore), "should left only one batch")
}

func TestIsPoolFull(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 3, eventMux, 3)
	ast.Equal(nil, err, "new txPool fail")

	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	txPool.AddNewTx(tx, true, false)
	txPool.batchedTxs["a"] = true
	ast.Equal(false, txPool.IsPoolFull(), "store two transactions in total, expect not full")
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	txPool.AddNewTx(tx2, true, false)
	ast.Equal(true, txPool.IsPoolFull(), "store three transactions in total, expect full")
}

func TestGenerateTxBatch(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	err = txPool.generateTxBatch()
	ast.Equal(ErrPoolEmpty, err, "pool is empty, expect ErrPoolEmpty")
}

func TestNewTxBatch(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 3, eventMux, 1)
	ast.Equal(nil, err, "new txPool fail")

	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	t.Logf("tx hash is: %s", tx.GetHash().Hex())
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	t.Logf("tx2 hash is: %s", tx2.GetHash().Hex())
	tx3 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test3"),
		Id:              uint64(1),
		TransactionHash: []byte("hash3"),
	}
	t.Logf("tx3 hash is: %s", tx3.GetHash().Hex())

	txPool.AddNewTx(tx, true, false)
	if len(txPool.txPool) != 0 || len(txPool.txPoolHash) != 0 || len(txPool.batchStore) != 1 {
		t.Error("should generate one batch")
	}
	txPool.addTxs([]*types.Transaction{tx2, tx3})
	txPool.newTxBatch()
	txPool.newTxBatch()
	//if len(txPool.txPool) != 1 || len(txPool.txPoolHash) != 1 || len(txPool.batchStore) != 2 {
	//	t.Error("should generate one batch again")
	//}
	for _, batch := range txPool.batchStore {
		t.Log(batch.BatchHash)
		t.Logf("batch hash list is: %v", batch.TxHashList)
	}
}

func TestRemoveTxPoolTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txPool.txPool["1"] = &types.Transaction{}
	txPool.txPool["2"] = &types.Transaction{}
	txPool.txPool["3"] = &types.Transaction{}
	txPool.txPool["4"] = &types.Transaction{}
	txPool.txPoolHash = []string{"1", "2", "3", "4"}
	txPool.removeTxPoolTxs([]string{"2", "4"})
	if len(txPool.txPool) != 2 || len(txPool.txPoolHash) != 2 {
		t.Error("should remove two transactions")
	}
	if txPool.txPoolHash[0] != "1" || txPool.txPoolHash[1] != "3" {
		t.Error("should be ordered")
	}
}

func TestGetBatchById(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	batch1 := &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}}
	id := hash(batch1)
	batch1.BatchHash = id
	txPool.batchStore = append(txPool.batchStore, batch1)
	batch, index, err := txPool.getBatchById(id)
	ast.Nil(err, "error should be nothing")
	ast.Equal(0, index, "index should be 0")
	ast.Equal(id, batch.BatchHash, "Get a wrong batch")

	_, _, err = txPool.getBatchById("a")
	ast.Equal(ErrNoBatch, err, "should not get a batch")
}

func TestPostTxBatch(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 1, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	sub := txPool.queue.Subscribe(TxHashBatch{})
	go func() {
		for {
			select {
			case obj := <-sub.Chan():
				switch ev := obj.Data.(type) {
				case TxHashBatch:
					t.Logf("receive TxHashBatch: %v", ev)
					ast.Equal("a", ev.BatchHash, "Batch hash should be 'a'")
				default:
					t.Error("invalid event type.")
				}
			}
		}
	}()

	batch1 := TxHashBatch{BatchHash: "a"}
	batch2 := TxHashBatch{BatchHash: "b"}
	batch3 := TxHashBatch{BatchHash: "c"}
	txPool.postTxBatch(batch1)
	time.Sleep(time.Nanosecond)

	txPool.pendingBatches = []*TxHashBatch{&batch2}
	txPool.postTxBatch(batch3)
	ast.Equal(0, len(txPool.pendingBatches), "pendingBatches should be nil.")
}
