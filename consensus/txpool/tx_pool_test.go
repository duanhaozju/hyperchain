//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"github.com/stretchr/testify/assert"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"testing"
	"time"
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
	_, err = txPool.ReturnFetchTxs("1", []string{"1"})
	ast.NotEqual(nil, err, "should not find this batch")
	txBatch := &TxHashBatch{
		TxHashList: []string{"1", "2", "3", "4"},
		TxList: []*types.Transaction{
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test1"),
				Id:              uint64(1),
				TransactionHash: []byte("hash1"),
			},
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test2"),
				Id:              uint64(1),
				TransactionHash: []byte("hash2"),
			},
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test3"),
				Id:              uint64(1),
				TransactionHash: []byte("hash3"),
			},
			&types.Transaction{
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
	txBatches, err := txPool.ReturnFetchTxs(txBatch.BatchHash, []string{"5"})
	ast.NotEqual(nil, err, "should not find this transanction")

	txBatches, err = txPool.ReturnFetchTxs(txBatch.BatchHash, []string{"2", "4"})
	ast.Equal(nil, err, "should find these transanctions")
	ast.Equal(2, len(txBatches), "should get two transanctions")
	ast.Equal("hash2", string(txBatches[0].TransactionHash), "should find transanctions whose hash values are 'hash2' and 'hash4'")
	ast.Equal("hash4", string(txBatches[1].TransactionHash), "should find transanctions whose hash values are 'hash2' and 'hash4'")
}

func TestGotMissingTxs(t *testing.T) {
	ast := assert.New(t)
	namespace := "1"
	common.InitRawHyperLogger(namespace)
	common.SetLogLevel(namespace, "hyperdb", "ERROR")
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(namespace, 100, eventMux, 10)
	ast.Equal(nil, err, "new txPool fail")

	txs := []*types.Transaction{
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
	txPool.missingTxs["2"] = []string{txs[0].GetHash().Hex(), txs[1].GetHash().Hex()}
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
	hashBatch1 := &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}, TxList: txlist}
	txPool.batchStore = append(txPool.batchStore, hashBatch1)
	id := hash(txPool.batchStore[0])
	txs, _, err := txPool.GetTxsByHashList(id, []string{})
	ast.Equal(txlist, txs, "This batch already exists, should return itself")

	missingBatchId := "miss"
	missingBatchHash := []string{"hello", "world"}
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
	err = txPool.RemoveBatchedTxs(hashList)
	ast.Equal(nil, err, "remove batched transactions failed")
	ast.Equal(1, len(txPool.batchStore), "remove two batch, expect 1")

	hashList = []string{"2", "3"}
	err = txPool.RemoveBatchedTxs(hashList)
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
	err = txPool.RemoveOneBatchedTxs("4")
	ast.Equal(ErrNoBatch, err, "should not find the batch whose hash is '4'")

	err = txPool.RemoveOneBatchedTxs("2")
	if err != nil || txPool.batchStore[1].BatchHash != "3" {
		t.Error("the batch should be removed")
	}
}

func TestGetTxsBack(t *testing.T) {
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
	err = txPool.GetTxsBack([]string{batch1.BatchHash, "1"})
	ast.Equal(ErrNoBatch, err, "should not move because there is no batch whose hash is '1'")
	ast.Equal(2, len(txPool.batchStore), "should not move because there is no batch whose hash is '1'")

	err = txPool.GetTxsBack([]string{batch1.BatchHash, batch2.BatchHash})
	if err != nil || txPool.batchStore != nil {
		t.Error("should move all batched batches")
	}
	if txPool.txPoolHash[0] != tx1.GetHash().Hex() {
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

	err = txPool.GetOneTxsBack("1")
	ast.Equal(ErrNoBatch, err, "should not move because there is no batch whose hash is '1'")

	err = txPool.GetOneTxsBack(batch1.BatchHash)
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
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	tx3 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test3"),
		Id:              uint64(1),
		TransactionHash: []byte("hash3"),
	}
	txPool.AddNewTx(tx, true, false)
	if len(txPool.txPool) != 0 || len(txPool.txPoolHash) != 0 || len(txPool.batchStore) != 1 {
		t.Error("should generate one batch")
	}
	txPool.addTxs([]*types.Transaction{tx2, tx3})
	txPool.newTxBatch()
	if len(txPool.txPool) != 1 || len(txPool.txPoolHash) != 1 || len(txPool.batchStore) != 2 {
		t.Error("should generate one batch again")
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

	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}})
	h := hash(txPool.batchStore[0])
	batch, err := txPool.getBatchById(h)
	ast.Nil(err, err)
	ast.Equal("1", batch.BatchHash, "Get a wrong batch")

	batch, err = txPool.getBatchById("a")
	ast.Equal(ErrNoBatch, err, "should not get a batch")
}
